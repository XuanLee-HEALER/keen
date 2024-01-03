package pvd

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gitea.fcdm.top/lixuan/keen"
	"gitea.fcdm.top/lixuan/keen/util"
	"gitea.fcdm.top/lixuan/keen/ylog"
	"github.com/cnyjp/fcdmpublic/model"
)

const (
	C_ERR_EXIT = 1
)

var LogNameReg = regexp.MustCompile(`(backup|restore|mount|umount|discover|refresh|illegal|pluginfo)_(.+)_(\d{14})\.log$`)

var SimpleArch ylog.Archive = func(fn string) (bool, string) {
	matches := LogNameReg.FindAllStringSubmatch(fn, -1)
	if len(matches) >= 1 {
		gs := matches[0]
		if len(gs) >= 2 {
			return true, gs[1]
		}
	}

	return false, ""
}

// GenLogName 根据命令、jobid和时间戳生成日志文件名称
func GenLogName(cmd, jobId string) string {
	return fmt.Sprintf("%s_%s_%s.log", cmd, jobId, util.LocalFormat(util.SERIAL_FMT, time.Now()))
}

// PluginInfoJson 将plugininfo的Config结构转换为json字符串，如果转换失败则返回空字符串结果
func PluginInfoJson(conf model.PluginConfig) string {
	bs, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		keen.Log.Error("failed to marshal the plugin configuration to json format: %v", err)
		return ""
	}

	return string(bs)
}

// LegalCaller 判断provider的调用者是否合法，根据进程树上是否存在fcdmconnector进程来判断
func LegalCaller() bool {
	pid := strconv.Itoa(os.Getpid())
	keen.Log.Debug("the ID of current process", pid)
	for {
		proc, err := util.QueryProcess(pid)
		if err != nil {
			keen.Log.Error("failed to get the ID of process [%s]: %v", pid, err)
			break
		}

		keen.Log.Debug("parent process status:\n%s", proc)

		if runtime.GOOS == "windows" {
			if strings.Contains(proc.Command, "fcdmconnector") {
				return true
			}
		} else {
			if proc.Command[1:len(proc.Command)-1] == "fcdmconnector" {
				return true
			}
		}

		pid = proc.PPID
	}

	return false
}
