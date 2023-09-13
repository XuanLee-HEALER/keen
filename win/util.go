// go:build windows

package win

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gitea.fcdm.top/lixuan/keen/datastructure"
	"gitea.fcdm.top/lixuan/keen/util"
	"golang.org/x/sys/windows/registry"
)

var trimSpace = func(r rune) bool { return r == ' ' || r == '\t' }

// GrantFullAccess 给特定路径赋予指定账户的全部访问权限
func GrantFullAccess(path, account string) error {
	var grandScript = `& {
	chcp 437 > $null; 
	$NewAcl = Get-Acl -Path "%s";
	$identity = "%s";
	$fileSystemRights = "FullControl";
	$type = "Allow";
	$fileSystemAccessRuleArgumentList = $identity, $fileSystemRights, $type;
	$fileSystemAccessRule = New-Object -TypeName System.Security.AccessControl.FileSystemAccessRule -ArgumentList $fileSystemAccessRuleArgumentList;
	$NewAcl.SetAccessRule($fileSystemAccessRule);
	Set-Acl -Path "%s" -AclObject $NewAcl};`
	_, err := PSExec(fmt.Sprintf(grandScript, PSEscape(path), account, PSEscape(path)))
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
	TotalPhysicalMemoryStr    string               `json:"CsTotalPhysicalMemoryStr"`
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
	var script string
	switch currentPSVer {
	case PSv2:
		script = `& {chcp 437 > $null; systeminfo.exe /FO 'csv'}`
		bs, err := PSExec(script)
		if err != nil {
			return err
		}
		sysinfo := util.ReadCSVOutput(bs)
		for i, h := range sysinfo[0] {
			switch h {
			case "Host Name":
				obj.ComputerName = sysinfo[1][i]
				obj.DNSHostName = sysinfo[1][i]
			case "OS Name":
				obj.WindowsProductName = sysinfo[1][i]
			case "System Locale":
				obj.Language = sysinfo[1][i]
			case "Domain":
				obj.Domain = sysinfo[1][i]
			case "Total Physical Memory":
				obj.TotalPhysicalMemoryStr = sysinfo[1][i]
			case "Time Zone":
				obj.TimeZone = sysinfo[1][i]
			}
		}
		return nil
	case PSv5:
		script = `& {chcp 437 > $null; Get-ComputerInfo -Property "WindowsProductName","WindowsInstallationType","CsDNSHostName","CsDomain","CsName","CsNetworkAdapters","CsNumberOfLogicalProcessors","CsNumberOfProcessors","CsTotalPhysicalMemory","CsWorkgroup","OsMuiLanguages","OsLanguage","TimeZone" | ConvertTo-Json}`
		return PSRetrieve(script, &obj)
	default:
		return UnsupportedPSVersionErr
	}
}

type ServiceInfo struct {
	ProcessID int    `json:"ProcessId"`
	Name      string `json:"Name"`
	State     string `json:"State"`
	PathName  string `json:"PathName"`
	StartName string `json:"StartName"`
}

// ServiceByFilter 根据过滤条件获取服务信息，返回单个对象
func ServiceByFilter(cond string) ([]ServiceInfo, error) {
	var script string
	switch currentPSVer {
	case PSv2:
		script = `& {chcp 437 > $null; Get-WmiObject -Class Win32_Service | Where-Object {%s} | Select-Object ProcessId,Name,State,PathName,StartName | ConvertTo-Csv}`
		bs, err := PSExec(script)
		if err != nil {
			return nil, err
		}

		serviceInfos := make([]ServiceInfo, 0)
		csv := util.ReadCSVOutput(bs)
		if len(csv) <= 0 {
			return serviceInfos, nil
		} else {
			for i := 1; i < len(csv); i++ {
				serviceInfo := ServiceInfo{}
				for j, c := range csv[i] {
					switch csv[0][j] {
					case "ProcessId":
						v, _ := strconv.Atoi(c)
						serviceInfo.ProcessID = v
					case "Name":
						serviceInfo.Name = c
					case "State":
						serviceInfo.State = c
					case "PathName":
						serviceInfo.PathName = c
					case "StartName":
						serviceInfo.StartName = c
					}
				}
				serviceInfos = append(serviceInfos, serviceInfo)
			}
		}

		return serviceInfos, nil
	case PSv5:
		script = `& {chcp 437 > $null; Get-CimInstance -ClassName Win32_Service | Where-Object {%s} | Select-Object ProcessId,Name,State,PathName,StartName | ConvertTo-Json}`
		var serviceInfo ServiceInfo
		err := PSRetrieve(script, &serviceInfo)
		if err != nil {
			var driverInfos []DriverInfo
			err = PSRetrieve(script, &driverInfos)
			return nil, err
		} else {
			serviceInfos := make([]ServiceInfo, 0)
			serviceInfos = append(serviceInfos, serviceInfo)
			return serviceInfos, nil
		}
	default:
		return nil, UnsupportedPSVersionErr
	}
}

