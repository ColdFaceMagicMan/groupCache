// ensures that each key is only fetched once
package singleFlight

import "sync"

// 一个请求对应一个call
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex       //guard m
	m  map[string]*call //key对应的正在处理请求的call
}

// 处理请求，同一时间相同key的请求只被fetch一次，其他等待结果即可
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	//有其他单位正在执行该请求,等待其结束后返回结果
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)

	g.m[key] = c
	c.wg.Add(1)
	g.mu.Unlock()

	c.val, c.err = fn()

	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err

}
