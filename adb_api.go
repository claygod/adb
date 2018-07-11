package adb

// Account database
// Main
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/claygod/adb/account"
	"github.com/claygod/adb/batcher"
	"github.com/claygod/adb/logname"
	"github.com/claygod/adb/wal"
)

type Adb struct {
	counter  int64
	accounts *Accounts
	//answers    *Answers
	state int64
	path  string
	//queue      *Queue
	//queuesPool [256]*queue.Queue
	batcher  *batcher.Batcher
	wal      *wal.Wal
	snapshot *Snapshoter
	ch       chan *batcher.Task
	ch2      chan *batcher.Task
	time     *time.Time
	symbol   *Symbol
}

func New(path string) (*Adb, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("'%s' does non exists", path)
	}

	// проверка на то, как закончилась последняя сессия
	// проводится ДО запуска базы !!!

	symbol := newSymbol()
	// fileName := "start.txt"
	ln := logname.New(8)
	wal, err := wal.New(path, ln, symbol.Separator1, logExt) //ToDo: del fileName
	if err != nil {
		return nil, err
	}
	ch := make(chan *batcher.Task, 1024)
	ch2 := make(chan *batcher.Task, 1024)
	//q := newQueue(sizeBucket * 16)
	b := batcher.New(wal, ch, ch2, logExt)

	adb := &Adb{
		accounts: newAccounts(symbol),
		//answers:  newAnswers(),
		//queue:   q,
		state:    stateClosed,
		path:     path,
		batcher:  b,
		wal:      wal,
		snapshot: newSnapshoter(symbol, path),
		ch:       ch,
		ch2:      ch2,
		time:     &time.Time{},
		symbol:   symbol,
	}

	b.SetBatchSize(sizeBucket) //.Start()
	return adb, nil
}

func (a *Adb) Start() {
	a.batcher.Start()
	atomic.StoreInt64(&a.state, stateOpen)
}

func (a *Adb) Stop() {
	atomic.StoreInt64(&a.state, stateClosed)
	a.batcher.Stop()
	a.saveToDisk()
}

func (a *Adb) Save() {
	a.Stop()
	a.saveToDisk()
	a.snapshot.clean()
	a.Start()
}

func (a *Adb) Load() error {
	a.Stop()
	defer a.Start()
	// тут получить имя текущего спапа
	// загрузить его и "догнать" логами
	snapsList, err := a.listFiles(snapExt, a.path)
	if err != nil {
		return err
	}
	ln := len(snapsList)
	switch {
	case ln == 0:
		return a.loadFromDisk(a.path + "adb" + dbExt)
	case ln == 1:
		return a.loadSnap(snapsList[ln-1])
	case ln > 1:
		return a.loadSnap(snapsList[ln-2])
	}
	return nil
	//a.Start()
}

/*
func (a *Adb) exeOrders(ords []*Order) error {
	for _, ord := range ords {
		a.transactionUnsafe(ord)
	}
	return nil
}
*/

/*
ToDo: что делать, если транзакция проведена, а ответ не успели отправить,
тогда получатель может не знать, что операция проведена (успешно/неуспешно).
Возможен двухфазовый подход - сначала отдаём номер транзакции и потом получаем,
исполняем и сохраняем номер транзакции с её результатом. При получении лучше
проверять, а следовательно, лучше давать номер и хэш, чтобы потом проверять.

Запрос на проверку транзакции можно сделать дорогим ждя запрашивающего через PoW
чтобы отмести спам и усложнить атаку на базу данных. Хотя скорее нет, т.к.
пусть лучше снаружи принимает решение программист, защищать ли базу.
*/

func (a *Adb) Transaction(order *Order) (*Answer, error) {
	//fmt.Println(order)
	num := atomic.AddInt64(&a.counter, 1)
	ans := a.doTransaction(order, num)
	// runtime.Gosched()
	//time.Sleep(1 * time.Microsecond)
	//fmt.Println(a.getAnswer(num, ans))
	return a.getAnswer(num, ans), nil
}

func (a *Adb) doTransaction(order *Order, num int64) *Answer {
	ans := &Answer{code: 0}
	tsk := a.getTask(order, ans)
	a.ch <- tsk
	return ans
}

func (a *Adb) getAnswer(num int64, ans *Answer) *Answer { // , a **Answer
	runtime.Gosched()
	for i := 0; ; i++ { //  i := 0; i < 1500000; i++
		if atomic.LoadInt64(&ans.code) > 0 {
			return ans //.(*Answer)
		}
		//time.Sleep(1000000 * time.Microsecond) //time.Duration(u) *
		runtime.Gosched()
	}
	fmt.Printf("\r\n- не найден - %d \r\n", num) // ToDo !!
	return nil
}

/*
type Query struct {
	num   int64
	order *Order
	log   []byte
}
*/

type Answer struct {
	code    int64
	balance map[string]map[string]account.Balance
}

type Order struct {
	Hash    string
	Debit   []*Part
	Credit  []*Part
	Block   []*Part
	Unblock []*Part
}

type Part struct {
	Id     string
	Key    string
	Amount uint64
}

/*
type Wal struct {
	sync.Mutex
}

func newWal() *Wal {
	return &Wal{}
}

func (w *Wal) Log(s string) error {
	return nil
}

func (w *Wal) Save() error {
	return nil
}

func (w *Wal) Close() error {
	return nil
}
*/
