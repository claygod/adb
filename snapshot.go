package adb

// Account database
// Snapshot
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"sync/atomic"
)

type Snapshoter struct {
	path         string
	state        int64
	lastSnapshot int64
	accounts     *Accounts
}

func newSnapshoter(symbol *Symbol, path string) *Snapshoter {
	return &Snapshoter{
		path:     path,
		state:    stateClosed,
		accounts: newAccounts(symbol),
	}
}

func (s *Snapshoter) start() {

}

func (s *Snapshoter) stop() {

}

func (s *Snapshoter) worker() {
	for {
		if atomic.LoadInt64(&s.state) == stateClosed {
			return
		}
	}
}
