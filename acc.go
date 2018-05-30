package adb

// Account database
// Main
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"bytes"
	// "encoding/gob"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	// "time"
	// "unsafe"

	"github.com/claygod/batcher"
	"github.com/claygod/queue"
	"github.com/claygod/transaction"
)

// Hasp state
const (
	stateClosed int64 = iota
	stateOpen
)
const permitError int64 = -2147483647
const sizeBucket int = 32

type Reception struct {
	// sync.Mutex
	counter int64
	store   *Store //sync.Map
	//bucket     *Bucket
	workerStop int64
	tCore      *transaction.Core
	queue      *queue.Queue
	queuesPool [256]*queue.Queue
	batcher    *batcher.Batcher
	wal        *Wal
}

func NewReception(tc *transaction.Core) *Reception {
	wal := newWal()
	q := queue.New(sizeBucket * 32)
	// b := NewBucket()
	r := &Reception{
		store:   newStore(),
		tCore:   tc,
		queue:   q,
		batcher: batcher.New(wal),
		wal:     wal,
	}

	for i := 0; i < 256; i++ {
		r.queuesPool[i] = queue.New(sizeBucket)
	}

	go r.worker(0)
	// go r.worker(1)
	//time.Sleep(100000 * time.Microsecond)
	//go r.worker(1)
	return r
}

func (r *Reception) ExeTransaction(t *Transaction) *Answer {
	num := atomic.AddInt64(&r.counter, 1)
	r.DoTransaction(t, num)
	//time.Sleep(1 * time.Microsecond)
	return r.GetAnswer(num)
}

func (r *Reception) DoTransaction(t *Transaction, num int64) int64 { // , a **Answer
	// num := atomic.AddInt64(&r.counter, 1)
	var trGob bytes.Buffer // Stand-in for the network.

	// Create an encoder and send a value.
	/*
		enc := gob.NewEncoder(&trGob)
		err := enc.Encode(t)
		if err != nil {
			r.store.Store(num, &Answer{code: 404})
			fmt.Printf("\r\n- отбросили из-за ошибки кодирования - %d \r\n", num)
		}
	*/
	q := &Query{
		num: num,
		t:   t,
		log: trGob.Bytes(),
	}
	/*
		if !r.queuesPool[uint8(num)].PushTail(q) {
			r.store.Store(num, &Answer{code: 404})
			fmt.Printf("\r\n- отбросили - %d \r\n", num)
		}
	*/
	if !r.queue.PushTail(q) {
		r.store.Store(num, &Answer{code: 404})
		fmt.Printf("\r\n- отбросили - %d \r\n", num)
	}

	return num
}

func (r *Reception) GetAnswer(num int64) *Answer { // , a **Answer
	for { //  i := 0; i < 1500000; i++
		if a1, ok := r.store.Load(num); ok {
			go r.store.Delete(num)
			return a1 //.(*Answer)
		}
		/*
			u := i
			if u > 10000 {
				u = 10000
			}
		*/
		// time.Sleep(10 * time.Microsecond) //time.Duration(u) *
		runtime.Gosched()
	}
	fmt.Printf("\r\n- не найден - %d \r\n", num)
	return nil
}

func (r *Reception) worker(level int) {
	//var shift uint8
	var wg sync.WaitGroup
	//var an *Answer = &Answer{code: 200}
	for {
		//shift++
		//b := r.queuesPool[shift].PopHeadList(sizeBucket)
		b := r.queue.PopHeadList(sizeBucket)
		//b := r.queue.PopAll()

		if len(b) == 0 {
			// fmt.Printf("\r\n- воркер получил пустую очередь- \r\n")
			runtime.Gosched()
			// time.Sleep(30 * time.Microsecond)
			continue
		}
		if len(b) > 5 {
			// fmt.Printf("\r\n- размер очереди- %d \r\n", len(b))
		}
		//var logBuf bytes.Buffer
		for _, q1 := range b {
			q := q1.(*Query)
			wg.Add(1)
			go r.handler(q.t, &wg, q.num, q.log)
			//r.store.Store(q.num, an) //
			// тут сохраняем лог
		}
		wg.Wait()

		if !r.wal.Save() || r.workerStop == 1 {
			return
		}
		//shift++
	}
}

func (r *Reception) handler(t *Transaction, wg *sync.WaitGroup, num int64, log []byte) {
	r.store.Store(num, &Answer{code: 200})
	// if ok
	r.wal.Log(num, log)
	// dummy
	wg.Done()
}

type Answer struct {
	code int64
}

type Query struct {
	num int64
	t   *Transaction
	log []byte
	// a   **Answer
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

type Store struct {
	sync.Mutex
	arr map[int64]*Answer
}

func newStore() *Store {
	return &Store{
		arr: make(map[int64]*Answer),
	}
}

func (s *Store) Load(key int64) (*Answer, bool) {
	s.Lock()
	// defer s.RUnlock()
	a, ok := s.arr[key]
	s.Unlock()
	return a, ok
}

func (s *Store) Store(key int64, a *Answer) {
	s.Lock()
	s.arr[key] = a
	s.Unlock()
}

func (s *Store) Delete(key int64) {
	s.Lock()
	delete(s.arr, key)
	s.Unlock()
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
