//go:build windows

package win

import (
	"encoding/json"
	"os/exec"
	"strings"
)

const (
	POWERSHELL = "powershell.exe"
)

// PSRetrieve 使用powershell内置的cmdlet，结果为json字符串，直接unmarshal到传入的对象指针中
func PSRetrieve(xcmd string, obj any) error {
	pshell, err := exec.LookPath(POWERSHELL)
	if err != nil {
		return err
	}

	cmd := exec.Command(pshell, "-Command", xcmd)
	bs, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	err = json.Unmarshal(bs, obj)
	if err != nil {
		return err
	}

	return nil
}

// Exec 使用powershell运行命令，返回输出内容
func PSExec(xcmd string) ([]byte, error) {
	pshell, err := exec.LookPath(POWERSHELL)
	if err != nil {
		return nil, err
	}

	c := exec.Command(pshell, "-Command", xcmd)
	bs, err := c.CombinedOutput()
	if err != nil {
		return bs, err
	}

	return bs, nil
}

// PSEscape 转移PowerShell中的特殊字符
func PSEscape(s string) string {
	s = strings.ReplaceAll(s, "$", "`$")
	s = strings.ReplaceAll(s, " ", "` ")
	return s
}
