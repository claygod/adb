package adb

// Account database
// Main
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	// "time"

	"github.com/claygod/adb/account"
	"github.com/claygod/adb/batcher"
	"github.com/claygod/adb/wal"
)

type Reception struct {
	counter  int64
	accounts *Accounts
	//answers    *Answers
	workerStop int64
	queue      *Queue
	//queuesPool [256]*queue.Queue
	batcher *batcher.Batcher
	wal     *wal.Wal
	ch      chan *batcher.Task
}

func NewReception(patch string) (*Reception, error) {
	wal, err := wal.New(patch, walSeparator) //newWal()
	if err != nil {
		return nil, err
	}
	ch := make(chan *batcher.Task, 256)
	q := newQueue(sizeBucket * 16)
	b := batcher.New(wal, q, ch)

	r := &Reception{
		accounts: newAccounts(),
		//answers:  newAnswers(),
		queue:   q,
		batcher: b,
		wal:     wal,
		ch:      ch,
	}
	//b.SetBatchSize(sizeBucket).Start(batcher.Sync)
	b.SetBatchSize(sizeBucket).StartChain(batcher.Sync)
	//b.SetBatchSize(sizeBucket * 8).StartChain(batcher.Sync)

	//for i := 0; i < 256; i++ {
	//	r.queuesPool[i] = queue.New(sizeBucket)
	//}
	return r, nil
}

func (r *Reception) ExeTransaction(order *Order) *Answer {
	num := atomic.AddInt64(&r.counter, 1)
	ans := r.DoTransaction(order, num)
	runtime.Gosched()
	//time.Sleep(1 * time.Microsecond)
	return r.GetAnswer(num, ans)
}

func (r *Reception) DoTransaction(order *Order, num int64) *Answer {
	//fmt.Println(" @001@ ", num)
	ans := &Answer{code: 0}

	//qClosure := r.getClosure(logBytes, order, num, ans)
	tsk := r.getTask(order, ans)

	r.ch <- tsk

	//if !r.queue.AddTransaction(&qClosure) {
	//	ans.code = 404
	//}

	return ans
}

func (r *Reception) GetAnswer(num int64, ans *Answer) *Answer { // , a **Answer
	runtime.Gosched()
	// return &Answer{code: 404}
	//fmt.Println(" @031@ ", num)
	for i := 0; ; i++ { //  i := 0; i < 1500000; i++
		//fmt.Println(" @032@ ", num)
		if atomic.LoadInt64(&ans.code) > 0 {
			//go r.answers.Delete(num)
			//fmt.Println(" @032@ ура, ответ получен! на шаге ", i, " код=", ans.code)
			return ans //.(*Answer)
		}
		//time.Sleep(1000000 * time.Microsecond) //time.Duration(u) *
		runtime.Gosched()
	}
	fmt.Printf("\r\n- не найден - %d \r\n", num)
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

type Wal struct {
	sync.Mutex
}

func newWal() *Wal {
	return &Wal{}
}

func (w *Wal) Log(key int64, b []byte) error {
	return nil
}

func (w *Wal) Save() error {
	return nil
}

func (w *Wal) Close() error {
	return nil
}
