package pvd

import "github.com/cnyjp/fcdmpublic/model"

func MakeFCDMLocalApp(name string,
	displayName string,
	size int64,
) model.Application {
	return model.Application{
		Name:        name,
		DisplayName: displayName,
		Size:        size,
		Available:   true,
	}
}
