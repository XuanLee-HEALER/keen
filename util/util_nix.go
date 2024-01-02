//go:build linux && aix && darwin

package util

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	xgods "github.com/XuanLee-HEALER/gods-keqing"
)

func Sync(_ string) {
	syscall.Sync()
}

type MemSeg struct {
	Layer int
	Start uint64
	End   uint64
	Usage string
	Sub   []MemSeg
}

func prefixStr(pre string, seg MemSeg) string {
	strb := strings.Builder{}
	strb.WriteString(fmt.Sprintf("%sstart: 0x%09x end: 0x%09x usage: %s", pre, seg.Start, seg.End, seg.Usage))
	for _, m := range seg.Sub {
		strb.WriteString("\n")
		strb.WriteString(prefixStr(pre+"\t", m))
	}
	return strb.String()
}

func (seg MemSeg) String() string {
	return prefixStr("", seg)
}

// QuerySystemMemory 获取操作系统内存分配信息，源数据为/proc/iomem信息
func QuerySystemMemory() ([]*MemSeg, error) {
	f, err := os.Open("/proc/iomem")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	recToSeg := func(t string) MemSeg {
		ct := 0
		for _, r := range t {
			if r == ' ' {
				ct++
			} else {
				break
			}
		}
		ly := ct / 2
		t = t[ct:]
		segs := strings.Split(t, " : ")
		subsegs := strings.Split(segs[0], "-")
		s, _ := strconv.ParseUint(subsegs[0], 16, 64)
		e, _ := strconv.ParseUint(subsegs[1], 16, 64)
		return MemSeg{ly, s, e, segs[1], nil}
	}

	merge := func(l int, sck *xgods.Stack) int {
		curl := sck.Peek().(*MemSeg).Layer
		for curl > l {
			tarr := make([]MemSeg, 0)
			for {
				e1 := sck.Remove().(*MemSeg)
				e2 := sck.Peek().(*MemSeg)
				if e1.Layer == e2.Layer {
					tarr = append(tarr, *e1)
					continue
				} else if e1.Layer > e2.Layer {
					tarr = append(tarr, *e1)
					for i, j := 0, len(tarr)-1; i < j; i, j = i+1, j-1 {
						tarr[i], tarr[j] = tarr[j], tarr[i]
					}
					e2.Sub = tarr
					curl = e2.Layer
					break
				}
			}
		}
		return curl
	}

	res := make([]*MemSeg, 0)
	sck := xgods.NewStack()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ln := scanner.Text()
		seg := recToSeg(ln)
		h := sck.Peek()
		if h == nil {
			sck.Push(&seg)
		} else {
			if old := h.(*MemSeg).Layer; old < seg.Layer {
				sck.Push(&seg)
			} else if old == seg.Layer {
				if old == 0 {
					res = append(res, sck.Remove().(*MemSeg))
					sck.Push(&seg)
				} else {
					sck.Push(&seg)
				}
			} else if old > seg.Layer {
				old = merge(seg.Layer, &sck)
				if old == 0 {
					res = append(res, sck.Remove().(*MemSeg))
				}
				sck.Push(&seg)
			}
		}
	}

	merge(0, &sck)
	res = append(res, sck.Remove().(*MemSeg))

	return res, nil
}
