package adb

// ADB
// Accounts
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"

	"bytes"
	"strings"
	// "sync"

	"github.com/claygod/adb/account"
)

type Accounts struct {
	//sync.Mutex
	data   map[string]*account.Account
	symbol *Symbol
}

/*
newwAccounts - create new Accounts.
*/
func newAccounts(symbol *Symbol) *Accounts {
	return &Accounts{
		data:   make(map[string]*account.Account),
		symbol: symbol,
	}
}

func (a *Accounts) Account(id string) *account.Account {
	//a.Lock()
	//defer a.Unlock()
	acc, ok := a.data[id]
	if !ok {
		return nil
	}
	return acc
}

func (a *Accounts) AddAccount(id string) bool {
	//a.Lock()
	//defer a.Unlock()
	if _, ok := a.data[id]; ok {
		return false
	}
	a.data[id] = account.New() // WalSimbolSeparator1, WalSimbolSeparator2, "*"
	return true
}
func (a *Accounts) Export() string {
	//a.Lock()
	//defer a.Unlock()
	var buf bytes.Buffer

	for id, acc := range a.data {
		buf.WriteString(id)
		// buf.WriteString(WalSimbolSeparator1)
		buf.WriteString(acc.Export(a.symbol.Separator1, a.symbol.Separator2, a.symbol.Separator3))
		buf.WriteString("\n")
	}

	return buf.String()
}

func (a *Accounts) Import(str string) {
	//a.Lock()
	//defer a.Unlock()
	accs := strings.Split(str, "\n")
	for _, accStr := range accs {
		key := a.ejectKey(accStr, a.symbol.Separator2)
		acc := account.New()
		acc.Import(a.symbol.Separator1, a.symbol.Separator2, a.symbol.Separator3, accStr)
		a.data[key] = acc
	}
}

func (a *Accounts) ejectKey(str string, separator2 string) string {
	subs := strings.SplitN(str, separator2, 2)
	// fmt.Println(" ++^^++ ", subs)
	return subs[0]
}
