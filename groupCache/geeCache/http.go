package geeCache

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"consistenthash"
	pb "groupCachePb"

	"github.com/golang/protobuf/proto"
)

// prefix
const defaultBasePath = "/_groupCache/"

const defaultReplicas = 50

type HTTPPool struct {
	//监听地址
	self string
	//基本前缀
	basePath string

	//guard peers和httpGetters
	mu sync.Mutex

	//通过key挑选节点
	peers *consistenthash.Map

	// keyed by e.g. "http://10.0.0.2:8008"
	//每个key对应一个HTTPGetter
	httpGetters map[string]*HTTPGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool served unexpected path" + r.URL.Path)
	}

	//use like /basePath/groupName/key
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
	}
	groupName := parts[0]
	key := parts[1]
	group := GetGroup(groupName)

	if group == nil {
		http.Error(w, "group"+groupName+"not found", http.StatusNotFound)
	}
	v, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := proto.Marshal(&pb.Response{Value: v.ByteSlice()})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println(r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "octet-stream")
	w.Write(body)
}

func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)

	p.httpGetters = make(map[string]*HTTPGetter)
	for _, peer := range peers {
		p.httpGetters[peer] = &HTTPGetter{baseURL: peer + p.basePath}
	}

}

func (p *HTTPPool) PickPeer(key string) (ProtoGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.peers == nil {
		return nil, false
	}

	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		return p.httpGetters[peer], true
	}

	return nil, false
}

type HTTPGetter struct {
	baseURL string
}

func (h *HTTPGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		//QueryEscape escapes the string so it can be safely placed inside a URL query.
		//转义掉一些在query中的危险符号（变为%加上hex的形式）
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("response read error:%v", err)
	}

	if err := proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("protobuf decode error:%v", err)
	}

	return nil
}
