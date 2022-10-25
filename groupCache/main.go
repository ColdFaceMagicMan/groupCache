package main

import (
	"flag"
	"fmt"
	"geeCache"
	"log"
	"net/http"
)

var db = map[string]string{
	"coffee": "sold out",
	"cola":   "pepsi",
	"snack":  "cheetos",
}

// curl "http://localhost:8000/api?key=cola"
// curl "http://localhost:8000/api?key=coffee"
// curl "http://localhost:8000/api?key=snack"
func createGroup() *geeCache.Group {
	return geeCache.NewGroup("testCache", 2<<10, geeCache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("get from db :", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("key %s not found", key)

		}))

}

func startCacheServer(addr string, peer_names []string, group *geeCache.Group) {
	peers := geeCache.NewHTTPPool(addr)
	peers.Set(peer_names...)
	group.RegisterPeers(peers)
	log.Println("cache running at" + addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startApiServer(addr string, group *geeCache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := group.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))

	log.Println("api server running at " + addr)

	//去除”http://“, 否则listen tcp报错
	log.Fatal(http.ListenAndServe(addr[7:], nil))
}

func main() {
	var port int
	var startApiOrNot bool
	flag.IntVar(&port, "port", 8001, "cache server port")
	flag.BoolVar(&startApiOrNot, "api", false, "start api server or not")
	//Parse parses the command-line flags from os.Args[1:]. Must be called after all flags are defined and before flags are accessed by the program.
	flag.Parse()

	apiAddr := "http://localhost:8000"
	peerAddrsMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var peerAddrs []string
	for _, addr := range peerAddrsMap {
		peerAddrs = append(peerAddrs, addr)
	}

	groupCache := createGroup()

	if startApiOrNot {
		go startApiServer(apiAddr, groupCache)
	}

	startCacheServer(peerAddrsMap[port], []string(peerAddrs), groupCache)

}
