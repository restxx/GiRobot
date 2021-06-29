package logic

import (
	cfg "GiantQA/GiRobot/Cfg"
	logger "GiantQA/GiRobot/Logger"
	utils2 "GiantQA/GiRobot/utils"
	"GiantQA/common"
	csproto "GiantQA/proto"
	"GiantQA/proto/gate"
	"encoding/json"
	"fmt"
	"github.com/letterbaby/manzo/network"
	"github.com/letterbaby/manzo/rand"
	"github.com/letterbaby/manzo/utils"
	"sort"
	"time"
)

//登录前流程
func (self *DBTRobotV2) getServerList() bool {

	self.CreateNode("SerList")
	url := cfg.Cfg.GetString("Login.ServerGroup")
	Str := cfg.Cfg.GetString("Login.PostStr")
	version := cfg.Cfg.GetInt32("Login.SerVersion")
	postStr := fmt.Sprintf(Str, version, self.CliData.Account)

	code, Dat, err := utils.HttpRequest(url, postStr, nil, "POST", common.GLOBAL_LOGIN_TIMEOUT)
	if code != 200 || err != nil {
		self.CloseNode("SerList", 2)
		return false
	}
	fmt.Printf("%s", Dat)
	_ = json.Unmarshal(Dat, self.ServerGroup)

	{ // 校验是指定的zoneId 是否可用
		zoneId := self.cfgZoneId() //cfg.Cfg.GetInt("Login.ZoneId")
		var Servers *common.Servers
		for _, zone := range self.ServerGroup.Servers {
			if zone.ZoneId == zoneId && zone.Status == 1 {
				Servers = &zone
				break
			}
		}
		if Servers == nil {
			self.CloseNode("SerList", 2)
			return false
		}
	}

	self.CloseNode("SerList", 1)
	return true
}

func (self *DBTRobotV2) getServers() *common.Servers {

	defer func() {
		if err := recover(); err != nil {
			logger.Error("[%s]捕获异常:[%v]", self.CliData.Account, err)
		}
	}()

	zoneId := self.cfgZoneId()

	var Servers *common.Servers
	for _, zone := range self.ServerGroup.Servers {
		if zone.ZoneId == zoneId && zone.Status == 1 {
			Servers = &zone
			break
		}
	}
	if Servers == nil {
		panic("无可用Zone服务器！！")
	}
	return Servers
}

//   worldId 策略
func (self *DBTRobotV2) GetWorldId(IsNewLogin bool) bool {
	// 如果是“新帐号重登录” 所有world永远没有角色
	// 跟据newLoginCnt 自动取模
	Servers := self.getServers()
	if Servers == nil {
		return false
	}
	if IsNewLogin {
		cnt := self.CliData.GetNewLoginCnt()
		mod := cnt % len(Servers.Worlds)
		self.CliData.MpData["WorldId"] = Servers.Worlds[mod].WorldId
		return true
	} else { //普通登录
		// 如果都无角色或都有角色 跟据序号取模
		roles := Servers.Roles
		if len(roles) == 0 || len(roles) == len(Servers.Worlds) {
			// clientId 初始化的时候已经加过skip
			mod := self.CliData.ClientId % len(Servers.Worlds)
			self.CliData.MpData["WorldId"] = Servers.Worlds[mod].WorldId
			return true
		} else if len(roles) > len(Servers.Worlds) {
			// 如果合过服
			mod := self.CliData.ClientId % len(Servers.Worlds)
			self.CliData.MpData["WorldId"] = roles[mod].WorldId
			return true
		} else { // 存在无角色的world则优先进入
			var roleMap = make(map[int]int)
			for _, role := range roles {
				roleMap[role.WorldId] = 1
			}
			for _, world := range Servers.Worlds {
				if roleMap[world.WorldId] != 1 {
					self.CliData.MpData["WorldId"] = world.WorldId
					return true
				}
			}
		}
	}
	return false
}

func (self *DBTRobotV2) GetMiniWorldId() bool {
	servers := self.getServers()
	if servers == nil {
		return false
	}
	if len(servers.Worlds) > 1 {
		sort.Slice(servers.Worlds, func(i, j int) bool {
			return servers.Worlds[i].WorldId < servers.Worlds[j].WorldId
		})
	}
	self.CliData.MpData["WorldId"] = servers.Worlds[0].WorldId
	return true
}

