package TCPNet

import (
	logger "github.com/restxx/GiRobot/Logger"
	"time"
)

// 发送心跳
func defaultPing(conn *TcpConn) {
	conn.SendPak([]byte("ping"))
}

// 默认处理函数
func defaultMainProc() {
	for {
		time.Sleep(time.Second)
	}
}

// 连接服务器成功后会被调用
func defaultOnConnected(conn *TcpConn) {
	logger.Trace("Default Connected:%v", conn.SessionId)
}

// 清理用户数据资源
func defaultOnClose(conn *TcpConn) {
	if conn.TConn == nil {
		return
	}
	logger.Info("Default Closed:%v From:%v", conn.SessionId, conn.TConn.RemoteAddr)
}
