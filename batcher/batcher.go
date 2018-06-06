package batcher

// Batcher
// API
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

const batchSize int64 = 256

const (
	stateRun int64 = iota
	stateStop
)

const Sync bool = true
const Async bool = false

type Batcher struct {
	batchSize int64
	barrier   int64
	wal       Wal
	queue     Queue
	wg        sync.WaitGroup
}

func New(wal Wal, queue Queue) *Batcher {
	return &Batcher{
		batchSize: batchSize,
		barrier:   stateStop,
		wal:       wal,
		queue:     queue,
		wg:        sync.WaitGroup{},
	}
}

func (b *Batcher) Start(mode bool) *Batcher {
	if atomic.CompareAndSwapInt64(&b.barrier, stateStop, stateRun) {
		if mode == Sync {
			fmt.Println(" @053-Sync@ ")
			go b.workerSync()
		} else {
			fmt.Println(" @053-Async@ ")
			go b.workerAsync()
		}
	}
	return b
}

func (b *Batcher) startCut() *Batcher {
	if atomic.CompareAndSwapInt64(&b.barrier, stateStop, stateRun) {
		go b.workerCut()
	}
	return b
}

func (b *Batcher) Stop() *Batcher {
	atomic.StoreInt64(&b.barrier, stateStop)
	return b
}

func (b *Batcher) SetBatchSize(size int64) *Batcher {
	atomic.StoreInt64(&b.batchSize, stateStop)
	return b
}

func (b *Batcher) workerSync() {
	for {
		batch := b.queue.GetBatch(b.batchSize)
		// fmt.Println(" @052@ ")

		if len(batch) == 0 {
			runtime.Gosched()
			continue
		}
		for _, in := range batch {
			b.inputProcessSync(in)
		}
		if b.wal.Save() != nil || b.barrier == stateStop {
			return
		}
	}
}

func (b *Batcher) workerAsync() {
	var wg sync.WaitGroup
	for {
		//fmt.Println(" @051@ ")
		batch := b.queue.GetBatch(b.batchSize)
		//fmt.Println(" @052@ ")

		if len(batch) == 0 {
			runtime.Gosched()
			continue
		}
		for _, in := range batch {
			wg.Add(1)
			b.inputProcessAsync(in, &wg)
		}
		//fmt.Println(" @054@ ", len(batch))
		wg.Wait()
		//fmt.Println(" @055@ ", len(batch))
		if b.wal.Save() != nil || b.barrier == stateStop {
			return
		}
	}
}

func (b *Batcher) workerCut() {
	for {
		b.work()
		if b.wal.Save() != nil || b.barrier == stateStop {
			return
		}
	}
}

func (b *Batcher) work() {
	var wg sync.WaitGroup = b.wg
	batch := b.queue.GetBatch(b.batchSize)
	if len(batch) == 0 {
		runtime.Gosched()
		return
	}
	for _, in := range batch {
		wg.Add(1)
		b.inputProcessAsync(in, &wg) //
	}
	wg.Wait()
}

/*
inputProcessAsync

Return:
	- key (int64) - number executed tasks
	- answer ([]byte) - gob-serialize
*/
func (b *Batcher) inputProcessAsync(in *func() (int64, []byte), wg *sync.WaitGroup) {
	//fmt.Println(" @053@ ")
	b.wal.Log((*in)())
	//fmt.Println(" @053@ ___")
	wg.Done()
}

func (b *Batcher) inputProcessSync(in *func() (int64, []byte)) {
	fmt.Println(" @053--1@ ")
	b.wal.Log((*in)())
	fmt.Println(" @053--2@ ")
}

/*
type Handler interface {
	Do(*Input)
}
*/
type Queue interface {
	GetBatch(int64) []*func() (int64, []byte) //Input
}

type Wal interface {
	Log(int64, []byte) error // key, log
	Save() error
	Close() error
}

/*
type Output struct {
	code    int64
	context interface{}
}
*/
/*
type Input struct {
	number  int64
	handler Handler
	log     []byte
	context interface{}
}
*/
