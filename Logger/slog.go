package logger

import (
	"fmt"
	"github.com/cihub/seelog"
	"os"
	"time"
)

type slog struct {
	logger seelog.LoggerInterface
}

func (s *slog) DEBUG(format string, a ...interface{}) {
	s.logger.Debugf(format, a...)
}

func (s *slog) DEBUGV(a ...interface{}) {
	s.logger.Debug(a...)
}

func (s *slog) INFO(format string, a ...interface{}) {
	s.logger.Infof(format, a...)
}

func (s *slog) INFOV(a ...interface{}) {
	s.logger.Info(a...)
}

func (s *slog) TRACE(format string, a ...interface{}) {
	s.logger.Tracef(format, a...)
}

func (s *slog) TRACEV(a ...interface{}) {
	s.logger.Trace(a...)
}

func (s *slog) WARN(format string, a ...interface{}) {
	_ = s.logger.Warnf(format, a...)
}

func (s *slog) WARNV(a ...interface{}) {
	_ = s.logger.Warn(a...)
}

func (s *slog) ERROR(format string, a ...interface{}) {
	_ = s.logger.Errorf(format, a...)
}

func (s *slog) ERRORV(a ...interface{}) {
	_ = s.logger.Error(a...)
}

func (s *slog) CRT(format string, a ...interface{}) {
	_ = s.logger.Criticalf(format, a...)
}

func (s *slog) CRTV(a ...interface{}) {
	_ = s.logger.Critical(a...)
}

func (s *slog) Flush() {
	s.logger.Flush()
}

func tpslogName(CaseName, Prefix string, totalNum int) string {
	date := time.Now().Format("2006-01-02_15-04-05")
	return fmt.Sprintf("%s_%s_%s_%d", date, CaseName, Prefix, totalNum)
}

var ins *slog

func BuildLogger(logDir, CaseName, Prefix string, totalNum int, logLevel string) *slog {
	ins = &slog{}

	tpslog := tpslogName(CaseName, Prefix, totalNum)
	_ = os.Mkdir(logDir, 0777)
	realFile := logDir + "/normal.log"
	realErrorFile := logDir + "/error.log"
	realDebugFile := logDir + "/debug.log"
	realWarnTpsFile := logDir + "/" + tpslog + ".log"

	sConfig := fmt.Sprintf(`
	<seelog type="asynctimer" asyncinterval="1000"  minlevel="%v">
		<outputs formatid="main">  
			<filter levels="info">   
				<console />    
				<rollingfile type="size"  filename="%v" maxsize="20480000" maxrolls="100" />    
			</filter>
			<filter levels="critical,error">
				<console />   
				<rollingfile type="size"  filename="%v" maxsize="40960000" maxrolls="500" />   
			</filter>
			<filter levels="debug">
				<console />   
				<rollingfile type="size" filename="%v" maxsize="40960000" maxrolls="500" />   
			</filter>
			<filter levels="warn" formatid="tpslog">
				<console />   
				<rollingfile type="size" filename="%v" maxsize="40960000" maxrolls="500" />   
			</filter>
		</outputs>
		<formats>
			<format id="main" format="[%%Date(2006-01-02 15:04:05.000)] [%%File:%%Line] [%%LEVEL] %%Msg%%n"/>
			<format id="tpslog" format="%%Msg%%n"/>
		</formats>
	</seelog>
	`, logLevel, realFile, realErrorFile, realDebugFile, realWarnTpsFile)

	ins.logger, _ = seelog.LoggerFromConfigAsString(sConfig)

	_ = ins.logger.SetAdditionalStackDepth(2)
	_ = seelog.UseLogger(ins.logger)
	return ins
}
