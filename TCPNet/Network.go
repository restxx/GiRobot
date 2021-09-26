package TCPNet

import (
	"fmt"
	cfg "github.com/restxx/GiRobot/Cfg"
	"github.com/restxx/GiRobot/Meter"
	"github.com/restxx/GiRobot/stream"
	"github.com/restxx/GiRobot/utils"
	"log"
)

// 所有的callBack函数
type TCPNetwork struct {
	// 创建用户DATA
	// CreateUserData func() interface{}

	// 连接成功
	OnConnected func(conn *TcpConn)
	// 数据包进入
	OnHandler func(conn *TcpConn, msgBlock *utils.MsgBlock)
	// 连接关闭
	OnClose func(conn *TcpConn)

	MainTicker func(conn *TcpConn) // 主处理函数

	// 打包以及加密行为
	Package   func(msgBlocks *utils.MsgBlock, enKey []byte) (data []byte, err error)
	Unpacking func(pStream stream.IBuffIO, deKey []byte) (msgBlocks []*utils.MsgBlock, err error)

	// ping 心跳功能
	SendPing func(conn *TcpConn)
}

func NewEmptyTcp() *TCPNetwork {
	return &TCPNetwork{

		Package:   nil,
		Unpacking: nil,

		OnClose:     defaultOnClose,
		OnConnected: defaultOnConnected,

		OnHandler:  nil,
		MainTicker: nil,

		SendPing: defaultPing,
	}
}

func (tn *TCPNetwork) Dial(addr string, UD *utils.ClientData) (conn *TcpConn, err error) {

	log.Printf("Dial To :%v", addr)
	conn, err = NewDial(tn, addr)
	if err != nil {
		log.Printf("Dial Faild:%v", err)
		return
	}
	conn.UserData = UD
	conn.Meter = Meter.NewMtManager(cfg.GetProject(),
		fmt.Sprintf(cfg.GetNameFmt(), cfg.GetPrefix(), UD.ClientId+cfg.SkipNum()),
		(uint64)(UD.ClientId+cfg.SkipNum()), cfg.TimeOut())

	tn.OnConnected(conn)
	go conn.mainTicker()
	return
}
