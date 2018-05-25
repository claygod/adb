package adb

// Account database
// Test
// Copyright © 2017-2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	"sync"
	"testing"
	"time"

	"github.com/claygod/transaction"
)

func BenchmarkBucketAddQuerySequence(b *testing.B) {
	b.StopTimer()
	bkt := NewBucket()
	cnt := 1000000
	qArray := ForTestGenQueryArray(cnt)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if i == cnt {
			return
		}
		//go r.DoTransaction(tArray[i], aArray[i])
		bkt.AddQuery(qArray[i])
	}
}

func Benchmark12Sequence(b *testing.B) {
	b.StopTimer()
	tc := transaction.New()
	if !tc.Start() {
		// t.Error("Now the start is possible!")
	}
	r := NewReception(&tc)

	//cnt := 1000000
	// prepare
	//tArray := ForTestGenTransactionsArray(cnt)
	time.Sleep(1 * time.Second)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		//if i == cnt {
		//	return
		//}
		//go r.DoTransaction(tArray[i], aArray[i])
		r.ExeTransaction(&Transaction{}) // tArray[i])
	}
}

func Benchmark12Parallel(b *testing.B) {
	b.StopTimer()

	tc := transaction.New()
	if !tc.Start() {
		// t.Error("Now the start is possible!")
	}
	r := NewReception(&tc)

	//cnt := 1000000
	// prepare
	// tArray := ForTestGenTransactionsArray(cnt)
	time.Sleep(1 * time.Second)
	//i := 0
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r.ExeTransaction(&Transaction{}) // tArray[i]
			//i++
		}
	})
}

/*
func Benchmark11Sequence(b *testing.B) {
	b.StopTimer()
	tc := transaction.New()
	if !tc.Start() {
		// t.Error("Now the start is possible!")
	}
	r := NewReception(&tc)

	cnt := 1000000
	// prepare
	tArray := make([]*Transaction, 0, cnt)
	aArray := make([]**Answer, 0, cnt)

	for i := 0; i < cnt; i++ {
		p := &Answer{code: int64(i)}
		var a **Answer = &p
		p = nil
		aArray = append(aArray, a)

		tr := &Transaction{}
		tArray = append(tArray, tr)
	}

	go func(tArray []*Transaction, aArray []**Answer) {
		for i := 0; i < cnt; i++ {
			r.DoTransaction(tArray[i], aArray[i])
			// time.Sleep(1 * time.Microsecond)
		}
	}(tArray, aArray)
	time.Sleep(1 * time.Second)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if i == cnt {
			return
		}
		//go r.DoTransaction(tArray[i], aArray[i])
		r.GetAnswer(i, aArray[i])
	}
}
*/
func ForTestExeTransaction(r *Reception, t *Transaction, wg *sync.WaitGroup) {
	r.ExeTransaction(t)
	wg.Done()
}

func ForTestGenAnswersArray(cnt int) []**Answer {
	aArray := make([]**Answer, 0, cnt)
	for i := 0; i < cnt; i++ {
		p := &Answer{}
		var a **Answer = &p
		p = nil
		aArray = append(aArray, a)
	}
	return aArray
}

func ForTestGenTransactionsArray(cnt int) []*Transaction {
	tArray := make([]*Transaction, 0, cnt)
	for i := 0; i < cnt; i++ {
		tr := &Transaction{}
		tArray = append(tArray, tr)
	}
	return tArray
}

func ForTestGenQueryArray(cnt int) []*Query {
	qArray := make([]*Query, 0, cnt)
	for i := 0; i < cnt; i++ {
		tr := &Transaction{}
		//p := &Answer{}
		//var a **Answer = &p
		//p = nil

		q := &Query{
			t: tr,
			//a: a,
		}

		qArray = append(qArray, q)
	}
	return qArray
}
