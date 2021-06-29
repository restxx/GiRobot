package main

import (
	cfg "GiantQA/GiRobot/Cfg"
	logger "GiantQA/GiRobot/Logger"
	meter "GiantQA/GiRobot/Meter"
	"GiantQA/GiRobot/Robot"
	"GiantQA/GiRobot/utils"
	"GiantQA/common"
	"GiantQA/logic"
	"github.com/spf13/pflag"

	//manzologger "github.com/letterbaby/manzo/logger"
	"net/http"
	_ "net/http/pprof"
)

var curSvrInfo *common.DevServerInfo

// 处理自定义的参数
func setConfig(Var ...interface{}) {
	curSvrInfo = &common.DevServerInfo{Name: "", LoginUrl: cfg.ServerAddr(), Worlds: []int32{1}}
}

func init() {
	pflag.StringP("robot.CaseName", "C", "None", "填入相应TestCase名称(login, newLogin, battle, talk, mailsignal, equip, upSkill)")
}

func main() {
	// 每个项目的TestCase 列表各不相同 对应的-h 说明需自己定义

	cfg.InitConfig(setConfig, nil)

	logger.Init(logger.BuildLogger("./tmp", cfg.GetCaseName(), cfg.GetPrefix(), cfg.GetTotal(), cfg.GetLogLevel()))
	defer logger.Flush()

	go http.ListenAndServe("0.0.0.0:9999", nil)
	// go tool pprof -pdf http://localhost:9999/debug/pprof/profile > cpu.pdf

	//manzolog := &manzologger.Config{}
	//manzolog.Dir = "./tmp"
	//manzolog.File = true
	//manzolog.Rotating = true
	//manzolog.MaxSize = 200
	//manzolog.Name = "xxxxXXXXX.log"
	//manzologger.StartEx(manzolog)

	// 按配置策略启动
	Robot.Start(goRun, nil)
}

func goRun(CD *utils.ClientData, Var interface{}) {

	_meter := meter.NewMtManager(cfg.GetProject(), CD.Account, cfg.TimeOut())

	r := logic.NewDBTRobotV2()
	r.Init(curSvrInfo, _meter, CD)
	if IsNewLogin, ok := Var.(bool); ok {
		r.Run(IsNewLogin)
	}

}
