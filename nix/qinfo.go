// go:build linux amd64

package nix

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/user"
	"strconv"
	"strings"

	"gitea.fcdm.top/lixuan/keen/datastructure"
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

type MemSeg struct {
	Layer int
	Start uint64
	End   uint64
	Usage string
	Sub   []MemSeg
}

func prefixStr(pre string, seg MemSeg) string {
	strb := strings.Builder{}
	strb.WriteString(fmt.Sprintf("%sstart: 0x%09x end: 0x%09x usage: %s", pre, seg.Start, seg.End, seg.Usage))
	for _, m := range seg.Sub {
		strb.WriteString("\n")
		strb.WriteString(prefixStr(pre+"\t", m))
	}
	return strb.String()
}

func (seg MemSeg) String() string {
	return prefixStr("", seg)
}

func StatusSystemMemory() ([]*MemSeg, error) {
	f, err := os.Open("/proc/iomem")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	recToSeg := func(t string) MemSeg {
		ct := 0
		for _, r := range t {
			if r == ' ' {
				ct++
			} else {
				break
			}
		}
		ly := ct / 2
		t = t[ct:]
		segs := strings.Split(t, " : ")
		subsegs := strings.Split(segs[0], "-")
		s, _ := strconv.ParseUint(subsegs[0], 16, 64)
		e, _ := strconv.ParseUint(subsegs[1], 16, 64)
		return MemSeg{ly, s, e, segs[1], nil}
	}

	merge := func(l int, sck *datastructure.Stack) int {
		curl := sck.Peek().(*MemSeg).Layer
		for curl > l {
			tarr := make([]MemSeg, 0)
			for {
				e1 := sck.Remove().(*MemSeg)
				e2 := sck.Peek().(*MemSeg)
				if e1.Layer == e2.Layer {
					tarr = append(tarr, *e1)
					continue
				} else if e1.Layer > e2.Layer {
					tarr = append(tarr, *e1)
					for i, j := 0, len(tarr)-1; i < j; i, j = i+1, j-1 {
						tarr[i], tarr[j] = tarr[j], tarr[i]
					}
					e2.Sub = tarr
					curl = e2.Layer
					break
				}
			}
		}
		return curl
	}

	res := make([]*MemSeg, 0)
	sck := datastructure.NewStack()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ln := scanner.Text()
		seg := recToSeg(ln)
		h := sck.Peek()
		if h == nil {
			sck.Push(&seg)
		} else {
			if old := h.(*MemSeg).Layer; old < seg.Layer {
				sck.Push(&seg)
			} else if old == seg.Layer {
				if old == 0 {
					res = append(res, sck.Remove().(*MemSeg))
					sck.Push(&seg)
				} else {
					sck.Push(&seg)
				}
			} else if old > seg.Layer {
				old = merge(seg.Layer, &sck)
				if old == 0 {
					res = append(res, sck.Remove().(*MemSeg))
				}
				sck.Push(&seg)
			}
		}
	}

	merge(0, &sck)
	res = append(res, sck.Remove().(*MemSeg))

	return res, nil
}

func TotalMemory(seg []*MemSeg) uint64 {
	return seg[len(seg)-1].End
}
