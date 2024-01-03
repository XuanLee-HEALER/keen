package pvd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cnyjp/fcdmpublic/model"
)

var (
	Zh = Nation{ID: 0x0, Name: "zh_CN"}
	En = Nation{ID: 0x1, Name: "en_US"}
)

type Nation struct {
	ID   uint8
	Name string
}

func NewNation(id uint8, name string) Nation {
	return Nation{
		id, name,
	}
}

const (
	SUPPORTED_LINGUAL uint8 = 0xFF
)

type display struct {
	name    string
	desc    string
	options map[string]string
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
			if display.options != nil {
				w.WriteString(fmt.Sprintf("Option: %v", display.options))
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

	for id := range pkg.savedIds {
		curDisplay := pkg.displays[id][config.Name]
		config.I18n[pkg.nations[id].Name] = model.ConfigI18n{
			Name:    curDisplay.name,
			Desc:    curDisplay.desc,
			Options: curDisplay.options,
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

func (pkg *LangPackage) AddDisplay(nation Nation, oriName, name, desc string, options map[string]string) error {
	if pkg.NationExist(nation) {
		if pkg.displays[nation.ID] == nil {
			pkg.displays[nation.ID] = map[string]display{
				oriName: display{name, desc, options},
			}
		} else {
			pkg.displays[nation.ID][oriName] = display{name, desc, options}
		}

		return nil
	}
	return errors.New("Nation " + strconv.Itoa(int(nation.ID)) + " is not exist")
}
