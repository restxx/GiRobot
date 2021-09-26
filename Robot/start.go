package Robot

import (
	"fmt"
	"github.com/letterbaby/manzo/signal"
	cfg "github.com/restxx/GiRobot/Cfg"
	logger "github.com/restxx/GiRobot/Logger"
	"github.com/restxx/GiRobot/utils"
	"os"
	"sync"
	"syscall"
	"time"
)

func StopSvc() {
	stopAll = true
}

var stopAll bool

func IsStop() bool {
	return stopAll
}

func Start(goFunc func(*utils.ClientData, interface{}), Var interface{}) {
	stopAll = false
	TickerTime := cfg.GetPerMillisecond()
	tmTick := time.NewTicker(TickerTime * time.Millisecond)
	Wg := &sync.WaitGroup{}

	for i := 0; i < cfg.GetTotal(); i++ {
		select {
		case <-tmTick.C:
			{
				UD := utils.GetClientData(Wg)
				UD.Init(i)
				Wg.Add(1)

				UD.AddAction(utils.RELOGIN, func(Var interface{}) {
					ud := Var.(*utils.ClientData)
					logger.Debug("[%s] 收到[重登录]消息", ud.Account)
					go func() {
						if ud.DelayLogin {
							// 延时登录
							logger.Info("[%s]延时登录", ud.Account)
							time.Sleep(time.Second * cfg.GetOfflineSecond())
						}
						goFunc(ud, false)
					}()
				})

				UD.AddAction(utils.NEW_RELOGIN, func(i interface{}) {
					ud := i.(*utils.ClientData)
					logger.Debug("[%s] 收到[用新帐号重登录]消息", ud.Account)
					strFmt := cfg.GetNameFmt()
					PreFix := cfg.GetPrefix()
					cnt := ud.AddNewLoginCnt()
					// 得到原始名字
					ud.Account = fmt.Sprintf(strFmt, PreFix, ud.ClientId)
					// 带后缀的新名字
					ud.Account = fmt.Sprintf("%s_%d", ud.Account, cnt)
					go goFunc(ud, true)
				})

				UD.Trigger(utils.RELOGIN, UD)
			}
		} // end select
	} // end for
	Wg.Wait()
}

//func ReLogin(Once *sync.Once, UD *utils.ClientData, FnStop func()) {
//	Once.Do(func() {
//		UD.Wg.Add(1)
//		FnStop()
//		UD.DelayLogin = true
//		UD.Trigger(utils.RELOGIN, UD)
//	})
//}

func init() {

	go func() {
		h := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
		signal.Watch(h, func(s os.Signal) {
			logger.Info("Handle signal: %v", s)
			stopAll = true
		})
	}()
}
