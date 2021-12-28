package Meter

import (
	logger "github.com/restxx/GiRobot/Logger"
	"github.com/restxx/GiRobot/report"
	"math/rand"
	"testing"
	"time"
)

//logger.ClientInit(logger.BuildLogger("./tmp", "", "", 0, cfg.GetLogLevel()))
//defer logger.Flush()

func TestMeter(t *testing.T) {
	logger.Init(logger.BuildLogger("../../tmp", "aaa", "aaa", 1, "debug"))
	defer logger.Flush()

	mt := NewMtManager("fzdl", "00001", time.Second*4)
	mt.CreateNode("login")
	time.Sleep(time.Second)
	mt.CloseNode("login", 1)

	mt.CreateNode("login")
	time.Sleep(time.Second)
	mt.CloseNode("login", 1)

	time.Sleep(time.Second * 5)

}

func TestMeter2(t *testing.T) {
	logger.Init(logger.BuildLogger("../../tmp", "aaa", "aaa", 1, "debug"))
	defer logger.Flush()

	mt := NewMtManager("fzdl", "00001", time.Second*4)
	report.ClientInit("http://127.0.0.1:8002/logData", "5d373cd7ca059a35432476b960558580")

	flag := make(chan bool, 1)
	go func() {
		tk := time.NewTicker(time.Millisecond * 100)
		defer tk.Stop()

		rand.Seed(time.Now().UnixMilli())
		for {
			select {
			case <-tk.C:
				mt.CreateNode("login")
				time.Sleep(time.Millisecond * time.Duration(rand.Int63n(98)+1))
				mt.CloseNode("login", 1)
			}

		}

	}()

	<-flag
}
