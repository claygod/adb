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
	//separator1 string
	//separator2 string
	//separator3 string
	data map[string]*SubAccount
}

/*
New - create new Account.
*/
func New() *Account { // separator1 string, separator2 string, separator3 string
	return &Account{
		//separator1: separator1,
		//separator2: separator2,
		//separator3: separator3,
		data: make(map[string]*SubAccount),
	}
}

/*
func Import(args ...string) (*Account, error) { // separator1 string, separator2 string, separator3 string
	switch len(args) {
	case 3:
		return &Account{
			separator1: args[0],
			separator2: args[1],
			separator3: args[2],
			data:       make(map[string]*SubAccount),
		}, nil
	case 4:
		acc := &Account{
			separator1: args[0],
			separator2: args[1],
			separator3: args[2],
			data:       make(map[string]*SubAccount),
		}
		if err := acc.Import(args[3]); err != nil {
			return nil, err
		}
		return acc, nil

	default:
		return nil, fmt.Errorf("Invalid number of arguments")
	}

}
*/
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
	//fmt.Println(" +++++ ", subs)
	for i := 1; i < len(subs); i++ {
		key := a.ejectKey(subs[i], separator2)
		//fmt.Println(" ++", key, "+++++++++++++++++++++++++++ ")
		//fmt.Println(" ++", subs[i], "++ ")
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
	// fmt.Println(" ++^^++ ", subs)
	return subs[0]
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

/*
Тут есть два варианта по блокированию:
1) блокируется на конкретный хэш конкретная сумма, и блокированная сумма должна точно совпадать со списываемой потом
2) на хэш может приходить и блокироваться несколько сумм, и из них потом списывается

можно сделать два режима (в перспективе)
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
