package ylog_test

import (
	"fmt"
	"testing"
	"time"

	"gitea.fcdm.top/lixuan/keen/ylog"
)

func TestFileLogger(t *testing.T) {
	logger := ylog.YLogger{
		ConsoleColorful: true,
		FileLog:         true,
		FileLogDir:      "C:\\tlog",
		FileLevel:       ylog.Trace,
		FileSuffix:      "unittest",
		FileClean:       1 * time.Second,
	}

	flog := logger.InitLogger()
	flog.Println(ylog.TRACE, fmt.Sprintf("a \n%d", 123))
	flog.Println(ylog.TRACE, fmt.Sprintf("a \n%d", 123))
	flog.Println(ylog.INFO, "test logger1")
	flog.Println(ylog.WARN, "test logger2")
	flog.Println(ylog.ERROR, "test logger3")
}
