package util_test

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"gitea.fcdm.top/lixuan/keen/util"
)

func TestRegUnReg(t *testing.T) {
	m := util.NewTaskManager()

	wg := new(sync.WaitGroup)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			r := rand.Intn(5)
			time.Sleep(time.Duration(r) * time.Second)
			mid := m.Register()
			t.Logf("%d - %d", idx, mid)

			if r >= 2 {
				t.Logf("unreg goroutine(%d) - id(%d)", idx, mid)
				m.UnRegister(mid)
			}
		}(i)
	}

	wg.Wait()

	m.ListIdGroup()
}
