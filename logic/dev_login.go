package logic

import (
	logger "GiantQA/GiRobot/Logger"
	"GiantQA/common"
	"encoding/json"
	"fmt"
	"github.com/letterbaby/manzo/rand"
	"github.com/letterbaby/manzo/utils"
)

func (self *DBTRobotV2) oldRun() bool {
	if !self.auth() {
		self.ReLogin()
		return false
	}

	if !self.loginWorld() {
		self.ReLogin()
		return false
	}

	self.runGame()
	return true
}

// --------私有方法--------
func (self *DBTRobotV2) auth() bool {

	self.CreateNode("auth")
	defer self.CloseNode("auth", 2)

	//url := fmt.Sprintf("%s?api=local&account=%s&debug=1",
	url := fmt.Sprintf("%s?api=local&account=%s",
		self.SvrInfo.LoginUrl, self.CliData.Account)

	code, rt, err := utils.HttpRequest(url, "", nil, "GET", common.GLOBAL_LOGIN_TIMEOUT)
	if code != 200 || err != nil {
		logger.Error("Test2:userLogin uid:%v,url:%v", self.CliData.ClientId, err)
		return false
	}

	ar := &common.LoginResult{}
	err = json.Unmarshal(rt, ar)
	if err != nil {
		logger.Error("Test2:userLogin uid:%v,json:%v", self.CliData.ClientId, string(rt))
		return false
	}

	if ar.Error != 0 {
		logger.Error("Test2:userLogin uid:%v,json:%v", self.CliData.ClientId, string(rt))
		return false
	}

	self.userLoginAck = &common.UserLoginAck{}
	err = json.Unmarshal([]byte(ar.Msg), self.userLoginAck)
	if err != nil {
		logger.Error("Test2:userLogin uid:%v,json:%v", self.CliData.ClientId, string(rt))
		return false
	}

	self.CloseNode("auth", 1)
	return true
}

func (self *DBTRobotV2) loginWorld() bool {

	self.CreateNode("login")
	defer self.CloseNode("login", 2)

	w := int32(2)
	l := int32(len(self.SvrInfo.Worlds))
	if l > 0 {
		idx := rand.RandInt(0, l-1)
		w = self.SvrInfo.Worlds[idx]
	}

	url := fmt.Sprintf("%s?api=login&token=%s&worldId=%d",
		self.SvrInfo.LoginUrl, self.userLoginAck.Token, w)

	code, rt, err := utils.HttpRequest(url, "", nil, "GET", common.GLOBAL_LOGIN_TIMEOUT)
	if code != 200 || err != nil {
		logger.Error("Test2:loginWorld uid:%v,url:%v", self.CliData.ClientId, err)
		return false
	}

	ar := &common.LoginResult{}
	err = json.Unmarshal(rt, ar)
	if err != nil {
		logger.Error("Test2:loginWorld uid:%v,json:%v", self.CliData.ClientId, string(rt))
		return false
	}

	if ar.Error != 0 {
		logger.Error("Test2:loginWorld uid:%v,json:%v", self.CliData.ClientId, string(rt))
		return false
	}

	self.loginWorldAck = &common.LoginWorldAck{}
	err = json.Unmarshal([]byte(ar.Msg), self.loginWorldAck)
	if err != nil {
		logger.Error("Test2:loginWorld uid:%v,json:%v", self.CliData.ClientId, string(rt))
		return false
	}

	self.CloseNode("login", 1)
	return true
}
