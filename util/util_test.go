package util_test

import (
	"testing"

	"gitea.fcdm.top/lixuan/keen/util"
)

func TestReadCSVOutput(t *testing.T) {
	oriContent := `
"ProcessId","Name","State","PathName","StartName"
"1460","MSSQLSERVER","Running","""C:\Program Files\Microsoft SQL Server\MSSQL10_50.MSSQLSERVER\MSSQL\Binn\sqlservr.exe"" -sMSSQLSERVER","NT AUTHORITY\NETWORKSERVICE"
	`
	csv := util.ReadCSVOutput([]byte(oriContent))
	for i := range csv {
		for j := range csv[i] {
			print(csv[i][j], " ")
		}
		println()
	}
}
