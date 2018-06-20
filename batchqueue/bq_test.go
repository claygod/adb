package batchqueue

// Batchqueue
// Test
// Copyright © 2016 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"testing"
)

func TestUint8(t *testing.T) {
	var u uint8 = 100
	u += 200
	fmt.Println("Uint8 100+200=", u)

	var x uint64 = 100
	x += 200
	fmt.Println("Uint8(uint64) 100+200=", uint8(x))
	//if w.counter != 5 {
	//t.Error("Error in call 'WAL' (expected 5) - ", w.counter)
	//}
}

func TestPush(t *testing.T) {
	q := New(2)
	if !q.Push(10) {
		t.Error("Error Push")
	}
	q.Push(20)
	q.Push(30)
	if q.batches[3].data[0] != 30 {
		t.Error("Error Push|Batch ", q.batches[2])
	}
}

func TestPopBatch(t *testing.T) {
	q := New(2)
	if !q.Push(10) {
		t.Error("Error Push")
	}
	q.Push(20)
	q.Push(30)
	btch := q.PopBatch()
	if btch == nil {
		t.Error("Error PopBatch NIL ")
	}

	if btch.size != 0 {
		t.Error("Error PopBatch .size ", btch.size)
	}

	btch = q.PopBatch()
	//fmt.Println(q.cursor, q.PopBatch())
	btch = q.PopBatch()
	//btch = q.PopBatch()
	if btch.size != 2 {
		t.Error("Error PopBatch .size ", btch.size)
		// q.batches[0], q.batches[1], q.batches[2], q.batches[3], q.batches[4], q.batches[5]
	}

	if btch.data[0] != 10 {
		t.Error("Error PopBatch .data[0] ", btch.data[0])
	}
}
