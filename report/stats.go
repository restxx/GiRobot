package report

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogo/protobuf/sortkeys"
	logger "github.com/restxx/GiRobot/Logger"
	"math"
	"time"
)

const (
	SERVER_MODE = 1
	CLIENT_MODE = 2
)

const (
	MAXSIZE = 3000
)

type RequestStats struct {
	TestID  string
	entries map[string]*statsEntry
	totals  map[string]*statsEntry
	mode    uint8

	ClientReportChan chan *ClientReport
	LogDataChan      chan *LogData

	Ctx        context.Context
	CancelFunc context.CancelFunc
}

func NewRequestStats(testID string, mode uint8) (stats *RequestStats) {
	entries := make(map[string]*statsEntry)
	var totals map[string]*statsEntry
	if mode == SERVER_MODE {
		totals = make(map[string]*statsEntry)
	}

	stats = &RequestStats{
		TestID:  testID,
		entries: entries,
		totals:  totals,
		mode:    mode,
	}
	stats.ClientReportChan = make(chan *ClientReport, 200) // 接收每秒收集的数据
	stats.LogDataChan = make(chan *LogData, MAXSIZE)       // 实时接收 每秒统计一次

	stats.Ctx, stats.CancelFunc = context.WithCancel(context.Background())
	return stats
}

//服务端每30秒结算一次 平时处理客户端每秒发来的数据

func (r *RequestStats) ServRecvLoop() {
	tk := time.NewTicker(time.Second * 30)
	defer tk.Stop()

	for {
		select {
		case <-r.Ctx.Done():
			if r.mode == SERVER_MODE {
				t := r.CollectTotalsReportData()
				fmt.Println("all: ", t)
				sendData("http://211.159.200.183:6080/api/resall/", t)
			}
			return
		case msg := <-r.ClientReportChan:
			//fmt.Println(msg)
			for _, v := range msg.Data {
				//v.Case
				for _, result := range v.Result {
					//result.Tm
					r.logRequestNum(v.Case, result.Tm, int64(result.Count))
				}
				// 处理ErrCount
				r.logErrorNum(v.Case, v.ErrCount)
			}
		case <-tk.C:
			if r.mode == SERVER_MODE {
				//	每30秒结算一次
				t := r.CollectReportData()
				fmt.Println("30s:", t)
				sendData("http://211.159.200.183:6080/api/restmp/", t)
			}
		}
	}
}

func (s *RequestStats) LogRequest(name string, responseTime int64) {
	if s.mode == SERVER_MODE {
		s.getTotals(name).log(responseTime)
	}
	s.get(name).log(responseTime)
}

func (s *RequestStats) logRequestNum(name string, responseTime, num int64) {
	s.getTotals(name).logNum(responseTime, num)
	s.get(name).logNum(responseTime, num)
}

func (s *RequestStats) LogError(name string) {
	if s.mode == SERVER_MODE {
		s.getTotals(name).logError()
	}
	s.get(name).logError()
}

func (s *RequestStats) logErrorNum(name string, num int64) {
	s.getTotals(name).logErrorNum(num)
	s.get(name).logErrorNum(num)
}

func (s *RequestStats) get(name string) (entry *statsEntry) {
	entry, ok := s.entries[name]
	if !ok {
		newEntry := &statsEntry{
			Name:          name,
			ResponseTimes: make(map[int64]int64),
		}
		newEntry.reset()
		s.entries[name] = newEntry
		return newEntry
	}
	return entry
}

func (s *RequestStats) getTotals(name string) (entry *statsEntry) {
	entry, ok := s.totals[name]
	if !ok {
		newEntry := &statsEntry{
			Name:          name,
			ResponseTimes: make(map[int64]int64),
		}
		newEntry.reset()
		s.totals[name] = newEntry
		return newEntry
	}
	return entry
}

func (s *RequestStats) serializeStats() []byte {
	defer func() {
		for _, v := range s.entries {
			v.reset() // 清空数据结构
		}
		if err := recover(); err != nil {
			logger.Error("serializeStats Error [%v]", err)
		}
	}()

	entries := make([]interface{}, 0, len(s.entries))
	for _, v := range s.entries {
		if !(v.NumRequests == 0 && v.NumFailures == 0) {
			v.calcPercentiles()
			entries = append(entries, v)
		}
	}

	Entry := &RequestEntry{
		TestID: s.TestID,
		Data:   entries,
	}
	val, err := json.Marshal(Entry)
	if err != nil {
		logger.Error("serializeStats Error [%v]", err)
		panic(err)
	}
	return val
}

