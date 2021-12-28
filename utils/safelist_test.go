package utils

import (
	"fmt"
	"testing"
	"time"
)

type ST struct {
}

func TestNewSafeList(t *testing.T) {

	List := NewSafeList()

	go func() {
		tk := time.NewTicker(time.Second)
		defer tk.Stop()
		for {
			select {
			case <-tk.C:
				List.Put(&ST{})
			}
		}
	}()

	go func() {
		for {
			select {
			case <-List.Signal():
				fmt.Println(List.Pop())
				fmt.Println(List.Pop()) // <nil> no node
			}
		}
	}()

	time.Sleep(time.Second * 20)
}
