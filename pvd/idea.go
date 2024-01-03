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
	BackupAll() (BackupImage, error)
	BackupDataOnly() (BackupImage, error)
	BackupLogOnly() (BackupImage, error)
	Restore(backupSet BackupImage) error
	Mount(backupSet BackupImage) error
	UnMount(backupSet BackupImage) error
	GenFCDMApplicationName() model.Application // 每个应用在目标主机上的FCDM应用名称应当唯一
	GenFCDMApplicationVolumesName() []string   // 每个应用需要申请的所有设备列表的名称/标识符应当全局唯一
	ToFCDMApplication() model.Application
	Refresh()
}

type Provider interface {
	ParseBackupImage() (BackupImage, error)
	DiscoverApplications() ([]BackupApplication, error)
	FindApplication(appName string) (BackupApplication, error)
	HandleLang(lang *LangPackage)
	PlugInfo() string
}

func Do(pvd Provider, env FCDMArgument) int {
	switch env.Command {
	case model.CMD_DISCOVER:
		keen.Log.Info("start to discover applications")
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
		println(pvd.PlugInfo())
	}

	return 0
}

// Pre 处理环境变量中配置的有效性
func Pre(validators ...ConfigValidator) (FCDMArgument, bool) {
	env := NewFCDMArgument()
	baseValid := env.Validate()
	if !baseValid {
		keen.Log.Error("the validation of FCDM environment does not pass")
		return env, false
	}

	var isValidConfig bool
	for _, validator := range validators {
		key, ok := validator(env)
		if !ok {
			keen.Log.Warn("failed  to validate config [%s]", key)
			isValidConfig = false
			break
		}
		isValidConfig = ok
	}

	if !isValidConfig {
		keen.Log.Error("failed to valid configuration")
		return env, false
	}

	return env, true
}

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
