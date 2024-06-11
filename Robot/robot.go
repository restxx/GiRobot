package Robot

import (
	"context"
	cfg "github.com/restxx/GiRobot/Cfg"
	logger "github.com/restxx/GiRobot/Logger"
	"github.com/restxx/GiRobot/utils"
	"reflect"
	"time"
)

type Robot struct {
	CliData    *utils.ClientData
	InCome     chan interface{}
	Ctx        context.Context
	CancelFunc context.CancelFunc

	lastStep   int // 完成的步骤
	normalStep int // 当前的步骤
	MainTicker *time.Ticker
	MtMgr      utils.IMtManager

	MainAction func()
	HeartBeat  func()
	ProcessMsg func(Msg interface{})
	ReLogin    func()
}

type CheckFun interface{}

func (self *Robot) Wait(check CheckFun) bool {
	tm := time.NewTimer(30 * time.Second)
	defer tm.Stop()

	T1, M1 := self.MtMgr.GetTimer(1)
	T2, M2 := self.MtMgr.GetTimer(2)
	T3, M3 := self.MtMgr.GetTimer(3)

	for {
		select {
		case <-T1.C:
			self.MtMgr.StopNode(M1)
		case <-T2.C:
			self.MtMgr.StopNode(M2)
		case <-T3.C:
			self.MtMgr.StopNode(M3)
		case <-tm.C:
			logger.Error("请求长时间无返回机器人主动断开[%s]", self.CliData.Account)
			self.ReLogin()
			return false
		case <-self.Ctx.Done():
			return false
		case msg := <-self.InCome:
			// 先处理注册过的消息函数
			self.ProcessMsg(msg)
			// 反射调用Check函数
			in := []reflect.Value{reflect.ValueOf(msg)}
			rRet := reflect.ValueOf(check).Call(in)
			if rRet[0].Bool() {
				return true
			}
		} // end_select
	} // end_for
}

func (self *Robot) WaitTime(Millisecond int64) bool {

	tm := time.NewTimer(time.Duration(Millisecond) * time.Millisecond)
	defer tm.Stop()

	for {
		select {
		case <-tm.C:
			return true
		case <-self.Ctx.Done():
			return false
		case msg := <-self.InCome:
			// 处理注册过的消息函数
			self.ProcessMsg(msg)
		}
	}
}

func (self *Robot) SetCliData(data *utils.ClientData) {
	self.CliData = data
}

func (self *Robot) Account() string {
	return self.CliData.Account
}

func (self *Robot) ClientId() int {
	return self.CliData.ClientId
}

func (self *Robot) AddWg() {
	self.CliData.Wg.Add(1)
}

func (self *Robot) WgDone() {
	self.CliData.Wg.Done()
}

// 状态控制
func (self *Robot) SetNormalStep(n int) {
	self.normalStep = n
}

func (self *Robot) SetLastStep(n int) {
	self.lastStep = n
}

func (self *Robot) GetNormalStep() int {
	return self.normalStep
}

func (self *Robot) GetLastStep() int {
	return self.lastStep
}

func (self *Robot) MainLoop() {
	self.MainTicker = time.NewTicker(cfg.GetTickerTime() * time.Millisecond)
	defer self.MainTicker.Stop()
	defer logger.Debug("[%s]关闭mainloop协程", self.CliData.Account)

	beat := time.NewTicker(30 * time.Second)
	defer beat.Stop()

	for {
		select {
		case <-self.Ctx.Done():
			logger.Info("收到 Ctx.Done()-------")
			return
		default:
		}

		select {
		case <-beat.C:
			logger.Info("收到心跳 beat.C--------")
			// 心跳
			self.HeartBeat()
		case <-self.MainTicker.C:
			// 执行主动发包动作
			self.MainAction()
		case msg := <-self.InCome:
			self.ProcessMsg(msg)

		} // select end
	} //for end
}
