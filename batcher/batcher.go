package batcher

// Batcher
// API
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	// "fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const batchSize int64 = 4

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
	ch        chan *Task
}

func New(wal Wal, queue Queue, ch chan *Task) *Batcher {
	return &Batcher{
		batchSize: batchSize,
		barrier:   stateStop,
		wal:       wal,
		queue:     queue,
		wg:        sync.WaitGroup{},
		ch:        ch,
	}
}

func (b *Batcher) Start(mode bool) *Batcher {
	if atomic.CompareAndSwapInt64(&b.barrier, stateStop, stateRun) {
		if mode == Sync {
			//fmt.Println(" @053-Sync@ ")
			go b.workerSync()
		} else {
			//fmt.Println(" @053-Async@ ")
			go b.workerAsync()
		}
	}
	return b
}

func (b *Batcher) StartChain(mode bool) *Batcher {
	if atomic.CompareAndSwapInt64(&b.barrier, stateStop, stateRun) {
		if mode == Sync {
			//fmt.Println(" @053-Sync@ ")
			go b.workerSyncChan(b.ch)
		} else {
			//fmt.Println(" @053-Async@ ")
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
	atomic.StoreInt64(&b.batchSize, size)
	return b
}

func (b *Batcher) workerSync() {
	for {
		batch := b.queue.GetBatch(b.batchSize)
		//fmt.Println(" @длина батча@ -- ", len(batch))

		if len(batch) == 0 {
			runtime.Gosched()
			time.Sleep(1 * time.Microsecond)
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

func (b *Batcher) workerSyncChan(ch chan *Task) {
	for {
		// var i int64 = 0
		tasks := make([]*Task, 0, b.batchSize)
		for i := int64(0); i < b.batchSize; i++ {

			//t := <-ch
			//tasks = append(tasks, t)
			//f := *t.Main
			//f()
			//time.Sleep(1000 * time.Microsecond)

			select {
			case t := <-ch:
				tasks = append(tasks, t)
				f := *t.Main
				f()
			default:
				// runtime.Gosched()
				i = b.batchSize
				//fmt.Println(" @i---------------------------------@ ", i, " b.batchSize: ", b.batchSize)
				//break
			}

			//fmt.Println(" @i@ ", i, " b.batchSize: ", b.batchSize)
		}
		//fmt.Println(" i---------------------------------@----------------- ")
		//if len(tasks) > 1 {
		// fmt.Println(" @len batch@ ", len(tasks))
		//}

		if len(tasks) == 0 {
			runtime.Gosched()
			continue
		}

		for _, t := range tasks {
			f := *t.Finish
			f()
			// fmt.Println(" #i# ", i)
		}

		if b.barrier == stateStop { // b.wal.Save() != nil ||
			return
		}

		//time.Sleep(100 * time.Microsecond)
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
	(*in)()
	//fmt.Println(" @053--1@ ")
	//b.wal.Log((*in)())
	//fmt.Println(" @053--2@ ")
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

type Task struct {
	Main   *func()
	Finish *func()
}
