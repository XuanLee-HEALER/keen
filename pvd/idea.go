package pvd

import (
	"time"

	"github.com/cnyjp/fcdmpublic/model"
)

// Runner 发现出来的应用
type Runner interface {
	ToFCDMApplication() model.Application
	ToFCDMApplicationConfigs() []model.ConfigConfig
	Backup() (Holder, error)
	Restore(Holder) error
	Mount(Holder) error
	Umount(Holder) error
	UpdateConfig()
}

// Holder 应用制作出来的备份集/镜像，实质上Holder的内容就是备份集的元数据
type Holder interface {
	ToFCDMBackupResponse() model.BackupResponse
	WriteToFile(string) error
}

// Provider
type Provider interface {
	// Run 返回状态码，无错误为0，有错误为非0值
	Run() int
	// Discover 发现过程，返回主机上所有的应用
	Discover() ([]Runner, error)
	// Backup 备份过程，返回应用制作的备份镜像
	Backup() (Holder, error)
	// Restore 恢复过程，在应用中应用备份镜像
	Restore() error
	// Mount 挂载过程，在应用上将镜像挂载上
	Mount() error
	// Umount 卸载过程，卸载应用上的镜像
	Umount() error
	// AppInfo 二次发现过程
	AppInfo(appname string) (Runner, error)
	// PluginInfo Privider信息，副作用函数，仅打印
	PluginInfo()
}

// BackupSetMetadata 基础备份集元数据
//
// 备份集元数据需要说明：备份结果，备份出的数据大小，备份过程的开始、结束时间
//
// 通常需要在此基础上添加信息，根据不同的应用类型，需要补充备份集中文件的路径信息（分类），制作此备份集的应用参数的一份拷贝
type BackupSetMetadata struct {
	Status    bool      `json:"backup_status"`
	Size      uint64    `json:"backup_size"`
	Comment   string    `json:"comment"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	SpaceErr  error     `json:"space_error"`
}
