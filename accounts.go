package adb

// ADB
// Accounts
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"

	"github.com/claygod/adb/account"
)

type Accounts struct {
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
	acc, ok := a.data[id]
	if !ok {
		return nil
	}
	return acc
}

func (a *Accounts) AddAccount(id string) bool {
	if _, ok := a.data[id]; ok {
		return false
	}
	a.data[id] = account.New()
	return true
}
