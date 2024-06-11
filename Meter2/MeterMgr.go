package Meter2

import (
	logger "github.com/restxx/GiRobot/Logger"
	"github.com/restxx/GiRobot/report"
	"sync"
	"time"
)

//-------------Meter--------------------------------------------------------------------

type Meter struct {
	mt      *MtManager
	name    string
	StartTm time.Time
	status  uint8
	T       *time.Timer
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

//-----------MtManager-----------------------------------------------------------------

type MtManager struct {
	project string
	account string
	timeout time.Duration
	hMap    sync.Map //方便跟据name查找
	//最多只支持三层meter嵌套
	chLayer1 chan *Meter
	chLayer2 chan *Meter
	chLayer3 chan *Meter

	tk1 *time.Timer
	tk2 *time.Timer
	tk3 *time.Timer
}

// NewMtManager 每个客户一个Manager
func NewMtManager(project string, account string, timeout time.Duration) *MtManager {
	a := &MtManager{
		project:  project,
		account:  account,
		timeout:  timeout,
		chLayer1: make(chan *Meter, 1),
		chLayer2: make(chan *Meter, 1),
		chLayer3: make(chan *Meter, 1),

		tk1: time.NewTimer(0),
		tk2: time.NewTimer(0),
		tk3: time.NewTimer(0),
	}
	//必须初始化,但又不能泄露
	a.tk1.Stop()
	a.tk2.Stop()
	a.tk3.Stop()
	return a
}

func (mt *MtManager) newMeter(trans string) *Meter {
	return &Meter{mt: mt, name: trans, StartTm: time.Now(), T: time.NewTimer(mt.timeout * time.Second)}
}

func (m *MtManager) CreateNode(caseName string) {
	//判断chLayer1  2  3 哪个能接收
	if len(m.chLayer1) == 0 {
		t := m.newMeter(caseName)
		m.chLayer1 <- t
		m.tk1 = t.T
		m.hMap.Store(caseName, m.chLayer1)
	} else if len(m.chLayer2) == 0 {
		t := m.newMeter(caseName)
		m.chLayer2 <- t
		m.tk2 = t.T
		m.hMap.Store(caseName, m.chLayer2)
	} else if len(m.chLayer3) == 0 {
		t := m.newMeter(caseName)
		m.chLayer3 <- t
		m.tk3 = t.T
		m.hMap.Store(caseName, m.chLayer3)
	}
}

func (m *MtManager) CloseNode(caseName string, status uint8) {
	if v, ok := m.hMap.Load(caseName); ok {
		switch vv := v.(type) {
		case chan *Meter:
			meter := <-vv
			meter.status = status
			meter.T.Stop()
			meter.tps(meter.mt.project, meter.mt.account, meter.name, meter.status, time.Since(meter.StartTm).Milliseconds())
			m.hMap.Delete(caseName)
		}
	}

}

func (m *MtManager) GetTimer(idx int) (*time.Timer, interface{}) {
	switch idx {

	case 1:
		return m.tk1, m.chLayer1
	case 2:
		return m.tk2, m.chLayer2
	case 3:
		return m.tk3, m.chLayer3
	}
	return nil, nil
}

func (m *MtManager) StopNode(chLayer interface{}) {
	v := chLayer.(chan *Meter)
	if len(v) > 0 {
		meter := <-v
		meter.status = 3
		meter.T.Stop()
		meter.tps(meter.mt.project, meter.mt.account, meter.name, meter.status, time.Since(meter.StartTm).Milliseconds())
		m.hMap.Delete(meter.name)
	}
}
