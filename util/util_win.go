//go:build windows
// +build windows

package util

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"gitea.fcdm.top/lixuan/keen"
	"gitea.fcdm.top/lixuan/keen/datastructure"
	"gitea.fcdm.top/lixuan/keen/fp"
	"gitea.fcdm.top/lixuan/keen/ylog"
	"golang.org/x/sys/windows/registry"
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

func Sync(dir string) {
	err := IterDir(dir, func(s string, de fs.DirEntry) bool { return false }, func(s string, de fs.DirEntry) error {
		p := s
		wm := syscall.O_RDWR
		perm := de.Type().Perm()
		f, err := syscall.Open(p, wm, uint32(perm))
		if err != nil {
			return err
		}
		defer syscall.Close(f)
		err = syscall.Fsync(f)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		keen.Logger.Println(ylog.ERROR, fmt.Sprintf("failed to sync filesystem buffer to disk: %v", err))
	}
}

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
	keen.Logger.Println(ylog.TRACE, fmt.Sprintf("powershell retrieve object: \n%s", xcmd))
	pshell, err := exec.LookPath(POWERSHELL)
	if err != nil {
		return err
	}

	cmd := exec.Command(pshell, "-Command", xcmd)
	bs, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	keen.Logger.Println(ylog.TRACE, fmt.Sprintf("powershell retrieve object output: \n%s", string(bs)))

	err = json.Unmarshal(bs, obj)
	if err != nil {
		return err
	}

	return nil
}

// PSExec 使用powershell运行命令，返回输出内容
func PSExec(xcmd string) ([]byte, error) {
	keen.Logger.Println(ylog.TRACE, fmt.Sprintf("powershell execute command: \n%s", xcmd))
	pshell, err := exec.LookPath(POWERSHELL)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(pshell, "-Command", xcmd)
	bs, err := cmd.CombinedOutput()
	if err != nil {
		return bs, err
	}
	keen.Logger.Println(ylog.TRACE, fmt.Sprintf("powershell execute command output: \n%s", string(bs)))

	return bs, nil
}

// PSEscape 转移PowerShell中的特殊字符
func PSEscape(s string) string {
	s = strings.ReplaceAll(s, "$", "`$")
	s = strings.ReplaceAll(s, " ", "` ")
	return s
}

var ErrPSObjectNotFound error = errors.New("failed to find specific object")

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
	TotalPhysicalMemory       uint64               `json:"CsTotalPhysicalMemory"`
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