func (self *DBTRobotV2) loginZone() bool {

	self.CreateNode("loginZone")

	Servers := self.getServers()
	if Servers == nil {
		return false
	}
	loginUrl := fmt.Sprintf(cfg.Cfg.GetString("Login.loginZone"), Servers.LoginUrl,
		cfg.Cfg.GetInt32("Login.SerVersion"), self.ServerGroup.Token)
	fmt.Println(loginUrl)
	code, Dat, err := utils.HttpRequest(loginUrl, "", nil, "GET", common.GLOBAL_LOGIN_TIMEOUT)
	if code != 200 || err != nil {
		self.CloseNode("loginZone", 2)
		return false
	}
	Ret := &common.ZoneRet{}
	_ = json.Unmarshal(Dat, Ret)
	if Ret.Error != 0 {
		self.CloseNode("loginZone", 2)
		return false
	}
	tk := &common.ZoneToken{}
	_ = json.Unmarshal(([]byte)(Ret.Msg), tk)
	self.CloseNode("loginZone", 1)

	self.CreateNode("loginWorld")
	{
		loginWorld := fmt.Sprintf("%s?api=world&token=%s&worldId=%d", Servers.LoginUrl, tk.Token, self.CliData.GetInt("WorldId"))
		fmt.Println(loginWorld)
		code, Dat, err := utils.HttpRequest(loginWorld, "", nil, "GET", common.GLOBAL_LOGIN_TIMEOUT)
		if code != 200 || err != nil {
			self.CloseNode("loginWorld", 2)
			return false
		}
		fmt.Printf("%s\n", Dat)
		Ret := &common.ZoneRet{} //这里其实是worldRet
		_ = json.Unmarshal(Dat, Ret)
		if Ret.Error != 0 {
			self.CloseNode("loginWorld", 2)
			return false
		}

		fmt.Println(Ret.Msg)
		wdMsg := &common.WorldRetMsg{}
		_ = json.Unmarshal(([]byte)(Ret.Msg), wdMsg)
		self.loginWorldAck = &common.LoginWorldAck{Token: wdMsg.Token, Ip: wdMsg.IP, Port: int32(wdMsg.Port)}
	}
	self.CloseNode("loginWorld", 1)
	return true
}

// 登录游戏服流程-------------------------------------------------------------------------
func (self *DBTRobotV2) loginReq() {
	self.SetLastStep(utils2.LOGIN)
	self.SetNormalStep(utils2.LOGIN)

	self.CreateNode("LoginGate1")

	req := &gate.LoginReq{Token: self.loginWorldAck.Token,
		Account: self.CliData.Account, Version: "1.0",
		LoginType: gate.LoginType_LoginType_Login_Token,
	}
	msg := common.NewGateMessage(gate.ClientMsgType_ClientType_LoginReq, req)

	_ = self.gateSend(msg)
	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgRaw.MsgId == uint16(gate.ClientMsgType_ClientType_LoginResp) {
			if msgRaw.MsgData.(*gate.LoginResp).Result != gate.ResultType_Success {
				return false
			}
			self.CloseNode("LoginGate1", 1)
			return true
		}
		return false
	}) {
		self.CloseNode("LoginGate1", 2)
	}

	self.CreateNode("LoginGate2")
	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgRaw.MsgId == uint16(gate.ClientMsgType_ClientType_EnterGameResult) {
			if msgRaw.MsgData.(*gate.EnterGameResult).Result != 0 {
				return false
			}
			self.CloseNode("LoginGate2", 1)
			return true
		}
		return false
	}) {
		self.CloseNode("LoginGate2", 2)
	}
}

