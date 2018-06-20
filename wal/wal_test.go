package wal

// WAL
// Test
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"os"
	"testing"
)

const filePatch = "./test.txt"

func TestNewWal(t *testing.T) {
	w, err := New(filePatch, "@")
	if err != nil {
		t.Error(err)
	}
	w.Log("223:35432:USD:+:5")
	w.Log("224:35432:USD:-:2")
	w.Save()
	os.Remove(filePatch)
}
