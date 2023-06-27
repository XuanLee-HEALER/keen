package multitask

import (
	"container/list"
	"sync"
)

// PubChan 广播通知器，当需要在一个goroutine中同时通知多个goroutine时使用，不允许复制
type PubChan struct {
	chs   *list.List
	l     sync.Mutex
	next  int
	rec   map[int]*list.Element
	limit int
}

// NewPubChan 创建一个新的广播控制器
//
// in: i 指定允许同时存在的接收广播内容的数量
func NewPubChan(i int) *PubChan {
	return &PubChan{
		chs:   list.New(),
		l:     sync.Mutex{},
		rec:   make(map[int]*list.Element),
		limit: i,
	}
}

// Pub 广播
func (c *PubChan) Pub() {
	c.l.Lock()
	defer c.l.Unlock()
	for head := c.chs.Front(); head != nil; head = head.Next() {
		head.Value.(chan struct{}) <- struct{}{}
	}
}

// Sub 工作goroutine向控制器注册
//
// in:
//
// out: int 注册id，注销时使用 | <-chan struct{} 通知监听channel | bool 是否注册成功
func (c *PubChan) Sub() (int, <-chan struct{}, bool) {
	c.l.Lock()
	defer c.l.Unlock()
	if c.next < c.limit {
		tc := make(chan struct{}, 1)
		c.chs.PushBack(tc)
		c.rec[c.next] = c.chs.Back()
		c.next++
		return c.next, tc, true
	}
	return 0, nil, false
}

// UnSub 工作goroutine向控制器注销
//
// in: id 注册成功时返回的id
//
// out: bool 是否注销失败
func (c *PubChan) UnSub(id int) bool {
	c.l.Lock()
	defer c.l.Unlock()
	if v, ok := c.rec[id]; ok {
		c.chs.Remove(v)
		delete(c.rec, id)
		return true
	}
	return false
}
