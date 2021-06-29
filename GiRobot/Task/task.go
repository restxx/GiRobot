package Task

import (
	"GiantQA/GiRobot/utils"
	"sync/atomic"
	"time"
)

type TaskType int

const (
	TaskTypeSampleDelay = TaskType(1)
)

type Task interface {
	IsFinish(host interface{}) bool
	Continue(host interface{})

	AddRunStep()
	GetType() TaskType
	WhenCancel()
	WhenFinish()
}

type TaskMgr struct {
	tasks []Task
}

func (mgr *TaskMgr) TimeAction(host interface{}) {
	var delList []Task
	for _, v := range mgr.tasks {
		v.Continue(host)
		v.AddRunStep()
		if v.IsFinish(host) {
			v.WhenFinish()
			delList = append(delList, v)
			//util.DelSliceIndex(&mgr.tasks, k)
		}
	}

	if len(delList) > 0 {
		utils.DelSlice(&mgr.tasks, func(d interface{}) bool {
			for _, v := range delList {
				if v == d {
					return true
				}
			}
			return false
		})
	}
}

func (mgr *TaskMgr) AddTask(t Task, distinct bool) {
	if distinct {
		mgr.ClearByType(t.GetType())
	}
	mgr.tasks = append(mgr.tasks, t)
}

func (mgr *TaskMgr) ClearByType(t TaskType) {
	for k, v := range mgr.tasks {
		if v.GetType() == t {
			v.WhenCancel()
			utils.DelSliceIndex(&mgr.tasks, k)
		}
	}
}

func (mgr *TaskMgr) GetByType(t TaskType) Task {
	for _, v := range mgr.tasks {
		if v.GetType() == t {
			return v
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
var _BaseTaskUniqueID = uint64(1000)

func GenUniqueId() uint64 {
	return atomic.AddUint64(&_BaseTaskUniqueID, 1)
}

func NewBaseTask(tt TaskType) *BaseTask {
	t := &BaseTask{
		UniqueID:   GenUniqueId(),
		Type:       tt,
		step:       0,
		CancelFunc: nil,
		FinishFunc: nil,
		_timestamp: 0,
	}
	return t
}

type BaseTask struct {
	UniqueID   uint64
	Type       TaskType
	step       int
	CancelFunc func()
	FinishFunc func()

	_timestamp int64
}

func (dt *BaseTask) AddRunStep() {
	dt.step++
}
func (dt *BaseTask) GetType() TaskType {
	return dt.Type
}
func (dt *BaseTask) WhenCancel() {
	if dt.CancelFunc != nil {
		dt.CancelFunc()
	}
}
func (dt *BaseTask) WhenFinish() {
	if dt.FinishFunc != nil {
		dt.FinishFunc()
	}
}
func (dt *BaseTask) CheckTimestamp() bool {
	return dt._timestamp <= NowTimestampMillisecond()
}
func (dt *BaseTask) DelayTimestamp(millisecond int64) {
	dt._timestamp = NowTimestampMillisecond() + millisecond
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
var DelaySeconds = int64(0)

func NowTimestamp() int64 {
	return time.Now().Unix() + DelaySeconds
}

func NowTimestampMillisecond() int64 {
	return time.Now().UnixNano()/int64(time.Millisecond) + DelaySeconds*1000
}
