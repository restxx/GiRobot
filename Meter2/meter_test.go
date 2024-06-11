package Meter2

import (
	logger "github.com/restxx/GiRobot/Logger"
	"testing"
	"time"
)

func TestNewMtManager(t *testing.T) {
	logger.Init(logger.BuildLogger("../../tmp", "aaa", "aaa", 1, "debug"))
	defer logger.Flush()

	mt := NewMtManager("fzdl", "00001", 4)

	mt.CreateNode("test1")

	mt.CreateNode("test1_1")
	time.Sleep(1 * time.Second)

	mt.CreateNode("test1_1_1")
	time.Sleep(1 * time.Second)
	mt.CloseNode("test1_1_1", 1)

	mt.CloseNode("test1_1", 1)

	time.Sleep(5 * time.Second)
	mt.CloseNode("test1", 1)

}
