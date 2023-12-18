package pvd

import "gitea.fcdm.top/lixuan/keen/util"

type BackupImage interface {
}

type BackupApplication interface {
}

type Provider interface {
	DiscoverApplications() ([]BackupApplication, error)
	FindApplication(appName string) (BackupApplication, error)
}

func Run(clean util.CleanFunc) {

}
