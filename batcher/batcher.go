package batcher

// Batcher
// API
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	// "fmt"
	"runtime"
	"sync"
	"sync/atomic"
	// "time"
)

const batchSize int64 = 1024

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
	//queue     Queue
	wg  sync.WaitGroup
	ch  chan *Task
	ch2 chan *Task
}

func New(wal Wal, ch chan *Task, ch2 chan *Task) *Batcher {
	return &Batcher{
		batchSize: batchSize,
		barrier:   stateStop,
		wal:       wal,
		//queue:     queue,
		wg:  sync.WaitGroup{},
		ch:  ch,
		ch2: ch2,
	}
}

func (b *Batcher) StartChain(mode bool) *Batcher {
	if atomic.CompareAndSwapInt64(&b.barrier, stateStop, stateRun) {
		if mode == Sync {
			go b.workerSyncChan(b.ch, b.ch2)
		} else {

		}
	}
	return b
}

func (b *Batcher) Stop() *Batcher {
	atomic.StoreInt64(&b.barrier, stateStop)
	return b
}

func (b *Batcher) SetBatchSize(size int64) *Batcher {
	atomic.StoreInt64(&b.batchSize, size)
	return b
}

func (b *Batcher) workerSyncChan(ch chan *Task, ch2 chan *Task) {
	tasks := make([]*Task, b.batchSize, b.batchSize)
	for {
		counter := 0
		for i := int64(0); i < b.batchSize; i++ {
			select {
			case t := <-ch:
				tasks[i] = t
				f := *t.Main
				f()
				counter++
			default:
				i = b.batchSize
			}
		}

		if counter == 0 { // len(tasks)
			runtime.Gosched()
			continue
		}

		b.wal.Save()
		// fmt.Println(" @len batch@ ", len(tasks), b.batchSize)
		for i := 0; i < counter; i++ { //}_, t := range tasks {
			f := *tasks[i].Finish
			f()
		}

		if b.barrier == stateStop { // b.wal.Save() != nil ||
			return
		}
	}
}

type Queue interface {
	GetBatch(int64) []*func() (int64, []byte) //Input
}

type Wal interface {
	Log(string) error // key, log
	Save() error
	Close() error
}

type Task struct {
	Main   *func()
	Finish *func()
}
