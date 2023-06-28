package multitask

import (
	"reflect"
	"sync"

	"gitea.fcdm.top/lixuan/keen/common"
)

type FnParam func(args ...any) error
type FnRet func() (any, error)
type FnParamRet func(args ...any) (any, error)
type Fn func() error
type FnParamNoErr func(args ...any)
type FnRetNoErr func() any
type FnParamRetNoErr func(args ...any) any
type FnNoErr func()
type FnRevoke func() error
type FnRevokeParam func(args ...any) error

type TaskFn interface {
	FnParam | FnRet | FnParamRet | Fn | FnParamNoErr | FnRetNoErr | FnParamRetNoErr | FnNoErr
}

type TaskRevokeFn interface {
	FnRevoke | FnRevokeParam | interface{}
}

func Do[F TaskFn | TaskRevokeFn](f F, args []any) (any, error) {
	var (
		res any
		err error
	)
	switch ft := reflect.TypeOf(f); ft.Name() {
	case "FnParam", "FnRevokeParam":
		t := make([]reflect.Value, 0)
		t = append(t, reflect.ValueOf(args))
		tr := reflect.ValueOf(f).CallSlice(t)
		if vi := tr[0].Interface(); vi == nil {
			err = nil
		} else {
			err = vi.(error)
		}
	case "FnRet":
		tr := reflect.ValueOf(f).Call(nil)
		res = tr[0].Interface()
		if vi := tr[1].Interface(); vi == nil {
			err = nil
		} else {
			err = vi.(error)
		}
	case "FnParamRet":
		t := make([]reflect.Value, 0)
		t = append(t, reflect.ValueOf(args))
		tr := reflect.ValueOf(f).CallSlice(t)
		res = tr[0].Interface()
		if vi := tr[1].Interface(); vi == nil {
			err = nil
		} else {
			err = vi.(error)
		}
	case "Fn", "FnRevoke":
		tr := reflect.ValueOf(f).Call(nil)
		if vi := tr[0].Interface(); vi == nil {
			err = nil
		} else {
			err = vi.(error)
		}
	case "FnParamNoErr":
		t := make([]reflect.Value, 0)
		t = append(t, reflect.ValueOf(args))
		reflect.ValueOf(f).CallSlice(t)
	case "FnRetNoErr":
		tr := reflect.ValueOf(f).Call(nil)
		res = tr[0].Interface()
	case "FnParamRetNoErr":
		t := make([]reflect.Value, 0)
		t = append(t, reflect.ValueOf(args))
		tr := reflect.ValueOf(f).CallSlice(t)
		res = tr[0].Interface()
	case "FnNoErr":
		reflect.ValueOf(f).Call(nil)
	}

	return res, err
}

type Task[F TaskFn] struct {
	f    F
	args []any
}

func TaskList[F TaskFn](f F, args [][]any) []Task[F] {
	res := make([]Task[F], 0)

	for i := range args {
		ti := i
		res = append(res, Task[F]{f, args[ti]})
	}

	return res
}

type FailTasks[F TaskFn, R TaskRevokeFn] struct {
	taskList []Task[F]
	re       R
	out      chan any
	in       []chan bool
}

func NewFailTasks[F TaskFn](f ...Task[F]) *FailTasks[F, interface{}] {
	out := make(chan any, len(f))
	return &FailTasks[F, interface{}]{
		taskList: f,
		out:      out,
	}
}

func NewReFailTasks[F TaskFn, R TaskRevokeFn](r R, f ...Task[F]) *FailTasks[F, R] {
	in := make([]chan bool, len(f))
	out := make(chan any, len(f))
	return &FailTasks[F, R]{
		taskList: f,
		re:       r,
		out:      out,
		in:       in,
	}
}

func (t *FailTasks[F, R]) Out() <-chan any {
	return t.out
}

func (t *FailTasks[F, R]) Run() {
	wg := sync.WaitGroup{}
	for _, task := range t.taskList {
		wg.Add(1)
		go func(task Task[F]) {
			defer wg.Done()
			res, err := Do[F](task.f, task.args)
			if err != nil {
				t.out <- err
			} else {
				t.out <- res
			}
		}(task)
	}
	wg.Wait()
	defer close(t.out)
}

func (t *FailTasks[F, R]) RunRe() {
	taskSize := len(t.taskList)
	tout := make(chan any, taskSize)
	trout := make(chan any, taskSize)
	cache := make([]any, 0)
	trcache := make([]any, 0)

	for i, task := range t.taskList {
		t.in = append(t.in, make(chan bool, 1))
		go func(idx int, task Task[F]) {
			res, err := Do[F](task.f, task.args)
			if err != nil {
				tout <- err
			} else {
				tout <- res
			}
			for b := range t.in[idx] {
				if !b {
					_, err := Do[R](t.re, task.args)
					trout <- err
				}
			}
		}(i, task)
	}

	isRe := false
	for tv := range tout {
		if isRe {
			break
		}
		if common.IsErr(tv) {
			for _, ch := range t.in {
				ch <- false
			}
			counter := 0
			for err := range trout {
				counter++
				trcache = append(trcache, err)
				if counter == taskSize {
					for _, v := range trcache {
						t.out <- v
					}
					break
				}
			}
			isRe = true
		} else {
			cache = append(cache, tv)
		}
	}

	if !isRe {
		for _, v := range cache {
			t.out <- v
		}
	}

	defer close(t.out)
}
