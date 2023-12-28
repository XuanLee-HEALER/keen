package util

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"gitea.fcdm.top/lixuan/keen/datastructure"
)

/// 任务管理器

// TaskMonitor 任务监视器，适用场景为，有一组固定数量的任务，它们的执行结果分为成功或失败，当任务监视器启动（通常在一个独立的goroutine），它会接收任务传递的执行结果，当所有任务都汇报自己的执行结果后，监视器向注册的接收者广播结果/错误，并且清理自身资源。（可选）初始化一个计时器，当超时后，若仍然有任务未报告执行结果，则传递超时信息，并且清理资源
type TaskMonitor struct {
	counter     int
	stateTable  int16
	subscribers []chan<- []TaskStateSummaryMsg
	setter      chan TaskStateMsg
	overtime    int64
	mut         *sync.Mutex
}

type TaskStateMsg struct {
	Idx   int
	State bool
}

type TaskStateSummaryMsg struct {
	Overtime bool
	State    bool
}

func Report(ms []TaskStateSummaryMsg) ([]int, []bool, string) {
	o, s := make([]int, 0), make([]bool, len(ms))
	for i, m := range ms {
		if m.Overtime {
			o = append(o, i)
		} else {
			s[i] = m.State
		}
	}
	return o, s, fmt.Sprintf("Overtime: %v\nState set successfully: %v\n", o, s)
}

func NewTaskMonitor(n int, t int64) (*TaskMonitor, error) {
	if n < 0 || n > 16 {
		return nil, fmt.Errorf("the number of tasks is limited to (1~16)")
	}
	return &TaskMonitor{
		counter:     n,
		stateTable:  0,
		subscribers: make([]chan<- []TaskStateSummaryMsg, 0),
		setter:      make(chan TaskStateMsg, n),
		overtime:    t,
		mut:         new(sync.Mutex),
	}, nil
}

func (tm *TaskMonitor) Run() {
	res := make([]TaskStateSummaryMsg, tm.counter)
	ccounter := tm.counter
	if tm.overtime > 0 {
		rm := make(datastructure.Set[int])
		for i := 0; i < tm.counter; i++ {
			rm.Add(i)
		}
		curTimer := time.NewTimer(time.Duration(tm.overtime) * time.Second)
		for {
			select {
			case <-curTimer.C:
				{
					for ri := range rm {
						res[ri] = TaskStateSummaryMsg{Overtime: true}
					}

					tm.subscribe(res)
					break
				}
			case tsm := <-tm.setter:
				{
					ccounter--
					res[tsm.Idx] = TaskStateSummaryMsg{State: tsm.State}

					rm.Del(tsm.Idx)
					if ccounter == 0 {
						if b := curTimer.Stop(); !b {
							<-curTimer.C
						}

						tm.subscribe(res)
						break
					}
				}
			}
		}
	} else {
		for tms := range tm.setter {
			ccounter--
			tm.setState(tms)
			res[tms.Idx] = TaskStateSummaryMsg{State: tms.State}
			if ccounter == 0 {
				break
			}
		}

		tm.subscribe(res)
	}
}

func (tm *TaskMonitor) subscribe(summary []TaskStateSummaryMsg) {
	for _, ch := range tm.subscribers {
		ch <- summary
		close(ch)
	}
	close(tm.setter)
}

func (tm *TaskMonitor) setState(tsm TaskStateMsg) {
	var xi int16
	if tsm.State {
		xi = 1 << tsm.Idx
	}

	tm.mut.Lock()
	defer tm.mut.Unlock()

	tm.stateTable ^= xi
}

func (tm *TaskMonitor) Register(ch ...chan<- []TaskStateSummaryMsg) {
	tm.subscribers = append(tm.subscribers, ch...)
}

func (tm TaskMonitor) Send(msg TaskStateMsg) (err error) {
	defer func() {
		if xerr := recover(); xerr != nil {
			err = xerr.(error)
		}
	}()

	tm.setter <- msg
	return
}

type ExecutionStrategy int