func (self *DBTRobotV2) loginGame() {

	logger.Debug("[%s] loginGame", self.CliData.Account)
	self.CreateNode("loginGame")

	msg := &csproto.CommonMessage{}
	msg.Code = csproto.Cmd_GAME_LOGIN
	req := &csproto.LoginGameReq{Token: self.loginWorldAck.Token}
	msg.LoginGameReq = req
	_ = self.send(msg)

	var createRole bool
	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
			if msgData.Code == csproto.Cmd_GAME_LOGIN {
				req := msgData.LoginGameAck
				if req.CreateRole {
					createRole = true
				}
				return true
			}
		}
		return false
	}) {
		self.CloseNode("loginGame", 2)
		self.CloseNode("connect", 2)
		self.ReLogin()
		return
	}
	// 需要创角
	if createRole {
		self.createRole(1)
	}
	// 等待LOGIN_END
	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
			if msgData.Code == csproto.Cmd_GAME_LOGIN_END {
				self.CloseNode("loginGame", 1)
				return true
			}
		}
		return false
	}) {
		self.CloseNode("loginGame", 2)
		return
	}

	//if createRole {
	//	self.add4Skill()
	//}

	// 放弃领奖
	self.bossReward(true)

	if cfg.GetCaseName() == "login" {
		sec := cfg.GetALiveSecond()
		self.WaitTime(1000 * sec)
		self.ReLogin()
	} else if cfg.GetCaseName() == "newLogin" {
		sec := cfg.GetALiveSecond()
		self.WaitTime(1000 * sec)
		self.loginOnce.Do(func() {
			self.AddWg()
			self.Stop()
			self.CliData.Trigger(utils2.NEW_RELOGIN, self.CliData)
		}) //end_Once.Do
	} // end_if
	// 切换主状态为空闲
	self.SetNormalStep(utils2.IDLE)
}

func (self *DBTRobotV2) bossReward(giveUp bool) {
	msg := &csproto.CommonMessage{}
	msg.Code = csproto.Cmd_SCENE_BOSS_REWARD
	req := &csproto.SceneBossRewardReq{GiveUp: giveUp}
	msg.SceneBossRewardReq = req
	_ = self.send(msg)

	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
			if msgData.Code == csproto.Cmd_SCENE_BOSS_REWARD {
				return true
			}
		}
		return false
	}) {
	}
}

func (self *DBTRobotV2) createRole(charModId int32) {

	defer self.CloseNode("CreateRole", 2)

	msg := &csproto.CommonMessage{}
	msg.Code = csproto.Cmd_GAME_CREATE_ROLE
	req := &csproto.CreateRoleReq{}

	req.RoleName = fmt.Sprintf("%sw%d", self.CliData.Account, self.CliData.GetInt("WorldId"))
	req.CharacterId = charModId
	req.Face = ""
	msg.CreateRoleReq = req

	self.CreateNode("CreateRole")
	_ = self.send(msg)

	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
			if msgData.Code == csproto.Cmd_GAME_CREATE_ROLE {
				return true
			}
		}
		return false
	}) {
		self.CloseNode("CreateRole", 2)
		return
	}
	self.CloseNode("CreateRole", 1)
}

// GM--------------------------------------------------

func (self *DBTRobotV2) add4Skill() {

	_skill := []int32{110102, 110202, 110302, 110402, 110502, 110602, 110702, 110802, 110902, 111002}
	lsIdx := rand.RandIntN(0, 9, 4)

	for i, v := range lsIdx {
		{
			// 加技能
			msg := &csproto.CommonMessage{}
			msg.Code = csproto.Cmd_GAME_CMD
			req := &csproto.GCommandReq{}
			req.Cmd = fmt.Sprintf("AddItem %d 1", _skill[v])
			msg.GCommandReq = req
			_ = self.send(msg)
			if !self.Wait(func(msgRaw *network.RawMessage) bool {
				if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
					if msgData.Code != csproto.Cmd_GAME_GAIN_NTF {
						return false
					}
					msg := msgData.GetGainInfoNtf()
					for _, v := range msg.Datas {
						if v.Dt == csproto.GainType_SKILL {
							logger.Info("[%s] %s OK", self.Account(), req.Cmd)
							return true
						}
					}
				}
				return false
			}) {
				return
			}
		}

		{
			// 安装技能
			msg2 := &csproto.CommonMessage{}
			msg2.Code = csproto.Cmd_GAME_SET_SKILL
			req2 := &csproto.SetSkillReq{}
			req2.SkillId = _skill[v]
			req2.Idx = int32(i + 1)
			msg2.SetSkillReq = req2

			_ = self.send(msg2)
			if !self.Wait(func(msgRaw *network.RawMessage) bool {
				if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
					if msgData.Code == csproto.Cmd_GAME_SET_SKILL {
						logger.Info("[%s] SetSkill [%d] OK!", self.Account(), _skill[v])
						return true
					}
				}
				return false
			}) {
				return
			} // end if
		}
	} // end_for
}

