package lru

import "container/list"

type Cache struct {
	maxBytes  int64                         // 缓存容量，即最大可缓存的字节数
	nBytes    int64                         // 已用字节数
	ll        *list.List                    // 双向链表，存储缓存值作为链表节点
	cache     map[string]*list.Element      // 缓存map，存储缓存键值对
	OnEvicted func(key string, value Value) // 缓存淘汰回调处理函数
}

// entry 双向链表节点的数据结构
type entry struct {
	key   string
	value Value
}

// Value 节点的值
type Value interface {
	Len() int
}

func New(maxBytes int64, OnEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: OnEvicted,
	}
}

// Get 缓存获取方法
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 缓存淘汰策略
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add 缓存的添加或更新策略
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes > 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}
