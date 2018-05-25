package accounter

// Accounter
// Test
// Copyright © 2017-2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"testing"
	"time"
)

func TestCreditPrepare(t *testing.T) {

	r := NewReception()
	time.Sleep(10 * time.Millisecond)
	tr := &Transaction{}

	p := &Answer{}
	var a **Answer = &p
	p = nil

	r.DoTransaction(tr, a)

	fmt.Printf("Answer ? %v \r\n", a)
	for i := 0; i < 5; i++ {
		fmt.Printf(" ... Answer wait ...\r\n")
		if *a != nil {
			fmt.Printf("Answer OK!\r\n")
			break
		} else {
			fmt.Printf("Answer NO!\r\n")
		}
		time.Sleep(100 * time.Millisecond)
	}

	/*
		t.Error("Blah-blah-blah.")
	*/
}
