package adb

// ADB
// Accounts
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"

	"bytes"
	"sync"

	"github.com/claygod/adb/account"
)

type Accounts struct {
	sync.Mutex
	data map[string]*account.Account
}

/*
newwAccounts - create new Accounts.
*/
func newAccounts() *Accounts {
	return &Accounts{
		data: make(map[string]*account.Account),
	}
}

func (a *Accounts) Account(id string) *account.Account {
	a.Lock()
	defer a.Unlock()
	acc, ok := a.data[id]
	if !ok {
		return nil
	}
	return acc
}

func (a *Accounts) AddAccount(id string) bool {
	a.Lock()
	defer a.Unlock()
	if _, ok := a.data[id]; ok {
		return false
	}
	a.data[id] = account.New() // WalSimbolSeparator1, WalSimbolSeparator2, "*"
	return true
}
func (a *Accounts) Export() string {
	a.Lock()
	defer a.Unlock()
	var buf bytes.Buffer

	for id, acc := range a.data {
		buf.WriteString(id)
		buf.WriteString(WalSimbolSeparator1)
		buf.WriteString(acc.Export(WalSimbolSeparator1, WalSimbolSeparator2, "*"))
		buf.WriteString("\n")
	}

	return buf.String()
}
