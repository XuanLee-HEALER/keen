package pvd

import (
	"encoding/json"
	"strconv"
	"time"

	"gitea.fcdm.top/lixuan/keen"
	"gitea.fcdm.top/lixuan/keen/ylog"
	"github.com/cnyjp/fcdmpublic/model"
)

type BackupImage interface {
	Meta() string
	ToFCDMBackupImage() model.BackupResponse
}

// BackupApplication 应当使用指针来实现此接口，对于恢复、挂载、卸载、刷新等操作，应用状态会发生改变
type BackupApplication interface {
	BackupAll() (BackupImage, error)           // 产生全备份
	BackupDataOnly() (BackupImage, error)      // 仅数据
	BackupLogOnly() (BackupImage, error)       // 仅日志
	AppType() string                           // 支持不同类型的应用，以区分不同的配置项
	AppConfigurationList() []string            // 获取对应类型的配置项的key
	Restore(backupSet BackupImage) error       // 恢复镜像
	Mount(backupSet BackupImage) error         // 挂载镜像
	UnMount(backupSet BackupImage) error       // 卸载镜像
	GenFCDMApplicationName() model.Application // 每个应用在目标主机上的FCDM应用名称应当唯一
	GenFCDMApplicationVolumesName() []string   // 每个应用需要申请的所有设备列表的名称/标识符应当全局唯一
	ToFCDMApplication() model.Application      // 转换到FCDM应用规范
	Refresh()                                  // 刷新应用状态
}

type Provider interface {
	ValidConfig() bool                                         // 验证配置项/选项
	ParseBackupImage() (BackupImage, error)                    // 根据元数据文件路径来转换镜像结构体
	DiscoverApplications() ([]BackupApplication, error)        // 发现主机上所有应用
	FindApplication(appName string) (BackupApplication, error) // 在主机上查找指定应用名称的应用
	HandleLang(lang *LangPackage)                              // 处理配置项的多语言相关
	PlugInfo() string                                          // 返回插件信息
}

func ValidateConfig(pvd Provider) bool {
	keen.Log.Info("start to validate configuration")
	r := pvd.ValidConfig()
	if !r {
		keen.Log.Error("failed to validate the configuration for the current operation")
		return r
	}
	keen.Log.Info("validate configuration completely")

	return r
}