func (self *DBTRobotV2) Gm_AddExp(cmd string, num int64) {

	cmdString := fmt.Sprintf("%s %d", cmd, num)

	msg := &csproto.CommonMessage{}
	msg.Code = csproto.Cmd_GAME_CMD
	req := &csproto.GCommandReq{}
	req.Cmd = cmdString
	msg.GCommandReq = req

	_ = self.send(msg)
	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
			if msgData.Code != csproto.Cmd_GAME_GAIN_NTF {
				return false
			}
			msg := msgData.GetGainInfoNtf()
			for _, Ginfo := range msg.Datas {
				if Ginfo.Dt != csproto.GainType_ROLE_EXP {
					continue
				}
				if Ginfo.DtLong == num {
					self.RoleInfo.Level = 10
					return true
				}
			}
		} //断言
		return false
	}) {
	}
}

func (self *DBTRobotV2) _GM_SkillStress(cmd string, num int32) {

	sCmd := fmt.Sprintf("%s %d", cmd, num)

	msg := &csproto.CommonMessage{}
	msg.Code = csproto.Cmd_GAME_CMD
	req := &csproto.GCommandReq{}
	req.Cmd = sCmd
	msg.GCommandReq = req
	_ = self.send(msg)
	//if !self.Wait(func(msgRaw *network.RawMessage) bool {
	//	if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
	//		if msgData.Code != csproto.Cmd_GAME_GAIN_NTF {
	//			return false
	//		}
	//		msg := msgData.GetGainInfoNtf()
	//		var _count int32 = 0
	//		for _, v := range msg.Datas {
	//			if v.Dt != csproto.GainType_SKILL {
	//				continue
	//			}
	//			if v.Skills[0].Package != -1 && v.Skills[0].ChangeNum > 0 {
	//				_count = _count + v.Skills[0].ChangeNum
	//				self.Skills = append(self.Skills, v.Skills...)
	//			}
	//		}
	//		if _count == num {
	//			return true
	//		}
	//	}
	//	return false //lambda 返回
	//}) {
	//}

}

func (self *DBTRobotV2) _GM(cmd string, gt csproto.GainType) {
	msg := &csproto.CommonMessage{}
	msg.Code = csproto.Cmd_GAME_CMD
	req := &csproto.GCommandReq{}
	req.Cmd = cmd
	msg.GCommandReq = req
	_ = self.send(msg)
	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
			if msgData.Code == csproto.Cmd_GAME_GAIN_NTF {
				msg := msgData.GetGainInfoNtf()
				ok := false
				for _, v := range msg.Datas {
					if v.Dt == csproto.GainType_EQUIP {
						self.Equips = append(self.Equips, v.Equips...)
					} else if v.Dt == csproto.GainType_SKILL {
						if v.Skills[0].Package != -1 && v.Skills[0].Count > 0 {
							self.Skills = append(self.Skills, v.Skills...)
						} else {
							logger.Debug("Gm里Gain有Package==-1 || Count==0的技能 %v", v.Skills)
							return false
						}
						logger.Debug("GAIN_NTF [%v] skill=%v", self.CliData.Account, self.Skills)
					}
					if gt == v.Dt {
						ok = true
					}
				} //end_for
				return ok
			} // Code比较
		} // 断言
		return false
	}) {
		return
	}
}

func (self *DBTRobotV2) GM_Add_Currencys(cmd string, num int64) {
	cmdString := fmt.Sprintf("%s %d", cmd, num)

	msg := &csproto.CommonMessage{}
	msg.Code = csproto.Cmd_GAME_CMD
	req := &csproto.GCommandReq{}
	req.Cmd = cmdString
	msg.GCommandReq = req
	_ = self.send(msg)

	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
			if msgData.Code != csproto.Cmd_GAME_GAIN_NTF {
				return false
			}
			msg := msgData.GetGainInfoNtf()
			for _, Ginfo := range msg.Datas {
				if Ginfo.Dt != csproto.GainType_CURRENCY {
					continue
				}
				for _, currency := range Ginfo.Currencys {
					if currency.ChangeNum == int32(num) {
						return true
					}
				}
			}
		} //断言
		return false
	}) {
	}
}

