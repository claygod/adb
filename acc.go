package accounter

// Accounter
// Main
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	//"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type Reception struct {
	sync.Mutex
	counter    int64
	bucket     *Bucket
	bkt        unsafe.Pointer
	workerStop int64
}

func NewReception() *Reception {
	b := NewBucket()
	r := &Reception{
		bucket: NewBucket(),
		bkt:    unsafe.Pointer(b),
	}
	go r.worker()
	return r
}

func (r *Reception) DoTransaction(t *Transaction, a **Answer) {
	fmt.Print("###### DOOOOOOOOOOOOOOOOOOOOOOOOOOOOO Do Do 1 ######\r\n")
	atomic.AddInt64(&r.counter, 1)
	q := &Query{
		t: t,
		a: a,
	}
	r.bucket.AddQuery(q)
	fmt.Print("###### DOOOOOOOOOOOOOOOOOOOOOOOOOOOOO Do Do 2 ######\r\n")
	return
}

func (r *Reception) getBucket() *Bucket {
	r.Lock()
	oldBucket := r.bucket
	r.bucket = NewBucket()
	r.Unlock()
	return oldBucket
	/*
		newBucket := NewBucket()
		addr := unsafe.Pointer(&oldBucket)
		addr2 := unsafe.Pointer(&newBucket)
		fmt.Printf(" ---   %s >> %s\r\n", addr, addr2)
		atomic.StorePointer(&addr, addr2) //
		return oldBucket
	*/
}

func (r *Reception) worker() {
	for u := 0; u < 5000; u++ {
		// fmt.Print("###### worker ######\r\n")
		b := r.getBucket()
		//fmt.Printf(" ---   %d >> %s\r\n", len(b.arr), &b)
		b.Lock()
		if len(b.arr) == 0 {
			//fmt.Print("###### sleep ######\r\n")
			//runtime.Gosched()
			time.Sleep(10 * time.Millisecond)
			b.Unlock()
			continue
		}
		//fmt.Print("###### NO sleep ######\r\n")

		var wg sync.WaitGroup
		for _, q := range b.arr {
			fmt.Printf("++++++++++++++++++++++++++  %v\r\n", q)
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

type Transaction struct { // dummy
}

type Answer struct {
	code int64
}

type Query struct {
	t *Transaction
	a **Answer
}

type Bucket struct {
	sync.Mutex
	arr []*Query
}

func NewBucket() *Bucket {
	return &Bucket{
		arr: make([]*Query, 0, 10),
	}
}

func (b *Bucket) AddQuery(q *Query) {
	b.Lock()
	b.arr = append(b.arr, q)
	b.Unlock()
}
