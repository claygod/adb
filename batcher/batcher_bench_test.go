package batcher

// Batcher
// Bench
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	// "sync/atomic"
	"testing"
)

func BenchmarkWorkCycleSequence(b *testing.B) {
	b.StopTimer()
	batch := New(newMockWal(), newMockQueue(1)).SetBatchSize(1).Start()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		batch.work()
	}
}

func BenchmarkWork16x1Sequence(b *testing.B) {
	b.StopTimer()
	batch := New(newMockWal(), newMockQueue(16)).SetBatchSize(1).Start()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		batch.work()
	}
}

func BenchmarkWork16x8Sequence(b *testing.B) {
	b.StopTimer()
	batch := New(newMockWal(), newMockQueue(16)).SetBatchSize(8).Start()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		batch.work()
	}
}

func BenchmarkWork64x32Sequence(b *testing.B) {
	b.StopTimer()
	batch := New(newMockWal(), newMockQueue(64)).SetBatchSize(32).Start()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		batch.work()
	}
}
func BenchmarkWork64x64Sequence(b *testing.B) {
	b.StopTimer()
	batch := New(newMockWal(), newMockQueue(64)).SetBatchSize(64).Start()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		batch.work()
	}
}

/*
func BenchmarkWorkCycleParallel(b *testing.B) {
	b.StopTimer()
	batch := NewBatcher(newMockWal(), newMockQueue()).SetBatchSize(1).Start()
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			batch.work()
		}
	})
}
*/
