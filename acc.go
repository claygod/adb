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

	"github.com/claygod/adb/account"
	"github.com/claygod/adb/batcher"
	"github.com/claygod/adb/queue"
)

// Hasp state
const (
	stateClosed int64 = iota
	stateOpen
)
const sizeBucket int64 = 256

type Reception struct {
	counter    int64
	accounts   *Accounts
	answers    *Answers
	workerStop int64
	queue      *Queue
	//queuesPool [256]*queue.Queue
	batcher *batcher.Batcher
	wal     *Wal
}

func NewReception() *Reception {
	wal := newWal()
	q := newQueue(sizeBucket * 32)
	b := batcher.New(wal, q)
	r := &Reception{
		accounts: newAccounts(),
		answers:  newAnswers(),
		queue:    q,
		batcher:  b,
		wal:      wal,
	}
	b.SetBatchSize(sizeBucket * 8).Start(batcher.Sync)

	//for i := 0; i < 256; i++ {
	//	r.queuesPool[i] = queue.New(sizeBucket)
	//}
	return r
}

func (r *Reception) ExeTransaction(order *Order) *Answer {
	num := atomic.AddInt64(&r.counter, 1)
	r.DoTransaction(order, num)
	//time.Sleep(1 * time.Microsecond)
	return r.GetAnswer(num)
}

func (r *Reception) DoTransaction(order *Order, num int64) {
	fmt.Println(" @001@ ", num)

	logBytes, err := r.orderToLog(order)
	if err != nil {
		fmt.Println(" - ошибка кодирования лога ", num, err)
		r.answers.Store(num, &Answer{code: 404}) // отрицательный ответ
		return
	}
	qClosure := r.getClosure(logBytes, order, num)
	fmt.Println(" @002@ ", num)
	if !r.queue.PushTail(&qClosure) {
		r.answers.Store(num, &Answer{code: 404})
		fmt.Printf("\r\n- отбросили ---- %d \r\n", num)
	}
	fmt.Println(" @003@ ", num)
	//return 1
	return
}

func (r *Reception) getClosure(logBytes []byte, order *Order, num int64) func() (int64, []byte) {
	return func() (int64, []byte) {
		replyBalances := make(map[string]map[string]account.Balance)
		lenBlock := len(order.Block)
		lenUnblock := len(order.Unblock)
		lenCredit := len(order.Credit)
		lenDebit := len(order.Debit)
		// Block
		fmt.Println(" @e01@ начата Block")
		if lenBlock > 0 {
			if count, err := r.doBlock(order, replyBalances); err != nil {
				r.rollbackBlock(count, order)
				r.answers.Store(num, &Answer{code: 404})
				return num, []byte("") // тут логовое сообщение для ошибочной транзакции - оно должно быть пустым!
			}
		}
		// Unblock
		fmt.Println(" @e01@ начата Unblock")
		if lenUnblock > 0 {
			if count, err := r.doUnblock(order, replyBalances); err != nil {
				if lenBlock > 0 {
					r.rollbackBlock(len(order.Block), order)
				}
				r.rollbackUnblock(count, order)
				r.answers.Store(num, &Answer{code: 404})
				return num, []byte("") // тут логовое сообщение для ошибочной транзакции - оно должно быть пустым!
			}
		}
		// Credit
		fmt.Println(" @e01@ начата Credit")
		if lenCredit > 0 {
			if count, err := r.doCredit(order, replyBalances); err != nil {
				if lenBlock > 0 {
					r.rollbackBlock(len(order.Block), order)
				}
				if lenUnblock > 0 {
					r.rollbackUnblock(len(order.Block), order)
				}
				r.rollbackCredit(count, order)
				r.answers.Store(num, &Answer{code: 404})
				return num, []byte("") // тут логовое сообщение для ошибочной транзакции - оно должно быть пустым!
			}
		}
		// Debit
		fmt.Println(" @e01@ начата Debit")
		if lenDebit > 0 {
			if count, err := r.doDebit(order, replyBalances); err != nil {
				if lenBlock > 0 {
					r.rollbackBlock(len(order.Block), order)
				}
				if lenUnblock > 0 {
					r.rollbackUnblock(len(order.Block), order)
				}
				if lenCredit > 0 {
					r.rollbackCredit(len(order.Block), order)
				}
				r.rollbackDebit(count, order)
				r.answers.Store(num, &Answer{code: 404})
				return num, []byte("") // тут логовое сообщение для ошибочной транзакции - оно должно быть пустым!
			}
		}

		r.answers.Store(num, &Answer{code: 200, balance: replyBalances})
		fmt.Println(" замыкание запущено под номером: ", num)
		return num, logBytes
	}
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

func (r *Reception) doBlock(order *Order, replyBalances map[string]map[string]account.Balance) (int, error) {
	for i, part := range order.Block {
		acc := r.accounts.Account(part.Id)
		if acc == nil {
			return i - 1, fmt.Errorf("Account %s not found")
		}
		balance, err := acc.Balance(part.Key).Block(order.Hash, part.Amount)
		if err != nil {
			return i, err
		}
		r.balancesAddBalance(part.Id, part.Key, replyBalances, balance)
	}
	return 0, nil
}

func (r *Reception) doUnblock(order *Order, replyBalances map[string]map[string]account.Balance) (int, error) {
	for i, part := range order.Unblock {
		acc := r.accounts.Account(part.Id)
		if acc == nil {
			return i - 1, fmt.Errorf("Account %s not found")
		}
		balance, err := acc.Balance(part.Key).Unblock(order.Hash, part.Amount)
		if err != nil {
			return i, err
		}
		r.balancesAddBalance(part.Id, part.Key, replyBalances, balance)
	}
	return 0, nil
}

func (r *Reception) doCredit(order *Order, replyBalances map[string]map[string]account.Balance) (int, error) {
	for i, part := range order.Credit {
		acc := r.accounts.Account(part.Id)
		if acc == nil {
			return i - 1, fmt.Errorf("Account %s not found")
		}
		balance, err := acc.Balance(part.Key).Credit(order.Hash, part.Amount)
		if err != nil {
			return i, err
		}
		r.balancesAddBalance(part.Id, part.Key, replyBalances, balance)
	}
	return 0, nil
}

func (r *Reception) doDebit(order *Order, replyBalances map[string]map[string]account.Balance) (int, error) {
	for i, part := range order.Debit {
		acc := r.accounts.Account(part.Id)
		if acc == nil {
			return i - 1, fmt.Errorf("Account %s not found")
		}
		balance, err := acc.Balance(part.Key).Debit(part.Amount)
		if err != nil {
			return i, err
		}
		r.balancesAddBalance(part.Id, part.Key, replyBalances, balance)
	}
	return 0, nil
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

func (r *Reception) rollbackCredit(num int, order *Order) {
	for i := 0; i < num; i++ {
		r.accounts.Account(order.Block[i].Id).
			Balance(order.Block[i].Key).
			Debit(order.Block[i].Amount)
		r.accounts.Account(order.Block[i].Id).
			Balance(order.Block[i].Key).
			Block(order.Hash, order.Block[i].Amount)
	}
}
func (r *Reception) rollbackDebit(num int, order *Order) {
	for i := 0; i < num; i++ {
		r.accounts.Account(order.Block[i].Id).
			Balance(order.Block[i].Key).
			WriteOff(order.Block[i].Amount)
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

type Answer struct {
	code    int64
	balance map[string]map[string]account.Balance
}

type Query struct {
	num   int64
	order *Order
	log   []byte
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
