package utils

import (
	cfg "GiantQA/GiRobot/Cfg"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	RELOGIN     string = "重登录"
	NEW_RELOGIN string = "用新帐号重登录"
)

const (
	START = iota
	LOGIN
	IDLE
	BUSYING
)

var nseeduuid uint64 = 0

// 设置字节序
var Endian binary.ByteOrder

// uuId生成器
func CreateUUID(seed int) string {
	atomic.AddUint64(&nseeduuid, 1)

	vmd5 := md5.New()
	vmd5.Write([]byte(strconv.FormatInt(int64(seed), 10)))
	vmd5.Write([]byte("_"))
	vmd5.Write([]byte(strconv.FormatUint(nseeduuid, 36)))
	vmd5.Write([]byte("_"))
	vmd5.Write([]byte(time.Now().Format("2019-05-09 16:47:00")))

	return strings.ToUpper(hex.EncodeToString(vmd5.Sum(nil)))
}

// 调试用函数
func GetFunName(Func interface{}) string {
	// 将函数包装为反射值对象
	funcValue := reflect.ValueOf(Func)
	return runtime.FuncForPC(funcValue.Pointer()).Name()
}

func Sleep(n int32) {
	t := rand.Int31() % n
	time.Sleep(time.Duration(t) * time.Millisecond)
}

// Tcp_conn 公共接口
type IConn interface {
	// SendPak(pak interface{}) int
}

func SetEndian() {
	Endian = cfg.GetEndian()
}

////////////////////////////////////////////////////////////////
