package utils

import (
	"bytes"
	"encoding/binary"
	"sync"
)

// 消息通用数据结构
type MsgBlock struct {
	MID  interface{}
	SID  interface{}
	Data interface{}

	Pool MsgPool
}

func (m *MsgBlock) Write(buf []byte) {
	m.Data = append(m.Data.([]byte), buf...)
}

func NewMsgBlock(mId, sId, data interface{}) *MsgBlock {
	buf := &bytes.Buffer{}
	_ = binary.Write(buf, Endian, data)

	return &MsgBlock{MID: mId, SID: sId, Data: buf.Bytes()}
}

type MsgPool struct {
	p *sync.Pool
}

func NewPool() MsgPool {
	return MsgPool{p: &sync.Pool{
		New: func() interface{} {
			return &MsgBlock{Data: make([]byte, 0, 1024*64)}
		},
	}}
}

// Get retrieves a Buffer from the pool, creating one if necessary.
func (p MsgPool) Get() *MsgBlock {
	buf := p.p.Get().(*MsgBlock)
	buf.Data = buf.Data.([]byte)[:0]
	buf.Pool = p
	return buf
}

func (p MsgPool) Put(buf *MsgBlock) {
	p.p.Put(buf)
}
