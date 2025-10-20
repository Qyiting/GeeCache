package lfu

import (
	"reflect"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	lfu := New(int64(8), nil)
	lfu.Add("key1", String("1234"))
	if v, ok := lfu.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := lfu.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	cap := len(k1 + k2 + v1 + v2)
	lfu := New(int64(cap), nil)
	lfu.Add(k1, String(v1))
	lfu.Add(k2, String(v2))
	lfu.Add(k3, String(v3))

	if _, ok := lfu.Get(k1); ok || lfu.Len() != 2 {
		t.Fatalf("Removeoldest k1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lfu := New(int64(10), callback)
	lfu.Add("key1", String("123456"))
	lfu.Add("k2", String("k2"))
	lfu.Add("k3", String("k3"))
	lfu.Add("k4", String("k4"))

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Logf("evicted keys: %v", keys)
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}

func TestAdd(t *testing.T) {
	lfu := New(int64(0), nil)
	lfu.Add("key", String("1"))
	lfu.Add("key", String("111"))

	if lfu.nBytes != int64(len("key")+len("111")) {
		t.Fatal("expected 6 but got", lfu.nBytes)
	}
}

func TestLFUEviction(t *testing.T) {
	lfu := New(int64(30), nil)

	// 添加三个key
	lfu.Add("key1", String("val1")) // freq: 1
	lfu.Add("key2", String("val2")) // freq: 1
	lfu.Add("key3", String("val3")) // freq: 1

	// 访问key1和key2，增加它们的频率
	lfu.Get("key1") // key1 freq: 2
	lfu.Get("key2") // key2 freq: 2
	lfu.Get("key1") // key1 freq: 3

	// 添加新key，应该淘汰频率最低的key3
	lfu.Add("key4", String("val4"))

	// key3应该被淘汰
	if _, ok := lfu.Get("key3"); ok {
		t.Fatal("key3 should be evicted")
	}

	// key1, key2, key4应该还在
	if _, ok := lfu.Get("key1"); !ok {
		t.Fatal("key1 should still exist")
	}
	if _, ok := lfu.Get("key2"); !ok {
		t.Fatal("key2 should still exist")
	}
	if _, ok := lfu.Get("key4"); !ok {
		t.Fatal("key4 should exist")
	}
}

func TestFrequencyUpdate(t *testing.T) {
	lfu := New(int64(100), nil)

	lfu.Add("key1", String("value1"))

	// 检查初始频率
	if freq := lfu.freqMap["key1"]; freq != 1 {
		t.Fatalf("Expected frequency 1, got %d", freq)
	}

	// 访问key1，频率应该增加
	lfu.Get("key1")
	if freq := lfu.freqMap["key1"]; freq != 2 {
		t.Fatalf("Expected frequency 2, got %d", freq)
	}

	// 再次访问
	lfu.Get("key1")
	if freq := lfu.freqMap["key1"]; freq != 3 {
		t.Fatalf("Expected frequency 3, got %d", freq)
	}
}
