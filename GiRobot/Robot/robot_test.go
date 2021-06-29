package Robot

import (
	"context"
	"testing"
)

type TSRobot struct {
	Robot
}

type Message struct{}

type Message2 struct{}

func TestRobot_Wait(t *testing.T) {

	rob := TSRobot{}
	rob.InCome = make(chan interface{}, 1)
	rob.Ctx, rob.CancelFunc = context.WithCancel(context.Background())
	rob.ProcessMsg = func(Msg interface{}) {}

	m := &Message{}
	rob.InCome <- m

	rob.Wait(func(msg *Message) bool {
		return true
	})

	rob.WaitTime(1000 * 2)

	m2 := &Message2{}
	rob.InCome <- m2
	rob.Wait(func(msg *Message2) bool {
		return true
	})

}
