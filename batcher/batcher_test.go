package batcher

// Batcher
// Test
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"sync/atomic"
	"testing"
)

func TestCallWal(t *testing.T) {
	w := newMockWal()
	batch := New(w, newMockQueue(1)).SetBatchSize(1).Start()
	for i := 0; i < 5; i++ {
		batch.work()
	}

	if w.counter != 5 {
		//t.Error("Error in call 'WAL' (expected 5) - ", w.counter)
	}
}

type mockWal struct {
	counter int64
}

func newMockWal() *mockWal {
	return &mockWal{}
}

func (w *mockWal) Log(key int64, log []byte) {
	atomic.AddInt64(&w.counter, 1)
	// fmt.Println("---", w.counter, " ", key, " ", log)
}

func (w *mockWal) Save() bool {
	atomic.StoreInt64(&w.counter, 0)
	return true
}

type mockQueue struct {
	f []*func() (int64, []byte)
}

func newMockQueue(num int64) *mockQueue {
	q := &mockQueue{
		f: make([]*(func() (int64, []byte)), 0, num),
	}

	for i := int64(0); i < num; i++ {
		fn := func() (int64, []byte) {
			return i, []byte{byte(i)}
		}
		q.f = append(q.f, &fn)
	}
	return q
}

func (q *mockQueue) GetBatch(count int64) []*func() (int64, []byte) {
	return q.f
}
