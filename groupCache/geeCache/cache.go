package geeCache

import (
	"lru"
	"sync"
)

// cache封装*lru并对相关操作加锁
type cache struct {
	mu     sync.RWMutex
	lru    *lru.Cache
	nbytes int64 //计算所有key value占的用内存
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = &lru.Cache{
			//被删除时修改nbytes
			OnEvicted: func(key lru.Key, value interface{}) {
				val := value.(ByteView)
				c.nbytes -= int64(len(key.(string))) + int64(val.Len())
			},
		}
	}
	c.lru.Add(key, value)
	//len(string)会返回string的byte数
	c.nbytes += int64(len(key)) + int64(value.Len())
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), true
	}
	return
}

func (c *cache) removeOldest() {
	mu.Lock()
	defer mu.Unlock()
	if c.lru != nil {
		c.lru.RemoveOldest()
	}
}

// 统计该cache占用byte数
func (c *cache) bytes() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.nbytes
}
