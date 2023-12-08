package ylog2_test

import (
	"os"
	"path"
	"strings"
	"testing"

	"gitea.fcdm.top/lixuan/keen/ylog2"
)

func TestYLog2Level(t *testing.T) {
	t.Log(ylog2.TRACE)
	t.Log(ylog2.DEBUG)
	t.Log(ylog2.INFO)
	t.Log(ylog2.WARN)
	t.Log(ylog2.ERROR)
	t.Log(ylog2.FATAL)
}

func TestYLog2ConsoleLogger(t *testing.T) {
	console := ylog2.NewConsoleWriter(func(i int) bool { return i >= ylog2.TRACE })
	logger := ylog2.NewLogger(console)
	logger.Trace("test trace log message")
}

func TestSubPath(t *testing.T) {
	p1 := "p1"
	p2 := "p1/p2"
	p3 := "/p3/p2/p1"
	p4 := "C:\\"
	p5 := "/"
	p6 := "C:\\p1"
	p7 := "c:\\p2\\p1"
	samples := []string{p1, p2, p3, p4, p5, p6, p7}

	f := func(str string, lyr int) string {
		pstck := make([]string, 0)

		count := lyr
		for count > 0 {
			d, f := path.Split(str)
			pstck = append(pstck, f)
			if d == "" {
				break
			}

			count--
			str = d
		}

		for i, j := 0, len(pstck)-1; i < j; i, j = i+1, j-1 {
			pstck[i], pstck[j] = pstck[j], pstck[i]
		}

		return strings.Join(pstck, string(os.PathSeparator))
	}

	for _, p := range samples {
		t.Log(f(p, 1))
	}
}