func (self *DBTRobotV2) set4Skill() {

	_skills := []int32{110101, 110201, 110301, 110401, 110501, 110601, 110701, 110801, 110901, 111001}
	lsIdx := rand.RandIntN(0, 9, 4)

	for i, v := range lsIdx {

		msg := &csproto.CommonMessage{}
		msg.Code = csproto.Cmd_GAME_SET_SKILL
		req := &csproto.SetSkillReq{}
		req.SkillId = _skills[v]
		req.Idx = int32(i)
		msg.SetSkillReq = req

		_ = self.send(msg)
		self.Wait(func(msgRaw *network.RawMessage) bool {
			if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
				if msgData.Code == csproto.Cmd_GAME_SET_SKILL {
					return true
				}
			}
			return false
		})
	}
}

// Case ------------------------------------------------------------------------
func (self *DBTRobotV2) getUserSeasonInfo() {

	if self.GetNormalStep() < utils2.IDLE {
		return
	}
	msg := &csproto.CommonMessage{}
	msg.Code = csproto.Cmd_SEASON_GET_INFO
	msg.GetUserSeasonInfoReq = &csproto.GetUserSeasonInfoReq{}

	self.CreateNode("Rank")
	_ = self.send(msg)
	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
			if msgData.Code == csproto.Cmd_SEASON_GET_INFO {
				self.CloseNode("Rank", 1)
				return true
			}
		}
		return false
	}) {
		self.CloseNode("Rank", 2)
	}
}

func (self *DBTRobotV2) searchBoss() {

	now := time.Now().Unix()
	if self.bossTick > 0 {
		replay := now - self.bossTick
		if replay < self.replayTick {
			panic("xxxxxxxxxxxxxxxxxxxxxxx")
		}
	}

	// 升太高无法正常战斗升级
	//if self.RoleInfo.GetLevel() < 10{
	//	self.Gm_AddExp("AddItem 2", 4999)
	//	return
	//}

	self.SetLastStep(utils2.BUSYING)
	self.SetNormalStep(utils2.BUSYING)

	msg := &csproto.CommonMessage{}
	msg.Code = csproto.Cmd_SCENE_SEARCH_BOSS
	req := &csproto.SceneSearchBossReq{}
	msg.SceneSearchBossReq = req

	self.CreateNode("SearchBoss")
	//打印账号
	logger.Debug("DBTRobot:SearchBoss conn:%v,uid:%v", self.Conn, self.Account())

	_ = self.send(msg)

	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
			if msgData.Code == csproto.Cmd_BATTLE_UPDATE_NTF {
				frames := msgData.BattleUpdateNtf.Frames
				if len(frames) > 0 {
					dt := (frames[len(frames)-1].FrameNumber * common.GLOBAL_BATTLE_DT) / 1000
					// fmt.Println("FrameNumber = ", frames[len(frames)-1].FrameNumber)
					self.replayTick = int64(dt / 2)
					self.bossTick = now
				}
				self.BattleCount++ //战斗计数

				self.CloseNode("SearchBoss", 1)
				return true
			}
		}
		return false
	}) {
		self.CloseNode("SearchBoss", 2)
		return
	}

	// 战斗需延时
	self.WaitTime(1000 * self.replayTick)
	self.bossReward(true)

	if self.BattleCount%5 == 0 {
		self.getUserSeasonInfo()
	}

	self.WaitTime(1000 * 2)
	self.SetNormalStep(utils2.IDLE)
}

