package utils

import (
	"sync"
)

var (
	pools  [18]sync.Pool
	pool64 *sync.Pool
)

func init() {
	var i uint
	for i = 6; i < 24; i++ {
		n := 1 << i
		pools[bytePoolNum(n)].New = func() interface{} {
			return make([]byte, 0, n)
		}
	}
	pool64 = &pools[0]
}

func Put(b []byte) {
	if b == nil {
		return
	}

	c := cap(b)
	if c < 64 {
		return
	}

	pn := bytePoolNum((c + 2) >> 1)
	if pn != -1 {
		pools[pn].Put(b[0:0])
	}
}

func Get(n int) []byte {
	if n <= 64 {
		return pool64.Get().([]byte)[0:n]
	}

	pn := bytePoolNum(n)
	if pn != -1 {
		return pools[pn].Get().([]byte)[0:n]
	} else {
		return make([]byte, n)
	}
}

func bytePoolNum(i int) int {
	if i <= 64 {
		return 0
	} else if i <= 128 {
		return 1
	} else if i <= 256 {
		return 2
	} else if i <= 512 {
		return 3
	} else if i <= 1024 {
		return 4
	} else if i <= 2048 {
		return 5
	} else if i <= 4096 {
		return 6
	} else if i <= 8192 {
		return 7
	} else if i <= 16384 {
		return 8
	} else if i <= 32768 {
		return 9
	} else if i <= 65536 {
		return 10
	} else if i <= 131072 {
		return 11
	} else if i <= 262144 {
		return 12
	} else if i <= 524288 {
		return 13
	} else if i <= 524288*2 {
		return 14
	} else if i <= 524288*4 {
		return 15
	} else if i <= 524288*8 {
		return 16
	} else if i <= 524288*16 {
		return 17
	} else {
		return -1
	}
}
