package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash Hash
	//每个节点拥有的虚拟节点数，通过虚拟节点使节点分布更均匀，
	//同时分散一个节点下线时后一个节点的压力
	replicas int

	//排序后的虚拟节点值,用于顺序遍历hashMap
	keys []int

	//虚拟节点对应的真实节点
	hashMap map[int]string
}

// hashFunc传入nil则默认crc32.ChecksumIEEE
func New(replicas int, hashFunc Hash) *Map {
	if hashFunc == nil {
		hashFunc = crc32.ChecksumIEEE
	}
	m := &Map{
		hash:     hashFunc,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	return m
}

// keys为节点名称
func (m *Map) Add(keys ...string) {

	for _, key := range keys {
		//若key为a，则加入0a，1a，2a...
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *Map) IsEmpty() bool {
	return len(m.keys) == 0
}

func (m *Map) Get(key string) string {
	if m.IsEmpty() {
		return ""
	}

	//没有转型风险，即使变成负数仍能正常使用
	hash := int(m.hash([]byte(key)))

	//Search uses binary search to find and return the smallest index i in [0, n) at which f(i) is true
	index := sort.Search(len(m.keys), func(i int) bool {

		//did a bug
		return m.keys[i] >= hash
	})

	//到了最后一个节点后面，返回第一个

	if index == len(m.keys) {
		index = 0
	}

	return m.hashMap[m.keys[index]]
}
