package ylog2

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"gitea.fcdm.top/lixuan/keen/util"
	"github.com/fatih/color"
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
type Archive = func(fn string) (bool, string)

type LogMessage struct {
	level int8
	msg   string
}

type LogWriter interface {
	msg(LogMessage)
	enable(int8) bool
	flush()
	clean()
	io.Writer
}

type ConsoleWriter struct {
	levelFilter func(int8) bool
	colored     bool
	colors      []*color.Color
	curMsg      LogMessage
	ai          atomic.Bool
}

func NewConsoleWriter(level LevelFilter, colourful bool) *ConsoleWriter {
	res := new(ConsoleWriter)
	res.ai = atomic.Bool{}
	res.levelFilter = level
	if colourful {
		traceC := color.New(color.FgCyan)
		debugC := color.New(color.FgGreen)
		infoC := color.New(color.FgWhite)
		warnC := color.New(color.FgYellow)
		errorC := color.New(color.FgRed)
		fatalC := color.New(color.FgMagenta)
		res.colored = colourful
		res.colors = []*color.Color{traceC, debugC, infoC, warnC, errorC, fatalC}
		for _, c := range res.colors {
			c.EnableColor()
		}
	}
	return res
}

func (w *ConsoleWriter) msg(m LogMessage) {
	for !w.ai.CompareAndSwap(false, true) {
	}
	w.curMsg = m
}

func (w *ConsoleWriter) enable(l int8) bool {
	return w.levelFilter(l)
}

func (w *ConsoleWriter) Write(bs []byte) (int, error) {
	var (
		n   int
		err error
	)
	if w.colored {
		n, err = w.colors[w.curMsg.level].Fprint(os.Stdout, string(bs))
	} else {
		n, err = os.Stdout.Write(bs)
	}
	w.ai.CompareAndSwap(true, false)
	return n, err
}

func (w *ConsoleWriter) flush() {
	os.Stdout.Sync()
}

func (w *ConsoleWriter) clean() {
	if w.colored {
		for _, c := range w.colors {
			c.DisableColor()
		}
	}
}

type FileWriter struct {
	logDir      string
	file        *os.File
	expire      time.Duration
	archive     Archive
	levelFilter LevelFilter
}

func DeleteExpiredLogAndArchive(logDir string, expire time.Duration, archive Archive) {
	var (
		isArc bool
		arcMp map[string][]string = make(map[string][]string)
	)

	if archive != nil {
		isArc = true
	}

	n := time.Now()
	st := n.Add(-expire)
	absPath, _ := filepath.Abs(logDir)
	infof("delete expired log files in [%s]: motification datetime <= %s\n", absPath, st.Format(LOG_TIME_FMT))
	err := filepath.WalkDir(absPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == absPath {
			return nil
		}

		if d.IsDir() {
			return filepath.SkipDir
		}

		fi, err := d.Info()
		if err != nil {
			return err
		}

		if expire != 0 && fi.ModTime().Before(st) {
			err = os.Remove(path)
			if err != nil {
				errorf("failed to delete log file [%s]: %v\n", path, err)
			}
		} else {
			if isArc {
				if b, dst := archive(d.Name()); b {
					arcMp[dst] = append(arcMp[dst], path)
				}
			}
		}

		return nil
	})

	if err != nil {
		errorf("error occured while clean expired log files: %v\n", err)
	}

	for dir, fs := range arcMp {
		ndir := filepath.Join(absPath, dir)
		err := os.Mkdir(ndir, os.ModeDir)
		if err != nil {
			if !os.IsExist(err) {
				errorf("failed to create archive directory [%s]: %v\n", ndir, err)
				continue
			}
		}

		for _, f := range fs {
			np := filepath.Join(ndir, filepath.Base(f))
			err := os.Rename(f, np)
			if err != nil {
				errorf("failed to move the log file from [%s] to [%s]: %v\n", f, np, err)
			}
		}
	}

}

func NewFileWriter(logPath string, filename string, level LevelFilter, expire time.Duration, archive Archive) (*FileWriter, error) {
	f := filepath.Join(logPath, filename)
	absf, err := filepath.Abs(f)
	if err != nil {
		return nil, err
	}

	pd := filepath.Dir(absf)
	dir, err := os.Stat(pd)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(pd, os.ModeDir|700)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if !dir.IsDir() {
		return nil, fmt.Errorf("the path of log directory [%s] is existed and it is a file path", absf)
	}

	// check expired log and archive log
	DeleteExpiredLogAndArchive(logPath, expire, archive)

	fs, err := os.Create(absf)
	if err != nil {
		errorf("failed to create log file (%s): %v\n", f, err)
		return nil, err
	}

	wr := new(FileWriter)
	wr.logDir = logPath
	wr.file = fs
	wr.expire = expire
	wr.archive = archive
	wr.levelFilter = level
	return wr, nil
}

func (w *FileWriter) msg(LogMessage) {}

func (w *FileWriter) enable(l int8) bool {
	return w.levelFilter(l)
}

func (w *FileWriter) flush() {
	// golang对于文件写入不使用缓冲
}

func (w *FileWriter) Write(p []byte) (n int, err error) {
	// write log
	return w.file.Write(p)
}

func (w *FileWriter) clean() {
	if err := w.file.Close(); err != nil {
		errorf("failed to close the log file: %v\n", err)
	}
}

type Logger struct {
	writers []LogWriter
}

func NewLogger(writers ...LogWriter) Logger {
	l := Logger{
		writers: writers,
	}

	return l
}

func callInfo() (string, int) {
	_, fn, ln, ok := runtime.Caller(3)
	if !ok {
		errorf("failed to retrieve caller information\n")
	}
	return fn, ln
}

func groupInfo(level int, msg string, args ...any) string {
	t := time.Now()
	ts := t.Format(LOG_TIME_FMT)
	fn, ln := callInfo()
	fn = SubPath(fn, 2)
	l := parseLogLevel(level)
	msg = fmt.Sprintf(LOG_MSG_FORMAT, ts, l, fn, ln, fmt.Sprintf(msg, args...))

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
			writer.msg(msg)
			_, err := writer.Write([]byte(msg.msg))
			if err != nil {
				errorf("failed to write log: %v\n", err)
			}
			writer.flush()
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
	util.ExitWith(1, log.Clean)
}

func (log *Logger) Clean() {
	for _, w := range log.writers {
		w.clean()
	}
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

func infof(fmtStr string, args ...any) {
	fmt.Fprintf(os.Stdout, fmtStr, args...)
}

func errorf(fmtStr string, args ...any) {
	fmt.Fprintf(os.Stderr, fmtStr, args...)
}
