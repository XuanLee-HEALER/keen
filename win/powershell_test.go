package win_test

import (
	"testing"
	"time"

	"gitea.fcdm.top/lixuan/keen/win"
	"gitea.fcdm.top/lixuan/keen/ylog"
)

func init() {
	logger := ylog.YLogger{
		ConsoleLevel:    ylog.Trace,
		ConsoleColorful: true,
		FileLog:         true,
		FileLogDir:      "C:\\tlog",
		FileClean:       3 * time.Second,
	}
	win.SetupLogger(logger.InitLogger())
}

func TestPSVersionTable(t *testing.T) {
	v, err := win.PSVersionTable()
	if err != nil {
		t.Fatalf("failed to fetch powershell version: %v", err)
	}
	t.Log("version:", v)
}
