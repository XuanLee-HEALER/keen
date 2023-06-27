package multitasktest_test

import (
	"sync"
	"testing"
	"time"

	"gitea.fcdm.top/lixuan/keen/multitask"
)

func TestPubSub(t *testing.T) {
	const LIMIT = 5
	p := multitask.NewPubChan(5)

	go func() {
		time.Sleep(2 * time.Second)
		p.Pub()
	}()

	wg := sync.WaitGroup{}
	for i := 0; i < LIMIT+1; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if id, ch, ok := p.Sub(); ok {
				defer p.UnSub(id)
				<-ch
				println("successed")
			} else {
				println("failed")
			}
		}(i)
	}

	wg.Wait()
}
