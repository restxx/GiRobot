package logic

import (
	cfg "GiantQA/GiRobot/Cfg"
	logger "GiantQA/GiRobot/Logger"
	meter "GiantQA/GiRobot/Meter"
	GiRobot "GiantQA/GiRobot/Robot"
	utils2 "GiantQA/GiRobot/utils"
	"GiantQA/common"
	csproto "GiantQA/proto"
	"GiantQA/proto/gate"
	"context"
	"fmt"
	"github.com/letterbaby/manzo/network"
	"github.com/letterbaby/manzo/utils"
	"net"
	"sync"
	"sync/atomic"
)

var (
	netcfg *network.Config
)

const MAXCHAN = 2000

func init() {
	netcfg = &network.Config{}
	parser := network.NewProtocParser(-1)
	_ = parser.Register(uint16(csproto.Cmd_NONE), csproto.CommonMessage{})
	_ = parser.Register(uint16(gate.ClientMsgType_ClientType_EnterGameResult), gate.EnterGameResult{})
	_ = parser.Register(uint16(gate.ClientMsgType_ClientType_LoginResp), gate.LoginResp{})
	_ = parser.Register(uint16(gate.ClientMsgType_ClientType_LoginReq), gate.LoginReq{})

	netcfg.Parser = parser
}

type DBTRobotV2 struct {
	GiRobot.Robot
	Meter *meter.MtManager

	Conn        network.IConn
	Parser      network.IMessage
	IsClosed    int32 //链接断开标记
	SvrInfo     *common.DevServerInfo
	seq         uint32 //发送计数
	BattleCount uint64 // 每5次拉一下排名

	//new 登录
	ServerGroup *common.ServerGroup

	//lastStep   int // 完成的步骤
	//normalStep int // 当前的步骤
	//
	//MainTicker *time.Ticker

	//战斗回放
	replayTick int64
	bossTick   int64

	userLoginAck  *common.UserLoginAck
	loginWorldAck *common.LoginWorldAck
	RoleInfo      *csproto.RoleInfoNtf
	MailInfo      *csproto.MailInfoNtf
	CharInfo      *csproto.RoleCharNtf

	Equips []*csproto.EquipInfo
	Skills []*csproto.SkillInfo

	sync.Once            // stop函数专用
	loginOnce *sync.Once //重登录专用
}

func NewDBTRobotV2() *DBTRobotV2 {
	rob := &DBTRobotV2{
		ServerGroup: &common.ServerGroup{},
		loginOnce:   &sync.Once{},
		seq:         0,
		BattleCount: 0,
		IsClosed:    0,
	}
	rob.InCome = make(chan interface{}, MAXCHAN)
	rob.Ctx, rob.CancelFunc = context.WithCancel(context.Background())
	rob.SetNormalStep(utils2.START)
	rob.SetLastStep(utils2.START)

	rob.Robot.ProcessMsg = rob.ProcessMsg
	rob.Robot.ReLogin = rob.ReLogin
	rob.Robot.MainAction = rob.mainActive
	rob.Robot.HeartBeat = rob.heartBeat
	return rob
}

func (self *DBTRobotV2) Init(info *common.DevServerInfo, manager *meter.MtManager, data *utils2.ClientData) {
	self.SvrInfo = info
	self.Meter = manager
	self.SetCliData(data)
}

func (self *DBTRobotV2) AddSeq() uint32 {
	self.seq++
	return self.seq
}

// 关闭机器人和Conn
func (self *DBTRobotV2) Stop() {
	self.Once.Do(func() {
		atomic.StoreInt32(&self.IsClosed, 1)
		if self.Conn != nil {
			self.Conn.Close()
			logger.Debug("[%s]主动关闭网络连接", self.CliData.Account)
		}
		if self.MainTicker != nil {
			self.MainTicker.Stop()
		}
		self.CancelFunc()
		self.WgDone()
	})
}

