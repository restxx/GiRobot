package report

import (
	"encoding/json"
	"fmt"
	logger "github.com/restxx/GiRobot/Logger"
	"net/http"
	"strings"
	"time"
)

var ClientReqStats *RequestStats

// 负责将收集到的压测数据每秒发送到chartServer

func sendData(url, strJson string) {

	defer func() {
		if err := recover(); err != nil {
			logger.Error("url=[%s] sendData Error %v", url, strJson)
			return
		}
	}()

	resp, err := http.Post(url, "application/json", strings.NewReader(strJson))
	defer resp.Body.Close()
	if err != nil {
		logger.ErrorV(err)
	}
}

func AddLogData(name string, sTime int64, status uint8) {
	data := LogDataPool.Get().(*LogData)
	data.Name = name
	data.Tm = sTime
	data.Status = status
	ClientReqStats.LogDataChan <- data
}

func ClientInit(servAddr, testID string) {

	fmt.Println(servAddr, testID)

	ClientReqStats = NewRequestStats(testID, CLIENT_MODE)

	go func() {
		tk := time.NewTicker(time.Second)
		defer tk.Stop()

		for {
			select {
			case logData := <-ClientReqStats.LogDataChan:
				if logData.Status == 1 {
					ClientReqStats.LogRequest(logData.Name, logData.Tm)
				} else {
					ClientReqStats.LogError(logData.Name)
				}
				LogDataPool.Put(logData)
			case <-tk.C:
				t := ClientReqStats.SerializePreSec()
				bstr, err := json.Marshal(t)
				if err == nil {
					fmt.Println("每秒:", string(bstr))
					go sendData(servAddr, string(bstr))
				}

			}
		}
	}()
}
