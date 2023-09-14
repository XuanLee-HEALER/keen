package win_test

import (
	"encoding/json"
	"testing"

	"gitea.fcdm.top/lixuan/keen/win"
)

func print(obj any, t *testing.T) {
	bs, _ := json.MarshalIndent(obj, "", "  ")
	t.Log("\n", string(bs))
}

func TestHostComputerInfo(t *testing.T) {
	err := win.SetupPowerShellVersion()
	if err != nil {
		t.Fatal(err)
	}

	computerInfo, err := win.HostComputerInfo()
	if err != nil {
		t.Fatalf("failed to retrieve computer information: %v", err)
	}

	print(computerInfo, t)
}

func TestServiceByFilter(t *testing.T) {
	err := win.SetupPowerShellVersion()
	if err != nil {
		t.Fatal(err)
	}

	cond0 := `$_.Name -eq 'AJRouter'`
	cond1 := `$_.Name -eq 'MSSQLSERVER'`
	cond2 := `$_.Name -like 'A*'`
	conds := []string{cond0, cond1, cond2}

	for _, cond := range conds {
		t.Log("current condition: ", cond)
		services, err := win.ServiceByFilter(cond)
		if err != nil {
			t.Errorf("failed to retrieve service information: %v", err)
		}

		if len(services) > 0 {
			print(services, t)
		}
	}
}
