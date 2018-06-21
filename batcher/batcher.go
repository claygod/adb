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
const logExt string = ".txt"

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
	wg   sync.WaitGroup
	ch   chan *Task
	ch2  chan *Task
	time time.Time
}

func New(wal Wal, ch chan *Task, ch2 chan *Task) *Batcher {

	b := &Batcher{
		batchSize: batchSize,
		barrier:   stateStop,
		wal:       wal,
		//queue:     queue,
		wg:   sync.WaitGroup{},
		ch:   ch,
		ch2:  ch2,
		time: time.Now(),
	}
	b.wal.Filename(b.TimeToString(b.GetTime()))
	return b
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
	//fileName := strconv.FormatUint((uint64(time.Now().Unix())>>8)<<8, 10)
	// walTime := b.GetTime() //(uint64(b.time.Unix()) >> 8) << 8
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
		// fileName := strconv.FormatUint((uint64(time.Now().Unix())>>8)<<8, 10)

		//if wt := b.GetTime(); wt != walTime {
		//	b.wal.ChangeFilename(b.TimeToString(b.GetTime()))//b.wal.ChangeFilename(strconv.FormatUint(wt, 10) + logExt)
		//}
	}
}

func (b *Batcher) GetTime() uint64 {
	return (uint64(b.time.Unix()) >> 8) << 8
}

func (b *Batcher) TimeToString(wt uint64) string {
	return strconv.FormatUint(wt, 10) + logExt
}

type Queue interface {
	GetBatch(int64) []*func() (int64, []byte) //Input
}

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
