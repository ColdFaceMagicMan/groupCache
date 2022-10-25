package lru

import (
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {
	lru := New(0)
	lru.Add("key1", 1234)
	if value, ok := lru.Get("key1"); !ok || value != 1234 {
		t.Fatalf("key1 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}

}

func TestRemoveOldest(t *testing.T) {
	lru := New(1)
	lru.Add("key1", 1234)
	lru.Add("key2", 1324)
	if _, ok := lru.Get("key1"); ok {
		t.Fatalf("RemoveOldest failed")
	}
}

func TestEvict(t *testing.T) {
	evictedKeys := make([]Key, 0)
	onEvictedFun := func(key Key, value interface{}) {
		evictedKeys = append(evictedKeys, key)
	}

	lru := New(20)
	lru.OnEvicted = onEvictedFun
	for i := 0; i < 22; i++ {
		lru.Add(fmt.Sprintf("myKey%d", i), 1234)
	}

	if len(evictedKeys) != 2 {
		t.Fatalf("got %d evicted keys; want 2", len(evictedKeys))
	}
	if evictedKeys[0] != Key("myKey0") {
		t.Fatalf("got %v in first evicted key; want %s", evictedKeys[0], "myKey0")
	}
	if evictedKeys[1] != Key("myKey1") {
		t.Fatalf("got %v in second evicted key; want %s", evictedKeys[1], "myKey1")
	}
}
