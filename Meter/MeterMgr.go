package Meter

import (
	"context"
	logger "github.com/restxx/GiRobot/Logger"
	"sync"
	"time"
)

// 单个机器人timeout时间内同名事务最大数
const MAXSIZE = 200

type MeteTrans struct {
	mt      *MtManager
	name    string
	StartTm time.Time
	T       *time.Timer
	status  uint8

	Ctx    context.Context
	Cancel context.CancelFunc
}

func (m *MeteTrans) TPS(project string, account string, transName string, status uint8, spendtime float64) {
	now := time.Now().Format("2006-01-02 15:04:05.000")
	// str := fmt.Sprintf(`{"projectId": %s, "logTime": "%s", "account": "%s","testCase": "%s", "testResult": %d, "responseTime": %d}`,
	logger.Warn(`{"ID":"%s","LT":"%s","AC":"%s","TC":"%s","RT":%d,"TM":%d}`,
		project, now, account, transName, status, uint32(spendtime))
}

func NewMeteTrans(mt *MtManager, trans string) *MeteTrans {
	m := &MeteTrans{mt: mt, name: trans, StartTm: time.Now(), T: time.NewTimer(mt.timeout * time.Second)}
	m.Ctx, m.Cancel = context.WithCancel(context.Background())

	go func(meter *MeteTrans) {
		select {
		case <-meter.T.C: // 超时退出
			// 超时写logTrans
			meter.mt.CloseNode(trans, 3)
		case <-meter.Ctx.Done(): // 主动关闭
			// 写TransLog millisecond
			milli := float64((time.Now().Sub(m.StartTm)) / 10e5)
			meter.TPS(meter.mt.project, meter.mt.account, meter.name, meter.status, milli)
		}
		// 关闭计时
		meter.T.Stop()
	}(m)
	return m
}

type MtManager struct {
	project string
	account string
	// uid     uint64
	timeout time.Duration
	hMap    sync.Map
	lock    *sync.Mutex
}

// 每个客户一个Manager
var _lock sync.Mutex

func NewMtManager(project string, account string, timeout time.Duration) *MtManager {
	_lock.Lock()
	defer _lock.Unlock()
	return &MtManager{project: project, account: account, timeout: timeout, lock: &sync.Mutex{}}
}

func (m *MtManager) CreateNode(trans string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	var ch chan *MeteTrans
	iv, ok := m.hMap.Load(trans)

	if !ok { // 当前chan不存在
		ch = make(chan *MeteTrans, MAXSIZE)
		m.hMap.Store(trans, ch)
	} else {
		ch = iv.(chan *MeteTrans)
	}
	ch <- NewMeteTrans(m, trans)
}

func (m *MtManager) CloseNode(trans string, status uint8) {
	m.lock.Lock()
	defer m.lock.Unlock()

	iv, ok := m.hMap.Load(trans)
	if !ok {
		return
	}
	ch, ok := iv.(chan *MeteTrans)
	if !ok {
		return
	}
	select {
	case mete := <-ch:
		mete.status = status
		if status != 3 { // 如果是正常结束
			mete.Cancel()
			return
		} else { // 超时退出
			milli := float64((time.Now().Sub(mete.StartTm)) / 10e5)
			mete.TPS(mete.mt.project, mete.mt.account, mete.name, mete.status, milli)
		}
	default:
		return
	}
}