func (self *DBTRobotV2) CreateNode(name string) {
	self.Meter.CreateNode(name)
}

func (self *DBTRobotV2) CloseNode(name string, status uint8) {
	self.Meter.CloseNode(name, status)
}

func (self *DBTRobotV2) Run(IsNewLogin bool) {
	if cfg.Cfg.GetBool("Login.devLogin") {
		self.oldRun()
		return
	}
	self.new_Run(IsNewLogin)
}

func (self *DBTRobotV2) new_Run(IsNewLogin bool) {

	if !self.getServerList() {
		self.ReLogin()
		return
	}
	if cfg.Cfg.GetBool("Login.FirstWorld") {
		self.GetMiniWorldId()
	} else if !self.GetWorldId(IsNewLogin) {
		self.ReLogin()
		return
	}
	logger.Info("\n[%s]-GetWorldId=[%d]-----------", self.CliData.Account, self.CliData.GetInt("WorldId"))
	if !self.loginZone() {
		self.ReLogin()
		return
	}
	self.runGame()
}

// --------私有方法--------
func (self *DBTRobotV2) runGame() {
	// 重置！！！！！！！！

	address := fmt.Sprintf("%s:%d", self.loginWorldAck.Ip, self.loginWorldAck.Port)
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		logger.Error("DBTRobotV2:battle addr:%v", err)
		return
	}

	self.CreateNode("connect")
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		self.CloseNode("connect", 2)
		logger.Error("DBTRobotV2:battle addr:%v", err)
		return
	}
	self.CloseNode("connect", 1)
	conn.SetNoDelay(true)
	self.Conn = &network.Conn{}

	self.Conn.Init(netcfg, conn)

	logger.Debug("DBTRobotV2:connect conn:%v,uid:%v", self.Conn, self.CliData.ClientId)

	self.SetNormalStep(utils2.LOGIN)

	logger.Debug("[%s] 启动recv  mainloop协程", self.CliData.Account)
	go self.recv()
	go self.MainLoop()

}

func (self *DBTRobotV2) recv() {
	defer utils.CatchPanic()
	defer self.Stop()
	defer logger.Debug("[%s]关闭Recv协程", self.CliData.Account)

	logger.Debug("DBTRobot:recv uid:%v,conn:%v", self.CliData.ClientId, self.Conn)

	for { // 这里必须登录完成处于空闲状态
		if GiRobot.IsStop() && self.GetNormalStep() == utils2.IDLE {
			logger.Info("-------Recv---Stop--------%v, %v", GiRobot.IsStop(), self.GetNormalStep())
			self.Stop()
			return
		}
		if atomic.LoadInt32(&self.IsClosed) > 0 {

			return
		}
		msg, err := self.Conn.RecvMsg()
		if err != nil {
			if !(cfg.GetCaseName() == "newLogin" || cfg.GetCaseName() == "login" || GiRobot.IsStop()) {
				// 如果是测试登录 不要记录
				logger.Error("DBTRobotV2:recv conn:%v,recvMsg:%v", self.Conn, err)
			}
			logger.Debug("DBTRobotV2:recv conn:%v,recvMsg:%v", self.Conn, err)
			return
		}
		if msg.MsgId == uint16(csproto.Cmd_NONE) {
			msgdata := msg.MsgData.(*csproto.CommonMessage)
			logger.Debug("DBTRobotV2 [%s] Recv CMD=[%v] conn=%v", self.CliData.Account, msgdata.Code, self.Conn)
			if msgdata.ErrorCode == csproto.ErrorCode_SUCCESS { //|| csproto.ErrorCode_WORLD_MERGING {
				self.InCome <- msg
			} else {
				logger.Error("DBTRobotV2 [%s] Recv errorCode=[%v], CMD=[%v]", self.CliData.Account, msgdata.ErrorCode, msgdata.Code)
			}
		} else {
			// 非CommonMessage类型
			fmt.Printf("[%s]RawMsg id=[%v]\n", self.CliData.Account, gate.ClientMsgType(msg.MsgId))
			self.InCome <- msg
		}
	} // for end
}

