// go:build linux amd64

package nix

import (
	"bufio"
	"container/list"
	"fmt"
	"io"
	"os"
	"os/user"
	"strconv"
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

type ProcessStatus struct {
	PID       string
	Command   string
	State     string
	StateDesc string
	PPID      string
	PGrpID    string
}

func (p ProcessStatus) String() string {
	return fmt.Sprintf("PID: %s\nCommand: %s\nState: %s\nStateDescription: %s\nPPID: %s\nProcessGroup: %s", p.PID, p.Command, p.State, p.StateDesc, p.PPID, p.PGrpID)
}

// StatusProcess 查询指定pid对应的进程状态，信息来源为/proc/<pid>/stat文件
func StatusProcess(pid string) (ProcessStatus, error) {
	res := ProcessStatus{}
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

// UidGid 获取指定用户名对应的UID和GID
func UidGid(username string) (string, string, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return "", "", err
	}

	if u == nil {
		return "", "", fmt.Errorf("user %s is nil", username)
	}

	return u.Uid, u.Gid, nil
}

type MemInfo *list.List
type MemSeg struct {
	Start uint64
	End   uint64
	Usage string
}

func StatusSystemMemory() (MemInfo, error) {
	f, err := os.Open("/proc/iomem")
	if err != nil {
		return nil, err
	}

	recToSeg := func(t string) MemSeg {
		segs := strings.Split(t, " : ")
		subsegs := strings.Split(segs[0], "-")
		s, _ := strconv.ParseUint(subsegs[0], 16, 64)
		e, _ := strconv.ParseUint(subsegs[1], 16, 64)
		return MemSeg{s, e, segs[1]}
	}

	scanner := bufio.NewScanner()

	return nil, nil
}
