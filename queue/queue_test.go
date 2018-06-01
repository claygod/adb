package queue

// Queue
// Test
// Copyright © 2016-2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"testing"
)

func TestLenQueue(t *testing.T) {
	q := New(10)
	q.PushTail(1)
	q.PushTail(2)
	if q.LenQueue() != 2 {
		t.Error("LenQueue 2 != ", q.LenQueue())
	}
}

func TestSizeQueue10(t *testing.T) {
	q := New(10)
	q.PushTail(1)
	if q.SizeQueue() != 10 {
		t.Error("SizeQueue 10 != ", q.LenQueue())
	}
}

func TestSizeQueue20(t *testing.T) {
	q := New(10)
	for i := 0; i < 5; i++ {
		q.PushTail(i)
	}
	if q.SizeQueue() != 20 {
		t.Error("SizeQueue 20 != ", q.LenQueue())
	}
}

/*
func TestClear(t *testing.T) {
	q := New(10)
	for i := 0; i < 30; i++ {
		q.PushTail(i)
	}
	for i := 0; i < 30; i++ {
		q.PopTail()
	}
	if q.SizeQueue() != 10 {
		t.Error("Clear not correct! SizeQueue 10 != ", q.SizeQueue())
	}
}
*/
func TestDatabase(t *testing.T) {
	q := New(10)
	for i := 0; i < 100; i++ {
		q.PushTail(i)
	}
	for i := 99; i >= 0; i-- {
		m, _ := q.PopTail()
		if m.(int) != i {
			t.Error("Error in database", i, " != ", m)
		}
	}
}
func TestPopHead(t *testing.T) {
	q := New(10)
	q.PushTail(100)
	q.PushHead(200)
	q.PushTail(300)

	if m, _ := q.PopHead(); m.(int) != 200 {
		t.Error("Error PopHead: 200 != ", m.(int))
	}
	if m, _ := q.PopHead(); m.(int) != 100 {
		t.Error("Error PopHead: 100 != ", m.(int))
	}
	if m, _ := q.PopHead(); m.(int) != 300 {
		t.Error("Error PopHead: 300 != ", m.(int))
	}
}

func TestPopTail(t *testing.T) {
	q := New(10)
	q.PushTail(100)
	q.PushHead(200)
	q.PushTail(300)

	if m, _ := q.PopTail(); m.(int) != 300 {
		t.Error("Error PopTail: 300 != ", m.(int))
	}
	if m, _ := q.PopTail(); m.(int) != 100 {
		t.Error("Error PopTail: 100 != ", m.(int))
	}
	if m, _ := q.PopTail(); m.(int) != 200 {
		t.Error("Error PopTail: 200 != ", m.(int))
	}
}

func TestQueuePushList(t *testing.T) {
	q := New(10)
	for i := 0; i < 10; i++ {
		q.PushTail(i)
	}
	lst := q.PopHeadList(3)
	if len(lst) != 3 {
		t.Error("Not received 3 items, although there are 10 in the database.")
	}
	if lst[0].(int) != 0 {
		t.Error("Received an incorrect answer.")
	}
}

func TestQueuePushList2(t *testing.T) {
	q := New(10)
	for i := 0; i < 10; i++ {
		q.PushTail(i)
	}
	lst := q.PopHeadList(13)
	if len(lst) != 10 {
		t.Error("Amount is incorrect.")
	}

}
