package keen

import (
	"log"

	"gitea.fcdm.top/lixuan/keen/ylog"
)

var Logger *log.Logger

// SetLogger 设置keen使用的logger，如果不设置则使用默认logger
func SetLogger(l *log.Logger) {
	if l != nil {
		Logger = l
	} else {
		logConfig := ylog.YLogger{
			ConsoleLevel:    ylog.Debug,
			ConsoleColorful: true,
		}
		Logger = logConfig.InitLogger()
	}
}
