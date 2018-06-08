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
	if q.batches[1].data[0] != 30 {
		t.Error("Error Push|Batch ", q.batches[1])
	}
}
