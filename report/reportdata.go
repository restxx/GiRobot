package report

import "sync"

// 压测端发来的每秒请求数据 序列化及反序列化

type ClientReport struct {
	TestID string  `json:"TestID"`
	Data   []*Data `json:"Data"`
}

type Data struct {
	Case     string    `json:"Case"`
	Result   []*Result `json:"Result"`
	ErrCount int64     `json:"ErrCount"`
}

type Result struct {
	Tm    int64 `json:"Tm"`
	Count int   `json:"Count"`
}

//用于序列化 requestStats  用于30秒或全部周期内的数据

type RequestEntry struct {
	TestID string        `json:"test_id"`
	Data   []interface{} `json:"data"`
}

type LogData struct {
	Name   string `json:"name"`
	Tm     int64  `json:"Tm"`
	Status uint8  `json:"status"`
}

var LogDataPool *sync.Pool

func init() {
	LogDataPool = &sync.Pool{
		New: func() interface{} {
			return &LogData{}
		},
	}
}
