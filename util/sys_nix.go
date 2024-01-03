//go:build linux || aix || darwin
// +build linux aix darwin

package util

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var processStateDescription map[string]string = map[string]string{
	"R": "Running",
	"S": "Sleeping in an interruptible wait",
	"D": "Waiting in uninterruptible disk sleep",
	"Z": "Zombie",
	"T": "Stopped (on a signal) or (before Linux2.6.33) trace stopped",
	"t": "Tracing stop (Linux 2.6.33 onward)",
	"W": "Paging (only before Linux 2.6.0) | Waking (Linux 2.6.33 to 3.13 only)",
	"X": "Dead (from Linux 2.6.0 onward)",
	"x": "Dead (Linux 2.6.33 to 3.13 only)",
	"K": "Wakekill (Linux 2.6.33 to 3.13 only)",
	"P": "Parked (Linux 3.9 to 3.13 only)",
	"I": "Idle (Linux 4.14 onward)",
}

type ProcessStat struct {
	PID       string
	Command   string
	State     string
	StateDesc string
	PPID      string
	PGrpID    string
}

func (p ProcessStat) String() string {
	return fmt.Sprintf("PID: %s\nCommand: %s\nState: %s\nStateDescription: %s\nPPID: %s\nProcessGroup: %s", p.PID, p.Command, p.State, p.StateDesc, p.PPID, p.PGrpID)
}

// queryProcess 查询指定pid对应的进程状态，源数据为/proc/<pid>/stat文件
func QueryProcess(pid string) (ProcessStat, error) {
	res := ProcessStat{}
	path := fmt.Sprintf("/proc/%s/stat", pid)
	f, err := os.Open(path)
	if err != nil {
		return res, err
	}
	defer f.Close()

	bs, err := io.ReadAll(f)
	if err != nil {
		return res, err
	}

	segs := strings.Split(string(bs), " ")
	res.PID = segs[0]
	res.Command = segs[1]
	res.State = segs[2]
	res.StateDesc = processStateDescription[segs[2]]
	res.PPID = segs[3]
	res.PGrpID = segs[4]

	return res, nil
}
