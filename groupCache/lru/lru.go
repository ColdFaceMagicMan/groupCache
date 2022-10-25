package lru

import "container/list"

// 最基本lrucache结构，无锁，由cache封装加锁
type Cache struct {
	// zero means no limit
	MaxEntries int

	//key被删除时的回调函数
	OnEvicted func(key Key, value interface{})

	//链表元素为{key,value}
	ll    *list.List
	cache map[interface{}]*list.Element
}

type Key interface{}

type entry struct {
	key   Key
	value interface{}
}

func New(maxEntries int) *Cache {
	return &Cache{
		MaxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[interface{}]*list.Element),
	}
}

func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		//将使用的ele移到链表头（lru）
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}

// 移除链表尾元素，即上次使用时间最早的元素
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}

}

// 从链表和cache中删除元素，调用回调函数
func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

func (c *Cache) Add(key Key, value interface{}) {
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}

	//存在则更新
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		ele.Value.(*entry).value = value
		return
	}

	ele := c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele
	//超出空间上限
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}