func (s *RequestStats) serializeTotalsStats() []byte {
	defer func() {
		for _, v := range s.totals {
			v.reset() // 清空数据结构
		}
		if err := recover(); err != nil {
			logger.Error("serializeTotalsStats Error [%v]", err)
		}
	}()

	entries := make([]interface{}, 0, len(s.totals))
	for _, v := range s.totals {
		if !(v.NumRequests == 0 && v.NumFailures == 0) {
			v.calcPercentiles()
			entries = append(entries, v)
		}
	}

	Entry := &RequestEntry{
		TestID: s.TestID,
		Data:   entries,
	}
	val, err := json.Marshal(Entry)
	if err != nil {
		logger.Error("serializeTotalsStats Error [%v]", err)
		panic(err)
	}
	return val
}

func (s *RequestStats) CollectReportData() string {
	return string(s.serializeStats())
}

func (s *RequestStats) CollectTotalsReportData() string {
	return string(s.serializeTotalsStats())
}

func (r *RequestStats) Stop() {
	r.CancelFunc()
}

func (r *RequestStats) SerializePreSec() *ClientReport {

	defer func() {
		for _, v := range r.entries {
			v.reset()
		}
	}()

	ret := &ClientReport{
		TestID: r.TestID,
		Data:   make([]*Data, 0, len(r.entries)),
	}
	for _, v := range r.entries {
		if !(v.NumRequests == 0 && v.NumFailures == 0) {
			ret.Data = append(ret.Data, v.CollectPreSec())
		}
	}
	return ret
}

//-------------------------------------------------------------------

func (s *statsEntry) CollectPreSec() *Data {
	data := &Data{
		Case: s.Name, ErrCount: s.NumFailures,
		Result: make([]*Result, 0, len(s.ResponseTimes)),
	}
	for k, v := range s.ResponseTimes {
		data.Result = append(data.Result, &Result{Tm: k, Count: int(v)})
	}
	return data
}

type statsEntry struct {
	// Name (trans) of this stats entry
	Name string `json:"name"`
	// The number of requests made
	NumRequests int64 `json:"num_requests"`
	// Number of failed request
	NumFailures int64 `json:"num_failures"`
	// Total sum of the response times
	TotalResponseTime int64 `json:"total_response_time"`
	// Minimum response time
	MinResponseTime int64 `json:"min_response_time"`
	// Maximum response time
	MaxResponseTime int64 `json:"max_response_time"`
	// A {response_time => count} dict that holds the response time distribution of all the requests
	// The keys (the response time in ms) are rounded to store 1, 2, ... 9, 10, 20. .. 90,
	// 100, 200 .. 900, 1000, 2000 ... 9000, in order to save memory.
	// This dict is used to calculate the median and percentile response times.
	ResponseTimes map[int64]int64 `json:"-"`
	// Time of the first request for this entry
	StartTime int64 `json:"start_time"`
	// Time of the last request for this entry
	LastRequestTimestamp int64 `json:"last_request_timestamp"`

	NumRequestPerSec float64 `json:"num_request_per_sec"`
	AvgResponseTime  int64   `json:"avg_response_time"`
	Cost50           int64   `json:"cost50"`
	Cost75           int64   `json:"cost75"`
	Cost90           int64   `json:"cost90"`
	Cost95           int64   `json:"cost95"`
	Cost99           int64   `json:"cost99"`
}

func (s *statsEntry) reset() {
	s.StartTime = time.Now().Unix()
	s.NumRequests = 0
	s.NumFailures = 0
	s.TotalResponseTime = 0
	s.ResponseTimes = make(map[int64]int64)
	s.MinResponseTime = 0
	s.MaxResponseTime = 0
	s.LastRequestTimestamp = time.Now().Unix()

	s.NumRequestPerSec = 0.0
	s.AvgResponseTime = 0
	s.Cost50 = 0
	s.Cost75 = 0
	s.Cost90 = 0
	s.Cost95 = 0
	s.Cost99 = 0
}

func (s *statsEntry) log(responseTime int64) {
	s.NumRequests++

	s.logTimeOfRequest()
	s.logResponseTime(responseTime)
}

func (s *statsEntry) logNum(responseTime, num int64) {
	s.NumRequests += num

	s.logTimeOfRequest()
	s.logResponseTimeNum(responseTime, num)
}

func (s *statsEntry) logTimeOfRequest() {
	key := time.Now().Unix()
	s.LastRequestTimestamp = key
}

func round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

