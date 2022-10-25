package geeCache

// A ByteView holds an immutable view of bytes.
// 只读的byte,存byte[]或string之一,执行操作是先检查b是否为nil，不为nil就使用b
type ByteView struct {
	//If b is non-nil, b is used, else s is used.
	b []byte
	s string
}

func (v ByteView) Len() int {
	if v.b != nil {
		return len(v.b)
	}
	return len(v.s)
}

func (v ByteView) ByteSlice() []byte {
	if v.b != nil {
		return cloneBytes(v.b)
	}
	return []byte(v.s)
}

func (v ByteView) String() string {
	if v.b != nil {
		return string(v.b)
	}
	return v.s
}

// 返回输入[]byte的clone
func cloneBytes(b []byte) []byte {
	clone := make([]byte, len(b))
	copy(clone, b)
	return clone
}
