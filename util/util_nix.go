//go:build linux && aix && darwin
// +build linux,aix,darwin

package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"gitea.fcdm.top/lixuan/keen"
	"gitea.fcdm.top/lixuan/keen/datastructure"
	"gitea.fcdm.top/lixuan/keen/ylog"
)

func Sync(_ string) {
	syscall.Sync()
}

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

// QuerySystemMemory 获取操作系统内存分配信息，源数据为/proc/iomem信息
func QuerySystemMemory() ([]*MemSeg, error) {
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

// ExecCmd 执行命令，返回顺序为标准输出、标准错误和执行错误对象
func ExecCmd(path string,
	args []string,
	envs map[string]string,
	dir string,
	uid string,
	gid string,
	stats []string) ([]byte, []byte, error) {
	envstrs := make([]string, 0)
	for k, v := range envs {
		envstrs = append(envstrs, k+"="+v)
	}
	iuid, err := strconv.Atoi(uid)
	if err != nil {
		return nil, nil, err
	}
	igid, err := strconv.Atoi(gid)
	if err != nil {
		return nil, nil, err
	}
	cmd := exec.Cmd{
		Path: path,
		Args: args,
		Dir:  dir,
		Env:  envstrs,
		SysProcAttr: &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: uint32(iuid),
				Gid: uint32(igid),
			},
			Setpgid: true,
		},
	}

	keen.Logger.Println(ylog.Debug, fmt.Sprintf("exec command: %s", cmd.String()))

	outp, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	oute, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}
	buf1, buf2 := make([]byte, 64*1024), make([]byte, 64*1024)
	outbs, errbs := make([]byte, 0), make([]byte, 0)
	go func() {
		for {
			n, err := outp.Read(buf1)
			if err != nil {
				return
			}
			outbs = append(outbs, buf1[:n]...)
		}
	}()
	go func() {
		for {
			n, err := oute.Read(buf2)
			if err != nil {
				return
			}
			outbs = append(outbs, buf2[:n]...)
		}
	}()
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}

	for _, stat := range stats {
		_, err := in.Write([]byte(stat + "\n"))
		if err != nil {
			return nil, nil, err
		}
	}
	err = in.Close()
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	if err := cmd.Wait(); err != nil {
		return outbs, errbs, err
	} else {
		return outbs, errbs, nil
	}
}