func (s *statsEntry) logResponseTime(responseTime int64) {
	s.TotalResponseTime += responseTime

	if s.MinResponseTime == 0 {
		s.MinResponseTime = responseTime
	}

	if responseTime < s.MinResponseTime {
		s.MinResponseTime = responseTime
	}

	if responseTime > s.MaxResponseTime {
		s.MaxResponseTime = responseTime
	}

	var roundedResponseTime int64

	// to avoid to much data that has to be transferred to the master node when
	// running in distributed mode, we save the response time rounded in a dict
	// so that 147 becomes 150, 3432 becomes 3400 and 58760 becomes 59000
	// see also locust's stats.py
	if responseTime < 100 {
		roundedResponseTime = responseTime
	} else if responseTime < 1000 {
		roundedResponseTime = int64(round(float64(responseTime), .5, -1))
	} else if responseTime < 10000 {
		roundedResponseTime = int64(round(float64(responseTime), .5, -2))
	} else {
		roundedResponseTime = int64(round(float64(responseTime), .5, -3))
	}

	_, ok := s.ResponseTimes[roundedResponseTime]
	if !ok {
		s.ResponseTimes[roundedResponseTime] = 1
	} else {
		s.ResponseTimes[roundedResponseTime]++
	}
}

func (s *statsEntry) logResponseTimeNum(responseTime, num int64) {
	s.TotalResponseTime += responseTime * num

	if s.MinResponseTime == 0 {
		s.MinResponseTime = responseTime
	}

	if responseTime < s.MinResponseTime {
		s.MinResponseTime = responseTime
	}

	if responseTime > s.MaxResponseTime {
		s.MaxResponseTime = responseTime
	}

	var roundedResponseTime int64

	// to avoid to much data that has to be transferred to the master node when
	// running in distributed mode, we save the response time rounded in a dict
	// so that 147 becomes 150, 3432 becomes 3400 and 58760 becomes 59000
	// see also locust's stats.py
	if responseTime < 100 {
		roundedResponseTime = responseTime
	} else if responseTime < 1000 {
		roundedResponseTime = int64(round(float64(responseTime), .5, -1))
	} else if responseTime < 10000 {
		roundedResponseTime = int64(round(float64(responseTime), .5, -2))
	} else {
		roundedResponseTime = int64(round(float64(responseTime), .5, -3))
	}

	_, ok := s.ResponseTimes[roundedResponseTime]
	if !ok {
		s.ResponseTimes[roundedResponseTime] = num
	} else {
		s.ResponseTimes[roundedResponseTime] += num
	}
}

func (s *statsEntry) logError() {
	s.logTimeOfRequest() //这里有争议要不要更改最后请求时间字段
	s.NumFailures++
}

func (s *statsEntry) logErrorNum(num int64) {
	s.logTimeOfRequest() //这里有争议要不要更改最后请求时间字段
	s.NumFailures += num
}

func (s *statsEntry) calcPercentiles() {
	//计算50线 75线 。。。99线
	Int64s := make([]int64, 0, len(s.ResponseTimes))
	for respTime, _ := range s.ResponseTimes {
		Int64s = append(Int64s, respTime)
	}
	sortkeys.Int64s(Int64s)

	// {0.50, 0.75, 0.90, 0.95, 0.99}
	P50Idx := int64(float64(s.NumRequests) * 0.50)
	P75Idx := int64(float64(s.NumRequests) * 0.75)
	P90Idx := int64(float64(s.NumRequests) * 0.90)
	P95Idx := int64(float64(s.NumRequests) * 0.95)
	P99Idx := int64(float64(s.NumRequests) * 0.99)

	var count int64 = 0
	for _, respTime := range Int64s {
		count += s.ResponseTimes[respTime]
		if count >= P50Idx && s.Cost50 == 0 {
			s.Cost50 = respTime
		}
		if count >= P75Idx && s.Cost75 == 0 {
			s.Cost75 = respTime
		}
		if count >= P90Idx && s.Cost90 == 0 {
			s.Cost90 = respTime
		}
		if count >= P95Idx && s.Cost95 == 0 {
			s.Cost95 = respTime
		}
		if count >= P99Idx && s.Cost99 == 0 {
			s.Cost99 = respTime
		}
	}

	// 计算平均耗时
	if s.NumRequests == 0 { // 一次都没有成功
		s.AvgResponseTime = 0
	} else {
		s.AvgResponseTime = s.TotalResponseTime / s.NumRequests
	}
	// 计算tps
	var totalSec = int64(1)
	if s.StartTime != s.LastRequestTimestamp {
		totalSec = s.LastRequestTimestamp - s.StartTime
	}
	s.NumRequestPerSec = float64(s.NumRequests) / float64(totalSec)
	s.ResponseTimes = nil
}

func (s *statsEntry) serialize() map[string]interface{} {
	var result map[string]interface{}
	val, err := json.Marshal(s)

	fmt.Println(string(val)) // 改这里

	if err != nil {
		return nil
	}
	err = json.Unmarshal(val, &result)
	if err != nil {
		return nil
	}
	return result
}

func (s *statsEntry) getStrippedReport() map[string]interface{} {
	s.calcPercentiles()
	report := s.serialize()
	s.reset()
	return report
}
