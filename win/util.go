// go:build windows

package win

import (
	"fmt"
	"strconv"
	"strings"

	"gitea.fcdm.top/lixuan/keen/datastructure"
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
	_, err := PSExec(fmt.Sprintf(GRAND_SCRIPT, PSEscape(path), account, PSEscape(path)))
	if err != nil {
		return err
	}
	return nil
}

type ComputerInfo struct {
	WindowsProductName        string               `json:"WindowsProductName"`
	WindowsInstallationType   string               `json:"WindowsInstallationType"`
	DNSHostName               string               `json:"CsDNSHostName"`
	Domain                    string               `json:"CsDomain"`
	ComputerName              string               `json:"CsName"`
	NetworkAdapters           []NetworkAdapterInfo `json:"CsNetworkAdapters"`
	NumberOfLogicalProcessors int                  `json:"CsNumberOfLogicalProcessors"`
	NumberOfProcessors        int                  `json:"CsNumberOfProcessors"`
	TotalPhysicalMemory       int64                `json:"CsTotalPhysicalMemory"`
	Workgroup                 string               `json:"CsWorkgroup"`
	MultiLanguage             []string             `json:"OsMuiLanguages"`
	Language                  string               `json:"OsLanguage"`
	TimeZone                  string               `json:"TimeZone"`
}

type NetworkAdapterInfo struct {
	Description      string `json:"Description"`
	ConnectionID     string `json:"ConnectionID"`
	DHCPEnabled      bool   `json:"DHCPEnabled"`
	DHCPServer       string `json:"DHCPServer"`
	ConnectionStatus int    `json:"ConnectionStatus"`
	IPAddresses      string `json:"IPAddresses"`
}

func HostComputerInfo(obj *ComputerInfo) error {
	COMPUTER_INFO_SCRIPT := `& {chcp 437 > $null; Get-ComputerInfo -Property "WindowsProductName","WindowsInstallationType","CsDNSHostName","CsDomain","CsName","CsNetworkAdapters","CsNumberOfLogicalProcessors","CsNumberOfProcessors","CsTotalPhysicalMemory","CsWorkgroup","OsMuiLanguages","OsLanguage","TimeZone" | ConvertTo-Json}`
	return PSRetrieve(COMPUTER_INFO_SCRIPT, &obj)
}

type ServiceInfo struct {
	ProcessID int    `json:"ProcessId"`
	Name      string `json:"Name"`
	State     string `json:"State"`
	PathName  string `json:"PathName"`
	StartName string `json:"StartName"`
}

// ServicesByFilter 根据过滤条件获取服务信息，返回对象数组
func ServicesByFilter(cond string, obj *[]ServiceInfo) error {
	SERVICES_SCRIPT := `& {chcp 437 > $null; Get-CimInstance -ClassName Win32_Service | Where-Object {%s} | Select-Object ProcessId,Name,State,PathName,StartName | ConvertTo-Json}`
	return PSRetrieve(fmt.Sprintf(SERVICES_SCRIPT, cond), &obj)
}

// ServiceByFilter 根据过滤条件获取服务信息，返回单个对象
func ServiceByFilter(cond string, obj *ServiceInfo) error {
	SERVICES_SCRIPT := `& {chcp 437 > $null; Get-CimInstance -ClassName Win32_Service | Where-Object {%s} | Select-Object ProcessId,Name,State,PathName,StartName | ConvertTo-Json}`
	return PSRetrieve(fmt.Sprintf(SERVICES_SCRIPT, cond), &obj)
}

type ProcessInfo struct {
	ProcessId       int `json:"ProcessId"`
	ParentProcessId int `json:"ParentProcessId"`
}

// ProcessInfoByName 根据进程名称获取进程信息
func ProcessInfoByName(pname string, obj *ProcessInfo) error {
	PROCESS_SCRIPT := `& {chcp 437 > $null; Get-CimInstance -ClassName Win32_Process -Filter "Name = '%s'" | Select-Object ProcessId,ParentProcessId | ConvertTo-Json}`
	return PSRetrieve(fmt.Sprintf(PROCESS_SCRIPT, pname), &obj)
}

// ProcessInfoById 根据进程ID获取进程信息
func ProcessInfoById(pid uint, obj *ProcessInfo) error {
	PROCESS_SCRIPT := `& {chcp 437 > $null; Get-CimInstance -ClassName Win32_Process -Filter "ProcessId = %d" | Select-Object ProcessId,ParentProcessId | ConvertTo-Json}`
	return PSRetrieve(fmt.Sprintf(PROCESS_SCRIPT, pid), &obj)
}

type ProductVersionInfo struct {
	ProductVersion string `json:"ProductVersion"`
}

// SQLServerProductId 获取SQL Server进程的产品ID
func SQLServerProductId(pid uint, obj *ProductVersionInfo) error {
	PRODUCT_VER_SCRIPT := `& {chcp 437 > $null; Get-Process -Id %d | Select-Object ProductVersion | ConvertTo-Json}`
	return PSRetrieve(fmt.Sprintf(PRODUCT_VER_SCRIPT, pid), &obj)
}

type VolumeInfo struct {
	UniqueId string `json:"UniqueId"`
}

// VolumeInfoByPath 获取指定路径的卷ID
func VolumeInfoByPath(path string, obj *VolumeInfo) error {
	VOL_INFO_SCRIPT := `& {chcp 437 > $null; Get-Volume -FilePath "%s" | Select-Object -Property UniqueId | ConvertTo-Json}`
	return PSRetrieve(fmt.Sprintf(VOL_INFO_SCRIPT, path), &obj)
}

type DriverInfo struct {
	Name string `json:"Name"`
}

// DriverInfos 获取主机上的盘符信息
func DriverInfos(obj *[]DriverInfo) error {
	DRIVER_INFO_SCRIPT := `& {chcp 437 > $null; Get-PSDrive | Select-Object -Property Name | ConvertTo-Json}`
	return PSRetrieve(DRIVER_INFO_SCRIPT, &obj)
}

type TcpInfo struct {
	LocalAddress string `json:"LocalAddress"`
	LocalPort    int    `json:"LocalPort"`
	State        int    `json:"State"`
}

// Listening 根据PID查询监听的TCP/IP端口号
func Listening(pid uint, tcpInfos *[]TcpInfo) error {
	const LISTEN_SCRIPT = `& {chcp 437 > $null; Get-NetTCPConnection -OwningProcess %d | Select-Object LocalAddress,LocalPort,State | ConvertTo-Json}`
	err := PSRetrieve(fmt.Sprintf(LISTEN_SCRIPT, pid), tcpInfos)
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

// StartService 启动一个服务
func StartService(sname string) ([]byte, error) {
	START_SVC_SCRIPT := `& {chcp 437 > $null; Start-Service -Name "%s"}`
	return PSExec(fmt.Sprintf(START_SVC_SCRIPT, sname))
}

// StopService 结束一个服务
func StopService(sname string) ([]byte, error) {
	STOP_SVC_SCRIPT := `& {chcp 437 > $null; Stop-Service -Name "%s"}`
	return PSExec(fmt.Sprintf(STOP_SVC_SCRIPT, sname))
}
