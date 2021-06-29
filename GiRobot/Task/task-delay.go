package Task

func NewSampleDelayTask(millisecond int64, f func(interface{})) *SampleDelayTask {
	t := &SampleDelayTask{}
	t.BaseTask = NewBaseTask(TaskTypeSampleDelay)
	t.f = f
	t.finish = false
	t.DelayTimestamp(millisecond)
	return t
}

type SampleDelayTask struct {
	*BaseTask
	f      func(interface{})
	finish bool
}

func (t *SampleDelayTask) Continue(host interface{}) {
	if !t.CheckTimestamp() {
		return
	}
	t.f(host)
	t.finish = true
}

func (t *SampleDelayTask) IsFinish(host interface{}) bool {
	return t.finish
}
