package adb

// Account database
// Main
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/claygod/adb/account"
	"github.com/claygod/adb/batcher"
)

type Reception struct {
	counter    int64
	accounts   *Accounts
	answers    *Answers
	workerStop int64
	queue      *Queue
	//queuesPool [256]*queue.Queue
	batcher *batcher.Batcher
	wal     *Wal
}

func NewReception() *Reception {
	wal := newWal()
	q := newQueue(sizeBucket * 32)
	b := batcher.New(wal, q)
	r := &Reception{
		accounts: newAccounts(),
		answers:  newAnswers(),
		queue:    q,
		batcher:  b,
		wal:      wal,
	}
	b.SetBatchSize(sizeBucket * 8).Start(batcher.Sync)

	//for i := 0; i < 256; i++ {
	//	r.queuesPool[i] = queue.New(sizeBucket)
	//}
	return r
}

func (r *Reception) ExeTransaction(order *Order) *Answer {
	num := atomic.AddInt64(&r.counter, 1)
	r.DoTransaction(order, num)
	//time.Sleep(1 * time.Microsecond)
	return r.GetAnswer(num)
}

func (r *Reception) DoTransaction(order *Order, num int64) {
	fmt.Println(" @001@ ", num)

	logBytes, err := r.orderToLog(order)
	if err != nil {
		fmt.Println(" - ошибка кодирования лога ", num, err)
		r.answers.Store(num, &Answer{code: 404}) // отрицательный ответ
		return
	}
	qClosure := r.getClosure(logBytes, order, num)
	fmt.Println(" @002@ ", num)
	if !r.queue.AddTransaction(&qClosure) {
		r.answers.Store(num, &Answer{code: 404})
		fmt.Printf("\r\n- отбросили ---- %d \r\n", num)
	}
	fmt.Println(" @003@ ", num)
	//return 1
	return
}

func (r *Reception) GetAnswer(num int64) *Answer { // , a **Answer
	fmt.Println(" @031@ ", num)
	for { //  i := 0; i < 1500000; i++
		fmt.Println(" @032@ ", num)
		if a1, ok := r.answers.Load(num); ok {
			go r.answers.Delete(num)
			fmt.Println(" @032@ ура, ответ получен! я", num)
			return a1 //.(*Answer)
		}
		/*
			u := i
			if u > 10000 {
				u = 10000
			}
		*/
		time.Sleep(1000000 * time.Microsecond) //time.Duration(u) *
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

func (w *Wal) Log(key int64, b []byte) {
}

func (w *Wal) Save() bool {
	return true
}
