package adb

// Account database
// Main
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	// "unsafe"

	"github.com/claygod/adb/account"
	"github.com/claygod/adb/batcher"
	"github.com/claygod/adb/queue"
	// "github.com/claygod/adb/transaction"
)

// Hasp state
const (
	stateClosed int64 = iota
	stateOpen
)

// const permitError int64 = -2147483647
const sizeBucket int64 = 256

type Reception struct {
	// sync.Mutex
	counter  int64
	accounts *Accounts
	answers  *Answers //sync.Map
	//bucket     *Bucket
	workerStop int64
	//tCore      *transaction.Core
	queue *Queue
	//queuesPool [256]*queue.Queue
	batcher *batcher.Batcher
	wal     *Wal
}

func NewReception() *Reception {
	wal := newWal()
	q := newQueue(sizeBucket * 32) // queue.New(sizeBucket * 32)
	b := batcher.New(wal, q)       // .SetBatchSize(sizeBucket * 8).Start()
	// b := NewBucket()
	r := &Reception{
		accounts: newAccounts(),
		answers:  newAnswers(),
		//tCore:    tc,
		queue:   q,
		batcher: b,
		wal:     wal,
	}
	b.Start(batcher.Sync)

	//for i := 0; i < 256; i++ {
	//	r.queuesPool[i] = queue.New(sizeBucket)
	//}

	//go r.worker(0)
	// go r.worker(1)
	//time.Sleep(100000 * time.Microsecond)
	//go r.worker(1)
	return r
}

func (r *Reception) ExeTransaction(order *Order) *Answer {
	num := atomic.AddInt64(&r.counter, 1)
	r.DoTransaction(order, num)
	//time.Sleep(1 * time.Microsecond)
	return r.GetAnswer(num)
}

func (r *Reception) DoTransaction(order *Order, num int64) { // , a **Answer
	// num := atomic.AddInt64(&r.counter, 1)
	// var orderGob bytes.Buffer // Stand-in for the network.

	// Create an encoder and send a value.
	/*
		enc := gob.NewEncoder(&orderGob)
		err := enc.Encode(order)
		if err != nil {
			r.store.Store(num, &Answer{code: 404})
			fmt.Printf("\r\n- отбросили из-за ошибки кодирования - %d \r\n", num)
		}
	*/
	fmt.Println(" @001@ ", num)

	logBytes, err := r.orderToLog(order)
	if err != nil {
		fmt.Println(" - ошибка кодирования лога ", num, err)
		r.answers.Store(num, &Answer{code: 404}) // отрицательный ответ
		return
	}
	qLambda := func() (int64, []byte) {
		replyBalances := make(map[string]map[string]account.Balance)
		//toLog := ""
		fmt.Println(" @e01@ начат цикл")
		for i := 0; i < len(order.Credit); i++ {
			acc := r.accounts.Account(order.Credit[i].Id)
			if acc == nil {
				r.Rollback(i, order)
				fmt.Println(" @e01@ --", num)
				return num, []byte("") // тут логовое сообщение для ошибочной транзакции
			}

			balance, err := acc.Balance(order.Credit[i].Key).WriteOff(order.Credit[i].Amount)
			if err != nil {
				r.Rollback(i+1, order)
				fmt.Println(" @e02@ --", num)
				return num, []byte("") // тут логовое сообщение для ошибочной транзакции
			}

			r.balancesAddBalance(order.Credit[i].Id, order.Credit[i].Key, replyBalances, balance)
			r.answers.Store(num, &Answer{code: 200}) // возвращаем положительный ответ
		}
		fmt.Println(" замыкание запущено под номером: ", num)
		return num, logBytes
	}
	fmt.Println(" @002@ ", num)
	if !r.queue.PushTail(&qLambda) {
		r.answers.Store(num, &Answer{code: 404})
		fmt.Printf("\r\n- отбросили ---- %d \r\n", num)
	}
	fmt.Println(" @003@ ", num)
	//return 1
	return
}

func (r *Reception) balancesAddBalance(id string, key string, balances map[string]map[string]account.Balance, balance account.Balance) {
	if _, ok := balances[id]; !ok {
		balances[id] = make(map[string]account.Balance)
	}
	balances[id][key] = balance
}

func (r *Reception) orderToLog(order *Order) ([]byte, error) {
	var orderGob bytes.Buffer
	enc := gob.NewEncoder(&orderGob)
	err := enc.Encode(order)
	if err != nil {
		return nil, err
	}
	return orderGob.Bytes(), nil
}

func (r *Reception) rollbackBlock(num int, order *Order) {
	for i := 0; i < num; i++ {
		r.accounts.Account(order.Block[i].Id).
			Balance(order.Block[i].Key).
			Unblock(order.Hash, order.Block[i].Amount)
	}
}

