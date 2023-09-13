package ylog_test

import (
	"testing"
	"time"

	"gitea.fcdm.top/lixuan/keen/ylog"
)

func TestFileLogger(t *testing.T) {
	logger := ylog.YLogger{
		ConsoleColorful: true,
		FileLog:         true,
		FileLogDir:      ".",
		FileLevel:       ylog.Info,
		FileClean:       1 * time.Second,
	}

	flog := logger.InitLogger()
	flog.Println(ylog.INFO, "test logger1")
	flog.Println(ylog.WARN, "test logger2")
	flog.Println(ylog.ERROR, "test logger3")
}
