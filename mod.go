package keen

import "gitea.fcdm.top/lixuan/keen/ylog2"

var consoleWriter ylog2.ConsoleWriter = *ylog2.NewConsoleWriter(func(i int8) bool { return i >= ylog2.DEBUG }, false)
var Log ylog2.Logger = ylog2.NewLogger(&consoleWriter)
