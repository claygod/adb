package account

// Account
// API
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
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

type Balance struct {
	available uint64
	blocked   uint64
}

type SubAccount struct {
	Balance
	//available uint64
	//blocked   uint64
	blocks map[string]uint64
}

/*
newAccount - create new Account.
*/
func newSubAccount() *SubAccount {
	return &SubAccount{
		blocks: make(map[string]uint64),
	}
}

func (s *SubAccount) Debit(amount uint64) (Balance, error) {
	newAviable := s.available + amount
	if newAviable < s.available {
		return s.Balance, fmt.Errorf("Overflow error: there is %d, add %d, get %d. (Debit operation)", s.available, amount, newAviable)
	}
	s.available = newAviable

	return s.Balance, nil
}

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

//func (a *Account) balance() Balance {
//	return Balance{a.available, a.blocked}
//}
