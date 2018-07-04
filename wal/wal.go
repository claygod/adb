package wal

// WAL
// API
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"bytes"
	//"io/ioutil"
	"os"
	//"sort"
	//"strings"
	"sync"
	"time"
)

type Wal struct {
	m         sync.Mutex
	file      *os.File
	buf       bytes.Buffer
	time      time.Time
	separator string
	path      string
}

func New(path string, fileName string, separator string) (*Wal, error) {
	//file, err := os.Create(patch + fileName)
	//if err != nil {
	//	return nil, err
	//}
	w := &Wal{
		m: sync.Mutex{},
		//file:      file,
		time:      time.Now(),
		separator: separator,
		path:      path,
	}
	return w, nil
}

func (w *Wal) Log(s string) error {

	var buf bytes.Buffer
	//if _, err := buf.WriteString(w.time.String()); err != nil {
	//	return err
	//}
	if _, err := buf.WriteString(s); err != nil {
		return err
	}
	if _, err := buf.WriteString("\n"); err != nil {
		return err
	}
	w.m.Lock()
	w.file.WriteString(buf.String())
	w.m.Unlock()
	return nil
}

func (w *Wal) Save() error {
	w.m.Lock()
	defer w.m.Unlock()
	return w.file.Sync()
}

func (w *Wal) Filename(fileName string) error {
	w.m.Lock()
	w.file.Close()
	file, err := os.Create(w.path + fileName)
	if err != nil {
		return err
	}
	w.file = file
	w.m.Unlock()
	return nil
}

func (w *Wal) Close() error {
	w.m.Lock()
	defer w.m.Unlock()
	return w.file.Close()
}

/*
func (w *Wal) ListLogFiles(logExt string) ([]string, error) {
	// No parallel mode!
	//w.m.Lock()
	//defer w.m.Unlock()
	files, err := ioutil.ReadDir(w.path)
	if err != nil {
		return nil, err
	}
	logs := make([]string, 0)
	for _, fileName := range files {
		if fns := fileName.Name(); strings.HasSuffix(fns, logExt) {
			logs = append(logs, fns)
		}
	}
	if len(logs) > 0 {
		sort.Strings(logs)
	}
	return logs, nil
}
*/