func (r *Reception) rollbackUnblock(num int, order *Order) {
	for i := 0; i < num; i++ {
		r.accounts.Account(order.Block[i].Id).
			Balance(order.Block[i].Key).
			Block(order.Hash, order.Block[i].Amount)
	}
}

func (r *Reception) rollbackCdedit(num int, order *Order) {
	for i := 0; i < num; i++ {
		r.accounts.Account(order.Block[i].Id).
			Balance(order.Block[i].Key).
			Debit(order.Block[i].Amount)
		r.accounts.Account(order.Block[i].Id).
			Balance(order.Block[i].Key).
			Block(order.Hash, order.Block[i].Amount)
	}
}

func (r *Reception) Rollback(num int, order *Order) {
	for i := 0; i < num; i++ {
		r.accounts.Account(order.Credit[i].Id).
			Balance(order.Credit[i].Key).
			Debit(order.Credit[i].Amount)
	}
}

func (r *Reception) GetAnswer(num int64) *Answer { // , a **Answer
	fmt.Println(" @031@ ", num)
	for { //  i := 0; i < 1500000; i++
		fmt.Println(" @032@ ", num)
		if a1, ok := r.answers.Load(num); ok {
			go r.answers.Delete(num)
			fmt.Println(" @032@ ура, ответ получен! я", num)
			return a1 //.(*Answer)
		}
		/*
			u := i
			if u > 10000 {
				u = 10000
			}
		*/
		time.Sleep(1000000 * time.Microsecond) //time.Duration(u) *
		runtime.Gosched()
	}
	fmt.Printf("\r\n- не найден - %d \r\n", num)
	return nil
}

/*
func (r *Reception) worker(level int) {
	//var shift uint8
	var wg sync.WaitGroup
	//var an *Answer = &Answer{code: 200}
	for {
		//shift++
		//b := r.queuesPool[shift].PopHeadList(sizeBucket)
		// ///////// b := r.queue.PopHeadList(int(sizeBucket))
		//b := r.queue.PopAll()

		if len(b) == 0 {
			// fmt.Printf("\r\n- воркер получил пустую очередь- \r\n")
			runtime.Gosched()
			// time.Sleep(30 * time.Microsecond)
			continue
		}
		if len(b) > 5 {
			// fmt.Printf("\r\n- размер очереди- %d \r\n", len(b))
		}
		//var logBuf bytes.Buffer
		for _, q1 := range b {
			q := q1.(*Query)
			wg.Add(1)
			go r.handler(q.t, &wg, q.num, q.log)
			//r.store.Store(q.num, an) //
			// тут сохраняем лог
		}
		wg.Wait()

		if !r.wal.Save() || r.workerStop == 1 {
			return
		}
		//shift++
	}
}
*/
func (r *Reception) handler(order *Order, wg *sync.WaitGroup, num int64, log []byte) {
	r.answers.Store(num, &Answer{code: 200})
	// if ok
	r.wal.Log(num, log)
	// dummy
	wg.Done()
}

type Answer struct {
	code    int64
	balance map[string]map[string]account.Balance
}

type Query struct {
	num   int64
	order *Order
	log   []byte
	// a   **Answer
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

type Answers struct {
	sync.Mutex
	arr map[int64]*Answer
}

func newAnswers() *Answers {
	return &Answers{
		arr: make(map[int64]*Answer),
	}
}

func (a *Answers) Load(key int64) (*Answer, bool) {
	a.Lock()
	// defer s.RUnlock()
	an, ok := a.arr[key]
	a.Unlock()
	return an, ok
}

func (a *Answers) Store(key int64, an *Answer) {
	a.Lock()
	a.arr[key] = an
	a.Unlock()
}

func (a *Answers) Delete(key int64) {
	a.Lock()
	delete(a.arr, key)
	a.Unlock()
}

type Wal struct {
	sync.Mutex
}

func newWal() *Wal {
	return &Wal{}
}

func (w *Wal) Log(key int64, b []byte) {
}

func (w *Wal) Save() bool {
	return true
}

type Queue struct {
	q *queue.Queue
	f []*func() (int64, []byte)
}

func newQueue(num int64) *Queue {
	q := &Queue{
		f: make([]*(func() (int64, []byte)), 0, num),
		q: queue.New(),
	}

	for i := int64(0); i < num; i++ {
		fn := func() (int64, []byte) {
			return i, []byte{byte(i)}
		}
		q.f = append(q.f, &fn)
	}
	return q
}

func (q *Queue) GetBatch(count int64) []*func() (int64, []byte) {
	qlsInterface := q.q.PopHeadList(int(count))
	qlsFunctions := make([](*func() (int64, []byte)), 0, len(qlsInterface))

	for _, qli := range qlsInterface {
		qlsFunctions = append(qlsFunctions, qli.(*func() (int64, []byte)))
	}
	return qlsFunctions
}

func (q *Queue) PushTail(qLambda *func() (int64, []byte)) bool { // Mock !
	return q.q.PushTail(qLambda)
}
