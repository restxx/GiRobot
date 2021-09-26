package utils

import "sync"

type _msgMap struct {
	mMap sync.Map
}

var MessageMap *_msgMap = &_msgMap{}

func (m *_msgMap) Bind(cmd interface{}, newFunc func() interface{}) error {
	m.mMap.Store(cmd, newFunc)
	return nil
}

// 返回消息结构
//func (m *_msgMap) GetByMsgID(cmd interface{}) (IMessage, bool) {
//
//	in, ok := m.mMap.Load(cmd)
//	if !ok {
//		return nil, false
//	}
//	newFunc, _ := in.(func() interface{})
//	return newFunc().(IMessage), true
//}
