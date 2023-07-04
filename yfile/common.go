package yfile

import (
	"errors"
	"io"
	"os"
	"sync"
)

type CopyError struct {
	Op    string
	FName string
	Err   error
}

func RCopy(dst []*os.File, src []*os.File) CopyError {
	const (
		BUFFER_SIZE = 32 * 1024
		BUFFER_LEN  = 32
	)

	var err CopyError
	terr := make(chan CopyError, 1)
	ntfs := make([]chan struct{}, 0)
	notify := func() {
		for _, ntf := range ntfs {
			ntf <- struct{}{}
		}
	}
	wg := new(sync.WaitGroup)

	reader := func(ch chan<- []byte, errc chan<- CopyError, ntf <-chan struct{}, src *os.File) {
		defer src.Close()
		defer wg.Done()

		buf := make([]byte, BUFFER_SIZE)
		isErr := false
	out:
		for {
			select {
			case <-ntf:
				close(ch)
				break out
			default:
				if !isErr {
					if n, err := src.Read(buf); n == 0 && errors.Is(err, io.EOF) {
						close(ch)
						break out
					} else if err != nil {
						errc <- CopyError{"read", src.Name(), err}
						isErr = true
					} else {
						ch <- buf[:n]
					}
				}
			}
		}
	}
	writer := func(ch <-chan []byte, errc chan<- CopyError, ntf <-chan struct{}, dst *os.File) {
		defer wg.Done()
		defer dst.Close()
		defer dst.Sync()
		isErr := false

	out:
		for {
			select {
			case <-ntf:
				break out
			case bs, ok := <-ch:
				if !isErr {
					if !ok {
						break out
					} else {
						if _, err := dst.Write(bs); err != nil {
							errc <- CopyError{"write", dst.Name(), err}
							isErr = true
						}
					}
				}
			}
		}
	}

	for i := range src {
		wg.Add(2)
		i := i
		ch := make(chan []byte, BUFFER_LEN)
		ntf1, ntf2 := make(chan struct{}, 1), make(chan struct{}, 1)
		ntfs = append(ntfs, ntf1, ntf2)
		go reader(ch, terr, ntf1, src[i])
		go writer(ch, terr, ntf2, dst[i])
	}

	go func() {
		for e := range terr {
			notify()
			err = e
			break
		}
	}()

	wg.Wait()

	return err
}
