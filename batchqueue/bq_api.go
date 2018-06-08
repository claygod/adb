package batchqueue

// Batchqueue
// API
// Copyright © 2016 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>
import (
	"runtime"
	"sync/atomic"
)

//"unsafe"

const countBatches uint64 = 256

// const sizeBatch int64 = 1024

type Batchqueue struct {
	cursor      uint64 // младшая tail, старшая head
	counterPush uint64
	counterHead uint64
	counterTail uint64
	//batch     *Batch
	batches   [countBatches]*Batch
	sizeBatch int64
}

// New - create new Batchqueue.
func New(size int64) *Batchqueue {
	q := &Batchqueue{sizeBatch: size}
	for i := uint64(0); i < countBatches; i++ {
		q.batches[i] = newBatch(size)
	}
	return q
}

func (q *Batchqueue) incrementTail() bool {
	for {
		oldCursor := atomic.LoadUint64(&q.cursor)
		cursorHead := uint8(oldCursor >> 32)
		cursorTail := uint8(oldCursor) + 1
		delta := int(cursorTail) - int(cursorHead)
		if 1 < delta || delta < -10 {
			// if cursorTail < cursorHead && cursorTail+10 >= cursorHead{
			return false
		}
		newCursor := uint64(cursorHead)<<32 + uint64(cursorTail)
		if atomic.CompareAndSwapUint64(&q.cursor, oldCursor, newCursor) {
			return true
		}
	}
}

func (q *Batchqueue) incrementHead() bool {
	// голова может толкать (стдвигать хвост- с ограничениями а хвост толкать не может
	for {
		oldCursor := atomic.LoadUint64(&q.cursor)
		cursorHead := uint8(oldCursor >> 32)
		cursorTail := uint8(oldCursor)
		delta := int(cursorTail+1) - int(cursorHead)
		if cursorHead+1 == cursorTail {
			// if cursorTail < cursorHead && cursorTail+10 >= cursorHead{
			return false
		}
		newCursor := uint64(cursorHead)<<32 + uint64(cursorTail)
		if atomic.CompareAndSwapUint64(&q.cursor, oldCursor, newCursor) {
			return true
		}
	}
}

func (q *Batchqueue) Push(item interface{}) bool {
	for {
		curCounterTail := atomic.LoadUint64(&q.counterTail)
		if q.batches[uint8(curCounterTail)].add(item) {
			return true
		}
		runtime.Gosched()
		/*
			curCounterHead := atomic.LoadUint64(&q.counterHead)
			if curCounterTail >= curCounterTail+countBatches-5 {
				runtime.Gosched()
				continue
			}
			if atomic.CompareAndSwapUint64(&q.counterTail, curCounterTail, curCounterTail+1) {
				//q.batches[uint8(curCounter+countBatches/2)] = newBatch(q.sizeBatch)
			}
		*/
	}
}

func (q *Batchqueue) PopBatch() *Batch {
	curCounter := atomic.AddUint64(&q.counterHead, 1)
	atomic.StoreInt64(&q.batches[curCounter-1].limit, 0)
	atomic.AddUint64(&q.counterTail, 1)
	q.batches[uint8(curCounter-2)] = newBatch(q.sizeBatch) // +countBatches/2
	return q.batches[curCounter-1]
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
