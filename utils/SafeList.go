package utils

import (
	"errors"
	"runtime"
	"sync/atomic"
	"unsafe"
)

type ISafeList interface {
	Put(data interface{})
	Pop() (interface{}, error)
	Signal() chan bool
}

// _SafeListNode 节点
type _SafeListNode struct {
	next  unsafe.Pointer
	value interface{}
}

func newNode(data interface{}) unsafe.Pointer {
	return unsafe.Pointer(&_SafeListNode{
		nil,
		data,
	})
}

// _SafeList 安全链表
type _SafeList struct {
	head unsafe.Pointer
	tail unsafe.Pointer

	C chan bool
}

// NewSafeList 新创建一个列表
func NewSafeList() ISafeList {

	node := unsafe.Pointer(newNode(nil))
	return &_SafeList{
		node,
		node,
		make(chan bool, 1),
	}
}

// Put 放入
func (sl *_SafeList) Put(data interface{}) {
	newNode := newNode(data)
	var tail unsafe.Pointer

	for {
		tail = sl.tail
		next := (*_SafeListNode)(tail).next

		if next != nil {
			atomic.CompareAndSwapPointer(&sl.tail, tail, next)
		} else {
			if atomic.CompareAndSwapPointer(&(*_SafeListNode)(sl.tail).next, nil, newNode) {
				break
			}
		}
		runtime.Gosched()
	}

	atomic.CompareAndSwapPointer(&sl.tail, tail, newNode)

	if len(sl.C) == 0 {
		sl.C <- true
	}
}

var errNoNode = errors.New("no node")

// Pop 拿出
func (sl *_SafeList) Pop() (interface{}, error) {

	for {

		head := sl.head
		tail := sl.tail

		next := (*_SafeListNode)(head).next

		if head == tail {
			if next == nil {
				return nil, errNoNode
			}
			atomic.CompareAndSwapPointer(&sl.tail, tail, next)
		} else {
			if atomic.CompareAndSwapPointer(&sl.head, head, next) {
				return (*_SafeListNode)(next).value, nil
			}
		}

		runtime.Gosched()
	}

}

func (sl *_SafeList) Signal() chan bool {
	return sl.C
}

// IsEmpty 是否为空
func (sl *_SafeList) IsEmpty() bool {
	head := sl.head
	tail := sl.tail

	next := (*_SafeListNode)(head).next
	if head == tail {
		if next == nil {
			return true
		}
	}

	return false
}
