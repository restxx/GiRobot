package utils

import (
	"fmt"
	cfg "github.com/restxx/GiRobot/Cfg"
	"github.com/restxx/GiRobot/event"
	"sync"
)

var _lock sync.Mutex

func GetClientData(wg *sync.WaitGroup) *ClientData {
	_lock.Lock()
	defer _lock.Unlock()
	return &ClientData{Wg: wg,
		_dispatcher: event.NewDispatcher(),
		Exit:        make(chan int),
		MpData:      make(map[string]interface{}),
	}
}

func (ud *ClientData) Init(idx int) {
	skip := cfg.SkipNum()
	strFmt := cfg.GetNameFmt()
	PreFix := cfg.GetPrefix()
	ud.ClientId = skip + idx
	ud.Account = fmt.Sprintf(strFmt, PreFix, ud.ClientId)
	ud.CaseId = cfg.GetCaseID()
	ud.DelayLogin = false
}

const (
	TP_LOGINSERVER = iota
	TP_GAMESERVER
	IN_SCENE
)

type Event struct {
	Name string
	Args []interface{}
}

// 单用户数据结构
type ClientData struct {
	Wg          *sync.WaitGroup
	Exit        chan int
	CaseId      int
	Account     string
	ClientId    int
	_dispatcher *event.Dispatcher
	newLoginCnt int //用于计数
	DelayLogin  bool
	//gopher-lua相关变量
	//LState *lua.LState
	//Mapper *gluamapper.Mapper

	// 行为树相关变量
	// BlockBoard *core.Blackboard

	// 以下为自定义字段
	MpData map[string]interface{}
}

func (ud *ClientData) Trigger(event string, source interface{}) {
	ud._dispatcher.Trigger(event, source)
}

func (ud *ClientData) AddAction(event string, action event.Action) int {
	return ud._dispatcher.AddAction(event, action, 0)
}

func (ud *ClientData) AddNewLoginCnt() int {
	ud.newLoginCnt += 1
	return ud.newLoginCnt
}

func (ud *ClientData) GetNewLoginCnt() int {
	return ud.newLoginCnt
}

func (ud *ClientData) SetExit() {
	ud.Exit <- 1
}

func (ud *ClientData) WaitExit() {
	<-ud.Exit
}

func (ud *ClientData) GetInt(key string) int {
	if ret, ok := ud.MpData[key].(int); ok {
		return ret
	}
	return -1
}

//func (ud *ClientData) Tbl2MsgBLK(table *lua.LTable) *MsgBlock {
//	mid := table.RawGetString("M_nRuntimeTypeId")
//	msgId := uint32(mid.(lua.LNumber))
//	// 这个有锁支持并行的
//	IMsg, _ := MessageMap.GetByMsgID(msgId)
//	if err := ud.Mapper.Map(table, IMsg); err != nil {
//		logger.ErrorV(err)
//		return nil
//	}
//	return NewMsgBlock(msgId, 0, IMsg.ToBytes())
//}

// -----------------------------------------------------
