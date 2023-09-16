package ylog

import (
	"bytes"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gitea.fcdm.top/lixuan/keen/fsop"
	"github.com/fatih/color"
)

type LogLevel uint8

const (
	Trace LogLevel = 0
	Debug LogLevel = 1
	Info  LogLevel = 2
	Warn  LogLevel = 3
	Error LogLevel = 4
	TRACE          = "<TRACE>"
	DEBUG          = "<DEBUG>"
	INFO           = "<INFO> "
	WARN           = "<WARN> "
	ERROR          = "<ERROR>"
)

func parseLogLevel(l string) LogLevel {
	switch l {
	case "<TRACE>":
		return Trace
	case "<DEBUG>":
		return Debug
	case "<INFO>":
		return Info
	case "<WARN>":
		return Warn
	case "<ERROR>":
		return Error
	}
	return Trace
}

const (
	LOG_FILE_DIR           = "C:\\ProgramData\\sqlpvd\\log"
	LOG_EXPIRE             = 7 * 24 * 60 * 60 * time.Second
	LOG_FILENAME_PREFIX    = "sqlpvd_"
	LOG_FILENAME_TIMESTAMP = "20060102150405"
	LOG_FLAG               = log.LstdFlags | log.Lshortfile
	CONSOLE_LOG_LEVEL      = INFO
	FILE_LOG_LEVEL         = TRACE
)

type YLogger struct {
	ConsoleLevel    LogLevel
	ConsoleColorful bool
	FileLog         bool
	FileLogDir      string
	FileLevel       LogLevel
	FileSuffix      string
	FileClean       time.Duration
}

func (l YLogger) InitLogger() *log.Logger {
	wrs := YLogWriter{}
	writer1 := NewConsoleWriter(l.ConsoleLevel, l.ConsoleColorful)
	wrs.console = writer1
	if l.FileLog {
		if l.FileLogDir == "" {
			l.FileLogDir = LOG_FILE_DIR
		}
		if l.FileClean == 0 {
			l.FileClean = LOG_EXPIRE
		}
		if l.FileSuffix == "" {
			l.FileSuffix = "DEFAULT"
		}
		err := CleanOldFileLogs(l.FileLogDir, l.FileClean)
		if err == nil {
			writer2, err := NewFileWriter(l.FileLogDir, l.FileSuffix, l.FileLevel)
			if err == nil {
				wrs.file = writer2
			}
		}
	}

	return log.New(wrs, "", LOG_FLAG)
}

type YLogWriter struct {
	console io.Writer
	file    io.Writer
}

func (w YLogWriter) Write(p []byte) (n int, err error) {
	if w.console != nil {
		n, err := w.console.Write(p)
		if err != nil {
			return n, err
		}
	}
	if w.file != nil {
		n, err := w.file.Write(p)
		if err != nil {
			return n, err
		}
	}

	return 0, nil
}

type ConsoleWriter struct {
	out         io.Writer
	level       LogLevel
	enableColor bool
	colorSet    map[LogLevel]*color.Color
}

func (w ConsoleWriter) Write(p []byte) (n int, err error) {
	if runtime.GOOS == "windows" {
		p = bytes.Replace(p, []byte("\n"), []byte("\r\n"), -1)
	}

	segs := bytes.Split(p, []byte(" "))
	level := parseLogLevel(string(segs[3]))
	if level >= w.level {
		if w.enableColor {
			return w.out.Write([]byte(w.colorSet[level].Sprint(string(p))))
		} else {
			return w.out.Write(p)
		}
	}

	return 0, nil
}

func NewConsoleWriter(level LogLevel, enableColor bool) ConsoleWriter {
	colorSet := map[LogLevel]*color.Color{
		Trace: color.New(color.FgHiWhite),
		Debug: color.New(color.FgHiGreen),
		Info:  color.New(color.FgBlue),
		Warn:  color.New(color.FgYellow),
		Error: color.New(color.FgRed),
	}
	return ConsoleWriter{
		out:         os.Stdout,
		level:       level,
		enableColor: enableColor,
		colorSet:    colorSet,
	}
}

func CleanOldFileLogs(dir string, expireTime time.Duration) error {
	if !fsop.IsDir(dir) {
		return nil
	}

	nt := time.Now()

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if dir == path {
			return nil
		}

		if d.IsDir() {
			return fs.SkipDir
		}

		if strings.HasPrefix(d.Name(), LOG_FILENAME_PREFIX) {
			sec := strings.Split(d.Name(), "_")
			// parse时间的时区要设置成local，否则为UTC时间
			t, err := time.ParseInLocation(LOG_FILENAME_TIMESTAMP, sec[1], time.Local)

			if err == nil {
				elapse := nt.Sub(t)
				if elapse >= expireTime {
					err := os.Remove(path)
					if err != nil {
						return err
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

type FileWriter struct {
	f     *os.File
	level LogLevel
}

func (w FileWriter) Write(p []byte) (n int, err error) {
	if runtime.GOOS == "windows" {
		p = bytes.Replace(p, []byte("\n"), []byte("\r\n"), -1)
	}

	segs := bytes.Split(p, []byte(" "))
	level := parseLogLevel(string(segs[3]))
	if level >= w.level {
		return w.f.Write(p)
	}

	return 0, nil
}

func NewFileWriter(logDir string, suffix string, level LogLevel) (FileWriter, error) {
	nowTimeStr := time.Now().Format(LOG_FILENAME_TIMESTAMP)
	filename := filepath.Join(logDir, LOG_FILENAME_PREFIX+nowTimeStr+"_"+suffix+".log")

	if !fsop.IsDir(logDir) {
		err := os.MkdirAll(logDir, os.ModeDir)
		if err != nil {
			return FileWriter{}, err
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return FileWriter{}, err
	}

	return FileWriter{
		f:     f,
		level: level,
	}, nil
}
