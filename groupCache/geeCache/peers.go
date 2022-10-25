package geeCache

import pb "groupCachePb"

// peer要实现Get方法，通过groupname和key在该peer中找到value
// 使用protobuf通信
type ProtoGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}

// 通过key 找出包含该key的peer
type PeerPicker interface {
	PickPeer(key string) (peer ProtoGetter, ok bool)
}