func (self *DBTRobotV2) Talk(Msg string) {

	self.SetLastStep(utils2.BUSYING)
	self.SetNormalStep(utils2.BUSYING)

	msg := &csproto.CommonMessage{}
	msg.Code = csproto.Cmd_WORLD_CHAT
	req := &csproto.WorldChatReq{}
	req.DestId = 0
	req.Msg = fmt.Sprintf("aa测试压力测试 %s 测试压力测试aa", self.Account())
	msg.WorldChatReq = req

	self.CreateNode("Talk")
	// 喊话需判断是不是自己发出的
	_ = self.send(msg)

	if !self.Wait(
		func(msgRaw *network.RawMessage) bool {
			if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
				if msgData.Code != csproto.Cmd_WORLD_CHAT_NTF {
					return false
				}
				for _, v := range msgData.WorldChatNtf.Msgs {
					if v.SrcRole.RoleId == self.RoleInfo.RoleId {
						logger.Debug("[%s] Chat", self.Account())
						self.CloseNode("Talk", 1)
						return true
					}
				}
			} //断言
			return false
		}) {
		self.CloseNode("Talk", 2)
	}
	// 喊话需延时
	self.WaitTime(1000 * 10)
	self.SetNormalStep(utils2.IDLE)
}

func (self *DBTRobotV2) _sendmail(RoleIDs, Content string, ItemId, num int) bool {
	// curl -X POST -H "Content-Type:application/json" -d '{"head":{"Cmdid":4128,"Seqid":104951,"ServiceName":"FZDL","SendTime":1617952551,"Version":0,"Authenticate":"49612052a1cd4e1b66ca65d1ba11ad7d"},"body":{"AreaId":0,"PlatId":0,"Partition":201,"MailTitle":"QA压测","MailAddress":"系统","Roles":"X,X","MailContent":"XXXXX","Attchment":[{"ItemId":%d,"Num":%d}]}}' http://idip:port/

	//defer self.CloseNode("sendMail", 2)
	//self.CreateNode("sendMail")

	head := make(map[string]interface{})
	head["Content-Type"] = "application/json"
	postFmt := `{"head":{"Cmdid":4128,"Seqid":104951,"ServiceName":"FZDL","SendTime":%d,"Version":0,"Authenticate":"49612052a1cd4e1b66ca65d1ba11ad7d"},"body":{"AreaId":0,"PlatId":0,"Partition":%d,"MailTitle":"QA压测","MailAddress":"系统","Roles":"%s","MailContent":"%s","Attchment":[{"ItemId":%d,"Num":%d}]}}`
	postStr := fmt.Sprintf(postFmt, uint32(time.Now().Unix()), self.cfgZoneId(), RoleIDs, Content, ItemId, num)
	fmt.Println(postStr)
	url := cfg.Cfg.GetString("Login.sendMail")

	code, Dat, err := utils.HttpRequest(url, postStr, head, "POST", common.GLOBAL_LOGIN_TIMEOUT)
	if code != 200 || err != nil {
		logger.Error("帐号[%s] Http 发送邮件失败 Code=[%d] Err[%v]", self.Account(), code, err)
		return false
	}
	sendRet := &common.GMMailRet{}
	_ = json.Unmarshal(Dat, sendRet)
	if sendRet.Head.Result != 0 {
		logger.Error("帐号[%s] 发送邮件返回失败结构 %v", self.Account(), sendRet)
		return false
	}
	fmt.Printf("%s--------", Dat)
	// 等待邮件Gain返回
	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {

			if msgData.Code == csproto.Cmd_GAME_GAIN_NTF {
				msg := msgData.GetGainInfoNtf()

				for _, v := range msg.Datas {
					if v.Dt == csproto.GainType_MAIL {
						if v.Mail.Body == Content {
							//self.CloseNode("sendMail", 1)
							return true //匿名函数返回
						}
						logger.Error("[%s] Mail.Boyd[%s]----Content[%s]", self.Account(), v.Mail.Body, Content)
					}
				}
			} //end_if
		} //断言
		return false //匿名函数返回
	}) {
		return false //if 返回
	}
	return true
}

