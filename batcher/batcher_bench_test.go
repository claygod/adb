package batcher

// Batcher
// Bench
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

//"fmt"
// "sync/atomic"
//"testing"

/*
func BenchmarkWorkCycleSequence(b *testing.B) {
	b.StopTimer()
	ch := make(chan *Task, 100)
	batch := New(newMockWal(), newMockQueue(1), ch).SetBatchSize(1).Start(Sync)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		batch.work()
	}
}

func BenchmarkWork16x1Sequence(b *testing.B) {
	b.StopTimer()
	ch := make(chan *Task, 100)
	batch := New(newMockWal(), newMockQueue(16), ch).SetBatchSize(1).Start(Sync)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		batch.work()
	}
}

func BenchmarkWork16x8Sequence(b *testing.B) {
	b.StopTimer()
	ch := make(chan *Task, 100)
	batch := New(newMockWal(), newMockQueue(16), ch).SetBatchSize(8).Start(Sync)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		batch.work()
	}
}

func BenchmarkWork64x32Sequence(b *testing.B) {
	b.StopTimer()
	ch := make(chan *Task, 100)
	batch := New(newMockWal(), newMockQueue(64), ch).SetBatchSize(32).Start(Async)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		batch.work()
	}
}
func BenchmarkWork64x64Sequence(b *testing.B) {
	b.StopTimer()
	ch := make(chan *Task, 100)
	batch := New(newMockWal(), newMockQueue(64), ch).SetBatchSize(64).Start(Sync)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		batch.work()
	}
}
*/
/*
func BenchmarkWorkCycleParallel(b *testing.B) {
	b.StopTimer()
	batch := NewBatcher(newMockWal(), newMockQueue()).SetBatchSize(1).Start(Sync)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			batch.work()
		}
	})
}
*/
