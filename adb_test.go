package adb

// Account database
// Test
// Copyright © 2017-2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	"testing"
	//"time"
	// "github.com/claygod/adb/transaction"
)

func TestTime1000000Trans(t *testing.T) {

	//tc := transaction.New()
	//if !tc.Start() {
	// t.Error("Now the start is possible!")
	//}
	r, _ := New(filePatch)
	r.accounts.AddAccount("111")
	r.accounts.AddAccount("112")
	r.accounts.Account("111").Balance("USD").Debit(9)
	r.accounts.Account("112").Balance("USD").Debit(9)
	p1 := &Part{Id: "111", Key: "USD", Amount: 0}
	p2 := &Part{Id: "112", Key: "USD", Amount: 5}
	minus := []*Part{p1}
	plus := []*Part{p2}
	r.ExeTransaction(&Order{
		Credit: minus,
		Debit:  plus,
	})
}

/*
func TestTime1000000Trans(t *testing.T) {
	tc := transaction.New()
	if !tc.Start() {
		// t.Error("Now the start is possible!")
	}
	r := NewReception(&tc)

	cnt := 100
	// prepare
	tArray := ForTestGenTransactionsArray(cnt)
	// time.Sleep(1 * time.Second)

	for i := 0; i < cnt; i++ {
		r.ExeTransaction(tArray[i])
	}
}


func TestCreditPrepare(t *testing.T) {
	tc := transaction.New()
	if !tc.Start() {
		t.Error("Now the start is possible!")
	}
	r := NewReception(&tc)

	time.Sleep(10 * time.Millisecond)
	tr := &Transaction{}

	//p := &Answer{}
	//var a **Answer = &p
	//p = nil

	num := r.DoTransaction(tr, 0)

	fmt.Printf("Answer ?  \r\n")
	for i := 0; i < 5; i++ {
		fmt.Printf(" ... Answer wait ...\r\n")
		if a := r.GetAnswer(num); a != nil {
			fmt.Printf("Answer OK!\r\n")
			break
		} else {
			fmt.Printf("Answer NO!\r\n")
		}
		time.Sleep(100 * time.Millisecond)
	}


}
*/
/*
	t.Error("Blah-blah-blah.")
*/
