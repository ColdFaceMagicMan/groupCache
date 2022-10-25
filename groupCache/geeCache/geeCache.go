package geeCache

import (
	"fmt"
	"log"
	"sync"

	pb "groupCachePb"
	"singleFlight"
)

// 搜索的key不存在时的回调函数
type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

// “定义一个函数类型 F，并且实现接口 A 的方法，然后在这个方法中调用自己。
// 这是 Go 语言中将其他函数（参数返回值定义与 F 一致）转换为接口 A 的常用技巧。”
// 接口实现者可以传入结构体或直接传入函数,例如GetterFunc(func(key string) ([]byte, error){})
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache cache

	peers      PeerPicker //peer挑选器
	cacheBytes int64      //缓存上限

	loadGroup filghtGroup // loadGroup ensures that each key is only fetched once
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

type filghtGroup interface {
	Do(key string, fn func() (interface{}, error)) (interface{}, error)
}

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:       name,
		getter:     getter,
		mainCache:  cache{},
		cacheBytes: cacheBytes,

		loadGroup: &singleFlight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// 不在本地缓存中则查找peers，peer没有则调用getLocally调入缓存
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is requierd")
	}
	if v, ok := g.mainCache.get(key); ok {
		return v, nil
	}

	return g.load(key)
}

// 从远程节点获取
func (g *Group) getFromPeer(peer ProtoGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil

}

// 先查找远程节点
func (g *Group) load(key string) (val ByteView, err error) {

	viewi, err := g.loadGroup.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				val, err := g.getFromPeer(peer, key)
				if err == nil {
					return val, nil
				}
				log.Println(key + "get from peer failed")
			}
		}

		return g.getLocally(key)
	})
	if err == nil {
		return viewi.(ByteView), nil
	}

	return

}

func (g *Group) getLocally(name string) (ByteView, error) {
	bytes, err := g.getter.Get(name)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(name, value)
	return value, nil
}

// 统计cache使用情况，超出则remove一些
func (g *Group) populateCache(key string, value ByteView) {
	//防止陷入死循环
	if g.cacheBytes <= 0 {
		return
	}
	g.mainCache.add(key, value)

	for {
		mainBytes := g.mainCache.bytes()
		if mainBytes <= g.cacheBytes {
			return
		}
		//后期加入多个cache时可以选择cache进行删除
		victim := &g.mainCache
		victim.removeOldest()
	}
}

// 初始化peers接口，一个group只能调用一次
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("Register called more than once")
	}
	g.peers = peers
}
