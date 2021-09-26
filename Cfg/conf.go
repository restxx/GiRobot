package Cfg

import (
	"encoding/binary"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"time"
)

var Cfg *viper.Viper

func InitConfig(Fn func(...interface{}), Var ...interface{}) {
	Cfg = viper.New()

	cfgName := pflag.String("cfg", "config.toml", "配置文件名称")
	pflag.Parse()
	Cfg.SetConfigFile(*cfgName)
	err := Cfg.ReadInConfig()
	Cfg.BindPFlags(pflag.CommandLine)
	if err != nil {
		panic("读取配置文件错误... " + err.Error())
	}
	Cfg.WatchConfig()
	Cfg.OnConfigChange(func(in fsnotify.Event) {
		if Fn != nil {
			Fn(Var)
		}
	})
	// 私有配置
	if Fn != nil {
		Fn(Var)
	}
}

func GetProject() string {
	return Cfg.GetString("robot.Project")
}

func GetCaseName() string {
	return Cfg.GetString("robot.CaseName")
}

func GetCaseID() int {
	return Cfg.GetInt("robot.caseId")
}

func GetPrefix() string {
	return Cfg.GetString("login.Prefix")
}

func GetTotal() int {
	return Cfg.GetInt("Login.Total")
}

func GetLogLevel() string {
	return Cfg.GetString("robot.LogLevel")
}

func GetTickerTime() time.Duration {
	return Cfg.GetDuration("robot.TickerTime")
}

func GetNameFmt() string {
	return Cfg.GetString("login.NameFmt")
}

func SkipNum() int {
	return Cfg.GetInt("Login.SkipNum")
}

func TimeOut() time.Duration {
	return Cfg.GetDuration("robot.Timeout")
}

func GetPerMillisecond() time.Duration {
	return Cfg.GetDuration("Login.PerMillisecond")
}

func GetOfflineSecond() time.Duration {
	return Cfg.GetDuration("robot.OfflineSecond")
}

func GetALiveSecond() int64 {
	return Cfg.GetInt64("robot.ALiveSecond")
}

func Path() string {
	return Cfg.GetString("robot.Path")
}

func ServerAddr() string {
	return Cfg.GetString("login.ServerAddr")
}

func GetEndian() binary.ByteOrder {
	if "123" == Cfg.GetString("robot.Endian") {
		return binary.LittleEndian
	} else {
		return binary.BigEndian
	}
}

func init() {
	// pflag.StringP("robot.CaseName", "C", "None", "填入相应TestCase名称(login, newLogin, battle, talk, mailsignal, equip, upSkill)")
	pflag.StringP("Login.total", "T", "1", "机器人总启动数")
	pflag.StringP("Login.PreFix", "P", "Dgh", "帐号前缀")
	pflag.StringP("Login.SkipNum", "S", "0", "跳过帐号个数")
}
