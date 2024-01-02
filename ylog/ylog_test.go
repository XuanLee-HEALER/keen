package ylog_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"gitea.fcdm.top/lixuan/keen/ylog"
)

func TestylogLevel(t *testing.T) {
	t.Log(ylog.TRACE)
	t.Log(ylog.DEBUG)
	t.Log(ylog.INFO)
	t.Log(ylog.WARN)
	t.Log(ylog.ERROR)
	t.Log(ylog.FATAL)
}

func TestylogConsoleLogger(t *testing.T) {
	console := ylog.NewConsoleWriter(func(i int8) bool { return i >= ylog.TRACE }, false)
	logger := ylog.NewLogger(console)
	logger.Trace("test trace log message")
}

func TestSubPath(t *testing.T) {
	p1 := "p1"
	p2 := "p1/p2"
	p3 := "/p3/p2/p1"
	p4 := "C:\\"
	p5 := "/"
	p6 := "C:\\p1"
	p7 := "c:\\p2\\p1"
	samples := []string{p1, p2, p3, p4, p5, p6, p7}

	for _, p := range samples {
		t.Log(ylog.SubPath(p, 1))
	}
}

func TestConsoleLogger(t *testing.T) {
	console := ylog.NewConsoleWriter(func(i int8) bool { return i >= ylog.TRACE }, true)
	logger := ylog.NewLogger(console)

	logger.Trace("test trace log message")
	logger.Debug("test debug log message")
	logger.Info("test info log message")
	logger.Warn("test warn log message")
	logger.Error("test error log message")
	logger.Fatal("test fatal log message")
}

func TestConcurrentConsoleLogger(t *testing.T) {
	console := ylog.NewConsoleWriter(func(i int8) bool { return i >= ylog.TRACE }, true)
	logger := ylog.NewLogger(console)
	ch := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func(idx int) {
			for i := 0; i < 20; i++ {
				logger.Trace("goroutine(%d) => test trace log message", idx)
				logger.Debug("goroutine(%d) => test debug log message", idx)
				logger.Info("goroutine(%d) => test info log message", idx)
				logger.Warn("goroutine(%d) => test warn log message", idx)
				logger.Error("goroutine(%d) => test error log message", idx)
			}
			ch <- struct{}{}
		}(i)
	}

	counter := 0
	for range ch {
		counter++
		if counter == 10 {
			break
		}
	}
	logger.Clean()
}

func TestFileLogger(t *testing.T) {
	file, err := ylog.NewFileWriter(".", "test.log", func(i int8) bool { return i >= ylog.TRACE }, 0, nil)
	if err != nil {
		t.Error(err)
	}
	logger := ylog.NewLogger(file)

	logger.Trace("test trace log message")
	logger.Debug("test debug log message")
	logger.Info("test info log message")
	logger.Warn("test warn log message")
	logger.Error("test error log message")
	// logger.Fatal("test fatal log message")

	logger.Clean()
}

func TestFileLoggerExpireAndArchive(t *testing.T) {
	for i := 0; i < 3; i++ {
		go func(idx int) {
			file, err := ylog.NewFileWriter("./testlogger", fmt.Sprintf("test-%d.log", idx), func(i int8) bool { return i >= ylog.TRACE }, 30*time.Second, func(fn string) (bool, string) {
				if strings.HasPrefix(fn, "test") {
					return true, "test"
				}
				return false, ""
			})
			if err != nil {
				t.Error(err)
			}
			logger := ylog.NewLogger(file)

			logger.Trace("test trace log message")
			logger.Debug("test debug log message")
			logger.Info("test info log message")
			logger.Warn("test warn log message")
			logger.Error("test error log message")

			logger.Clean()
		}(i)
		time.Sleep(5 * time.Second)
	}
}

func TestConcurrentFileLogger(t *testing.T) {
	file, err := ylog.NewFileWriter("./testlogger", "test-2.log", func(i int8) bool { return i >= ylog.TRACE }, 0, func(fn string) (bool, string) {
		if strings.HasPrefix(fn, "test") {
			return true, "test"
		}
		return false, ""
	})
	if err != nil {
		t.Error(err)
	}
	logger := ylog.NewLogger(file)

	ch := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func(idx int) {
			for i := 0; i < 20; i++ {
				logger.Trace("goroutine(%d) => test trace log message", idx)
				logger.Debug("goroutine(%d) => test debug log message", idx)
				logger.Info("goroutine(%d) => test info log message", idx)
				logger.Warn("goroutine(%d) => test warn log message", idx)
				logger.Error("goroutine(%d) => test error log message", idx)
			}
			ch <- struct{}{}
		}(i)
	}

	counter := 0
	for range ch {
		counter++
		if counter == 10 {
			break
		}
	}

	logger.Clean()
}

func TestMultiLogger(t *testing.T) {
	console := ylog.NewConsoleWriter(func(i int8) bool { return i >= ylog.TRACE }, true)
	file, err := ylog.NewFileWriter(".", "test-0.log", func(i int8) bool { return i >= ylog.TRACE }, 0, nil)
	if err != nil {
		t.Error(err)
	}
	logger := ylog.NewLogger(console, file)

	for i := 0; i < 10; i++ {
		logger.Trace("test trace log message")
		logger.Debug("test debug log message")
		logger.Info("test info log message")
		logger.Warn("test warn log message")
		logger.Error("test error log message")
		// logger.Fatal("test fatal log message")
	}
}
