package queue

// Queue
// API
// Copyright © 2016-2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	// "fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

const sizeBlockDefault int = 1000
const sizeQueueMax int = 100000
const trialLimit int = 20000000

// Queue - main struct.
type Queue struct {
	m         *sync.Mutex
	hasp      int32
	db        []interface{}
	dbReserve []interface{}
	head      int
	tail      int
	sizeQueue int
	sizeBlock int
	sizeIn    int
}

// New - create new queue.
// The optional argument: the initial size of the queue.
func New(args ...int) *Queue {
	var sizeBlock int
	if len(args) > 0 {
		sizeBlock = args[0]
	} else {
		sizeBlock = sizeBlockDefault
	}
	q := Queue{
		m:         &sync.Mutex{},
		hasp:      0,
		db:        make([]interface{}, sizeBlock),
		dbReserve: make([]interface{}, sizeBlock),
		head:      sizeBlock / 2,
		tail:      sizeBlock / 2,
		sizeQueue: sizeBlock,
		sizeBlock: sizeBlock, // nil,
		sizeIn:    sizeBlock,
	}
	// q.unlock() // q.hasp = 0
	return &q
}

// PushTail - Insert element in the tail queue
func (q *Queue) PushTail(n interface{}) bool {
	// q.lock()
	q.m.Lock()
	defer q.m.Unlock()
	if q.sizeQueue >= sizeQueueMax { //  || !q.lock()
		//q.m.Unlock()
		//q.unlock()
		return false
	}
	q.db[q.tail] = n
	q.tail++
	if q.tail >= q.sizeQueue {
		q.db = append(q.db, make([]interface{}, q.sizeBlock)...)
		q.sizeQueue += q.sizeBlock
	}
	//q.unlock() // q.hasp = 0
	//q.m.Unlock()
	return true
}

// PushHead - Paste item in the queue head
func (q *Queue) PushHead(n interface{}) bool {
	q.m.Lock()
	defer q.m.Unlock()
	if q.sizeQueue >= sizeQueueMax { //  || !q.lock()
		//q.m.Unlock()
		return false
	}
	q.head--
	if q.head == 0 {
		newDb := make([]interface{}, q.sizeQueue+q.sizeBlock)
		copy(newDb[q.sizeBlock:], q.db)
		q.db = newDb
		q.head += q.sizeBlock
		q.tail += q.sizeBlock
		q.sizeQueue = q.sizeQueue + q.sizeBlock
	}
	q.db[q.head] = n
	//q.unlock() // q.hasp = 0
	//q.m.Unlock()
	return true
}

// PopHead - Get the first element of the queue
func (q *Queue) PopHead() (interface{}, bool) {
	q.m.Lock()
	defer q.m.Unlock()
	var n interface{}
	//if !q.lock() {
	//	return n, false
	//}
	if q.tail == q.head {
		//q.unlock() // q.hasp = 0
		return n, false
	}
	n, q.db[q.head] = q.db[q.head], nil
	q.head++
	if q.head == q.tail { //  && q.sizeQueue >= q.sizeBlock*3
		q.clean()
	}
	//q.unlock() // q.hasp = 0
	return n, true
}

func (q *Queue) PopHeadList(num int) []interface{} {
	//q.lock()
	q.m.Lock()
	defer q.m.Unlock()
	//if !q.lock() {
	//	return make([]interface{}, 0), false
	//}
	if q.tail == q.head {
		//q.unlock() // q.hasp = 0
		//q.m.Unlock()
		//q.unlock()
		return make([]interface{}, 0)
	}
	end := q.head + num
	if end > q.tail {
		end = q.tail
	}
	out := make([]interface{}, end-q.head)
	copy(out, q.db[q.head:end])
	q.head = end
	if q.head == q.tail { //  && q.sizeQueue >= q.sizeBlock*3
		q.clean()
	}
	// q.unlock() // q.hasp = 0
	//q.m.Unlock()
	//q.unlock()
	return out
}

func (q *Queue) PopAll() []interface{} {
	ndb := make([]interface{}, q.sizeIn)
	q.m.Lock()
	defer q.m.Unlock()
	out := q.db[q.head:q.tail]

	q.hasp = 0
	q.db = ndb
	q.head = q.sizeIn / 2
	q.tail = q.sizeIn / 2
	q.sizeQueue = q.sizeIn
	q.sizeBlock = q.sizeIn
	//q.m.Unlock()
	return out
}

// PopTail - Get the item from the queue tail
func (q *Queue) PopTail() (interface{}, bool) {
	var n interface{}
	//if !q.lock() {
	//	return n, false
	//}
	q.m.Lock()
	defer q.m.Unlock()
	if q.head == q.tail {
		//q.unlock() // q.hasp = 0
		return n, false
	}
	q.tail--
	n, q.db[q.tail] = q.db[q.tail], nil
	if q.head == q.tail { // && q.sizeQueue >= q.sizeBlock*3
		q.clean()
	}
	//q.unlock() // q.hasp = 0
	return n, true
}

// LenQueue - The number of elements in the queue
func (q *Queue) LenQueue() int {
	q.m.Lock()
	defer q.m.Unlock()
	//q.lock()
	ln := q.tail - q.head
	//q.unlock() // q.hasp = 0
	return ln
}

// SizeQueue - The size reserved for queue
func (q *Queue) SizeQueue() int {
	q.m.Lock()
	defer q.m.Unlock()
	//q.lock()
	ln := q.sizeQueue
	//q.unlock() // q.hasp = 0
	return ln
}

func (q *Queue) cleanAlternative() {
	// fmt.Print("\r\n------------ CLEAN!!\r\n")
	if q.dbReserve == nil {
		q.db = make([]interface{}, q.sizeBlock)
	} else {
		q.db = q.dbReserve // make([]interface{}, q.sizeBlock)
	}

	q.head = q.sizeBlock / 2
	q.tail = q.sizeBlock / 2
	q.sizeQueue = q.sizeBlock
	q.dbReserve = nil
	go q.genDbReserve()
}

// cleanAndReplace - Resetting the queue (not thread-safe, is called only after the lock)
func (q *Queue) cleanAndReplace() {
	q.db = make([]interface{}, q.sizeBlock)
	q.head = q.sizeBlock / 2
	q.tail = q.sizeBlock / 2
	q.sizeQueue = q.sizeBlock
}

func (q *Queue) clean() {
	if q.sizeQueue >= sizeQueueMax/2 { // q.sizeBlock*3
		q.db = make([]interface{}, q.sizeBlock)
		q.head = q.sizeBlock / 2
		q.tail = q.sizeBlock / 2
		q.sizeQueue = q.sizeBlock
	} else {
		q.head = q.sizeQueue / 2
		q.tail = q.sizeQueue / 2
	}

}

func (q *Queue) genDbReserve() {
	q.dbReserve = make([]interface{}, q.sizeBlock)

}

// lock - block queue
func (q *Queue) lock() bool {
	for { // i := trialLimit; i > 0; i--
		if q.hasp == 0 && atomic.CompareAndSwapInt32(&q.hasp, 0, 1) {
			break
		}
		//if i == 0 {
		//	return false
		//}
		runtime.Gosched()
	}
	return true
}

func (q *Queue) unlock() {
	atomic.StoreInt32(&q.hasp, 0)
	// q.hasp = 0
}
