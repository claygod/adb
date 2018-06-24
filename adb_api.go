package adb

// Account database
// Main
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"io/ioutil"
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
	path  string
	//queue      *Queue
	//queuesPool [256]*queue.Queue
	batcher *batcher.Batcher
	wal     *wal.Wal
	ch      chan *batcher.Task
	ch2     chan *batcher.Task
	time    *time.Time
}

func New(path string) (*Adb, error) {
	// ToDo: exists dir ?
	fileName := "start.txt"
	wal, err := wal.New(path, fileName, WalSimbolSeparator1) //newWal()
	if err != nil {
		return nil, err
	}
	ch := make(chan *batcher.Task, 1024)
	ch2 := make(chan *batcher.Task, 1024)
	//q := newQueue(sizeBucket * 16)
	b := batcher.New(wal, ch, ch2)

	adb := &Adb{
		accounts: newAccounts(),
		//answers:  newAnswers(),
		//queue:   q,
		state:   stateClosed,
		path:    path,
		batcher: b,
		wal:     wal,
		ch:      ch,
		ch2:     ch2,
		time:    &time.Time{},
	}

	b.SetBatchSize(sizeBucket) //.Start()
	return adb, nil
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

func (a *Adb) Load() {
	a.Stop()
	a.loadFromDisk()
	a.Start()
}

func (a *Adb) saveToDisk() error {
	file, err := os.Create(a.path + "adb.txt")
	if err != nil {
		return err
	}
	_, err = file.WriteString(a.accounts.Export())
	if err != nil {
		return err
	}
	return nil
}

func (a *Adb) loadFromDisk() error {
	file, err := ioutil.ReadFile(a.path + "adb.txt")
	if err != nil {
		//fmt.Println(err)
		panic(err)
		// ToDo: read snapshots & etc.
		_, err := os.Create(a.path + "adb.txt")
		if err != nil {
			return err
		}
		return nil
	}
	a.accounts.Import(string(file))
	return nil
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
