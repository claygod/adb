package adb

// Account database
// Main
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"
	// "time"

	"github.com/claygod/adb/account"
	"github.com/claygod/adb/batcher"
)

// Hasp state
const (
	stateClosed int64 = iota
	stateOpen
)
const sizeBucket int64 = 256

const (
	WalSimbolSeparator1 string = "|"
	WalSimbolSeparator2 string = ";"
	WalSimbolBlock      string = "B"
	WalSimbolUnblock    string = "U"
	WalSimbolCredit     string = "C"
	WalSimbolDebit      string = "D"
)

func (a *Adb) getTask(order *Order, ans *Answer) *batcher.Task {
	t := &batcher.Task{}
	f1 := func() {
		ans.code *= -1
		return
	}
	t.Finish = &f1
	f2 := func() string {
		replyBalances := make(map[string]map[string]account.Balance)
		lenBlock := len(order.Block)
		lenUnblock := len(order.Unblock)
		lenCredit := len(order.Credit)
		lenDebit := len(order.Debit)
		// Block
		//fmt.Println(" @e01@ начата Block")
		if lenBlock > 0 {
			if count, err := a.doBlock(order, replyBalances); err != nil {
				a.rollbackBlock(count, order)
				//r.answers.Store(num, &Answer{code: 404})
				ans.code = -404
				return "" // тут логовое сообщение для ошибочной транзакции - оно должно быть пустым!
			}
		}
		// Unblock
		//fmt.Println(" @e01@ начата Unblock")
		if lenUnblock > 0 {
			if count, err := a.doUnblock(order, replyBalances); err != nil {
				if lenBlock > 0 {
					a.rollbackBlock(len(order.Block), order)
				}
				a.rollbackUnblock(count, order)
				ans.code = -404
				return "" // тут логовое сообщение для ошибочной транзакции - оно должно быть пустым!
			}
		}
		// Credit
		//fmt.Println(" @e01@ начата Credit")
		if lenCredit > 0 {
			if count, err := a.doCredit(order, replyBalances); err != nil {
				if lenBlock > 0 {
					a.rollbackBlock(len(order.Block), order)
				}
				if lenUnblock > 0 {
					a.rollbackUnblock(len(order.Block), order)
				}
				a.rollbackCredit(count, order)
				ans.code = -404
				return "" // тут логовое сообщение для ошибочной транзакции - оно должно быть пустым!
			}
		}
		// Debit
		//fmt.Println(" @e01@ начата Debit")
		if lenDebit > 0 {
			if count, err := a.doDebit(order, replyBalances); err != nil {
				if lenBlock > 0 {
					a.rollbackBlock(len(order.Block), order)
				}
				if lenUnblock > 0 {
					a.rollbackUnblock(len(order.Block), order)
				}
				if lenCredit > 0 {
					a.rollbackCredit(len(order.Block), order)
				}
				a.rollbackDebit(count, order)
				ans.code = -404
				return "" // тут логовое сообщение для ошибочной транзакции - оно должно быть пустым!
			}
		}

		//r.answers.Store(num, &Answer{code: 200, balance: replyBalances})
		ans.code = -200
		ans.balance = replyBalances
		//fmt.Println(" замыкание запущено под номером: ", num)
		// r.wal.Log(r.orderForWal(order)) //

		return a.orderForWal(order)
	}
	t.Main = &f2
	return t
}

func (a *Adb) balancesAddBalance(id string, key string, balances map[string]map[string]account.Balance, balance account.Balance) {
	if _, ok := balances[id]; !ok {
		balances[id] = make(map[string]account.Balance)
	}
	balances[id][key] = balance
}

func (a *Adb) orderToLog(order *Order) ([]byte, error) {
	var orderGob bytes.Buffer
	enc := gob.NewEncoder(&orderGob)
	err := enc.Encode(order)
	if err != nil {
		return nil, err
	}
	return orderGob.Bytes(), nil
}

