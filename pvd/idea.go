package pvd

import (
	"github.com/cnyjp/fcdmpublic/model"
)

type BackupImage interface {
	Meta() string
	ToFCDMBackupImage() model.BackupResponse
}

type BackupApplication interface {
	Backup() (BackupImage, error)
	Restore(backupSet BackupImage) error
	Mount(backupSet BackupImage) error
	UnMount(backupSet BackupImage) error
	ToFCDMApplication() model.Application
	Refresh()
}

type Provider interface {
	ParseBackupImage(path string) BackupImage
	DiscoverApplications() ([]BackupApplication, error)
	FindApplication(appName string) (BackupApplication, error)
	Run() int
}
