package datastructure

import (
	"fmt"
	"strings"

	"github.com/emirpasic/gods/maps/treemap"
)

/**
DirTree

模拟目录树结构，支持插入和导航操作
*/

type SimpleFile struct {
	DirName    string
	FileName   string
	Size       uint64
	User       string
	Group      string
	Access     string
	ModifyTime string
	CreateTime string
}

func (f SimpleFile) IsDir() bool {
	return f.FileName == ""
}

type DirTree DirNode
type DirNode struct {
	file *SimpleFile
	subs *treemap.Map
}

func NewDirTree() DirTree {
	return DirTree(DirNode{nil, treemap.NewWithStringComparator()})
}

func NewDir(dirName string) *DirTree {
	return (*DirTree)(&DirNode{
		file: &SimpleFile{
			DirName: dirName,
		},
		subs: treemap.NewWithStringComparator(),
	})
}

func NewFile(f *SimpleFile) *DirTree {
	return (*DirTree)(&DirNode{
		file: f,
	})
}

func (t DirTree) IsDir() bool {
	return t.IsRoot() || t.file.IsDir()
}

func (t DirTree) IsRoot() bool {
	return t.file == nil
}

func (t *DirTree) Navigate(path ...string) *DirTree {
	if t.IsRoot() && len(path) == 0 {
		return t
	}

	lt := t
	for _, p := range path {
		if sub, ok := lt.subs.Get(p); ok && sub.(*DirTree).file.IsDir() {
			lt = sub.(*DirTree)
		} else {
			return nil
		}
	}

	return lt
}

func (t *DirTree) AddDir(d *DirTree, path ...string) bool {
	od := t.Navigate(path...)
	if od != nil && od.IsDir() {
		od.subs.Put(d.file.DirName, d)
		return true
	}
	return false
}

func (t *DirTree) AddFile(f *DirTree, path ...string) bool {
	od := t.Navigate(path...)
	if od != nil && od.IsDir() {
		od.subs.Put(f.file.FileName, f)
		return true
	}
	return false
}

func (t DirTree) show(indent int) string {
	sb := new(strings.Builder)
	idnt := strings.Repeat(" ", indent)
	if indent == 0 {
		sb.WriteString("-\n")
	} else {
		sb.WriteString(fmt.Sprintf("%s- %s\n", idnt, t.file.DirName))
	}

	indent += 2
	idnt = strings.Repeat(" ", indent)
	for iter := t.subs.Iterator(); iter.Next(); {
		nm, sub := iter.Key(), iter.Value()
		if sub.(*DirTree).IsDir() {
			sb.WriteString(sub.(*DirTree).show(indent))
		} else {
			sb.WriteString(fmt.Sprintf("%s- %s\n", idnt, nm))
		}
	}

	return sb.String()
}

func (t DirTree) String() string {
	return t.show(0)
}
