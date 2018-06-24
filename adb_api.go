package adb

// Account database
// Main
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"os"
	"runtime"
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
	state int64
	patch string
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
		state:   stateClosed,
		patch:   patch,
		batcher: b,
		wal:     wal,
		ch:      ch,
		ch2:     ch2,
		time:    &time.Time{},
	}

	b.SetBatchSize(sizeBucket) //.Start()

	return r, nil
}

func (a *Adb) Start() {
	a.batcher.Start()
	atomic.StoreInt64(&a.state, stateOpen)
}

func (a *Adb) Stop() {
	atomic.StoreInt64(&a.state, stateClosed)
	a.batcher.Stop()
	a.saveToDisk()
}

func (a *Adb) Save() {
	a.Stop()
	a.saveToDisk()
	a.Start()
}

func (a *Adb) saveToDisk() error {
	file, err := os.Create(a.patch + "adb.txt")
	if err != nil {
		return err
	}
	_, err = file.WriteString(a.accounts.Export())
	if err != nil {
		return err
	}
	return nil
}

func (a *Adb) load() {
}

func (a *Adb) Transaction(order *Order) (*Answer, error) {
	//fmt.Println(order)
	num := atomic.AddInt64(&a.counter, 1)
	ans := a.doTransaction(order, num)
	// runtime.Gosched()
	//time.Sleep(1 * time.Microsecond)
	//fmt.Println(a.getAnswer(num, ans))
	return a.getAnswer(num, ans), nil
}

func (a *Adb) doTransaction(order *Order, num int64) *Answer {
	ans := &Answer{code: 0}
	tsk := a.getTask(order, ans)
	a.ch <- tsk
	return ans
}

func (a *Adb) getAnswer(num int64, ans *Answer) *Answer { // , a **Answer
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

/*
type Query struct {
	num   int64
	order *Order
	log   []byte
}
*/

type Answer struct {
	code    int64
	balance map[string]map[string]account.Balance
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
