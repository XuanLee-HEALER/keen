package pvd

import (
	"gitea.fcdm.top/lixuan/keen/util"
	"github.com/cnyjp/fcdmpublic/model"
)

type BackupImage interface {
}

type BackupApplication interface {
	Backup() (BackupImage, error)
	Restore(backupSet BackupImage) error
	Mount(backupSet BackupImage) error
	UnMount(backupSet BackupImage) error
	ToFCDMApplication() model.Application
}

type Provider interface {
	DiscoverApplications() ([]BackupApplication, error)
	FindApplication(appName string) (BackupApplication, error)
}

func Run(clean util.CleanFunc) {

}
