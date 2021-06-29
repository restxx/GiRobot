package LuaVM

import (
	"bufio"
	"errors"
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
	luar "layeh.com/gopher-luar"
	"os"
	"sync"
)

type handlemainLua struct {
	once           *sync.Once
	handleMainCode *lua.FunctionProto
	mainCode       *lua.FunctionProto
}

var _handleMainLua *handlemainLua = &handlemainLua{once: &sync.Once{}, handleMainCode: nil, mainCode: nil}

var _lock sync.Mutex

var NewHandleMainLua = newHandleMainLua

func newHandleMainLua(handlePath, mainPath string) (L *lua.LState, err error) {
	_lock.Lock()
	defer _lock.Unlock()

	_handleMainLua.once.Do(func() {
		_handleMainLua.handleMainCode, _ = compileLua(handlePath)
		_handleMainLua.mainCode, _ = compileLua(mainPath)
	})
	L = lua.NewState()
	err = compiledFile(L, _handleMainLua.handleMainCode)
	err = compiledFile(L, _handleMainLua.mainCode)
	return
}

type CoreLuaPool struct {
	p *sync.Pool
}

func NewCoreLuaPool(handlePath, mainPath string) CoreLuaPool {
	return CoreLuaPool{p: &sync.Pool{
		New: func() interface{} {
			L, _ := newHandleMainLua(handlePath, mainPath)
			return L
		},
	}}
}

func (m CoreLuaPool) Get() *lua.LState {
	L := m.p.Get().(*lua.LState)
	return L
}

func (m CoreLuaPool) Put(L *lua.LState) {
	m.p.Put(L)
}

// compileLua reads the passed lua file from disk and compiles it.
func compileLua(filePath string) (*lua.FunctionProto, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	chunk, err := parse.Parse(reader, filePath)
	if err != nil {
		return nil, err
	}
	proto, err := lua.Compile(chunk, filePath)
	if err != nil {
		return nil, err
	}
	return proto, nil
}

// compiledFile takes a FunctionProto, as returned by compileLua, and runs it in the LState. It is equivalent
// to calling DoFile on the LState with the original source file.
func compiledFile(L *lua.LState, proto *lua.FunctionProto) error {
	L.Push(L.NewFunctionFromProto(proto))
	return L.PCall(0, lua.MultRet, nil)
}

func CallGlobal(l *lua.LState, fnName string, args ...interface{}) (err error) {

	// 组合参数列表
	var lpValues []lua.LValue
	argsArr := []interface{}(args)
	for _, v := range argsArr {
		lpValues = append(lpValues, luar.New(l, v))
	}

	fn := l.GetGlobal(fnName)
	if fn.Type() != lua.LTFunction {
		err = errors.New(fmt.Sprintf("Unknow Lua Function:%v", fnName))
		return
	}

	err = l.CallByParam(lua.P{
		Fn:      fn,
		NRet:    0,
		Protect: true},
		lpValues...)
	return
}

func CallGlobalRet(l *lua.LState, fnName string, args ...interface{}) (lv lua.LValue, err error) {
	// 组合参数列表
	var lpValues []lua.LValue
	argsArr := []interface{}(args)
	for _, v := range argsArr {
		lpValues = append(lpValues, luar.New(l, v))
	}

	fn := l.GetGlobal(fnName)
	if fn.Type() != lua.LTFunction {
		err = errors.New(fmt.Sprintf("Unknow Lua Function:%v", fnName))
		return
	}

	err = l.CallByParam(lua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true},
		lpValues...)
	lv = l.Get(-1)
	l.Pop(1)
	return
}