func (self *DBTRobotV2) mailSignal() {

	self.SetLastStep(utils2.BUSYING)
	self.SetNormalStep(utils2.BUSYING)
	// 发送邮件-----------------------------------------------------------
	sRoleId := fmt.Sprintf("%d", self.RoleInfo.RoleId)
	mailContext := fmt.Sprintf("%s:%s:%d", sRoleId, sRoleId, uint32(time.Now().Unix()))
	if !self._sendmail(sRoleId, mailContext, 1, 998) {
		logger.Error("[%s] _sendmail Error", self.Account())
		self.ReLogin()
		return
	}
	self.WaitTime(1000 * self.cfgSendDelay())

	// 打开邮件--------------------------------------------------
	mails := self.MailInfo.GetMails()
	if len(mails) == 0 {
		self.WaitTime(2000)
		self.SetNormalStep(utils2.IDLE)
		return
	}
	mail := mails[0]
	mailOpen := func() *csproto.CommonMessage {
		msg := &csproto.CommonMessage{}
		msg.Code = csproto.Cmd_MAIL_OPEN
		req := &csproto.OpenMailReq{}
		req.Id = self.MailInfo.Mails[0].Id
		msg.OpenMailReq = req
		return msg
	}
	// 处理附件
	mailatch := func() *csproto.CommonMessage {
		msg := &csproto.CommonMessage{}
		msg.Code = csproto.Cmd_MAIL_ATCH
		req := &csproto.MailAtchReq{}

		if self.MailInfo != nil {
			l := len(self.MailInfo.Mails)
			if l > 0 {
				req.Id = self.MailInfo.Mails[0].Id
			}
		}

		msg.MailAtchReq = req
		return msg
	}

	// 删除邮件
	maildel := func() *csproto.CommonMessage {
		msg := &csproto.CommonMessage{}
		msg.Code = csproto.Cmd_MAIL_DEL
		req := &csproto.DelMailReq{}

		if self.MailInfo != nil {
			l := len(self.MailInfo.Mails)
			if l > 0 {
				req.Id = self.MailInfo.Mails[0].Id
			}
		}

		msg.DelMailReq = req
		return msg
	}

	if mail.Opened == false {
		logger.Debug("[%s] openMail", self.Account())
		msg := mailOpen()
		self.CreateNode("openMail")
		_ = self.send(msg)

		if !self.Wait(func(msgRaw *network.RawMessage) bool {
			if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
				if msgData.Code == csproto.Cmd_MAIL_OPEN {
					self.CloseNode("openMail", 1)
					return true
				}
			}
			return false
		}) {
			self.CloseNode("openMail", 2)
		}
	} else { // end open
		logger.Debug("[%s] haven`t mail need opened！", self.Account())
	}

	if len(mail.Items) > 0 {
		logger.Debug("[%s] atchMail", self.Account())
		msg := mailatch()
		self.CreateNode("atchMail")
		_ = self.send(msg)

		if !self.Wait(func(msgRaw *network.RawMessage) bool {
			if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
				if msgData.Code == csproto.Cmd_MAIL_ATCH {
					self.CloseNode("atchMail", 1)
					return true
				}
			}
			return false
		}) {
			self.CloseNode("atchMail", 2)
		}
	} else { //end ath
		logger.Debug("[%s] haven`t mail need attach！", self.Account())
	}

	logger.Debug("[%s] DelMail", self.Account())
	msg := maildel()
	self.CreateNode("DelMail")
	_ = self.send(msg)

	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
			if msgData.Code == csproto.Cmd_MAIL_DEL {
				self.MailInfo.Mails = msgData.DelMailAck.Mails
				self.CloseNode("DelMail", 1)
				return true
			}
		}
		return false
	}) {
		self.CloseNode("DelMail", 2)
	}

	self.WaitTime(1000 * self.cfgDelDelay())
	self.SetNormalStep(utils2.IDLE)
}

const EID int32 = 201001210

func (self *DBTRobotV2) _onEquip() bool {

	logger.Debug("_on--------------------------------------------------------")

	self.CreateNode("Equip")

	msg := &csproto.CommonMessage{}
	msg.Code = csproto.Cmd_EQUIP_ON

	ssid := int32(0)
	for _, v := range self.Equips {
		if v.CfgId == EID {
			ssid = v.Id
			break
		}
	}

	if ssid == 0 {
		panic("xxxxxxxxxxxxxxxxxxxxxxxxx")
	}
	req := &csproto.EquipOnReq{}
	req.EquipId = ssid
	req.EquipPos = int32(csproto.EquipPos_Clothes)
	req.HeroId = self.CharInfo.Chars[0].GetBase().GetId()

	msg.EquipOnReq = req

	_ = self.send(msg)
	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
			if msgData.Code == csproto.Cmd_GAME_GAIN_NTF {
				ntf := msgData.GainInfoNtf
				for _, v := range ntf.Datas {
					if v.Dt == csproto.GainType_EQUIP {
						return true
					}
				}
			}
		}
		return false
	}) {
		self.CloseNode("Equip", 2)
		return false
	}

	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
			if msgData.Code == csproto.Cmd_EQUIP_ON {
				return true
			}
		}
		return false
	}) {
		self.CloseNode("Equip", 2)
		return false
	}

	self.CloseNode("Equip", 1)

	self.WaitTime(1000)
	self.SetNormalStep(utils2.IDLE)
	return true
}

