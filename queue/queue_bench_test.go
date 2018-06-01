package queue

// Queue
// Bench
// Copyright © 2016-2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"testing"
)

func BenchmarkPushTail(b *testing.B) {
	b.StopTimer()
	q := New()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		q.PushTail(i)
	}
}

func BenchmarkPushTailParallel(b *testing.B) {
	b.StopTimer()
	q := New()
	i := 0
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			q.PushTail(i)
			i++
		}
	})
}

func BenchmarkPushHeadLimit(b *testing.B) {
	b.StopTimer()
	limit := 10000000
	q := New(150000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		q.PushHead(i)
		if i > limit {
			break
		}
	}
}

func BenchmarkPushHead(b *testing.B) {
	b.StopTimer()
	q := New()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		q.PushHead(i)
	}
}

func BenchmarkPopHead(b *testing.B) {
	b.StopTimer()
	q := New()
	for i := 0; i < 10000000; i++ {
		q.PushTail(i)
	}
	b.StartTimer()
	for i := 0; i < 1000000; i++ {
		q.PopHead()
	}
}

func BenchmarkPopTail(b *testing.B) {
	b.StopTimer()
	q := New()
	for i := 0; i < 10000000; i++ {
		q.PushTail(i)
	}
	b.StartTimer()
	for i := 0; i < 1000000; i++ {
		q.PopTail()
	}
}

func BenchmarkQueueList(b *testing.B) {
	b.StopTimer()
	q := New()
	for i := 0; i < 10000000; i++ {
		q.PushTail(i)
	}
	b.StartTimer()
	for i := 0; i < 1000000; i++ {
		q.PopHeadList(10)
	}
}

func BenchmarkQueueListParallel(b *testing.B) {
	b.StopTimer()
	q := New()
	for i := 0; i < 10000000; i++ {
		q.PushTail(i)
	}
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			q.PopHeadList(2)
		}
	})
}
