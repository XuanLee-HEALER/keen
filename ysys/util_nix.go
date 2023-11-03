// go:build aix darwin linux

package ysys

import (
	"io/fs"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

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

// DirSize 计算目录所有文件总大小
func DirSize(root string) (uint64, error) {
	var total uint64
	return total, filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if path == root {
			return nil
		}
		if !d.IsDir() {
			fi, xerr := d.Info()
			if xerr != nil {
				return xerr
			}
			total += uint64(fi.Size())
		}

		return nil
	})
}