func HostComputerInfo() (ComputerInfo, error) {
	var script string
	switch currentPSVer {
	case PSv2, PSv4:
		script = `& {chcp 437 > $null; systeminfo.exe /FO 'LIST'}`
		comp := ComputerInfo{}
		bs, err := PSExec(script)
		if err != nil {
			return comp, err
		}
		sysinfo := ReadListOutput(bs)

		if len(sysinfo) <= 0 {
			return comp, ErrPSObjectNotFound
		}

		for _, info := range sysinfo {
			for k, v := range info {
				switch k {
				case "Host Name":
					comp.ComputerName = v
					comp.DNSHostName = v
				case "OS Name":
					comp.WindowsProductName = v
				case "System Locale":
					comp.Language = v
				case "Domain":
					comp.Domain = v
				case "Total Physical Memory":
					pn := v[:len(v)-3]
					pn = strings.ReplaceAll(pn, ",", "")
					pnn, err := strconv.ParseUint(pn, 10, 64)
					if err != nil {
						return comp, err
					}
					comp.TotalPhysicalMemory = pnn
					comp.TotalPhysicalMemoryStr = v
				case "Time Zone":
					comp.TimeZone = v
				}
			}
		}

		return comp, nil
	case PSv5:
		script = `& {chcp 437 > $null; Get-ComputerInfo -Property "WindowsProductName","WindowsInstallationType","CsDNSHostName","CsDomain","CsName","CsNetworkAdapters","CsNumberOfLogicalProcessors","CsNumberOfProcessors","CsTotalPhysicalMemory","CsWorkgroup","OsMuiLanguages","OsLanguage","TimeZone" | ConvertTo-Json}`
		var comp ComputerInfo
		err := PSRetrieve(script, &comp)
		return comp, err
	default:
		return ComputerInfo{}, ErrUnsupportedPSVersion
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
func ServiceByFilter(filterStr string) ([]ServiceInfo, error) {
	var script string
	switch currentPSVer {
	case PSv2:
		script = `& {chcp 437 > $null; Get-WmiObject -Class Win32_Service | Where-Object {%s} | Select-Object ProcessId,Name,State,PathName,StartName | Format-List}`
		script = fmt.Sprintf(script, filterStr)
		bs, err := PSExec(script)
		if err != nil {
			return nil, err
		}

		serviceInfos := make([]ServiceInfo, 0)
		svcInfos := ReadListOutput(bs)
		if len(svcInfos) <= 0 {
			return serviceInfos, ErrPSObjectNotFound
		} else {
			for _, info := range svcInfos {
				serviceInfo := ServiceInfo{}
				for k, v := range info {
					switch k {
					case "ProcessId":
						v, _ := strconv.Atoi(v)
						serviceInfo.ProcessID = v
					case "Name":
						serviceInfo.Name = v
					case "State":
						serviceInfo.State = v
					case "PathName":
						serviceInfo.PathName = v
					case "StartName":
						serviceInfo.StartName = v
					}
				}
				serviceInfos = append(serviceInfos, serviceInfo)
			}
		}

		return serviceInfos, nil
	case PSv4, PSv5:
		script = `& {chcp 437 > $null; Get-CimInstance -ClassName Win32_Service | Where-Object {%s} | Select-Object ProcessId,Name,State,PathName,StartName | ConvertTo-Json}`
		script = fmt.Sprintf(script, filterStr)
		var serviceInfo ServiceInfo
		err := PSRetrieve(script, &serviceInfo)
		if err != nil {
			var serviceInfos []ServiceInfo
			err = PSRetrieve(script, &serviceInfos)
			return serviceInfos, err
		} else {
			serviceInfos := make([]ServiceInfo, 0)
			serviceInfos = append(serviceInfos, serviceInfo)
			return serviceInfos, nil
		}
	default:
		return nil, ErrUnsupportedPSVersion
	}
}

// QueryProcess 根据pid查询进程信息
func QueryProcess(pid string) (ProcessInfo, error) {
	var res ProcessInfo
	procs, err := ProcessInfoByFilter("ProcessId=" + pid)
	if err != nil {
		return res, err
	}

	if len(procs) <= 0 {
		return res, fmt.Errorf("failed to find the process which process id is %s", pid)
	}

	return procs[0], nil
}

type ProcessInfo struct {
	ProcessId       int    `json:"ProcessId"`
	ParentProcessId int    `json:"ParentProcessId"`
	PID             string `json:"pid"`
	Command         string `json:"ExecutablePath"`
	State           string `json:"Status"`
	StateDesc       string `json:"state_desc"`
	PPID            string `json:"ppid"`
	PGrpID          string `json:"process_group_id"`
}

func (pi ProcessInfo) String() string {
	return fmt.Sprintf(`Process Information
process_id: %d
parent_process_id: %d
pid: %s
command: %s
state: %s
state_description: %s
ppid: %s
pgroup_id: %s
`, pi.ProcessId, pi.ParentProcessId, pi.PID, pi.Command, pi.State, pi.StateDesc, pi.PPID, pi.PGrpID)
}

// ProcessInfoByFilter 根据过滤信息获取进程信息
func ProcessInfoByFilter(filterStr string) ([]ProcessInfo, error) {
	var script string
	switch currentPSVer {
	case PSv2:
		script = `& {chcp 437 > $null; Get-WmiObject -Class Win32_Process -Filter "%s" | Select-Object ProcessId,ParentProcessId,ExecutablePath | Format-List}`
		script = fmt.Sprintf(script, filterStr)
		bs, err := PSExec(script)
		if err != nil {
			return nil, err
		}

		processInfos := make([]ProcessInfo, 0)
		procInfos := ReadListOutput(bs)
		if len(procInfos) <= 0 {
			return processInfos, ErrPSObjectNotFound
		} else {
			for _, info := range procInfos {
				processInfo := ProcessInfo{}
				for k, v := range info {
					switch k {
					case "ProcessId":
						xv, _ := strconv.Atoi(v)
						processInfo.PID = v
						processInfo.ProcessId = xv
					case "ParentProcessId":
						xv, _ := strconv.Atoi(v)
						processInfo.PPID = v
						processInfo.ParentProcessId = xv
					case "ExecutablePath":
						processInfo.Command = v
					}
				}
				processInfos = append(processInfos, processInfo)
			}
		}

		return processInfos, nil
	case PSv4, PSv5:
		script = `& {chcp 437 > $null; Get-CimInstance -ClassName Win32_Process -Filter "%s" | Select-Object ProcessId,ParentProcessId,ExecutablePath | ConvertTo-Json}`
		script = fmt.Sprintf(script, filterStr)
		var procInfo ProcessInfo
		err := PSRetrieve(script, &procInfo)
		if err != nil {
			var procInfos []ProcessInfo
			err = PSRetrieve(script, &procInfos)
			return procInfos, err
		} else {
			procInfos := make([]ProcessInfo, 0)
			procInfos = append(procInfos, procInfo)
			return procInfos, nil
		}
	default:
		return nil, ErrUnsupportedPSVersion
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

		prodInfos := ReadListOutput(bs)
		if len(prodInfos) <= 0 {
			return productVerInfo, ErrPSObjectNotFound
		}

		productVerInfo.ProductVersion = prodInfos[0]["ProductVersion"]
		return productVerInfo, nil
	case PSv4, PSv5:
		script = `& {chcp 437 > $null; Get-Process -Id %d | Select-Object ProductVersion | ConvertTo-Json}`
		var productVerInfo ProductVersionInfo
		err := PSRetrieve(fmt.Sprintf(script, pid), &productVerInfo)
		if err != nil {
			return productVerInfo, err
		}
		return productVerInfo, nil
	default:
		return ProductVersionInfo{}, ErrUnsupportedPSVersion
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
		script = `Get-WMIObject -Class Win32_Volume | Select-Object Caption,DeviceID | Format-List`
		var volumeInfo VolumeInfo
		bs, err := PSExec(script)
		if err != nil {
			return volumeInfo, err
		}

		driverToId := make(map[string]string)
		driverInfos := ReadListOutput(bs)

		for _, info := range driverInfos {
			var cap, dev string
			for k, v := range info {
				switch k {
				case "Caption":
					cap = v
				case "DeviceID":
					dev = v
				}
			}
			driverToId[cap] = dev
		}

		drivers := make([]string, 0)
		for k := range driverToId {
			if IsSubDir(k, path) {
				drivers = append(drivers, k)
			}
		}

		if len(drivers) <= 0 {
			return volumeInfo, ErrPSObjectNotFound
		}

		sort.Slice(drivers, func(i, j int) bool {
			return len(drivers[i]) > len(drivers[j])
		})

		volumeInfo.UniqueId = driverToId[drivers[0]]
		return volumeInfo, nil
	case PSv4:
		script := `& {chcp 437 > $null; Get-Volume -FilePath "%s" | Select-Object -Property ObjectId | Format-List}`
		var volumeInfo VolumeInfo
		bs, err := PSExec(fmt.Sprintf(script, path))
		if err != nil {
			return volumeInfo, err
		}

		volIds := ReadListOutput(bs)
		if len(volIds) <= 0 {
			return volumeInfo, ErrPSObjectNotFound
		}

		volumeInfo.UniqueId = volIds[0]["ObjectId"]

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
		return VolumeInfo{}, ErrUnsupportedPSVersion
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
				line = strings.TrimFunc(line, TrimSpace)
				line = strings.ReplaceAll(line, "\t", " ")
				segs := strings.Split(line, " ")
				segs = fp.Map[string, string](segs, func(e string) string { return strings.TrimFunc(e, TrimSpace) })
				segs = fp.Filter[string](segs, func(e string) bool { return e != "" })
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

		if len(tcpInfos) <= 0 {
			return tcpInfos, ErrPSObjectNotFound
		}

		return tcpInfos, nil
	case PSv4, PSv5:
		script = `& {chcp 437 > $null; Get-NetTCPConnection -OwningProcess %d | Select-Object LocalAddress,LocalPort,State | ConvertTo-Json}`
		var tcpInfos []TcpInfo
		err := PSRetrieve(fmt.Sprintf(script, pid), &tcpInfos)
		if err != nil {
			return nil, err
		}
		return tcpInfos, nil
	default:
		return nil, ErrUnsupportedPSVersion
	}
}

// StringRegVal 获取注册表下指定key的注册表项的值
//
// regK 注册表已有key | path key下的路径 | access 访问权限
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
		script = `& {chcp 437 > $null; Get-PSDrive | Select-Object -Property Name | Format-List}`
		bs, err := PSExec(script)
		if err != nil {
			return nil, err
		}

		driverInfos := make([]DriverInfo, 0)
		driverList := ReadListOutput(bs)
		if len(driverList) <= 0 {
			return driverInfos, ErrPSObjectNotFound
		} else {
			for _, driver := range driverList {
				driverInfo := DriverInfo{}
				for k, v := range driver {
					switch k {
					case "Name":
						driverInfo.Name = v
					}
				}
				driverInfos = append(driverInfos, driverInfo)
			}
		}

		return driverInfos, nil
	case PSv4, PSv5:
		script = `& {chcp 437 > $null; Get-PSDrive | Select-Object -Property Name | ConvertTo-Json}`
		var driverInfo DriverInfo
		err := PSRetrieve(script, &driverInfo)
		if err != nil {
			var driverInfos []DriverInfo
			err = PSRetrieve(script, &driverInfos)
			return driverInfos, err
		} else {
			driverInfos := make([]DriverInfo, 0)
			driverInfos = append(driverInfos, driverInfo)
			return driverInfos, nil
		}
	default:
		return nil, ErrUnsupportedPSVersion
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
