package account

// Account
// API
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type Account struct {
	data map[string]*SubAccount
}

/*
New - create new Account.
*/
func New() *Account {
	return &Account{
		data: make(map[string]*SubAccount),
	}
}

func (a *Account) Balance(id string) *SubAccount {
	acc, ok := a.data[id]
	if !ok {
		acc = newSubAccount()
		a.data[id] = acc
	}
	return acc
}

func (a *Account) Export(separator1 string, separator2 string, separator3 string) string {
	var buf bytes.Buffer
	for key, acc := range a.data {
		buf.WriteString(separator1)
		buf.WriteString(key)
		buf.WriteString(separator2)
		buf.WriteString(acc.Export(separator2, separator3))
	}
	return buf.String()
}

func (a *Account) Import(separator1 string, separator2 string, separator3 string, str string) error {
	subs := strings.Split(str, separator1)
	for i := 1; i < len(subs); i++ {
		key := a.ejectKey(subs[i], separator2)
		s, err := importSubAccount(subs[i], separator2, separator3)
		if err != nil {
			return err
		}
		a.data[key] = s
	}
	return nil
}

func (a *Account) ejectKey(str string, separator2 string) string {
	subs := strings.SplitN(str, separator2, 2)
	return subs[0]
}

type Balance struct {
	available uint64
	blocked   uint64
}

type SubAccount struct {
	Balance
	blocks map[string]uint64
}

/*
newAccount - create new Account.
*/
func newSubAccount() *SubAccount {
	return &SubAccount{
		Balance: Balance{},
		blocks:  make(map[string]uint64),
	}
}

func importSubAccount(str string, separator2 string, separator3 string) (*SubAccount, error) {
	s := newSubAccount()
	subs := strings.Split(str, separator2)

	available, err := strconv.ParseUint(subs[1], 10, 64)
	if err != nil {
		return nil, err
	}
	s.Balance.available = available

	blocked, err := strconv.ParseUint(subs[2], 10, 64)
	if err != nil {
		return nil, err
	}
	s.Balance.blocked = blocked

	for i := 3; i < len(subs); i++ {
		bl := strings.Split(subs[i], separator3)
		sum, err := strconv.ParseUint(bl[1], 10, 64)
		if err != nil {
			return nil, err
		}
		s.blocks[bl[0]] = sum
	}
	return s, nil
}

func (s *SubAccount) Export(separator1 string, separator2 string) string {
	var buf bytes.Buffer
	buf.WriteString(strconv.FormatUint(s.Balance.available, 10))
	buf.WriteString(separator1)
	buf.WriteString(strconv.FormatUint(s.Balance.blocked, 10))

	for k, u := range s.blocks {
		buf.WriteString(separator1)
		buf.WriteString(k)
		buf.WriteString(separator2)
		buf.WriteString(strconv.FormatUint(u, 10))
	}

	return buf.String()
}

func (s *SubAccount) Debit(amount uint64) (Balance, error) {
	newAviable := s.available + amount
	if newAviable < s.available {
		return s.Balance, fmt.Errorf("Overflow error: there is %d, add %d, get %d. (Debit operation)", s.available, amount, newAviable)
	}
	s.available = newAviable

	return s.Balance, nil
}

func (s *SubAccount) DebitUnsafe(amount uint64) {
	newAviable := s.available + amount
	s.available = newAviable
}

/*
Тут есть два варианта по блокированию:
1) блокируется на конкретный хэш конкретная сумма, и блокированная сумма должна точно совпадать со списываемой потом
2) на хэш может приходить и блокироваться несколько сумм, и из них потом списывается

можно сделать два режима (в перспективе)

И ещё момент - как быть с блокированием при сохранении..?
*/