func Do(pvd Provider, env FCDMArgument) int {
	switch env.Command {
	case model.CMD_DISCOVER:
		keen.Log.Info("start to discover applications")
		if !ValidateConfig(pvd) {
			return C_ERR_EXIT
		}

		apps, err := pvd.DiscoverApplications()
		if err != nil {
			keen.Log.Error("failed to discover applications in the target host: %v", err)
			return C_ERR_EXIT
		}

		keen.Log.Info("start to transform custom application to FCDM specific application")
		xapps := make([]model.Application, 0, len(apps))
		for _, app := range apps {
			xapps = append(xapps, app.ToFCDMApplication())
		}
		bs, _ := json.MarshalIndent(xapps, "", "  ")
		keen.Log.Debug("transform result:\n%s", string(bs))

		bs, err = json.Marshal(xapps)
		if err != nil {
			keen.Log.Error("failed to marshal the struct: %v", err)
			return C_ERR_EXIT
		}

		println(string(bs))
	case model.CMD_APPLICATION_INFO:
		keen.Log.Info("start to refresh application information")
		if !ValidateConfig(pvd) {
			return C_ERR_EXIT
		}

		appName := env.ApplicationName
		app, err := pvd.FindApplication(appName)
		if err != nil {
			keen.Log.Error("failed to find the specific application [%s]: %v", appName, err)
			return C_ERR_EXIT
		}

		keen.Log.Info("start to transform custom application to FCDM specific application")
		xapp := app.ToFCDMApplication()
		bs, _ := json.MarshalIndent(xapp, "", "  ")
		keen.Log.Debug("transform result:\n%s", string(bs))

		bs, err = json.Marshal(xapp)
		if err != nil {
			keen.Log.Error("failed to marshal the struct: %v", err)
			return C_ERR_EXIT
		}

		println(string(bs))
	case model.CMD_BACKUP:
		keen.Log.Info("start to backup the application")
		if !ValidateConfig(pvd) {
			return C_ERR_EXIT
		}

		keen.Log.Info("start to find the specific application")
		appName := env.ApplicationName
		app, err := pvd.FindApplication(appName)
		if err != nil {
			keen.Log.Error("failed to find the specific application [%s]: %v", appName, err)
			return C_ERR_EXIT
		}
		keen.Log.Info("find the specific application completely")

		bkType, _ := strconv.Atoi(env.BackupType)
		keen.Log.Trace("current backup type: [%s]", bkType)
		switch bkType {
		case model.BACKUP_TYPE_ALL:
			keen.Log.Info("start to backup all of the application")
			img, err := app.BackupAll()
			if err != nil {
				keen.Log.Error("failed to backup all of the application: %v", err)
				return C_ERR_EXIT
			}

			bs, _ := json.MarshalIndent(img, "", "  ")
			keen.Log.Debug("image information:\n%s", string(bs))

			bs, err = json.Marshal(img)
			if err != nil {
				keen.Log.Error("failed to marshal the struct: %v", err)
				return C_ERR_EXIT
			}

			println(string(bs))
		case model.BACKUP_TYPE_DB:
			keen.Log.Info("start to only backup data of the application")
			img, err := app.BackupDataOnly()
			if err != nil {
				keen.Log.Error("failed to only backup data of the application: %v", err)
				return C_ERR_EXIT
			}

			bs, _ := json.MarshalIndent(img, "", "  ")
			keen.Log.Debug("image information:\n%s", string(bs))

			bs, err = json.Marshal(img)
			if err != nil {
				keen.Log.Error("failed to marshal the struct: %v", err)
				return C_ERR_EXIT
			}

			println(string(bs))
		case model.BACKUP_TYPE_LOG:
			keen.Log.Info("start to only backup log of the application")
			img, err := app.BackupAll()
			if err != nil {
				keen.Log.Error("failed to only backup log of the application: %v", err)
				return C_ERR_EXIT
			}

			bs, _ := json.MarshalIndent(img, "", "  ")
			keen.Log.Debug("image information:\n%s", string(bs))

			bs, err = json.Marshal(img)
			if err != nil {
				keen.Log.Error("failed to marshal the struct: %v", err)
				return C_ERR_EXIT
			}

			println(string(bs))
		}
	case model.CMD_RESTORE:
		keen.Log.Info("start to restore the backup iamge to application")
		if !ValidateConfig(pvd) {
			return C_ERR_EXIT
		}

		keen.Log.Info("start to parse the image from meta file")
		img, err := pvd.ParseBackupImage()
		if err != nil {
			keen.Log.Error("failed to parse the backup image: %v", err)
			return C_ERR_EXIT
		}

		bs, _ := json.MarshalIndent(img, "", "  ")
		keen.Log.Info("parse result:\n%s", string(bs))

		keen.Log.Info("start to find the specific application")
		appName := env.ApplicationName
		app, err := pvd.FindApplication(appName)
		if err != nil {
			keen.Log.Error("failed to find the specific application [%s]: %v", appName, err)
			return C_ERR_EXIT
		}

		keen.Log.Info("start to restore the backup image")
		err = app.Restore(img)
		if err != nil {
			keen.Log.Error("failed to restore the backup image: %v", err)
			return C_ERR_EXIT
		}
		keen.Log.Info("restore the backup image completely")
	case model.CMD_MOUNT:
		keen.Log.Info("start to mount the backup iamge to application")
		if !ValidateConfig(pvd) {
			return C_ERR_EXIT
		}

		keen.Log.Info("start to parse the image from meta file")
		img, err := pvd.ParseBackupImage()
		if err != nil {
			keen.Log.Error("failed to parse the backup image: %v", err)
			return C_ERR_EXIT
		}

		bs, _ := json.MarshalIndent(img, "", "  ")
		keen.Log.Info("parse result:\n%s", string(bs))

		keen.Log.Info("start to find the specific application")
		appName := env.ApplicationName
		app, err := pvd.FindApplication(appName)
		if err != nil {
			keen.Log.Error("failed to find the specific application [%s]: %v", appName, err)
			return C_ERR_EXIT
		}

		keen.Log.Info("start to mount the backup image")
		err = app.Mount(img)
		if err != nil {
			keen.Log.Error("failed to mount the backup image: %v", err)
			return C_ERR_EXIT
		}
		keen.Log.Info("mount the backup image completely")
	case model.CMD_UMOUNT:
		keen.Log.Info("start to unmount the backup iamge to application")
		if !ValidateConfig(pvd) {
			return C_ERR_EXIT
		}

		keen.Log.Info("start to parse the image from meta file")
		img, err := pvd.ParseBackupImage()
		if err != nil {
			keen.Log.Error("failed to parse the backup image: %v", err)
			return C_ERR_EXIT
		}

		bs, _ := json.MarshalIndent(img, "", "  ")
		keen.Log.Info("parse result:\n%s", string(bs))

		keen.Log.Info("start to find the specific application")
		appName := env.ApplicationName
		app, err := pvd.FindApplication(appName)
		if err != nil {
			keen.Log.Error("failed to find the specific application [%s]: %v", appName, err)
			return C_ERR_EXIT
		}

		keen.Log.Info("start to unmount the backup image")
		err = app.UnMount(img)
		if err != nil {
			keen.Log.Error("failed to unmount the backup image: %v", err)
			return C_ERR_EXIT
		}
		keen.Log.Info("unmount the backup image completely")
	case model.CMD_PLUGIN_INFO:
		if !ValidateConfig(pvd) {
			return C_ERR_EXIT
		}

		println(pvd.PlugInfo())
	}

	return 0
}

// Pre 处理环境变量中配置的有效性
func Pre() (FCDMArgument, bool) {
	env := NewFCDMArgument()
	baseValid := env.Validate()
	if !baseValid {
		keen.Log.Error("the validation of FCDM environment does not pass")
		return env, false
	}

	return env, true
}

// ProviderLogger 一般Provider的Logger配置，控制台打印INFO级别以上日志，文件日志打印TRACE级别以上日志。文件日志为7天删除+按照命令类型归档
func ProviderLogger(logPath, fileName string) ylog.Logger {
	var logger ylog.Logger
	console := ylog.NewConsoleWriter(func(i int8) bool { return i >= ylog.INFO }, true)
	file, err := ylog.NewFileWriter(logPath, fileName, func(i int8) bool { return i >= ylog.TRACE }, 7*24*time.Hour, SimpleArch)
	if err != nil {
		ylog.Errorf("failed to create file logger: %v", err)
		logger = ylog.NewLogger(console)
	} else {
		logger = ylog.NewLogger(console, file)
	}
	return logger
}
