package lfu

import (
	"time"
)

// Cache lfu缓存结构
type Cache struct {
	maxBytes   int64                         // 最大字节数，即缓存容量
	nBytes     int64                         // 已用字节数
	cache      map[string]*Value             // key -> value映射
	freqMap    map[string]int64              // key -> 频率映射
	timeMap    map[string]int64              // key -> 时间戳映射
	freqToKeys map[int64]map[string]bool     // 频率 -> key集合
	minFreq    int64                         // 当前最小频率
	OnEvicted  func(key string, value Value) // 淘汰缓存的回调处理函数
}

// Value 节点的值接口
type Value interface {
	Len() int
}

func New(maxBytes int64, OnEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:   maxBytes,
		cache:      make(map[string]*Value),
		freqMap:    make(map[string]int64),
		timeMap:    make(map[string]int64),
		freqToKeys: make(map[int64]map[string]bool),
		minFreq:    1,
		OnEvicted:  OnEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if val, exists := c.cache[key]; exists {
		// 访问后更新频率
		c.updateFreq(key)
		return *val, true
	}
	return
}

func (c *Cache) Add(key string, value Value) {
	if val, exists := c.cache[key]; exists {
		// 更新已有key和value的字节统计
		c.nBytes += int64(value.Len()) - int64((*val).Len())
		c.cache[key] = &value
		// 更新频率
		c.updateFreq(key)
	} else {
		// 更新缓存值
		c.cache[key] = &value
		c.nBytes += int64(len(key)) + int64(value.Len())
		c.timeMap[key] = time.Now().UnixNano()
		c.setFreq(key, 1)
		c.minFreq = 1
	}
	// 超出容量之后的淘汰策略
	for c.maxBytes != 0 && c.nBytes > c.maxBytes {
		c.removeOldest()
	}
}

// setFreq 设置key的频率
func (c *Cache) setFreq(key string, freq int64) {
	c.freqMap[key] = freq
	// 确保该频率对应的子 map 已初始化
	if c.freqToKeys[freq] == nil {
		c.freqToKeys[freq] = make(map[string]bool)
	}
	c.freqToKeys[freq][key] = true
}

// updateFreq 更新key的访问频率
func (c *Cache) updateFreq(key string) {
	oldFreq := c.freqMap[key]
	newFreq := oldFreq + 1
	// 从旧频率集合删除
	c.removeFromFreqSet(key, oldFreq)
	// 加入新频率集合
	c.setFreq(key, newFreq)
	// 如果旧频率是最小频率且该频率集合为空，更新最小频率
	if oldFreq == c.minFreq && c.freqToKeys[oldFreq] != nil && len(c.freqToKeys[oldFreq]) == 0 {
		c.minFreq = newFreq
	}
}

// Len 返回缓存中的条目数
func (c *Cache) Len() int {
	return len(c.cache)
}

// Bytes 返回当前使用的字节数
func (c *Cache) Bytes() int64 {
	return c.nBytes
}

// removeFromFreqSet 从频率集合中移除key
func (c *Cache) removeFromFreqSet(key string, freq int64) {
	if freqSet, exists := c.freqToKeys[freq]; exists {
		delete(freqSet, key)
		if len(freqSet) == 0 {
			delete(c.freqToKeys, freq)
		}
	}
}

// removeOldest 按最低频率淘汰：同频率下淘汰最早插入的key
func (c *Cache) removeOldest() {
	if len(c.cache) == 0 {
		return
	}
	minFreqKeys := c.freqToKeys[c.minFreq]
	if len(minFreqKeys) == 0 {
		return
	}
	var keyToRemove string
	var earliestTime int64
	for k := range minFreqKeys {
		if keyToRemove == "" || c.timeMap[k] < earliestTime {
			keyToRemove = k
			earliestTime = c.timeMap[k]
		}
	}
	c.removeKey(keyToRemove)
}

func (c *Cache) removeKey(key string) {
	if val, exists := c.cache[key]; exists {
		// 更新字节数
		c.nBytes -= int64(len(key)) + int64((*val).Len())
		// 清理频率相关结构
		freq := c.freqMap[key]
		c.removeFromFreqSet(key, freq)
		delete(c.freqMap, key)
		delete(c.timeMap, key)
		// 删除缓存
		delete(c.cache, key)
		// 回调
		if c.OnEvicted != nil {
			c.OnEvicted(key, *val)
		}
	}
}
