package pvd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cnyjp/fcdmpublic/model"
)

type Nation struct {
	ID   uint8
	Name string
}

const (
	SUPPORTED_LINGUAL uint8 = 0xFF
)

type display struct {
	name        string
	desc        string
	optionsName map[string]string
}

type LangPackage struct {
	savedNation uint8
	savedIds    map[uint8]struct{}
	nations     [SUPPORTED_LINGUAL]Nation
	displays    [SUPPORTED_LINGUAL]map[string]display
}

func NewLangPackage() LangPackage {
	return LangPackage{
		0,
		make(map[uint8]struct{}),
		[SUPPORTED_LINGUAL]Nation{},
		[SUPPORTED_LINGUAL]map[string]display{},
	}
}

func (pkg LangPackage) String() string {
	w := strings.Builder{}
	w.WriteString("Language Package: \n")
	for id := range pkg.savedIds {
		w.WriteString(fmt.Sprintf("%s\n", pkg.nations[id].Name))
		for k, display := range pkg.displays[id] {
			w.WriteString(fmt.Sprintf("OriName: %s\tName: %s\tDesc: %s\n", k, display.name, display.desc))
			if display.optionsName != nil {
				for k, v := range display.optionsName {
					w.WriteString(fmt.Sprintf("%s -> %s\n", v, k))
				}
			}
		}
		w.WriteString("\n")
	}
	w.WriteString("\b")
	return w.String()
}

func (pkg LangPackage) ApplyMultiLingual(defaultNation Nation, config *model.ConfigConfig) {
	config.I18n = make(map[string]model.ConfigI18n)

	config.Desc = pkg.displays[defaultNation.ID][config.Name].desc
	config.Options = pkg.displays[defaultNation.ID][config.Name].optionsName

	for id := range pkg.savedIds {
		curDisplay := pkg.displays[id][config.Name]
		config.I18n[pkg.nations[id].Name] = model.ConfigI18n{
			Name:    curDisplay.name,
			Desc:    curDisplay.desc,
			Options: curDisplay.optionsName,
		}
	}
}

func (pkg LangPackage) NationExist(nation Nation) bool {
	if _, ok := pkg.savedIds[nation.ID]; !ok {
		return false
	}
	return true
}

func (pkg *LangPackage) AddNation(nations ...Nation) error {
	if len(nations)+int(pkg.savedNation) > int(SUPPORTED_LINGUAL) {
		return errors.New("too many nations")
	}

	for _, nation := range nations {
		if nation.ID >= SUPPORTED_LINGUAL {
			return errors.New("error nation ID: " + strconv.Itoa(int(nation.ID)))
		}
		if pkg.NationExist(nation) {
			return errors.New("nation " + strconv.Itoa(int(nation.ID)) + " existed")
		}
		pkg.savedIds[nation.ID] = struct{}{}
		pkg.nations[nation.ID] = nation
		pkg.savedNation++
	}

	return nil
}

func (pkg *LangPackage) AddDisplay(nation Nation, oriName, name, desc string, optionsName map[string]string) error {
	if pkg.NationExist(nation) {
		if pkg.displays[nation.ID] == nil {
			pkg.displays[nation.ID] = map[string]display{
				oriName: display{name, desc, optionsName},
			}
		} else {
			pkg.displays[nation.ID][oriName] = display{name, desc, optionsName}
		}

		return nil
	}
	return errors.New("Nation " + strconv.Itoa(int(nation.ID)) + " is not exist")
}
