package util

import (
	"fmt"
	"runtime"

	"gitea.fcdm.top/lixuan/keen"
	"gitea.fcdm.top/lixuan/keen/ylog"
	"gitea.fcdm.top/lixuan/keen/ysys"
)

func isNix() bool {
	return runtime.GOOS == "aix" || runtime.GOOS == "darwin" || runtime.GOOS == "linux"
}

func isWin() bool {
	return runtime.GOOS == "windows"
}

type ProcessInfo struct {
	PID         string `json:"pid"`              // pid
	Command     string `json:"command"`          // exec path
	State       string `json:"state"`            // state id(enum)
	StateDesc   string `json:"state_desc"`       // state description
	PPID        string `json:"ppid"`             // ppid
	ProcGroupId string `json:"process_group_id"` // linux:process group ID of the process
}

// Processe 根据指定pid获取进程信息
func Processe(pid string) (ProcessInfo, error) {
	res := ProcessInfo{}

	if isNix() {
		ps, err := ysys.QueryProcess(pid)
		if err != nil {
			keen.Logger.Println(ylog.ERROR, "failed to query process detail of pid(%s): %v", pid, err)
			return res, err
		}

		res.PID = ps.PID
		res.Command = ps.Command
		res.State = ps.State
		res.StateDesc = ps.StateDesc
		res.PPID = ps.PPID
		res.ProcGroupId = ps.PGrpID
	}
	if isWin() {

	}

	return res, nil
}

func TotalMemory() (uint64, error) {
	var res uint64
	if isNix() {
		mem, err := ysys.QuerySystemMemory()
		if err != nil {
			keen.Logger.Println(ylog.ERROR, fmt.Sprintf("failed to query memory detail: %v", err))
			return res, err
		}
		res = mem[len(mem)-1].End
	}
	if isWin() {

	}
	return res, nil
}
