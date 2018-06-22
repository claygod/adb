package account

// Account
// Test
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	//"strconv"
	"testing"
)

// var testCount int = 4000000000

func TestExport(t *testing.T) {
	acc := New() // "|", ";", "*"
	acc.Balance("USD").Debit(7)
	acc.Balance("USD").Block("abc", 2)
	//acc.Balance("USD").Block("def", 1)
	//acc.Balance("EUR").Debit(5)
	//acc.Balance("EUR").Block("abc", 1)
	fmt.Println(acc.Export("|", ";", "*"))
	if str := acc.Export("|", ";", "*"); str != "|USD;5;2;abc*2" {
		t.Error("Error in export formatting: ", str)
	}
}

func TestImport(t *testing.T) {
	acc := New() // "|", ";", "*"
	str := "777|USD;4;3;abc*2;def*1|EUR;4;1;abc*1"
	if err := acc.Import("|", ";", "*", str); err != nil {
		t.Error(err)
	}
	a, ok := acc.data["USD"]
	if !ok {
		t.Error("Not exported USD !")
	} else {
		if a.available != 4 || a.blocked != 3 {
			t.Error("Errror in balance")
		}
		sumb, ok := a.blocks["abc"]
		if !ok {
			t.Error("No hash in blocked")
		}
		if sumb != 2 {
			t.Error("Blocked, want=2, fact=", sumb)
		}
	}
}
