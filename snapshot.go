package adb

// Account database
// Snapshot
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
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

func (s *Snapshoter) start() error {
	files, err := ioutil.ReadDir(s.path)
	if err != nil {
		return err
	}
	snaps := make([]string, 0)
	logs := make([]string, 0)
	for _, fileName := range files {
		fns := fileName.Name()
		if strings.HasSuffix(fns, snapExt) {
			snaps = append(snaps, fns)
		}
		if strings.HasSuffix(fns, logExt) {
			logs = append(logs, fns)
		}
	}
	// При старте по идее не должно быть снапов и логов!
	if len(snaps) > 0 || len(logs) > 0 {
		sort.Strings(snaps)
		sort.Strings(logs)
		return fmt.Errorf("Logs: %v \nSnaps:%v", logs, snaps)
	}
	return nil
}

func (s *Snapshoter) clean() error {
	files, err := ioutil.ReadDir(s.path)
	if err != nil {
		return err
	}
	for _, fileName := range files {
		fns := fileName.Name()
		if strings.HasSuffix(fns, snapExt) || strings.HasSuffix(fns, logExt) {
			if err := os.Remove(s.path + fns); err != nil {
				return nil
			}
		}
	}
	return nil
}

func (s *Snapshoter) currentSnap() (string, error) {
	files, err := ioutil.ReadDir(s.path)
	if err != nil {
		return "", err
	}
	snaps := make([]string, 0)
	for _, fileName := range files {
		fns := fileName.Name()
		if strings.HasSuffix(fns, snapExt) {
			snaps = append(snaps, fns)
		}
	}

	if len(snaps) > 0 {
		sort.Strings(snaps)
		if ln := len(snaps); ln > 1 {
			return snaps[ln-2], nil
		} else {
			return snaps[ln-1], nil
		}
	}
	return "", fmt.Errorf("Snapshot does not exists")
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
