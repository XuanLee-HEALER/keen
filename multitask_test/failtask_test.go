package multitasktest_test

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"gitea.fcdm.top/lixuan/keen/common"
	"gitea.fcdm.top/lixuan/keen/multitask"
)

// func TestFailTasks(t *testing.T) {
// 	fail := multitask.NewFailTasks(5)

// 	worker := func(args ...any) error {
// 		if args[0].(int) == 2 {
// 			time.Sleep(2 * time.Second)
// 			return fmt.Errorf("test error")
// 		}
// 		f, err := os.Create("file" + strconv.Itoa(args[0].(int)))
// 		if err != nil {
// 			return err
// 		}
// 		f.Close()
// 		return nil
// 	}

// 	workerRevoke := func(args ...any) error {
// 		name := "file" + strconv.Itoa(args[0].(int))
// 		return os.Remove(name)
// 	}
// 	for i := 0; i < 4; i++ {
// 		fail.AddTask(multitask.NewTask(worker, workerRevoke), i)
// 	}

//		fail.Run()
//	}
var div1 multitask.FnParam = func(args ...any) error {
	_, d2 := args[0].(int), args[1].(int)
	if d2 == 0 {
		return fmt.Errorf("divident is 0")
	}
	return nil
}

var div2 multitask.FnRet = func() (any, error) {
	d1, d2 := rand.Intn(5), rand.Intn(2)
	if d2 == 0 {
		return 0, fmt.Errorf("divident is 0")
	}
	return d1 / d2, nil
}

var div3 multitask.FnParamRet = func(args ...any) (any, error) {
	d1, d2 := args[0].(int), args[1].(int)
	if d2 == 0 {
		return 0, fmt.Errorf("divident is 0")
	}
	return d1 / d2, nil
}

var div4 multitask.Fn = func() error {
	d2 := rand.Intn(2)
	if d2 == 0 {
		return fmt.Errorf("divident is 0")
	}
	return nil
}

var div5 multitask.FnParamNoErr = func(args ...any) {
	d1, d2 := args[0].(int), args[1].(int)
	println("do multitask.FnParamNoErr", "d1", d1, "d2", d2)
}

var div6 multitask.FnRetNoErr = func() any {
	println("do multitask.FnRetNoErr", "d1", 6, "d2", 2)
	return 6 / 2
}

var div7 multitask.FnParamRetNoErr = func(args ...any) any {
	d1, d2 := args[0].(int), args[1].(int)
	return d1 / d2
}

var div8 multitask.FnNoErr = func() {
	println("do multitask.FnNoErr")
}

func TestTaskFn(t *testing.T) {
	_, err := multitask.Do[multitask.FnParam](div1, []any{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("div1")
	r, _ := multitask.Do[multitask.FnRet](div2, nil)
	t.Log("div2", r.(int))
	r, err = multitask.Do[multitask.FnParamRet](div3, []any{6, 3})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("div3", r.(int))
	multitask.Do[multitask.Fn](div4, nil)
	t.Log("div4")
	multitask.Do[multitask.FnParamNoErr](div5, []any{7, 2})
	t.Log("div5")
	r, _ = multitask.Do[multitask.FnRetNoErr](div6, nil)
	t.Log("div6", r.(int))
	r, _ = multitask.Do[multitask.FnParamRetNoErr](div7, []any{6, 3})
	t.Log("div7", r.(int))
	multitask.Do[multitask.FnNoErr](div8, nil)
	t.Log("div8")
}

func TestTasks(t *testing.T) {
	args := make([][]any, 0)
	for i, j := 0, 9; i < 10; {
		ti, tj := i, j
		args = append(args, []any{ti, tj})
		i++
		j--
	}
	testFn := func(args ...any) (any, error) {
		d1, d2 := args[0].(int), args[1].(int)
		if d2 == 0 {
			return nil, fmt.Errorf("divident is zero")
		}
		return d1 / d2, nil
	}

	taskList := multitask.TaskList[multitask.FnParamRet](testFn, args)
	fail := multitask.NewFailTasks[multitask.FnParamRet](taskList...)
	go fail.Run()

	ch := fail.Out()
	for v := range ch {
		if !common.IsErr(v) {
			println(v.(int))
		} else {
			println("error")
		}
	}
}

func TestReTasks(t *testing.T) {
	wf := func(args ...any) error {
		name := args[0].(string)
		// n := rand.Intn(10)
		// if n < 3 {
		// 	return fmt.Errorf("toy error")
		// }
		f, err := os.Create(args[0].(string))
		if err != nil {
			return err
		}
		defer f.Close()

		f.WriteString("my name is " + name)
		return nil
	}

	re := func(args ...any) error {
		name := args[0].(string)
		return os.Remove(name)
	}

	args := make([][]any, 0)
	for i := 0; i < 10; i++ {
		args = append(args, []any{"files" + strconv.Itoa(i)})
	}

	taskList := multitask.TaskList[multitask.FnParam](wf, args)
	fail := multitask.NewReFailTasks[multitask.FnParam, multitask.FnRevokeParam](re, taskList...)

	go fail.RunRe()

	ch := fail.Out()
	counter := 0
	for range ch {
		counter++
	}

	if counter != 10 {
		t.FailNow()
	}
}
