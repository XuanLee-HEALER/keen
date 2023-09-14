//go:build windows

package win

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

type PSVersion uint

const (
	PSv1 PSVersion = 1
	PSv2 PSVersion = 2
	PSv3 PSVersion = 3
	PSv4 PSVersion = 4
	PSv5 PSVersion = 5
	PSv6 PSVersion = 6
	PSv7 PSVersion = 7
)

const (
	POWERSHELL = "powershell.exe"
)

var currentPSVer PSVersion
var ErrUnsupportedPSVersion error = errors.New("unsupported powershell version")

func SetupPowerShellVersion() error {
	v, err := PSVersionTable()
	if err != nil {
		return err
	}
	currentPSVer = v
	return nil
}

func PSVersionTable() (PSVersion, error) {
	PSVERSION := `foreach ($k in $PSVersionTable.Keys) {
	if ($k -eq 'PSVersion') {
		$PSVersionTable[$k].ToString();
		break;
	}
}`
	bs, err := PSExec(PSVERSION)
	if err != nil {
		return PSv1, err
	}

	mv := string(bs)[0]
	switch mv {
	case '1':
		return PSv1, nil
	case '2':
		return PSv2, nil
	case '3':
		return PSv3, nil
	case '4':
		return PSv4, nil
	case '5':
		return PSv5, nil
	case '6':
		return PSv6, nil
	case '7':
		return PSv7, nil
	default:
		return PSv1, errors.New("illegal powershell version")
	}
}

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
	println("command=>", cmd.String())
	println("output=>", string(bs))

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

	cmd := exec.Command(pshell, "-Command", xcmd)
	bs, err := cmd.CombinedOutput()
	if err != nil {
		return bs, err
	}
	println("command=>", cmd.String())
	println("output=>", string(bs))

	return bs, nil
}

// PSEscape 转移PowerShell中的特殊字符
func PSEscape(s string) string {
	s = strings.ReplaceAll(s, "$", "`$")
	s = strings.ReplaceAll(s, " ", "` ")
	return s
}
