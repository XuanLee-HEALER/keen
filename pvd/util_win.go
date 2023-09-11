// go:build windows

package pvd

import (
	"fmt"

	"gitea.fcdm.top/lixuan/keen/win"
	"golang.org/x/sys/windows/registry"
)

// GrantFullAccess 给特定路径赋予指定账户的全部访问权限
func GrantFullAccess(path, account string) error {
	const GRAND_SCRIPT = `& {
	chcp 437 > $null; 
	$NewAcl = Get-Acl -Path "%s";
	$identity = "%s";
	$fileSystemRights = "FullControl";
	$type = "Allow";
	$fileSystemAccessRuleArgumentList = $identity, $fileSystemRights, $type;
	$fileSystemAccessRule = New-Object -TypeName System.Security.AccessControl.FileSystemAccessRule -ArgumentList $fileSystemAccessRuleArgumentList;
	$NewAcl.SetAccessRule($fileSystemAccessRule);
	Set-Acl -Path "%s" -AclObject $NewAcl};`
	_, err := win.PSExec(fmt.Sprintf(GRAND_SCRIPT, win.PSEscape(path), account, win.PSEscape(path)))
	if err != nil {
		return err
	}
	return nil
}

type TcpInfo struct {
	LocalAddress string `json:"LocalAddress"`
	LocalPort    int    `json:"LocalPort"`
	State        int    `json:"State"`
}

// Listening 根据PID查询监听的TCP/IP端口号
func Listening(pid uint, tcpInfos *[]TcpInfo) error {
	const LISTEN_SCRIPT = `& {chcp 437 > $null; Get-NetTCPConnection -OwningProcess %d | Select-Object LocalAddress,LocalPort,State | ConvertTo-Json}`
	err := win.PSRetrieve(fmt.Sprintf(LISTEN_SCRIPT, pid), tcpInfos)
	if err != nil {
		return err
	}

	return nil
}

// StringRegVal 获取注册表下指定key的注册表项的值
func StringRegVal(regK registry.Key, path string, access uint32, ks ...string) ([]string, error) {
	res := make([]string, 0)

	key, err := registry.OpenKey(regK, path, access)
	if err != nil {
		return nil, fmt.Errorf("failed to open registry key [%s]: %v", path, err)
	}
	defer key.Close()

	for _, k := range ks {
		v, _, err := key.GetStringValue(k)
		if err != nil {
			return nil, fmt.Errorf("failed to find specific key %s: %v", k, err)
		}
		res = append(res, v)
	}

	return res, nil
}
