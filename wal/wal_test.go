package wal

// WAL
// Test
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"os"
	"strconv"
	"testing"
	"time"
)

const filePatch = "./log/"

func TestNewWal(t *testing.T) {
	fileName := strconv.FormatUint((uint64(time.Now().Unix())>>8)<<8, 10)
	w, err := New(filePatch, fileName+".txt", "@")
	if err != nil {
		t.Error(err)
	}
	w.Log("223:35432:USD:+:5")
	w.Log("224:35432:USD:-:2")
	w.Save()
	os.Remove(filePatch + fileName + ".txt")
}