type ProcessInfo struct {
	ProcessId       int `json:"ProcessId"`
	ParentProcessId int `json:"ParentProcessId"`
}

// ProcessInfoByFilter 根据过滤信息获取进程信息
func ProcessInfoByFilter(filterStr string) ([]ProcessInfo, error) {
	var script string
	switch currentPSVer {
	case PSv2:
		script = `& {chcp 437 > $null; Get-WmiObject -Class Win32_Process -Filter "%s" | Select-Object ProcessId,ParentProcessId | ConvertTo-Json}`
		bs, err := PSExec(script)
		if err != nil {
			return nil, err
		}

		procInfos := make([]ProcessInfo, 0)
		csv := util.ReadCSVOutput(bs)
		if len(csv) <= 0 {
			return procInfos, nil
		} else {
			for i := 1; i < len(csv); i++ {
				procInfo := ProcessInfo{}
				for j, c := range csv[i] {
					switch csv[0][j] {
					case "ProcessId":
						v, _ := strconv.Atoi(c)
						procInfo.ProcessId = v
					case "ParentProcessId":
						v, _ := strconv.Atoi(c)
						procInfo.ParentProcessId = v
					}
				}
				procInfos = append(procInfos, procInfo)
			}
		}

		return procInfos, nil
	case PSv5:
		script = `& {chcp 437 > $null; Get-CimInstance -ClassName Win32_Process -Filter "%s" | Select-Object ProcessId,ParentProcessId | ConvertTo-Json}`
		var procInfo ProcessInfo
		err := PSRetrieve(script, &procInfo)
		if err != nil {
			var procInfos []ProcessInfo
			err = PSRetrieve(script, &procInfos)
			return nil, err
		} else {
			procInfos := make([]ProcessInfo, 0)
			procInfos = append(procInfos, procInfo)
			return procInfos, nil
		}
	default:
		return nil, UnsupportedPSVersionErr
	}
}

type ProductVersionInfo struct {
	ProductVersion string `json:"ProductVersion"`
}

// SQLServerProductId 获取SQL Server进程的产品ID
func SQLServerProductId(pid uint) (ProductVersionInfo, error) {
	var script string
	switch currentPSVer {
	case PSv2:
		script = `& {chcp 437 > $null; Get-Process -Id %d | Select-Object ProductVersion | Format-List}`
		var productVerInfo ProductVersionInfo
		bs, err := PSExec(fmt.Sprintf(script, pid))
		if err != nil {
			return productVerInfo, err
		}

		scanner := bufio.NewScanner(bytes.NewReader(bs))
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimFunc(line, trimSpace)
			if line != "" {
				segs := strings.Split(line, ":")
				productVerInfo.ProductVersion = strings.TrimFunc(segs[1], trimSpace)
			}
		}
		return productVerInfo, nil
	case PSv5:
		script = `& {chcp 437 > $null; Get-Process -Id %d | Select-Object ProductVersion | ConvertTo-Json}`
		var productVerInfo ProductVersionInfo
		err := PSRetrieve(fmt.Sprintf(script, pid), &productVerInfo)
		if err != nil {
			return productVerInfo, err
		}
		return productVerInfo, nil
	default:
		return ProductVersionInfo{}, UnsupportedPSVersionErr
	}
}

type VolumeInfo struct {
	UniqueId string `json:"UniqueId"`
}

