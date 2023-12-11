package ylog2_test

import (
	"testing"

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
	console := ylog2.NewConsoleWriter(func(i int8) bool { return i >= ylog2.TRACE })
	logger := ylog2.NewLogger(true, console)
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
	console := ylog2.NewConsoleWriter(func(i int8) bool { return i >= ylog2.TRACE })
	logger := ylog2.NewLogger(true, console)

	logger.Trace("test trace log message")
	logger.Debug("test debug log message")
	logger.Info("test info log message")
	logger.Warn("test warn log message")
	logger.Error("test error log message")
	// logger.Fatal("test fatal log message")
}

func TestFileLogger(t *testing.T) {
	file, err := ylog2.NewFileWriter(".", "test.log", func(i int8) bool { return i >= ylog2.TRACE }, 0, func(pattern, des string) bool { return false })
	if err != nil {
		t.Error(err)
	}
	logger := ylog2.NewLogger(true, file)

	logger.Trace("test trace log message")
	logger.Debug("test debug log message")
	logger.Info("test info log message")
	logger.Warn("test warn log message")
	logger.Error("test error log message")
	// logger.Fatal("test fatal log message")
}

func TestMultiLogger(t *testing.T) {
	console := ylog2.NewConsoleWriter(func(i int8) bool { return i >= ylog2.TRACE })
	file, err := ylog2.NewFileWriter(".", "test.log", func(i int8) bool { return i >= ylog2.TRACE }, 0, func(pattern, des string) bool { return false })
	if err != nil {
		t.Error(err)
	}
	logger := ylog2.NewLogger(true, console, file)

	for i := 0; i < 10; i++ {
		logger.Trace("test trace log message")
		logger.Debug("test debug log message")
		logger.Info("test info log message")
		logger.Warn("test warn log message")
		logger.Error("test error log message")
		// logger.Fatal("test fatal log message")
	}
}
