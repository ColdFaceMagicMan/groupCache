package geeCache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	if v, _ := f.Get("key1"); !reflect.DeepEqual(v, []byte("key1")) {
		t.Errorf("Getter test failed")
	}
}

var db = map[string]string{
	"coffee": "sold out",
	"cola":   "pepsi",
	"snack":  "cheetos",
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	cache := NewGroup("testCache", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("get from db :", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(v), nil
			}
			return nil, fmt.Errorf("key %s not found", key)

		}))
	for key := range db {
		if v, err := cache.Get(key); err != nil || v.String() != db[key] {
			t.Errorf("get failed")
		}
	}
	for key := range db {
		if v, err := cache.Get(key); err != nil || v.String() != db[key] {
			t.Errorf("get failed")
		}
	}
	if _, err := cache.Get("unknown_ket"); err == nil {
		t.Errorf("get unknown shouldnt success")
	}
}
