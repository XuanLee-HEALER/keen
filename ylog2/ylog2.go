package ylog2

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
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
	LOG_TIME_FMT   = "2006-01-02 15:04:05"
	LOG_MSG_FORMAT = "[%s] [%-5s] [%s:%d] - %s"
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

type LevelFilter = func(int8) bool
type Archive = func(pattern string, des string) bool

type LogMessage struct {
	level int8
	msg   string
}

type LogWriter interface {
	enable(int8) bool
	flush()
	io.Writer
}

type ConsoleWriter struct {
	levelFilter func(int8) bool
}

func NewConsoleWriter(level LevelFilter) *ConsoleWriter {
	return &ConsoleWriter{
		levelFilter: level,
	}
}

func (w *ConsoleWriter) enable(l int8) bool {
	return w.levelFilter(l)
}

func (w *ConsoleWriter) Write(bs []byte) (int, error) {
	return os.Stdout.Write(bs)
}

func (w *ConsoleWriter) flush() {
	os.Stdout.Sync()
}

type FileWriter struct {
	logDir      string
	file        *os.File
	cfile       io.Writer
	expire      time.Duration
	archive     Archive
	levelFilter LevelFilter
}

func NewFileWriter(logPath string, filename string, level LevelFilter, expire time.Duration, archive Archive) (*FileWriter, error) {
	f := filepath.Join(logPath, filename)
	fs, err := os.Create(f)
	if err != nil {
		errorf("failed to create log file (%s): %v", f, err)
		return nil, err
	}

	wr := new(FileWriter)
	wr.logDir = logPath
	wr.file = fs
	wr.cfile = colorable.NewColorable(fs)
	wr.expire = expire
	wr.archive = archive
	wr.levelFilter = level
	return wr, nil
}

func (w *FileWriter) enable(l int8) bool {
	return w.levelFilter(l)
}

func (w *FileWriter) flush() {
	// golang对于文件写入不使用缓冲
}

func checkExpiredLogFile(logDir string, expire time.Duration) {

}

func checkArchiveLogFile(logDir string, archive Archive) {

}

func (w *FileWriter) Write(p []byte) (n int, err error) {
	// check expired log
	checkExpiredLogFile(w.logDir, w.expire)
	// check if archive log or not
	checkArchiveLogFile(w.logDir, w.archive)
	// write log
	if runtime.GOOS == "windows" {
		return w.cfile.Write(p)
	} else {
		return w.file.Write(p)
	}
}

type Logger struct {
	writers []LogWriter
	colored bool
	colors  []*color.Color
}

func NewLogger(colourful bool, writers ...LogWriter) Logger {
	l := Logger{
		writers: writers,
	}
	if colourful {
		traceC := color.New(color.FgCyan)
		debugC := color.New(color.FgGreen)
		infoC := color.New(color.FgWhite)
		warnC := color.New(color.FgYellow)
		errorC := color.New(color.FgRed)
		fatalC := color.New(color.FgMagenta)
		l.colored = colourful
		l.colors = []*color.Color{traceC, debugC, infoC, warnC, errorC, fatalC}
		for _, c := range l.colors {
			c.EnableColor()
		}
	}
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
	msg = xmsg(msg, args...)
	fn, ln := callInfo()
	fn = SubPath(fn, 2)
	l := parseLogLevel(level)
	msg = fmt.Sprintf(LOG_MSG_FORMAT, ts, l, fn, ln, msg)

	if runtime.GOOS == "windows" {
		msg += "\r\n"
	} else {
		msg += "\n"
	}

	return msg
}

func (log *Logger) log(msg LogMessage) {
	for _, writer := range log.writers {
		if writer.enable(msg.level) {
			if log.colored {
				_, err := log.colors[msg.level].Fprint(writer, msg.msg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to write log: %v", err)
				}
				writer.flush()
			} else {
				_, err := writer.Write([]byte(msg.msg))
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to write log: %v", err)
				}
				writer.flush()
			}
		}
	}
}

func (log *Logger) Trace(msg string, args ...any) {
	msg = groupInfo(TRACE, msg, args...)
	log.log(LogMessage{TRACE, msg})
}

func (log *Logger) Debug(msg string, args ...any) {
	msg = groupInfo(DEBUG, msg, args...)
	log.log(LogMessage{DEBUG, msg})
}

func (log *Logger) Info(msg string, args ...any) {
	msg = groupInfo(INFO, msg, args...)
	log.log(LogMessage{INFO, msg})
}

func (log *Logger) Warn(msg string, args ...any) {
	msg = groupInfo(WARN, msg, args...)
	log.log(LogMessage{WARN, msg})
}

func (log *Logger) Error(msg string, args ...any) {
	msg = groupInfo(ERROR, msg, args...)
	log.log(LogMessage{ERROR, msg})
}

func (log *Logger) Fatal(msg string, args ...any) {
	msg = groupInfo(FATAL, msg, args...)
	log.log(LogMessage{FATAL, msg})
	os.Exit(1)
}

// SubPath 截取路径的后n级目录
func SubPath(p string, n int) string {
	pstck := make([]string, 0)

	count := n
	for count > 0 {
		d, f := filepath.Split(p)
		// t.Logf("dir: %s, file: %s", d, f)
		pstck = append(pstck, f)
		if d == "" || d == "/" {
			break
		}
		d = d[:len(d)-1]

		count--
		p = d
	}

	for i, j := 0, len(pstck)-1; i < j; i, j = i+1, j-1 {
		pstck[i], pstck[j] = pstck[j], pstck[i]
	}

	return strings.Join(pstck, string(os.PathSeparator))
}

func errorf(fmtStr string, args ...any) {
	fmt.Fprintf(os.Stderr, fmtStr, args...)
}
