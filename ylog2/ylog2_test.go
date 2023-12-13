package ylog2_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"gitea.fcdm.top/lixuan/keen/ylog2"
)

func TestYLog2Level(t *testing.T) {
	t.Log(ylog2.TRACE)
	t.Log(ylog2.DEBUG)
	t.Log(ylog2.INFO)
	t.Log(ylog2.WARN)
	t.Log(ylog2.ERROR)
	t.Log(ylog2.FATAL)
}

func TestYLog2ConsoleLogger(t *testing.T) {
	console := ylog2.NewConsoleWriter(func(i int8) bool { return i >= ylog2.TRACE }, false)
	logger := ylog2.NewLogger(console)
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
		t.Log(ylog2.SubPath(p, 1))
	}
}

func TestConsoleLogger(t *testing.T) {
	console := ylog2.NewConsoleWriter(func(i int8) bool { return i >= ylog2.TRACE }, true)
	logger := ylog2.NewLogger(console)

	logger.Trace("test trace log message")
	logger.Debug("test debug log message")
	logger.Info("test info log message")
	logger.Warn("test warn log message")
	logger.Error("test error log message")
	logger.Fatal("test fatal log message")
}

func TestConcurrentConsoleLogger(t *testing.T) {
	console := ylog2.NewConsoleWriter(func(i int8) bool { return i >= ylog2.TRACE }, true)
	logger := ylog2.NewLogger(console)
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
	file, err := ylog2.NewFileWriter(".", "test.log", func(i int8) bool { return i >= ylog2.TRACE }, 0, nil)
	if err != nil {
		t.Error(err)
	}
	logger := ylog2.NewLogger(file)

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
			file, err := ylog2.NewFileWriter("./testlogger", fmt.Sprintf("test-%d.log", idx), func(i int8) bool { return i >= ylog2.TRACE }, 30*time.Second, func(fn string) (bool, string) {
				if strings.HasPrefix(fn, "test") {
					return true, "test"
				}
				return false, ""
			})
			if err != nil {
				t.Error(err)
			}
			logger := ylog2.NewLogger(file)

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
	file, err := ylog2.NewFileWriter("./testlogger", "test-2.log", func(i int8) bool { return i >= ylog2.TRACE }, 0, func(fn string) (bool, string) {
		if strings.HasPrefix(fn, "test") {
			return true, "test"
		}
		return false, ""
	})
	if err != nil {
		t.Error(err)
	}
	logger := ylog2.NewLogger(file)

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
	console := ylog2.NewConsoleWriter(func(i int8) bool { return i >= ylog2.TRACE }, true)
	file, err := ylog2.NewFileWriter(".", "test-0.log", func(i int8) bool { return i >= ylog2.TRACE }, 0, nil)
	if err != nil {
		t.Error(err)
	}
	logger := ylog2.NewLogger(console, file)

	for i := 0; i < 10; i++ {
		logger.Trace("test trace log message")
		logger.Debug("test debug log message")
		logger.Info("test info log message")
		logger.Warn("test warn log message")
		logger.Error("test error log message")
		// logger.Fatal("test fatal log message")
	}
}
