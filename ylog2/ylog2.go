package ylog2

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

const (
	TRACE = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)

const (
	LOG_TIME_FMT   = "2006-05-04 15:02:01"
	LOG_MSG_FORMAT = "[%s] [%s] [%s:%d] - %s"
)

func parseLogLevel(l int) string {
	switch l {
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return ""
	}
}

type LogWriter interface {
	enable(int) bool
	write(string)
	flush()
}

type ConsoleWriter struct {
	levelFilter func(int) bool
}

func NewConsoleWriter(fn func(int) bool) *ConsoleWriter {
	return &ConsoleWriter{
		levelFilter: fn,
	}
}

func (w *ConsoleWriter) enable(l int) bool {
	return w.levelFilter(l)
}

func (w *ConsoleWriter) write(str string) {
	os.Stdout.Write([]byte(str))
}

func (w *ConsoleWriter) flush() {
	os.Stdout.Sync()
}

type FileWriter struct {
}

type Logger struct {
	writers []LogWriter
}

func NewLogger(writers ...LogWriter) Logger {
	l := Logger{
		writers: make([]LogWriter, 0, len(writers)),
	}
	l.writers = append(l.writers, writers...)
	return l
}

func xmsg(str string, args ...any) string {
	if len(args) <= 0 {
		return str
	}
	return fmt.Sprintf(str, args...)
}

func callInfo() (string, int) {
	_, fn, ln, ok := runtime.Caller(3)
	if !ok {
		fmt.Fprintln(os.Stderr, "failed to retrieve caller information")
	}
	return fn, ln
}

func groupInfo(level int, msg string, args ...any) string {
	t := time.Now()
	ts := t.Format(LOG_TIME_FMT)
	msg = xmsg(msg, args)
	fn, ln := callInfo()
	l := parseLogLevel(level)
	msg = fmt.Sprintf(LOG_MSG_FORMAT, ts, l, fn, ln, msg)

	if runtime.GOOS == "windows" {
		msg += "\r\n"
	} else {
		msg += "\n"
	}

	return msg
}

func (log *Logger) Trace(msg string, args ...any) {
	msg = groupInfo(TRACE, msg, args...)
	for _, wr := range log.writers {
		println("pass")

		if wr.enable(TRACE) {
			println("pass2")

			wr.write(msg)
		}
	}
}

func (log *Logger) Debug(msg string, args ...any) {

}

func (log *Logger) Info(msg string, args ...any) {

}

func (log *Logger) Warn(msg string, args ...any) {

}

func (log *Logger) Error(msg string, args ...any) {

}

func (log *Logger) Fatal(msg string, args ...any) {

}

// func SubPath(xpath string, layer int) string {
// 	pstck := make([]string, 0)

// 	for {
// 		d, f := path.Split(xpath)
// 		println(d, f)
// 	}
// }
