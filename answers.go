package adb

// Account database
// Answers storage
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"sync"
)

type Answers struct {
	sync.Mutex
	arr map[int64]*Answer
}

func newAnswers() *Answers {
	return &Answers{
		arr: make(map[int64]*Answer),
	}
}

func (a *Answers) Load(key int64) (*Answer, bool) {
	a.Lock()
	// defer s.RUnlock()
	an, ok := a.arr[key]
	a.Unlock()
	return an, ok
}

func (a *Answers) Store(key int64, an *Answer) {
	a.Lock()
	a.arr[key] = an
	a.Unlock()
}

func (a *Answers) Delete(key int64) {
	a.Lock()
	delete(a.arr, key)
	a.Unlock()
}
