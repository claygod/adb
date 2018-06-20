package batchqueue

// Batchqueue
// API
// Copyright © 2016 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>
import (
	//"runtime"
	//"fmt"
	"sync/atomic"
)

//"unsafe"

const countBatches uint64 = 256

const (
	codeDoneNow = iota
	codeDoneBefore
	codeNotDone
)

// const sizeBatch int64 = 1024

type Batchqueue struct {
	cursor uint64 // младшая tail, старшая head
	//counterPush uint64
	//counterHead uint64
	//counterTail uint64
	//batch     *Batch
	batches   [countBatches]*Batch
	sizeBatch int64
}

// New - create new Batchqueue.
func New(size int64) *Batchqueue {
	q := &Batchqueue{sizeBatch: size}
	q.cursor = 2 // !!! важно
	for i := uint64(0); i < countBatches; i++ {
		q.batches[i] = newBatch(size)
	}
	return q
}

func (q *Batchqueue) incrementTail(cursorTailCur uint8) int {
	for {
		oldCursor := atomic.LoadUint64(&q.cursor)
		cursorHead := uint8(oldCursor >> 32)
		cursorTail := uint8(oldCursor) + 1

		if cursorTailCur != cursorTail-1 {
			//fmt.Println("TAIL: ", cursorTailCur, cursorTail)
			return codeDoneBefore
		}

		delta := int(cursorHead) - int(cursorTail)
		//fmt.Println("Delta: ", delta)
		if -2 < delta && 10 > delta {

			// if cursorTail < cursorHead && cursorTail+10 >= cursorHead{
			return codeNotDone
		}
		newCursor := uint64(cursorHead)<<32 + uint64(cursorTail)
		if atomic.CompareAndSwapUint64(&q.cursor, oldCursor, newCursor) {
			return codeDoneNow
		}
	}
}

func (q *Batchqueue) incrementHead(cursorHeadCur uint8) int {
	// голова может толкать (стдвигать хвост- с ограничениями а хвост толкать не может
	for {
		var newCursor uint64
		oldCursor := atomic.LoadUint64(&q.cursor)
		cursorHead := uint8(oldCursor >> 32)
		cursorTail := uint8(oldCursor)
		if cursorHeadCur != cursorHead {
			return codeDoneBefore
		}
		//delta := int(cursorTail+1) - int(cursorHead)
		if cursorHead+1 == cursorTail {
			// if cursorTail < cursorHead && cursorTail+10 >= cursorHead{
			//return codeNotDone
			newCursor = uint64(cursorHead+1)<<32 + uint64(cursorTail+1)
		}
		newCursor = uint64(cursorHead+1)<<32 + uint64(cursorTail)
		if atomic.CompareAndSwapUint64(&q.cursor, oldCursor, newCursor) {
			return codeDoneNow
		}
	}
}

func (q *Batchqueue) incrementHead2(cursorHeadCur uint8) int {
	for {
		var newCursor uint64
		oldCursor := atomic.LoadUint64(&q.cursor)
		cursorHead := uint8(oldCursor >> 32)
		cursorTail := uint8(oldCursor)
		//delta := int(cursorTail+1) - int(cursorHead)
		if cursorHead+1 == cursorTail {
			newCursor = uint64(cursorHead+1)<<32 + uint64(cursorTail+1)
		}
		newCursor = uint64(cursorHead+1)<<32 + uint64(cursorTail)
		if atomic.CompareAndSwapUint64(&q.cursor, oldCursor, newCursor) {
			return codeDoneNow
		}
	}
}

func (q *Batchqueue) Push(item interface{}) bool {
	for u := 0; u < 5; u++ {
		//curCounterTail := atomic.LoadUint64(&q.counterTail)

		oldCursor := atomic.LoadUint64(&q.cursor)
		//cursorHead := uint8(oldCursor >> 32)
		cursorTail := uint8(oldCursor)

		if q.batches[cursorTail].add(item) {
			return true
		} else {
			//if !q.incrementTail() {
			//	return false
			//}
			//fmt.Println("No-add-item ", item)
			switch q.incrementTail(cursorTail) {
			case codeDoneBefore:
				//fmt.Println("A codeDoneBefore")
				continue
			case codeDoneNow:
				//fmt.Println("A codeDoneNow")
				//q.batches[cursorHead-1] = newBatch(q.sizeBatch) // +countBatches/2
				continue //q.batches[cursorHead]
			case codeNotDone:
				//fmt.Println("A codeNotDone")
				return false //nil
			}
		}
		//	runtime.Gosched()
	}
	return false
}

func (q *Batchqueue) PopBatch() *Batch {
	for {
		oldCursor := atomic.LoadUint64(&q.cursor)
		cursorHead := uint8(oldCursor >> 32)
		//cursorTail := uint8(oldCursor)
		switch q.incrementHead(cursorHead) {
		case codeDoneBefore:
			continue
		case codeDoneNow:
			q.batches[cursorHead-1] = newBatch(q.sizeBatch) // +countBatches/2
			return q.batches[cursorHead]
			//case codeNotDone:
			//	return nil
		}
	}

	//if !q.incrementHead(cursorHead) {
	//	return nil
	//}
	/*
		curCounter := atomic.AddUint64(&q.counterHead, 1)
		atomic.StoreInt64(&q.batches[curCounter-1].limit, 0)
		atomic.AddUint64(&q.counterTail, 1)
		q.batches[uint8(curCounter-2)] = newBatch(q.sizeBatch) // +countBatches/2
		return q.batches[curCounter-1]
	*/
}

type Batch struct {
	limit int64
	size  int64
	data  []interface{}
}

func newBatch(limit int64) *Batch {
	b := &Batch{
		limit: limit,
		data:  make([]interface{}, limit, limit),
	}
	return b
}

func (b *Batch) add(item interface{}) bool {
	shift := atomic.AddInt64(&b.limit, -1)
	if shift < 0 {
		// atomic.AddInt64(&b.limit, 1)
		return false
	}
	b.data[b.size] = item
	atomic.AddInt64(&b.size, 1)
	return true
}