// VolumeInfoByPath 获取指定路径的卷ID
func VolumeInfoByPath(path string) (VolumeInfo, error) {
	var script string
	switch currentPSVer {
	case PSv2:
		script = `Get-WMIObject -Class Win32_Volume | Select-Object Caption,DeviceID | ConvertTo-Csv`
		var volumeInfo VolumeInfo
		bs, err := PSExec(script)
		if err != nil {
			return volumeInfo, err
		}

		driverToId := make(map[string]string)
		csv := util.ReadCSVOutput(bs)
		for j := 1; j < len(csv); j++ {
			driverToId[csv[j][0]] = csv[j][1]
		}

		drivers := make([]string, 0)
		for k := range driverToId {
			if util.IsSubDir(k, path) {
				drivers = append(drivers, k)
			}
		}

		if len(drivers) <= 0 {
			return volumeInfo, fmt.Errorf("no volume information of path [%s] found", path)
		}

		sort.Slice(drivers, func(i, j int) bool {
			return len(drivers[i]) > len(drivers[j])
		})

		volumeInfo.UniqueId = driverToId[drivers[0]]
		return volumeInfo, nil
	case PSv5:
		script := `& {chcp 437 > $null; Get-Volume -FilePath "%s" | Select-Object -Property UniqueId | ConvertTo-Json}`
		var volumeInfo VolumeInfo
		err := PSRetrieve(fmt.Sprintf(script, path), &volumeInfo)
		if err != nil {
			return volumeInfo, err
		}

		return volumeInfo, nil
	default:
		return VolumeInfo{}, UnsupportedPSVersionErr
	}
}

type TcpInfo struct {
	LocalAddress string `json:"LocalAddress"`
	LocalPort    int    `json:"LocalPort"`
	State        int    `json:"State"`
}

// Listening 根据PID查询监听的TCP/IP端口号
func Listening(pid uint) ([]TcpInfo, error) {
	var script string
	switch currentPSVer {
	case PSv2:
		script = `& {chcp 437 > $null; NETSTAT.EXE -ano | findstr.exe "%d"}`
		tcpInfos := make([]TcpInfo, 0)

		bs, err := PSExec(fmt.Sprintf(script, pid))
		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(bytes.NewReader(bs))
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				line = strings.TrimFunc(line, trimSpace)
				line = strings.ReplaceAll(line, "\t", " ")
				segs := strings.Split(line, " ")
				if segs[3] == "LISTENING" {
					tcpInfo := TcpInfo{}
					idx := strings.LastIndex(segs[1], ":")
					tcpInfo.LocalAddress = segs[1][:idx]
					xport := segs[1][idx+1:]
					port, _ := strconv.Atoi(xport)
					tcpInfo.LocalPort = port
					tcpInfo.State = 2
					tcpInfos = append(tcpInfos, tcpInfo)
				}
			}
		}

		return tcpInfos, nil
	case PSv5:
		script = `& {chcp 437 > $null; Get-NetTCPConnection -OwningProcess %d | Select-Object LocalAddress,LocalPort,State | ConvertTo-Json}`
		var tcpInfos []TcpInfo
		err := PSRetrieve(fmt.Sprintf(script, pid), &tcpInfos)
		if err != nil {
			return nil, err
		}
		return tcpInfos, nil
	default:
		return nil, UnsupportedPSVersionErr
	}
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

type DriverInfo struct {
	Name string `json:"Name"`
}

// DriveLetters 获取当前已被占用的盘符
func DriveLetters() ([]DriverInfo, error) {
	var script string
	switch currentPSVer {
	case PSv2:
		script = `& {chcp 437 > $null; Get-PSDrive | Select-Object -Property Name | ConvertTo-Csv}`
		bs, err := PSExec(script)
		if err != nil {
			return nil, err
		}

		driverInfos := make([]DriverInfo, 0)
		csv := util.ReadCSVOutput(bs)
		if len(csv) <= 0 {
			return driverInfos, nil
		} else {
			for i := 1; i < len(csv); i++ {
				driverInfo := DriverInfo{}
				for j, c := range csv[i] {
					switch csv[0][j] {
					case "Name":
						driverInfo.Name = c
					}
				}
				driverInfos = append(driverInfos, driverInfo)
			}
		}

		return driverInfos, nil
	case PSv5:
		script = `& {chcp 437 > $null; Get-PSDrive | Select-Object -Property Name | ConvertTo-Json}`
		var driverInfo DriverInfo
		err := PSRetrieve(script, &driverInfo)
		if err != nil {
			var driverInfos []DriverInfo
			err = PSRetrieve(script, &driverInfos)
			return nil, err
		} else {
			driverInfos := make([]DriverInfo, 0)
			driverInfos = append(driverInfos, driverInfo)
			return driverInfos, nil
		}
	default:
		return nil, UnsupportedPSVersionErr
	}
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
