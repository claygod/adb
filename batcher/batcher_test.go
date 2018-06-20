package batcher

// Batcher
// Test
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"sync/atomic"
	//"testing"
)

/*
func TestCallWal(t *testing.T) {
	w := newMockWal()
	ch := make(chan *Task, 100)
	batch := New(w, newMockQueue(1), ch).SetBatchSize(1).Start(Sync)
	for i := 0; i < 5; i++ {
		batch.work()
	}

	if w.counter != 5 {
		//t.Error("Error in call 'WAL' (expected 5) - ", w.counter)
	}
}
*/
type mockWal struct {
	counter int64
}

func newMockWal() *mockWal {
	return &mockWal{}
}

func (w *mockWal) Log(log []byte) error {
	atomic.AddInt64(&w.counter, 1)
	// fmt.Println("---", w.counter, " ", key, " ", log)
	return nil
}

func (w *mockWal) Save() error {
	atomic.StoreInt64(&w.counter, 0)
	return nil
}

func (w *mockWal) Close() error {
	atomic.StoreInt64(&w.counter, 0)
	return nil
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
