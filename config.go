package adb

// Account database
// Config
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

// Hasp state
const (
	stateClosed int64 = iota
	stateOpen
)
const sizeBucket int64 = 256

/*
const (
	WalSimbolSeparator1 string = "|"
	WalSimbolSeparator2 string = ";"
	WalSimbolSeparator3 string = "*"
	WalSimbolBlock      string = "B"
	WalSimbolUnblock    string = "U"
	WalSimbolCredit     string = "C"
	WalSimbolDebit      string = "D"
)
*/
type Symbol struct {
	Separator1 string
	Separator2 string
	Separator3 string
	Block      string
	Unblock    string
	Credit     string
	Debit      string
}

func newSymbol() *Symbol {
	return &Symbol{
		Separator1: "|",
		Separator2: ";",
		Separator3: "*",
		Block:      "B",
		Unblock:    "U",
		Credit:     "C",
		Debit:      "D",
	}
}
