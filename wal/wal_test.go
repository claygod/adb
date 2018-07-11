package wal

// WAL
// Test
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"os"
	//"strconv"
	"testing"
	//"time"

	"github.com/claygod/adb/logname"
)

const filePatch = "./log/"

func TestNewWal(t *testing.T) {
	//fileName := strconv.FormatUint((uint64(time.Now().Unix())>>8)<<8, 10)
	ln := logname.New(8)
	fileName := ln.GetName()
	w, err := New(filePatch, ln, "@", ".log")
	if err != nil {
		t.Error(err)
	}
	w.Log("223:35432:USD:+:5")
	w.Log("224:35432:USD:-:2")
	w.Save()
	os.Remove(filePatch + fileName + ".log")
}
