package Cfg

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestInitConfig(t *testing.T) {

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 10; i++ {
		fmt.Println(rnd.Int63()%20 + 1)
	}

	InitConfig(nil)

	fmt.Println(Cfg.Get("robot.ServerAddr"))
}
