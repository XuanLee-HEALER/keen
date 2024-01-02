package keen

import "gitea.fcdm.top/lixuan/keen/ylog"

var consoleWriter ylog.ConsoleWriter = *ylog.NewConsoleWriter(func(i int8) bool { return i >= ylog.DEBUG }, true)
var Log ylog.Logger = ylog.NewLogger(&consoleWriter)
