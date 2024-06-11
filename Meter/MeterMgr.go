package Meter

import (
	"context"
	logger "github.com/restxx/GiRobot/Logger"
	"github.com/restxx/GiRobot/report"
	"sync"
	"time"
)

const MAXSIZE = 200 //单个机器人timeout时间内同名事务最大数

type Meter struct {
	mt      *MtManager
	name    string
	StartTm time.Time
	T       *time.Timer
	status  uint8
	Ctx     context.Context
	Cancel  context.CancelFunc
}

func (m *Meter) tps(project string, account string, transName string, status uint8, spendtime int64) {
	now := time.Now()
	// str := fmt.Sprintf(`{"projectId": %s, "logTime": "%s", "account": "%s","testCase": "%s", "testResult": %d, "responseTime": %d}`,
	logger.Warn(`{"ID":"%s","LT":"%s","AC":"%s","TC":"%s","RT":%d,"TM":%d, "TS":"%d"}`,
		project, now.Format("2006-01-02 15:04:05.000"), account, transName, status, spendtime, int32(now.Unix()))

	if report.ClientReqStats != nil {
		report.AddLogData(transName, spendtime, status)
	}
}

func newMeter(mt *MtManager, trans string) *Meter {
	m := &Meter{mt: mt, name: trans, StartTm: time.Now(), T: time.NewTimer(mt.timeout * time.Second)}
	m.Ctx, m.Cancel = context.WithCancel(context.Background())

	go func(meter *Meter) {
		defer meter.T.Stop()

		select {
		case <-meter.T.C: // 超时退出
			// 超时写logTrans
			meter.mt.CloseNode(trans, 3)
		case <-meter.Ctx.Done(): // 主动关闭
			Milliseconds := time.Since(meter.StartTm).Milliseconds()
			meter.tps(meter.mt.project, meter.mt.account, meter.name, meter.status, Milliseconds)
		}
	}(m)
	return m
}

//-----------------------------------------------------------------------------------------

type MtManager struct {
	project string
	account string
	timeout time.Duration
	hMap    sync.Map
	lock    *sync.Mutex
	t       *time.Timer //并没使用
}

// NewMtManager 每个客户一个Manager
func NewMtManager(project string, account string, timeout time.Duration) *MtManager {
	a := &MtManager{
		project: project,
		account: account,
		timeout: timeout,
		lock:    &sync.Mutex{},
		t:       time.NewTimer(0), //并没使用
	}
	a.t.Stop()
	return a
}

func (m *MtManager) CreateNode(caseName string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	ch := make(chan *Meter, MAXSIZE)
	iv, ok := m.hMap.LoadOrStore(caseName, ch)
	if ok {
		close(ch) //已存在对应的ch
	}
	iv.(chan *Meter) <- newMeter(m, caseName)
}

func (m *MtManager) CloseNode(caseName string, status uint8) {
	m.lock.Lock()
	defer m.lock.Unlock()

	iv, ok := m.hMap.Load(caseName)
	if !ok {
		return
	}
	ch, ok := iv.(chan *Meter)
	if !ok {
		return
	}
	select {
	case meter := <-ch:
		meter.status = status
		if status != 3 { // 如果是正常结束
			meter.Cancel()
			return
		} else { // 超时退出
			Milliseconds := time.Since(meter.StartTm).Milliseconds()
			meter.tps(meter.mt.project, meter.mt.account, meter.name, meter.status, Milliseconds)
		}
	default:
		return
	}
}

// GetTimer  -----只为满足接口声明---------------
func (m *MtManager) GetTimer(int) (*time.Timer, interface{}) {
	return m.t, nil
}

func (m *MtManager) StopNode(interface{}) {}
