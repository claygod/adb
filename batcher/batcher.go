package batcher

// Batcher
// API
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	// "fmt"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const batchSize int64 = 1024

//const logExt string = ".txt"

const (
	stateRun int64 = iota
	stateStops
	stateStopped
)

const Sync bool = true
const Async bool = false

type Batcher struct {
	batchSize int64
	barrier   int64
	logExt    string
	wal       Wal
	wg        sync.WaitGroup
	ch        chan *Task
	ch2       chan *Task
	time      time.Time
}

func New(wal Wal, ch chan *Task, ch2 chan *Task, logExt string) *Batcher {

	b := &Batcher{
		batchSize: batchSize,
		barrier:   stateStopped,
		logExt:    logExt,
		wal:       wal,
		wg:        sync.WaitGroup{},
		ch:        ch,
		ch2:       ch2,
		time:      time.Now(),
	}
	b.wal.Filename(b.TimeToString(b.GetTime()))
	return b
}

func (b *Batcher) Start() *Batcher {
	for {
		if atomic.CompareAndSwapInt64(&b.barrier, stateStopped, stateRun) {
			go b.worker(b.ch, b.ch2)
			return b
		}
		runtime.Gosched()
	}
}

func (b *Batcher) Stop() *Batcher {
	for {
		if atomic.LoadInt64(&b.barrier) == stateStopped {
			return b
		}
		atomic.CompareAndSwapInt64(&b.barrier, stateRun, stateStops)
		runtime.Gosched()
	}
}

func (b *Batcher) SetBatchSize(size int64) *Batcher {
	atomic.StoreInt64(&b.batchSize, size)
	return b
}

func (b *Batcher) worker(ch chan *Task, ch2 chan *Task) {
	tasks := make([]*Task, b.batchSize, b.batchSize)
	for {
		counter := 0
		for i := int64(0); i < b.batchSize; i++ {
			select {
			case t := <-ch:
				tasks[i] = t
				f := *t.Main
				//log := f()
				b.wal.Log(f())
				counter++
			default:
				if atomic.CompareAndSwapInt64(&b.barrier, stateStops, stateStopped) {
					return
				}
				i = b.batchSize
			}
		}

		if counter == 0 { // len(tasks)
			runtime.Gosched()
			continue
		}
		b.wal.Save()
		for i := 0; i < counter; i++ {
			f := *tasks[i].Finish
			f()
		}
	}
}

func (b *Batcher) GetTime() uint64 {
	return (uint64(b.time.Unix()) >> 8) << 8
}

func (b *Batcher) TimeToString(wt uint64) string {
	return strconv.FormatUint(wt, 10) + b.logExt
}

/*
type Queue interface {
	GetBatch(int64) []*func() (int64, []byte) //Input
}
*/
type Wal interface {
	Log(string) error // key, log
	Save() error
	Filename(string) error
	Close() error
}

type Task struct {
	Main   *func() string
	Finish *func()
}
