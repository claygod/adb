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
	"unsafe"

	"github.com/claygod/transaction"
)

// Hasp state
const (
	stateClosed int64 = iota
	stateOpen
)
const permitError int64 = -2147483647

type Reception struct {
	// sync.Mutex
	counter    int64
	store      sync.Map
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
	num := r.DoTransaction(t)
	return r.GetAnswer(num)
}

func (r *Reception) DoTransaction(t *Transaction) int64 { // , a **Answer
	num := atomic.AddInt64(&r.counter, 1)
	q := &Query{
		num: num,
		t:   t,
		//a:   a,
	}
	for {
		if r.bucket.AddQuery(q) {
			break
		}
		runtime.Gosched()
	}
	return num
}

func (r *Reception) GetAnswer(num int64) *Answer { // , a **Answer
	for i := 0; i < 9500; i++ {
		if a1, ok := r.store.Load(num); ok {
			return a1.(*Answer)
		}
		runtime.Gosched()
		u := i
		if u > 10000 {
			u = 10000
		}
		time.Sleep(time.Duration(u) * 1 * time.Microsecond) //
	}
	fmt.Printf("\r\n- не найден - %d \r\n", num)
	return nil
}

func (r *Reception) getBucket() *Bucket {
	r.bucket.Close()
	oldBucket := r.bucket
	r.bucket = NewBucket()
	return oldBucket
}

func (r *Reception) worker(level int) {
	for {
		b := r.getBucket()
		ln := len(b.arr)
		if ln == 0 {
			time.Sleep(1 * time.Microsecond)
			continue
		}

		var wg sync.WaitGroup
		for _, q := range b.arr {
			wg.Add(1)
			go r.doTr(q.t, &wg)
			r.store.Store(q.num, &Answer{code: 200})
			// тут сохраняем лог
		}
		wg.Wait()

		if r.workerStop == 1 {
			return
		}
	}
}

func (r *Reception) doTr(t *Transaction, wg *sync.WaitGroup) {
	// dummy
	wg.Done()
}

type Answer struct {
	code int64
}

type Query struct {
	num int64
	t   *Transaction
	// a   **Answer
}

type Bucket struct {
	sync.Mutex
	counter int64
	arr     []*Query
}

func NewBucket() *Bucket {
	return &Bucket{
		arr: make([]*Query, 0, 1000),
	}
}

func (b *Bucket) AddQuery(q *Query) bool {
	if !b.Catch() {
		return false
	}
	defer b.Throw()
	b.Lock()
	b.arr = append(b.arr, q)
	b.Unlock()

	return true
}

func (b *Bucket) Catch() bool {
	if atomic.AddInt64(&b.counter, 1) < 1 {
		atomic.AddInt64(&b.counter, -1)
		return false
	}
	return true
}

func (b *Bucket) Throw() {
	atomic.AddInt64(&b.counter, -1)
}

func (b *Bucket) Close() {
	atomic.AddInt64(&b.counter, permitError)
	for {
		if atomic.LoadInt64(&b.counter) <= permitError {
			return
		}
	}
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
