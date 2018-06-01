package account

// Account
// Bench
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	// "sync/atomic"
	"strconv"
	"testing"
)

var testCount int = 4000000000

func BenchmarkDebit(b *testing.B) {
	b.StopTimer()
	a := newSubAccount()
	a.Debit(uint64(testCount))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		a.Debit(1)
	}
}

func BenchmarkCredit(b *testing.B) {
	b.StopTimer()
	a := newSubAccount()
	a.Debit(uint64(testCount))
	//for i := 0; i < testCount; i++ {
	//	a.Block(strconv.Itoa(i), uint64(i))
	//}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		a.WriteOff(1)
	}
}

func BenchmarkBlock(b *testing.B) {
	b.StopTimer()
	a := newSubAccount()
	a.Debit(uint64(testCount))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		a.Block(strconv.Itoa(i), uint64(i))
	}
}

func BenchmarkBlockUnblock(b *testing.B) {
	b.StopTimer()
	a := newSubAccount()
	a.Debit(uint64(testCount))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		a.Block(key, uint64(i))
		a.Unblock(key, uint64(i))
	}
}
