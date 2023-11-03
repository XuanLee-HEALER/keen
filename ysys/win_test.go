//go:build windows
// +build windows

package ysys_test

import (
	"encoding/json"
	"testing"
	"time"

	"gitea.fcdm.top/lixuan/keen/datastructure"
	"gitea.fcdm.top/lixuan/keen/ylog"
	"gitea.fcdm.top/lixuan/keen/ysys"
)

func initI() {
	logger := ylog.YLogger{
		ConsoleLevel:    ylog.Trace,
		ConsoleColorful: true,
		FileLog:         true,
		FileLogDir:      "C:\\tlog",
		FileClean:       3 * time.Second,
	}
	ysys.SetupLogger(logger.InitLogger())
}

func TestPSVersionTable(t *testing.T) {
	v, err := ysys.PSVersionTable()
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
	err := ysys.SetupPowerShellVersion()
	if err != nil {
		t.Fatal(err)
	}

	computerInfo, err := ysys.HostComputerInfo()
	if err != nil {
		t.Fatalf("failed to retrieve computer information: %v", err)
	}

	print(computerInfo, t)
}

func TestServiceByFilter(t *testing.T) {
	err := ysys.SetupPowerShellVersion()
	if err != nil {
		t.Fatal(err)
	}

	cond0 := `$_.Name -eq 'AJRouter' -and $_.State -ne 'Stopped'`
	cond1 := `$_.Name -eq 'MSSQLSERVER' -and $_.State -ne 'Stopped'`
	cond2 := `$_.Name -like 'A*' -and $_.State -ne 'Stopped'`
	conds := []string{cond0, cond1, cond2}

	for _, cond := range conds {
		t.Log("current condition: ", cond)
		services, err := ysys.ServiceByFilter(cond)
		if err != nil {
			t.Errorf("failed to retrieve service information: %v", err)
		}

		if len(services) > 0 {
			print(services, t)
		}
	}
}

func TestProcessInfoByFilter(t *testing.T) {
	err := ysys.SetupPowerShellVersion()
	if err != nil {
		t.Fatal(err)
	}

	cond0 := `Name = 'sqlservr.exe'`
	cond1 := `ProcessId = 1460`
	conds := []string{cond0, cond1}

	for _, cond := range conds {
		t.Log("current condition: ", cond)
		processes, err := ysys.ProcessInfoByFilter(cond)
		if err != nil {
			t.Errorf("failed to retrieve service information: %v", err)
		}

		if len(processes) > 0 {
			print(processes, t)
		}
	}
}

func TestSQLServerProductId(t *testing.T) {
	err := ysys.SetupPowerShellVersion()
	if err != nil {
		t.Fatal(err)
	}

	prod, err := ysys.SQLServerProductId(1460)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("product version", prod)
}

func TestVolumeInfoByPath(t *testing.T) {
	err := ysys.SetupPowerShellVersion()
	if err != nil {
		t.Fatal(err)
	}

	p1 := "F:\\go_projects\\keen\\ylog\\ylog.go"
	p2 := "C:\\Program Files\\Microsoft SQL Server\\MSSQL10_50.MSSQLSERVER\\MSSQL\\DATA\\TEST_DB.mdf"
	ps := []string{p1, p2}

	for _, p := range ps {
		vol, err := ysys.VolumeInfoByPath(p)
		if err != nil {
			t.Errorf("failed to fetch volume information: %v", err)
		}

		print(vol, t)
	}
}

func TestListening(t *testing.T) {
	err := ysys.SetupPowerShellVersion()
	if err != nil {
		t.Fatal(err)
	}

	var p1, p2 uint = 1460, 16440
	ps := []uint{p1, p2}

	for _, p := range ps {
		info, err := ysys.Listening(p)
		if err != nil {
			t.Errorf("failed to fetch tcp information: %v", err)
		}

		print(info, t)
	}
}

func TestDriveLetters(t *testing.T) {
	err := ysys.SetupPowerShellVersion()
	if err != nil {
		t.Fatal(err)
	}

	drivers, err := ysys.DriveLetters()
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

	avail := ysys.AvailableLetter(driverSet)

	t.Log("avail", string(avail))
}