func (s *SubAccount) Block(key string, amount uint64) (Balance, error) {
	if _, ok := s.blocks[key]; ok {
		return s.Balance, fmt.Errorf("This key is already taken.")
	}
	if s.available < amount {
		return s.Balance, fmt.Errorf("Blocking error - there is %d, but blocked %d.", s.available, amount)
	}

	newAviable := s.available - amount
	newBlocked := s.blocked + amount
	if newBlocked < s.blocked {
		return s.Balance, fmt.Errorf("Overflow error: there is %d, add %d, get %d. (Block operation)", s.blocked, amount, newBlocked)
	}
	s.blocks[key] = amount
	s.available = newAviable
	s.blocked = newBlocked
	return s.Balance, nil
}

func (s *SubAccount) BlockUnsafe(key string, amount uint64) {
	newAviable := s.available - amount
	newBlocked := s.blocked + amount
	s.blocks[key] = amount
	s.available = newAviable
	s.blocked = newBlocked
}

func (s *SubAccount) BlockNoFix(amount uint64) (Balance, error) {
	if s.available < amount {
		return s.Balance, fmt.Errorf("Blocking error - there is %d, but blocked %d.", s.available, amount)
	}
	newAviable := s.available - amount
	newBlocked := s.blocked + amount
	if newBlocked < s.blocked {
		return s.Balance, fmt.Errorf("Overflow error: there is %d, add %d, get %d. (Block operation)", s.blocked, amount, newBlocked)
	}
	s.available = newAviable
	s.blocked = newBlocked
	return s.Balance, nil
}

func (s *SubAccount) Unblock(key string, amount uint64) (Balance, error) {
	sum, ok := s.blocks[key]
	if !ok {
		return s.Balance, fmt.Errorf("This key is missing.")
	}
	if sum != amount {
		return s.Balance, fmt.Errorf("The amount does not match the blocked amount..")
	}
	newAviable := s.available + amount
	newBlocked := s.blocked - amount

	if newAviable < s.available {
		return s.Balance, fmt.Errorf("Overflow error: there is %d, add %d, get %d. (Unlock operation)", s.available, amount, newAviable)
	}
	delete(s.blocks, key)
	s.available = newAviable
	s.blocked = newBlocked
	return s.Balance, nil
}

func (s *SubAccount) UnblockUnsafe(key string, amount uint64) {
	newAviable := s.available + amount
	newBlocked := s.blocked - amount
	delete(s.blocks, key)
	s.available = newAviable
	s.blocked = newBlocked
}

func (s *SubAccount) UnblockNoFix(amount uint64) (Balance, error) {
	newAviable := s.available + amount
	newBlocked := s.blocked - amount
	if newAviable < s.available {
		return s.Balance, fmt.Errorf("Overflow error: there is %d, add %d, get %d. (Unlock operation)", s.available, amount, newAviable)
	}
	s.available = newAviable
	s.blocked = newBlocked
	return s.Balance, nil
}

func (s *SubAccount) Credit(key string, amount uint64) (Balance, error) {
	sum, ok := s.blocks[key]
	if !ok {
		return s.Balance, fmt.Errorf("This key is missing.")
	}
	if sum != amount {
		return s.Balance, fmt.Errorf("The amount does not match the blocked amount..")
	}
	newBlocked := s.blocked - amount

	delete(s.blocks, key)
	s.blocked = newBlocked
	return s.Balance, nil
}

func (s *SubAccount) CreditUnsafe(key string, amount uint64) {
	newBlocked := s.blocked - amount
	delete(s.blocks, key)
	s.blocked = newBlocked
}

func (s *SubAccount) WriteOff(amount uint64) (Balance, error) { //  Credit operation without intermediate blocking of funds.
	if s.available < amount {
		return s.Balance, fmt.Errorf("Blocking error - there is %d, but blocked %d.", s.available, amount)
	}
	newAviable := s.available - amount
	if newAviable > s.available {
		return s.Balance, fmt.Errorf("Overflow error: there is %d, add %d, get %d. (WriteOff operation)", s.available, amount, newAviable)
	}
	s.available = newAviable
	return s.Balance, nil
}