const (
	Direct   ExecutionStrategy = 0
	AllAsOne ExecutionStrategy = 1
)

type ExecutionPriority int

const (
	EP1  ExecutionPriority = 1
	EP2  ExecutionPriority = 2
	EP3  ExecutionPriority = 3
	EP4  ExecutionPriority = 4
	EP5  ExecutionPriority = 5
	EP6  ExecutionPriority = 6
	EP7  ExecutionPriority = 7
	EP8  ExecutionPriority = 8
	EP9  ExecutionPriority = 9
	EP10 ExecutionPriority = 10
)

type TaskType interface {
	Typo() TaskExecType
}

type TaskExecType int

const (
	Exec     TaskExecType = 1
	WaitExec TaskExecType = 2
	React    TaskExecType = 3
	Cycle    TaskExecType = 4
)

type ExecTask struct{}

func (t ExecTask) Typo() TaskExecType {
	return Exec
}

type WaitExecTask struct {
	WaitTime time.Duration
}

func (t WaitExecTask) Typo() TaskExecType {
	return WaitExec
}

type ReactExecTask struct {
	Predict func() bool
}

func (t ReactExecTask) Typo() TaskExecType {
	return React
}

type CycleExecTask struct {
	Cycle time.Duration
}

func (t CycleExecTask) Typo() TaskExecType {
	return Cycle
}

type Task interface {
	Typo() TaskExecType
	Priority() ExecutionPriority
	Params() []any
	Do(chan<- string, chan<- int) (<-chan any, <-chan error)
	Clean() error
}

type BatchTask interface {
	ExecutionStrategy() ExecutionStrategy
	Tasks() []Task
}

type TaskOwner struct {
	id         int64
	taskDetail map[int64]TaskDetail
}

type TaskDetail struct {
	t     Task
	bt    BatchTask
	out   chan any
	errCh chan error
}

type TaskManager struct {
	id      int64
	rwMut   *sync.RWMutex
	idGroup *list.List

	maxTaskNum  int
	taskPq      *datastructure.ConcurrentPriorityQueue
	taskQ       *datastructure.ConcurrentQueue
	batchTaskPq *datastructure.ConcurrentPriorityQueue
	batchTaskQ  *datastructure.ConcurrentQueue
	mut         *sync.Mutex
	q1          chan Task
	q2          chan Task
	q3          chan BatchTask
	q4          chan BatchTask
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		id:      0,
		rwMut:   new(sync.RWMutex),
		idGroup: list.New(),
	}
}

func (m *TaskManager) Register() int64 {
	m.rwMut.Lock()
	defer m.rwMut.Unlock()

	r := m.id
	m.idGroup.PushBack(r)
	m.id++

	return r
}

func (m *TaskManager) UnRegister(id int64) {
	m.rwMut.Lock()
	defer m.rwMut.Unlock()

	for b := m.idGroup.Front(); b != nil; b = b.Next() {
		if b.Value.(int64) == id {
			m.idGroup.Remove(b)
			break
		}
	}
}

func (m *TaskManager) validId(id int64) bool {
	m.rwMut.RLock()
	defer m.rwMut.RUnlock()

	for b := m.idGroup.Front(); b != nil; b = b.Next() {
		if b.Value.(int64) == id {
			return true
		}
	}

	return false
}

func (m *TaskManager) ListIdGroup() {
	m.rwMut.RLock()
	defer m.rwMut.RUnlock()

	for b := m.idGroup.Front(); b != nil; b = b.Next() {
		print(b.Value.(int64), " ")
	}

	println("\b\n")
}

func (m *TaskManager) Commit(id int64, task Task) error {
	if m.validId(id) {
		return &ErrInvalidId{id}
	}
	return nil
}

func (m *TaskManager) CommitBatch(id string, batchTask BatchTask) (string, []string, error) {
	return "", nil, nil
}

func (m *TaskManager) Run() {}

type ErrInvalidId struct {
	ErrId int64
}

func (e *ErrInvalidId) Error() string {
	return fmt.Sprintf("invalid id %d", e.ErrId)
}