func (self *DBTRobotV2) heartBeat() {

	if self.GetNormalStep() < utils2.IDLE {
		return
	}

	if atomic.LoadInt32(&self.IsClosed) > 0 {
		return
	}

	logger.Debug("Send Ping-----------")
	msg := &csproto.CommonMessage{}
	msg.Code = csproto.Cmd_PING
	_ = self.send(msg)
}

func (self *DBTRobotV2) send(msg *csproto.CommonMessage) error {

	rmsg := common.NewSCRawMessage(msg)
	rmsg.Seq = self.AddSeq()
	err := self.Conn.SendMsg(rmsg)
	//manzologger.Debug("Send msg.Cmd=[%v] Acc=[%s], %v", msg.Code, self.CliData.Account, self.Conn)
	if err != nil {
		logger.Error("Send msg.Cmd=[%v] Acc=[%s], %v", msg.Code, self.CliData.Account, err)
		self.ReLogin()
	}
	return err
}

func (self *DBTRobotV2) gateSend(msg *network.RawMessage) error {
	msg.Seq = self.AddSeq()
	err := self.Conn.SendMsg(msg)
	//manzologger.Debug("Gate Send msg.Cmd=[%v] Acc=[%s], %v", msg.MsgId, self.player.Account, self.Conn)
	if err != nil {
		logger.Error("Gate Send msg.Cmd=[%v] Acc=[%s], %v", msg.MsgId, self.CliData.Account, err)
		self.ReLogin()
	}
	return err
}

func (self *DBTRobotV2) ProcessMsg(Msg interface{}) {
	msg := Msg.(*network.RawMessage)
	if msg.MsgId == uint16(csproto.Cmd_NONE) {
		Msg := msg.MsgData.(*csproto.CommonMessage)
		_ = HandleMsg.Call(Msg.Code, self, Msg)
	} else {
		//处理非commonMessage
		_ = HandleMsg.Call(gate.ClientMsgType(msg.MsgId), self, msg)
	}
}

func (self *DBTRobotV2) mainActive() {
	if GiRobot.IsStop() && self.GetNormalStep() == utils2.IDLE {
		logger.Debug("-------mainloop-------Stop--------%v, %v", GiRobot.IsStop(), self.GetNormalStep())
		self.Stop()
		return
	}
	if self.GetLastStep() != self.GetNormalStep() {
		switch self.GetNormalStep() {
		case utils2.LOGIN:
			self.loginReq()
			self.loginGame()
		case utils2.IDLE: // 空闲
			self._case()
		}
	}
}

func (self *DBTRobotV2) _case() {
	switch cfg.GetCaseName() {
	case "talk":
		self.Talk("aaaaa")
	case "battle":
		self.searchBoss()
	case "mailsignal":
		self.mailSignal()
	case "equip":
		self.equip()
	case "upSkill":
		self.UpSkill()
	default:
		logger.Info("Start None........")
	}
}

func (self *DBTRobotV2) ReLogin() {
	self.loginOnce.Do(func() {
		self.AddWg()
		self.Stop()
		logger.Debug("ReLogin 下线帐号[%s]", self.Account())
		self.CliData.DelayLogin = true
		self.CliData.Trigger(utils2.RELOGIN, self.CliData)
	})
}

func (self *DBTRobotV2) cfgZoneId() int {
	return cfg.Cfg.GetInt("Login.ZoneId")
}

func (self *DBTRobotV2) cfgSendDelay() int64 {
	return cfg.Cfg.GetInt64("mail.SendDelay")
}

func (self *DBTRobotV2) cfgDelDelay() int64 {
	return cfg.Cfg.GetInt64("mail.DelDelay")
}
