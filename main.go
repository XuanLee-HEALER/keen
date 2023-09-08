package main

import (
	"time"

	"gitea.fcdm.top/lixuan/keen/ylog"
)

func main() {
	ylogger := ylog.YLogger{
		ConsoleLevel:    ylog.Info,
		ConsoleColorful: true,
		FileLog:         true,
		FileLogDir:      "F:\\go_projects\\keen",
		FileLevel:       ylog.Trace,
		FileClean:       30 * time.Second,
	}
	log := ylogger.InitLogger()
	log.Println(ylog.TRACE, "good luck")
	log.Println(ylog.DEBUG, "good luck")
	log.Println(ylog.INFO, "good luck")
	log.Println(ylog.WARN, "good luck")
	log.Println(ylog.ERROR, "good luck")
}
