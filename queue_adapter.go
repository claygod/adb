package adb

// Account database
// Queue adapter
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"github.com/claygod/adb/queue"
)

type Queue struct {
	q *queue.Queue
	// f []*func() (int64, []byte)
}

func newQueue(num int64) *Queue {
	q := &Queue{
		// f: make([]*(func() (int64, []byte)), 0, num),
		q: queue.New(),
	}
	/*
		for i := int64(0); i < num; i++ {
			fn := func() (int64, []byte) {
				return i, []byte{byte(i)}
			}
			q.f = append(q.f, &fn)
		}
	*/
	return q
}

func (q *Queue) GetBatch(count int64) []*func() (int64, []byte) {
	qlsInterface := q.q.PopHeadList(int(count))
	qlsFunctions := make([](*func() (int64, []byte)), 0, len(qlsInterface))

	for _, qli := range qlsInterface {
		qlsFunctions = append(qlsFunctions, qli.(*func() (int64, []byte)))
	}
	return qlsFunctions
}

func (q *Queue) AddTransaction(qClosure *func() (int64, []byte)) bool {
	return q.q.PushTail(qClosure)
}
