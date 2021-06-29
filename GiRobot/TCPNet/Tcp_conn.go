package TCPNet

import (
	cfg "GiantQA/GiRobot/Cfg"
	logger "GiantQA/GiRobot/Logger"
	"GiantQA/GiRobot/Meter"
	"GiantQA/GiRobot/stream"
	"GiantQA/GiRobot/utils"
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ReadWaitPing = 60 * time.Second
	WriteTicker  = 60 * time.Second
	queueCount   = 2048
)

type TcpConn struct {
	SessionId string

	TConn    net.Conn
	ownerNet *TCPNetwork

	send    chan *utils.MsgBlock // 将要发送的数据
	process chan *utils.MsgBlock // 收到的数据

	IsClosed int32

	Ctx    context.Context
	Cancel context.CancelFunc

	Once *sync.Once

	Meter    *Meter.MtManager
	UserData *utils.ClientData
	// B3TreeMgr *B3Tree.BTreeManager
}

func EmptyConn() *TcpConn {
	newConn := &TcpConn{
		send:     make(chan *utils.MsgBlock, queueCount),
		process:  make(chan *utils.MsgBlock, queueCount),
		IsClosed: 0,
	}
	newConn.Once = new(sync.Once)
	newConn.SessionId = utils.CreateUUID(2)
	newConn.Ctx, newConn.Cancel = context.WithCancel(context.Background())

	return newConn
}

func NewDial(tn *TCPNetwork, addr string) (conn *TcpConn, err error) {
	dconn, derr := net.DialTimeout("tcp", addr, time.Second*30)
	if derr != nil {
		logger.ErrorV(derr)
		err = derr
		return
	}
	conn = EmptyConn()
	conn.TConn = dconn
	conn.ownerNet = tn
	return
}

// 收数据
func (c *TcpConn) recvPump() {
	defer func() {
		p := recover()
		if p != nil {
			logger.ErrorV(p)
			c.Close()
		}
		c.Cancel()

	}()

	buf := make([]byte, 65535)
	var bStream stream.IBuffIO = &stream.Buffer{}

	for {
		_ = c.TConn.SetReadDeadline(time.Now().Add(ReadWaitPing))
		rn, err := c.TConn.Read(buf)
		if err != nil {
			logger.Error("Session %v Recv Error:%v", c.SessionId, err)
			return
		}
		if rn == 0 {
			logger.Error("Session %v Recv Len:%v", c.SessionId, rn)
			return
		}
		if atomic.LoadInt32(&c.IsClosed) == 1 {
			return
		}

		n, err := bStream.Write(buf[:rn])
		if rn != n || err != nil {
			logger.Error("Session %v bStream write err %v", c.SessionId, buf[:rn])
			return
		}
		// 解包把bufI0对像传进去
		msgBlocks, err := c.ownerNet.Unpacking(bStream, c.UserData.EncryptKey)
		if err == nil {
			for _, p := range msgBlocks {
				c.process <- p
			}
			if bStream.Empty() {
				bStream.Reset()
			}
		}
	}
}

func (c *TcpConn) mainTicker() {

	defer func() {
		if p := recover(); p != nil {
			logger.ErrorV(p)
		}
		c.Close()
		logger.InfoV("mainTicker defer done!!!")
	}()

	// 这两个goroutine不操作游戏数据
	go c.recvPump()  // 不操作游戏逻辑只收包转成结构 	c.process<-MsgBlock里
	go c.writePump() // 不操作游戏逻辑只发送缓冲区数据	<-c.send

	tickerTime := cfg.Cfg.GetDuration("robot.TickerTime")
	ticker := time.NewTicker(tickerTime * time.Millisecond)
	for {
		select {
		case <-c.Ctx.Done():
			logger.Info("ManiTicker %v session write done...", c.SessionId)
			ticker.Stop()
			return
		case mb := <-c.process:
			// 处理收到的MsgBlock
			// fmt.Println(mb.MID, mb.SID, mb.Data)
			c.ownerNet.OnHandler(c, mb)
		case <-ticker.C:
			// 执行主动逻辑 真实环境必须为IN_SCENE
			if c.ownerNet.MainTicker != nil {
				c.ownerNet.MainTicker(c)
			}
		}
	}
}

// 将msgBlock从chan队列里取出来
func (c *TcpConn) writePump() {

	defer func() {
		if p := recover(); p != nil {
			logger.ErrorV(p)
		}
		c.Close()
		logger.Info("write Pump defer done!!!")
	}()

	for {
		select {
		case <-c.Ctx.Done():
			logger.Info("%v session write done...", c.SessionId)
			return
		case mb, ok := <-c.send:
			if (!ok) || (atomic.LoadInt32(&c.IsClosed) == 1) {
				return
			}
			// 打包加密
			bData, err := c.ownerNet.Package(mb, c.UserData.EncryptKey)
			if err != nil {
				return
			}
			_ = c.TConn.SetWriteDeadline(time.Now().Add(WriteTicker))
			if _, err := c.TConn.Write(bData); err != nil {
				logger.Error("write Pump error:%v !!!", err)
				return
			}
		} // select end
	} // for end
}

// 放进来的数据是go struct 和包函id *msgBlock
func (c *TcpConn) SendPak(val interface{}) int {

	if atomic.LoadInt32(&c.IsClosed) == 1 {
		return -1
	}

	if mb, ok := val.(*utils.MsgBlock); ok {
		return c.sendData(mb)
	}
	return 0
}

// 只是把数据放入待发送队列
func (c *TcpConn) sendData(block *utils.MsgBlock) int {

	select {
	case c.send <- block:
		return 1
	case <-time.After(3 * time.Second):
		break
	default:
		break
	}
	return -1
}

func (c *TcpConn) Close() {
	c.Once.Do(func() {
		if c.TConn == nil {
			return
		}
		if atomic.LoadInt32(&c.IsClosed) == 0 {
			atomic.StoreInt32(&c.IsClosed, 1)
			// 清理用户数据
			c.ownerNet.OnClose(c)
			_ = c.TConn.Close() // 关闭连接
		}
	})
}
