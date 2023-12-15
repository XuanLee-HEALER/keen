package win

import (
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

func Volumes() {
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, _ := oleutil.CreateObject("InternetExplorer.Application")
	defer unknown.Release()

	ie, _ := unknown.QueryInterface(ole.IID_IDispatch)
	defer ie.Release()

	oleutil.MustPutProperty(ie, "Visible", true)
	oleutil.MustCallMethod(ie, "Navigate", "https://www.bing.com")

	for oleutil.MustGetProperty(ie, "Busy").Val != false {
		oleutil.SinkEvents(ie)
	}
}
