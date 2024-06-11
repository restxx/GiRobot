package Robot

import (
	"context"
	"fmt"
	logger "github.com/restxx/GiRobot/Logger"
	meter "github.com/restxx/GiRobot/Meter2"
	"math/rand"
	"testing"
	"time"
)

type TSRobot struct {
	Robot
}

type Message struct {
	Id int
}

type Message2 struct{}

func TestRobot_Wait(t *testing.T) {

	logger.Init(logger.BuildLogger("../../tmp", "aaa", "aaa", 1, "debug"))
	defer logger.Flush()

	rob := TSRobot{}
	rob.InCome = make(chan interface{}, 100)
	rob.Ctx, rob.CancelFunc = context.WithCancel(context.Background())
	rob.ProcessMsg = func(Msg interface{}) {}

	rob.MtMgr = meter.NewMtManager("fzdl", "00001", 4)

	rob.MtMgr.CreateNode("test1")

	go func() {
		for {
			time.Sleep(time.Second)
			id := rand.Intn(10)
			m := &Message{Id: id}
			rob.InCome <- m
		}
	}()

	rob.Wait(func(msg *Message) bool {
		fmt.Println(msg.Id)
		if msg.Id == 9 {
			rob.MtMgr.CloseNode("test1", 1)
			fmt.Println("-------find 9-------")
			return true
		}
		return false
	})

	fmt.Println("rob.WaitTime(1000 * 2)")
	rob.WaitTime(1000 * 2)

	m2 := &Message2{}
	rob.InCome <- m2
	rob.Wait(func(msg *Message2) bool {
		return true
	})
}
