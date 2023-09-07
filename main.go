package main

import (
	"time"

	"gitea.fcdm.top/lixuan/keen/ylog"
)

func main() {
	ylogger := ylog.YLogger{
		ConsoleLevel: ylog.Info,
		FileLog:      true,
		FileLogDir:   "F:\\go_projects\\keen\\mewo_log",
		FileLevel:    ylog.Trace,
		FileClean:    5 * time.Second,
	}
	log := ylogger.InitLogger()
	log.Println(ylog.TRACE, "good")
	log.Println(ylog.DEBUG, "good")
	log.Println(ylog.INFO, "good")
	log.Println(ylog.WARN, "good")
	log.Println(ylog.ERROR, "good")
}