func (self *DBTRobotV2) _offEquip() {

	logger.Debug("_off--------------------------------------------------------")

	self.CreateNode("OffEquip")

	msg := &csproto.CommonMessage{}
	msg.Code = csproto.Cmd_EQUIP_OFF

	req := &csproto.EquipOffReq{}
	req.HeroId = self.CharInfo.GetChars()[0].GetBase().GetId()
	req.EquipPos = int32(csproto.EquipPos_Clothes)

	msg.EquipOffReq = req

	_ = self.send(msg)

	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
			if msgData.Code == csproto.Cmd_GAME_GAIN_NTF {
				ntf := msgData.GainInfoNtf
				for _, v := range ntf.Datas {
					if v.Dt == csproto.GainType_EQUIP {
						return true
					}
				}
			}
		}
		return false
	}) {
		self.CloseNode("OffEquip", 2)
		return
	}

	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
			if msgData.Code == csproto.Cmd_EQUIP_OFF {
				return true
			}
		}
		return false
	}) {
		self.CloseNode("OffEquip", 2)
		return
	}
	self.CloseNode("OffEquip", 1)

	self.WaitTime(1000)

	self.SetNormalStep(utils2.IDLE)
}

func (self *DBTRobotV2) equip() {

	self.SetLastStep(utils2.BUSYING)
	self.SetNormalStep(utils2.BUSYING)

	// 判断r.Equips是否为有需要的两件装备
	// 没有则用GM命令塞两个
	flag := 0 // 是否有装备
	if self.Equips != nil {
		// 101100020   训练用剑2  // 101101030   训练用剑3
		for _, e := range self.Equips {
			if e.CfgId == EID {
				flag += 1
			}
		}
	}

	//简便写法  不要去判断长度
	if self.CharInfo.GetChars()[0].GetEquips() != nil {
		flag += 1
	}

	if flag < 1 {
		//GM塞装备
		cmd := fmt.Sprintf("AddItem %d 1", EID)
		self._GM(cmd, csproto.GainType_EQUIP)
		logger.DebugV(cmd, "-----------")
		self.SetNormalStep(utils2.IDLE)
		return
	}

	if self.CharInfo.GetChars()[0].GetEquips() == nil {
		self._onEquip()
	} else {
		self._offEquip()
	}
}

const SKILLCID int32 = 60011030 //60011040, 60011050, 60011060  都可用
func (self *DBTRobotV2) UpSkill() {

	self.SetLastStep(utils2.BUSYING)
	self.SetNormalStep(utils2.BUSYING)

	self.Skills = nil
	self._GM_SkillStress("SkillStress", 1)
	logger.Debug("GM [%v] SkillStress 1", self.Account())

	self.WaitTime(1000)

	msg := &csproto.CommonMessage{}
	msg.Code = csproto.Cmd_GAME_SECOND_SKILL_LVUP
	req := &csproto.SecondSkillLvUpReq{
		Heroid:   self.CharInfo.Chars[0].GetBase().GetId(),
		Skillcid: SKILLCID,
	}
	msg.SecondSkillLvUpReq = req

	self.CreateNode("UpSkill")
	_ = self.send(msg)

	if !self.Wait(func(msgRaw *network.RawMessage) bool {
		if msgData, ok := msgRaw.MsgData.(*csproto.CommonMessage); ok {
			if msgData.Code == csproto.Cmd_GAME_SECOND_SKILL_LVUP {
				self.CloseNode("UpSkill", 1)
				return true
			}
		}
		return false
	}) {
		self.CloseNode("UpSkill", 2)
	}
	self.WaitTime(1000)
	self.SetNormalStep(utils2.IDLE)
}
