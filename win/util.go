// go:build windows

package win

import (
	"fmt"
	"strconv"
	"strings"

	"gitea.fcdm.top/lixuan/keen/datastructure"
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
	_, err := PSExec(fmt.Sprintf(GRAND_SCRIPT, PSEscape(path), account, PSEscape(path)))
	if err != nil {
		return err
	}
	return nil
}

// Chcp 获取默认码表
func Chcp() (uint, error) {
	const CHCP_SCRIPT = `& {chcp}`
	bs, err := PSExec(CHCP_SCRIPT)
	if err != nil {
		return 0, err
	}

	segs := strings.Split(string(bs), ":")
	cp := strings.TrimSpace(segs[1])
	ui, err := strconv.ParseUint(cp, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint(ui), err
}

// SetChcp 设置码表
func SetCodePage(codePage uint) error {
	const CHCP_SCRIPT = `& {chcp %s}`
	_, err := PSExec(fmt.Sprintf(CHCP_SCRIPT, strconv.FormatUint(uint64(codePage), 10)))
	if err != nil {
		return err
	}
	return nil
}

// DriveLetters 获取当前已被占用的盘符
func DriveLetters(rcv any) error {
	const DRIVER_LETTERS = `& {chcp 437 > $null; Get-PSDrive | Select-Object -Property Name | ConvertTo-Json}`
	err := PSRetrieve(DRIVER_LETTERS, rcv)
	if err != nil {
		return err
	}

	return nil
}

// AvailableLetter 查找可用盘符
func AvailableLetter(usedDrivers datastructure.Set[byte]) byte {
	for l := 'D'; l <= 'Z'; l++ {
		if _, ok := usedDrivers[byte(l)]; !ok {
			return byte(l)
		}
	}

	return 0
}
