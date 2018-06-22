package adb

// Account database
// Main
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"runtime"
	//"sync"
	"sync/atomic"
	"time"

	"github.com/claygod/adb/account"
	"github.com/claygod/adb/batcher"
	"github.com/claygod/adb/wal"
)

type Adb struct {
	counter  int64
	accounts *Accounts
	//answers    *Answers
	workerStop int64
	//queue      *Queue
	//queuesPool [256]*queue.Queue
	batcher *batcher.Batcher
	wal     *wal.Wal
	ch      chan *batcher.Task
	ch2     chan *batcher.Task
	time    *time.Time
}

func New(patch string) (*Adb, error) {
	fileName := "start.txt"
	wal, err := wal.New(patch, fileName, WalSimbolSeparator1) //newWal()
	if err != nil {
		return nil, err
	}
	ch := make(chan *batcher.Task, 1024)
	ch2 := make(chan *batcher.Task, 1024)
	//q := newQueue(sizeBucket * 16)
	b := batcher.New(wal, ch, ch2)

	r := &Adb{
		accounts: newAccounts(),
		//answers:  newAnswers(),
		//queue:   q,
		batcher: b,
		wal:     wal,
		ch:      ch,
		ch2:     ch2,
		time:    &time.Time{},
	}
	//b.SetBatchSize(sizeBucket).Start(batcher.Sync)
	b.SetBatchSize(sizeBucket).StartChain(batcher.Sync)
	//b.SetBatchSize(sizeBucket * 8).StartChain(batcher.Sync)

	//for i := 0; i < 256; i++ {
	//	r.queuesPool[i] = queue.New(sizeBucket)
	//}
	return r, nil
}

func (a *Adb) ExeTransaction(order *Order) *Answer {
	num := atomic.AddInt64(&a.counter, 1)
	ans := a.DoTransaction(order, num)
	// runtime.Gosched()
	//time.Sleep(1 * time.Microsecond)
	return a.GetAnswer(num, ans)
}

func (a *Adb) DoTransaction(order *Order, num int64) *Answer {
	ans := &Answer{code: 0}
	tsk := a.getTask(order, ans)
	a.ch <- tsk
	return ans
}

func (a *Adb) GetAnswer(num int64, ans *Answer) *Answer { // , a **Answer
	runtime.Gosched()
	for i := 0; ; i++ { //  i := 0; i < 1500000; i++
		if atomic.LoadInt64(&ans.code) > 0 {
			return ans //.(*Answer)
		}
		//time.Sleep(1000000 * time.Microsecond) //time.Duration(u) *
		runtime.Gosched()
	}
	fmt.Printf("\r\n- не найден - %d \r\n", num) // ToDo !!
	return nil
}

type Answer struct {
	code    int64
	balance map[string]map[string]account.Balance
}

type Query struct {
	num   int64
	order *Order
	log   []byte
}

type Order struct {
	Hash    string
	Debit   []*Part
	Credit  []*Part
	Block   []*Part
	Unblock []*Part
}

type Part struct {
	Id     string
	Key    string
	Amount uint64
}

/*
type Wal struct {
	sync.Mutex
}

func newWal() *Wal {
	return &Wal{}
}

func (w *Wal) Log(s string) error {
	return nil
}

func (w *Wal) Save() error {
	return nil
}

func (w *Wal) Close() error {
	return nil
}
*/
