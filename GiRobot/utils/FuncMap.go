package utils

import (
	"errors"
	"reflect"
	"sync"
)

type MapHandler interface{}

type FuncMap struct {
	hMap sync.Map
}

func NewMap() *FuncMap { return &FuncMap{} }

func (m *FuncMap) Bind(cmd interface{}, fun MapHandler) error {
	if reflect.TypeOf(fun).Kind() != reflect.Func {
		return errors.New("handler must be a callable function")
	}
	m.hMap.Delete(cmd)
	m.hMap.Store(cmd, fun)
	return nil
}

func (m *FuncMap) buildCaller(cmd interface{}, args ...interface{}) (pFunc interface{}, in []reflect.Value, err error) {
	err = nil
	pFunc, ok := m.hMap.Load(cmd)
	if !ok {
		err = errors.New("Unknown handler ")
		return
	}

	// 获取函数指针的参数
	t := reflect.TypeOf(pFunc)
	// 数量和参数对不上
	if len(args) < t.NumIn() { // It panics if the type'login Kind is not Func.
		err = errors.New("Not enough arguments ")
		return
	}

	// 逐步压入参数
	in = make([]reflect.Value, t.NumIn())
	for idx, arg := range args {
		// argType := t.In(idx)
		// if t.In(idx) != reflect.TypeOf(arg) {
		// 	err = errors.New(fmt.Sprintf("Value not found for type %v", argType))
		// 	return
		// }
		in[idx] = reflect.ValueOf(arg) // 完成一个基本的CALL
	}
	return
}

// 无返回参数
func (m *FuncMap) Call(cmd interface{}, args ...interface{}) error {
	pFunc, in, err := m.buildCaller(cmd, args...)
	if err != nil {
		return err
	}
	// fmt.Println(runtime.FuncForPC(reflect.ValueOf(pFunc).Pointer()).Name())
	reflect.ValueOf(pFunc).Call(in)
	return nil
}
