package adb

// Account database
// Main
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/claygod/adb/account"
)

// Hasp state
const (
	stateClosed int64 = iota
	stateOpen
)
const sizeBucket int64 = 256

const walSeparator string = "|"

func (r *Reception) getClosure(logBytes []byte, order *Order, num int64) func() (int64, []byte) {
	return func() (int64, []byte) {
		replyBalances := make(map[string]map[string]account.Balance)
		lenBlock := len(order.Block)
		lenUnblock := len(order.Unblock)
		lenCredit := len(order.Credit)
		lenDebit := len(order.Debit)
		// Block
		//fmt.Println(" @e01@ начата Block")
		if lenBlock > 0 {
			if count, err := r.doBlock(order, replyBalances); err != nil {
				r.rollbackBlock(count, order)
				r.answers.Store(num, &Answer{code: 404})
				return num, []byte("") // тут логовое сообщение для ошибочной транзакции - оно должно быть пустым!
			}
		}
		// Unblock
		//fmt.Println(" @e01@ начата Unblock")
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
		//fmt.Println(" @e01@ начата Credit")
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
		//fmt.Println(" @e01@ начата Debit")
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
		//fmt.Println(" замыкание запущено под номером: ", num)
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
