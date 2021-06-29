package log

type logger interface {
	DEBUG(format string, a ...interface{})
	DEBUGV(a ...interface{})

	INFO(format string, a ...interface{})
	INFOV(a ...interface{})

	TRACE(format string, a ...interface{})
	TRACEV(a ...interface{})

	ERROR(format string, a ...interface{})
	ERRORV(a ...interface{})

	CRT(format string, a ...interface{})
	CRTV(a ...interface{})

	WARN(format string, a ...interface{})
	WARNV(a ...interface{})

	Flush()
}

var _ins logger

func Init(l logger) {
	_ins = l
}

func Debug(format string, a ...interface{}) {
	_ins.DEBUG(format, a...)
}

func Info(format string, a ...interface{}) {
	_ins.INFO(format, a...)
}

func Trace(format string, a ...interface{}) {
	_ins.TRACE(format, a...)
}

func Error(format string, a ...interface{}) {
	_ins.ERROR(format, a...)
}

func Warn(format string, a ...interface{}) {
	_ins.WARN(format, a...)
}

func WarnV(a ...interface{}) {
	_ins.WARNV(a...)
}

func Crt(format string, a ...interface{}) {
	_ins.CRT(format, a...)
}

func DebugV(a ...interface{}) {
	_ins.DEBUGV(a...)
}

func InfoV(a ...interface{}) {
	_ins.INFOV(a...)
}

func TraceV(a ...interface{}) {
	_ins.TRACEV(a...)
}

func ErrorV(a ...interface{}) {
	_ins.ERRORV(a...)
}

func CrtV(a ...interface{}) {
	_ins.CRTV(a...)
}

func Flush() {
	_ins.Flush()
}