func (a *Adb) orderForWal(order *Order) string {
	var buf bytes.Buffer

	for _, part := range order.Block {
		buf.WriteString(WalSimbolSeparator1)
		buf.WriteString(WalSimbolBlock)
		a.partToBuf(part, &buf)
	}
	for _, part := range order.Unblock {
		buf.WriteString(WalSimbolSeparator1)
		buf.WriteString(WalSimbolUnblock)
		a.partToBuf(part, &buf)
	}
	for _, part := range order.Credit {
		buf.WriteString(WalSimbolSeparator1)
		buf.WriteString(WalSimbolCredit)
		a.partToBuf(part, &buf)
	}
	for _, part := range order.Debit {
		buf.WriteString(WalSimbolSeparator1)
		buf.WriteString(WalSimbolDebit)
		a.partToBuf(part, &buf)
	}
	return buf.String()
}

func (a *Adb) partToBuf(part *Part, buf *bytes.Buffer) {
	buf.WriteString(WalSimbolSeparator2)
	buf.WriteString(part.Id)
	buf.WriteString(WalSimbolSeparator2)
	buf.WriteString(part.Key)
	buf.WriteString(WalSimbolSeparator2)
	buf.WriteString(strconv.FormatUint(part.Amount, 10))
}

func (a *Adb) doBlock(order *Order, replyBalances map[string]map[string]account.Balance) (int, error) {
	for i, part := range order.Block {
		acc := a.accounts.Account(part.Id)
		if acc == nil {
			return i - 1, fmt.Errorf("Account %s not found", part.Id)
		}
		balance, err := acc.Balance(part.Key).Block(order.Hash, part.Amount)
		if err != nil {
			return i, err
		}
		a.balancesAddBalance(part.Id, part.Key, replyBalances, balance)
	}
	return 0, nil
}

func (a *Adb) doUnblock(order *Order, replyBalances map[string]map[string]account.Balance) (int, error) {
	for i, part := range order.Unblock {
		acc := a.accounts.Account(part.Id)
		if acc == nil {
			return i - 1, fmt.Errorf("Account %s not found", part.Id)
		}
		balance, err := acc.Balance(part.Key).Unblock(order.Hash, part.Amount)
		if err != nil {
			return i, err
		}
		a.balancesAddBalance(part.Id, part.Key, replyBalances, balance)
	}
	return 0, nil
}

func (a *Adb) doCredit(order *Order, replyBalances map[string]map[string]account.Balance) (int, error) {
	for i, part := range order.Credit {
		acc := a.accounts.Account(part.Id)
		if acc == nil {
			return i - 1, fmt.Errorf("Account %s not found", part.Id)
		}
		balance, err := acc.Balance(part.Key).Credit(order.Hash, part.Amount)
		if err != nil {
			return i, err
		}
		a.balancesAddBalance(part.Id, part.Key, replyBalances, balance)
	}
	return 0, nil
}

func (a *Adb) doDebit(order *Order, replyBalances map[string]map[string]account.Balance) (int, error) {
	for i, part := range order.Debit {
		acc := a.accounts.Account(part.Id)
		if acc == nil {
			return i - 1, fmt.Errorf("Account %s not found", part.Id)
		}
		balance, err := acc.Balance(part.Key).Debit(part.Amount)
		if err != nil {
			return i, err
		}
		a.balancesAddBalance(part.Id, part.Key, replyBalances, balance)
	}
	return 0, nil
}

func (a *Adb) rollbackBlock(num int, order *Order) {
	for i := 0; i < num; i++ {
		a.accounts.Account(order.Block[i].Id).
			Balance(order.Block[i].Key).
			Unblock(order.Hash, order.Block[i].Amount)
	}
}

func (a *Adb) rollbackUnblock(num int, order *Order) {
	for i := 0; i < num; i++ {
		a.accounts.Account(order.Block[i].Id).
			Balance(order.Block[i].Key).
			Block(order.Hash, order.Block[i].Amount)
	}
}

func (a *Adb) rollbackCredit(num int, order *Order) {
	for i := 0; i < num; i++ {
		a.accounts.Account(order.Block[i].Id).
			Balance(order.Block[i].Key).
			Debit(order.Block[i].Amount)
		a.accounts.Account(order.Block[i].Id).
			Balance(order.Block[i].Key).
			Block(order.Hash, order.Block[i].Amount)
	}
}
func (a *Adb) rollbackDebit(num int, order *Order) {
	for i := 0; i < num; i++ {
		a.accounts.Account(order.Block[i].Id).
			Balance(order.Block[i].Key).
			WriteOff(order.Block[i].Amount)
	}
}
