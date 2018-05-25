package adb

// Account database
// Main
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"runtime"
	"sync"
	//"sync/atomic"
	"time"
	"unsafe"

	"github.com/claygod/transaction"
)

type Reception struct {
	sync.Mutex
	counter    int64
	bucket     *Bucket
	bkt        unsafe.Pointer
	workerStop int64
	tCore      *transaction.Core
}

func NewReception(tc *transaction.Core) *Reception {
	b := NewBucket()
	r := &Reception{
		bucket: NewBucket(),
		bkt:    unsafe.Pointer(b),
		tCore:  tc,
	}
	go r.worker(0)
	return r
}

func (r *Reception) ExeTransaction(t *Transaction) *Answer {
	p := &Answer{}
	var a **Answer = &p
	p = nil

	go r.DoTransaction(t, a)
	r.GetAnswer(777, a)
	return *a
}

func (r *Reception) DoTransaction(t *Transaction, a **Answer) {
	//fmt.Print("###### DOOOOOOOOOOOOOOOOOOOOOOOOOOOOO Do Do 1 ######\r\n")
	// atomic.AddInt64(&r.counter, 1)
	q := &Query{
		t: t,
		a: a,
	}
	r.bucket.AddQuery(q)
	//fmt.Print("###### DOOOOOOOOOOOOOOOOOOOOOOOOOOOOO Do Do 2 ######\r\n")
	return
}

func (r *Reception) GetAnswer(num int, a **Answer) *Answer {
	for i := 0; i < 5000; i++ {
		if *a != nil {
			//fmt.Print("ok ")
			return *a
		}
		runtime.Gosched()
		time.Sleep(time.Duration(i) * 10 * time.Microsecond)
	}
	fmt.Printf("\r\n- не найден - %d \r\n", num)
	return nil
}

func (r *Reception) getBucket() *Bucket {
	r.Lock()
	oldBucket := r.bucket
	r.bucket = NewBucket()
	r.Unlock()
	return oldBucket
}

func (r *Reception) worker(level int) {
	for { // u := 0; u < 5000; u++
		b := r.getBucket()
		b.Lock()
		ln := len(b.arr)
		if ln == 0 {
			if level != 0 {
				b.Unlock()
				return
			}
			time.Sleep(1 * time.Millisecond)
			b.Unlock()
			continue
		} else if ln > 100 {
			//go r.worker(level + 1)
		}

		// fmt.Printf("len(w)=%d \r\n", len(b.arr))
		var wg sync.WaitGroup
		for _, q := range b.arr {
			wg.Add(1)
			go r.doTr(q.t, &wg)
			*q.a = &Answer{code: 200}
			// тут сохраняем лог
		}
		wg.Wait()

		b.Unlock()
		if r.workerStop == 1 {
			return
		}
	}
}

func (r *Reception) doTr(t *Transaction, wg *sync.WaitGroup) {
	// dummy
	wg.Done()
}

/*
type Trans struct { // dummy
}
*/
type Answer struct {
	code int64
}

type Query struct {
	t *Transaction
	// t2 *transaction.Transaction
	a **Answer
}

type Bucket struct {
	sync.Mutex
	arr []*Query
}

func NewBucket() *Bucket {
	return &Bucket{
		arr: make([]*Query, 0, 1000),
	}
}

func (b *Bucket) AddQuery(q *Query) {
	b.Lock()
	b.arr = append(b.arr, q)
	b.Unlock()
}

type Transaction struct {
	plus  []*Part
	minus []*Part
}

type Part struct {
	id     int64
	key    string
	amount int64
}
