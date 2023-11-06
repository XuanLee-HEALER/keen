//go:build windows
// +build windows

package util_test

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"

	"gitea.fcdm.top/lixuan/keen/datastructure"
	"gitea.fcdm.top/lixuan/keen/util"
	"golang.org/x/sys/windows/registry"
)

func TestPSVersionTable(t *testing.T) {
	v, err := util.PSVersionTable()
	if err != nil {
		t.Fatalf("failed to fetch powershell version: %v", err)
	}
	t.Log("version:", v)
}

func print(obj any, t *testing.T) {
	bs, _ := json.MarshalIndent(obj, "", "  ")
	t.Log("\n", string(bs))
}

func TestHostComputerInfo(t *testing.T) {
	err := util.SetupPowerShellVersion()
	if err != nil {
		t.Fatal(err)
	}

	computerInfo, err := util.HostComputerInfo()
	if err != nil {
		t.Fatalf("failed to retrieve computer information: %v", err)
	}

	print(computerInfo, t)
}

func TestServiceByFilter(t *testing.T) {
	err := util.SetupPowerShellVersion()
	if err != nil {
		t.Fatal(err)
	}

	cond0 := `$_.Name -eq 'AJRouter' -and $_.State -ne 'Stopped'`
	cond1 := `$_.Name -eq 'MSSQLSERVER' -and $_.State -ne 'Stopped'`
	cond2 := `$_.Name -like 'A*' -and $_.State -ne 'Stopped'`
	conds := []string{cond0, cond1, cond2}

	for _, cond := range conds {
		t.Log("current condition: ", cond)
		services, err := util.ServiceByFilter(cond)
		if err != nil {
			t.Errorf("failed to retrieve service information: %v", err)
		}

		if len(services) > 0 {
			print(services, t)
		}
	}
}

func TestProcessInfoByFilter(t *testing.T) {
	err := util.SetupPowerShellVersion()
	if err != nil {
		t.Fatal(err)
	}

	cond0 := `Name = 'sqlservr.exe'`
	cond1 := `ProcessId = 1460`
	conds := []string{cond0, cond1}

	for _, cond := range conds {
		t.Log("current condition: ", cond)
		processes, err := util.ProcessInfoByFilter(cond)
		if err != nil {
			t.Errorf("failed to retrieve service information: %v", err)
		}

		if len(processes) > 0 {
			print(processes, t)
		}
	}
}

func TestSQLServerProductId(t *testing.T) {
	err := util.SetupPowerShellVersion()
	if err != nil {
		t.Fatal(err)
	}

	prod, err := util.SQLServerProductId(1460)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("product version", prod)
}

func TestVolumeInfoByPath(t *testing.T) {
	err := util.SetupPowerShellVersion()
	if err != nil {
		t.Fatal(err)
	}

	p1 := "F:\\go_projects\\keen\\ylog\\ylog.go"
	p2 := "C:\\Program Files\\Microsoft SQL Server\\MSSQL10_50.MSSQLSERVER\\MSSQL\\DATA\\TEST_DB.mdf"
	ps := []string{p1, p2}

	for _, p := range ps {
		vol, err := util.VolumeInfoByPath(p)
		if err != nil {
			t.Errorf("failed to fetch volume information: %v", err)
		}

		print(vol, t)
	}
}

func TestListening(t *testing.T) {
	err := util.SetupPowerShellVersion()
	if err != nil {
		t.Fatal(err)
	}

	var p1, p2 uint = 1460, 16440
	ps := []uint{p1, p2}

	for _, p := range ps {
		info, err := util.Listening(p)
		if err != nil {
			t.Errorf("failed to fetch tcp information: %v", err)
		}

		print(info, t)
	}
}

func TestDriveLetters(t *testing.T) {
	err := util.SetupPowerShellVersion()
	if err != nil {
		t.Fatal(err)
	}

	drivers, err := util.DriveLetters()
	if err != nil {
		t.Fatal(err)
	}

	print(drivers, t)

	driverSet := make(datastructure.Set[byte])
	for _, driver := range drivers {
		if len(driver.Name) <= 1 {
			driverSet.Add(driver.Name[0])
		}
	}

	avail := util.AvailableLetter(driverSet)

	t.Log("avail", string(avail))
}

func TestStringRegVal(t *testing.T) {
	const (
		NBU_PATH    = "SOFTWARE\\Veritas\\NetBackup\\CurrentVersion"
		NBU_CONFIG  = "SOFTWARE\\Veritas\\NetBackup\\CurrentVersion\\Config"
		NBU_INIPATH = "INI Directory"
		NBU_MASTER  = "EMMSERVER"
	)
	res, err := util.StringRegVal(registry.LOCAL_MACHINE, NBU_CONFIG, registry.QUERY_VALUE, NBU_MASTER)
	if err != nil {
		t.Errorf("failed to get master server: %v", err)
	}
	t.Logf("the hostname of master server: %v", res[0])

	res, err = util.StringRegVal(registry.LOCAL_MACHINE, NBU_PATH, registry.QUERY_VALUE, NBU_INIPATH)
	if err != nil {
		t.Errorf("failed to get nbu home: %v", err)
	}
	t.Logf("the hostname of nbu home: %v", res[0])

	cmd := exec.Command(filepath.Join(res[0], "bin", "bplist"))
	bs, err := cmd.CombinedOutput()
	if err != nil {
		t.Error(err)
	}
	t.Log(string(bs))
}
