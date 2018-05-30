package adb

// Account database
// Test
// Copyright © 2017-2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	// "fmt"
	"os"
	// "strconv"
	"sync"
	"testing"
	"time"

	"github.com/claygod/transaction"
)

func BenchmarkFsyncSequense(b *testing.B) {
	b.StopTimer()
	text := ForTestGenStringArray(256)
	fv, err := os.OpenFile("benchFsync.txt", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)

	}
	defer func() {
		fv.Close()
		os.Remove("./benchFsync.txt")
	}()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		fv.WriteString(text)
		fv.Sync()
	}
}

/*
func BenchmarkFsync16Sequense(b *testing.B) {
	b.StopTimer()
	fvs := make([]*os.File, 0, 16)

	for i := 0; i < 16; i++ {
		fv, _ := os.OpenFile("./benchFsync"+strconv.Itoa(i)+".txt", os.O_CREATE|os.O_WRONLY, 0666)
		fvs = append(fvs, fv)
	}
	text := ForTestGenStringArray(256)

	defer func(fvs []*os.File) {
		for i := 0; i < 16; i++ {
			fvs[i].Close()
			os.Remove("./benchFsync" + strconv.Itoa(i) + ".txt")
		}
	}(fvs)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		key := byte(i) >> 4
		fvs[key].WriteString(text)
		fvs[key].Sync()
	}
}

func BenchmarkFsync2Parallel(b *testing.B) {
	b.StopTimer()
	fvs := make([]*os.File, 0, 2)

	for i := 0; i < 2; i++ {
		fv, _ := os.OpenFile("./benchFsync"+strconv.Itoa(i)+".txt", os.O_CREATE|os.O_WRONLY, 0666)
		fvs = append(fvs, fv)
	}
	text := ForTestGenStringArray(256)

	defer func(fvs []*os.File) {
		for i := 0; i < 2; i++ {
			fvs[i].Close()
			os.Remove("./benchFsync" + strconv.Itoa(i) + ".txt")
		}
	}(fvs)

	i := 0
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := ((byte(i)) << 7) >> 7
			fvs[key].WriteString(text)
			fvs[key].Sync()
			i++
		}
	})
}

func BenchmarkFsync4Parallel(b *testing.B) {
	b.StopTimer()
	fvs := make([]*os.File, 0, 4)

	for i := 0; i < 4; i++ {
		fv, _ := os.OpenFile("./benchFsync"+strconv.Itoa(i)+".txt", os.O_CREATE|os.O_WRONLY, 0666)
		fvs = append(fvs, fv)
	}
	text := ForTestGenStringArray(256)

	defer func(fvs []*os.File) {
		for i := 0; i < 4; i++ {
			fvs[i].Close()
			os.Remove("./benchFsync" + strconv.Itoa(i) + ".txt")
		}
	}(fvs)

	i := 0
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := ((byte(i)) << 6) >> 6
			fvs[key].WriteString(text)
			fvs[key].Sync()
			i++
		}
	})
}

func BenchmarkFsync8Parallel(b *testing.B) {
	b.StopTimer()
	fvs := make([]*os.File, 0, 8)

	for i := 0; i < 8; i++ {
		fv, _ := os.OpenFile("./benchFsync"+strconv.Itoa(i)+".txt", os.O_CREATE|os.O_WRONLY, 0666)
		fvs = append(fvs, fv)
	}
	text := ForTestGenStringArray(256)

	defer func(fvs []*os.File) {
		for i := 0; i < 8; i++ {
			fvs[i].Close()
			os.Remove("./benchFsync" + strconv.Itoa(i) + ".txt")
		}
	}(fvs)

	i := 0
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := ((byte(i)) << 5) >> 5
			fvs[key].WriteString(text)
			fvs[key].Sync()
			i++
		}
	})
}

func BenchmarkFsync16Parallel(b *testing.B) {
	b.StopTimer()
	fvs := make([]*os.File, 0, 16)

	for i := 0; i < 16; i++ {
		fv, _ := os.OpenFile("./benchFsync"+strconv.Itoa(i)+".txt", os.O_CREATE|os.O_WRONLY, 0666)
		fvs = append(fvs, fv)
	}
	text := ForTestGenStringArray(256)

	defer func(fvs []*os.File) {
		for i := 0; i < 16; i++ {
			fvs[i].Close()
			os.Remove("./benchFsync" + strconv.Itoa(i) + ".txt")
		}
	}(fvs)

	i := 0
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := ((byte(i)) << 4) >> 4
			fvs[key].WriteString(text)
			fvs[key].Sync()
			i++
		}
	})
}

func BenchmarkFsync32Parallel(b *testing.B) {
	b.StopTimer()
	fvs := make([]*os.File, 0, 32)

	for i := 0; i < 32; i++ {
		fv, _ := os.OpenFile("./benchFsync"+strconv.Itoa(i)+".txt", os.O_CREATE|os.O_WRONLY, 0666)
		fvs = append(fvs, fv)
	}
	text := ForTestGenStringArray(256)

	defer func(fvs []*os.File) {
		for i := 0; i < 32; i++ {
			fvs[i].Close()
			os.Remove("./benchFsync" + strconv.Itoa(i) + ".txt")
		}
	}(fvs)

	i := 0
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := ((byte(i)) << 3) >> 3
			fvs[key].WriteString(text)
			fvs[key].Sync()
			i++
		}
	})
}
*/
// ------------------
/*
func BenchmarkDoTransactionParallel(b *testing.B) {
	b.StopTimer()

	tc := transaction.New()
	if !tc.Start() {
		// t.Error("Now the start is possible!")
	}
	r := NewReception(&tc)
	time.Sleep(1 * time.Second)
	//i := 0
	tr := &Transaction{}
	var i int64
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r.DoTransaction(tr, i)
			if r.queue.SizeQueue() > 500000 {
				r.queue.PopAll()
			}
			i++
		}
	})
}
*/
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
func ForTestGenStringArray(num int) string {
	str := "12345678:879932921731971397821:RTFTRTGGH:+:1230\r\n"
	var out string
	for i := 0; i < num; i++ {
		out += str
	}
	return out
}

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
