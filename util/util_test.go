package util_test

import (
	"testing"

	"gitea.fcdm.top/lixuan/keen/util"
)

func TestReadListOutput(t *testing.T) {
	ori := `
Host Name:                 DESKTOP-ECQU9SU
OS Name:                   Microsoft Windows 10 专业版
OS Version:                10.0.19045 N/A Build 19045
OS Manufacturer:           Microsoft Corporation
OS Configuration:          Standalone Workstation
OS Build Type:             Multiprocessor Free
Registered Owner:          a543748616@outlook.com
Registered Organization:   N/A
Product ID:                00331-20020-00000-AA069
Original Install Date:     9/27/2020, 12:38:08 PM
System Boot Time:          8/31/2023, 2:32:48 PM
System Manufacturer:       HASEE Computer
System Model:              N95TP6
System Type:               x64-based PC
Processor(s):              1 Processor(s) Installed.
                           [01]: Intel64 Family 6 Model 158 Stepping 10 GenuineIntel ~2808 Mhz
BIOS Version:              American Megatrends Inc. 1.05.01RHA1, 12/13/2017
Windows Directory:         C:\Windows
System Directory:          C:\Windows\system32
Boot Device:               \Device\HarddiskVolume2
System Locale:             zh-cn;Chinese (China)
Input Locale:              en-us;English (United States)
Time Zone:                 (UTC+08:00) Beijing, Chongqing, Hong Kong, Urumqi
Total Physical Memory:     16,256 MB
Available Physical Memory: 5,845 MB
Virtual Memory: Max Size:  42,456 MB
Virtual Memory: Available: 5,286 MB
Virtual Memory: In Use:    37,170 MB
Page File Location(s):     C:\pagefile.sys
Domain:                    WORKGROUP
Logon Server:              \\DESKTOP-ECQU9SU
Hotfix(s):                 37 Hotfix(s) Installed.
                           [01]: KB5029716
                           [02]: KB5029919
                           [03]: KB5028951
                           [04]: KB4561600
                           [05]: KB4562830
                           [06]: KB4577266
                           [07]: KB4577586
                           [08]: KB4580325
                           [09]: KB4586864
                           [10]: KB4589212
                           [11]: KB4593175
                           [12]: KB4598481
                           [13]: KB5000736
                           [14]: KB5003791
                           [15]: KB5011048
                           [16]: KB5011050
                           [17]: KB5012170
                           [18]: KB5015684
                           [19]: KB5030211
                           [20]: KB5006753
                           [21]: KB5007273
                           [22]: KB5011352
                           [23]: KB5011651
                           [24]: KB5014032
                           [25]: KB5014035
                           [26]: KB5014671
                           [27]: KB5015895
                           [28]: KB5016705
                           [29]: KB5018506
                           [30]: KB5020372
                           [31]: KB5022924
                           [32]: KB5023794
                           [33]: KB5025315
                           [34]: KB5028318
                           [35]: KB5028380
                           [36]: KB5029709
                           [37]: KB5005699
Network Card(s):           6 NIC(s) Installed.
                           [01]: ExpressVPN TUN Driver
                                 Connection Name: 本地连接
                                 Status:          Media disconnected
                           [02]: VMware Virtual Ethernet Adapter for VMnet1
                                 Connection Name: VMware Network Adapter VMnet1
                                 DHCP Enabled:    No
                                 IP address(es)
                                 [01]: 169.254.55.237
                           [03]: VMware Virtual Ethernet Adapter for VMnet8
                                 Connection Name: VMware Network Adapter VMnet8
                                 DHCP Enabled:    No
                                 IP address(es)
                                 [01]: 169.254.31.117
                           [04]: Intel(R) Dual Band Wireless-AC 3168
                                 Connection Name: WLAN
                                 Status:          Media disconnected
                           [05]: Realtek PCIe GbE Family Controller
                                 Connection Name: 以太网
                                 DHCP Enabled:    No
                                 IP address(es)
                                 [01]: 192.168.108.208
                           [06]: Bluetooth Device (Personal Area Network)
                                 Connection Name: 蓝牙网络连接
                                 Status:          Media disconnected
Hyper-V Requirements:      VM Monitor Mode Extensions: Yes
                           Virtualization Enabled In Firmware: Yes
                           Second Level Address Translation: Yes
                           Data Execution Prevention Available: Yes`

	lst := util.ReadListOutput([]byte(ori))
	for _, e := range lst {
		for k, v := range e {
			t.Log(k, "=>", v)
		}
	}
}
