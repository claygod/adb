package adb

// Account database
// File operation
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
)

func (a *Adb) loadSnap(snapName string) error {
	if err := a.loadFromDisk(snapName); err != nil {
		return err
	}
	logsList, err := a.listFiles(logExt, a.path)
	if err != nil {
		return err
	}
	logsList, err = a.filterLogs(logsList, snapName)
	if err != nil {
		return err
	}
	for _, logFileName := range logsList {
		ords, err := a.logsFileToOrders(logFileName)
		if err != nil {
			return err
		}
		if err := a.exeOrders(ords); err != nil {
			return err
		}
	}
	return nil
}

func (a *Adb) filterLogs(logs []string, snap string) ([]string, error) {
	logsOut := make([]string, 0, len(logs))
	snapNum, err := strconv.Atoi(strings.TrimRight(snap, "."+snapExt))
	if err != nil {
		return nil, err
	}
	for _, logFileName := range logs {
		logNum, err := strconv.Atoi(strings.TrimRight(logFileName, "."+logExt))
		if err != nil {
			return nil, err
		}
		if logNum >= snapNum {
			logsOut = append(logsOut, logFileName)
		}
	}
	return logsOut, nil
}

func (a *Adb) logsFileToOrders(path string) ([]*Order, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	strs := strings.Split(string(file), "\n")
	//if strs[len(strs)-1] != "\n" {
	//	return
	//}
	ords := make([]*Order, len(strs)-1, len(strs)-1)
	for i := 0; i < len(strs)-1; i++ {
		ord, err := a.logToOrder(strs[i])
		if err != nil {
			return nil, fmt.Errorf("Error in log's string '%s', %v", strs[i], err)
		}
		ords[i] = ord
	}

	return ords, nil
}

func (a *Adb) logToOrder(logStr string) (*Order, error) {
	ord := &Order{}
	subStrs := strings.Split(logStr, a.symbol.Separator1)
	ord.Hash = subStrs[0]
	for i := 1; i < len(subStrs); i++ {
		strs := strings.Split(subStrs[i], a.symbol.Separator2)
		p, err := a.arrStrToPart(strs)
		if err != nil {
			return nil, err
		}
		switch strs[0] {
		case a.symbol.Block:
			ord.Block = append(ord.Block, p)
		case a.symbol.Unblock:
			ord.Unblock = append(ord.Unblock, p)
		case a.symbol.Credit:
			ord.Credit = append(ord.Credit, p)
		case a.symbol.Debit:
			ord.Debit = append(ord.Debit, p)
		}
	}

	return ord, nil
}

func (a *Adb) arrStrToPart(strs []string) (*Part, error) {
	am, err := strconv.ParseUint(strs[3], 10, 64)
	if err != nil {
		return nil, err
	}

	return &Part{Id: strs[1], Key: strs[2], Amount: am}, nil
}

func (a *Adb) saveToDisk() error {
	file, err := os.Create(a.path + "adb" + dbExt)
	if err != nil {
		return err
	}
	_, err = file.WriteString(a.accounts.Export())
	if err != nil {
		return err
	}
	return nil
}

func (a *Adb) loadFromDisk(fileName string) error {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		// panic(err)
		// ToDo: read snapshots & etc.
		_, err := os.Create(fileName)
		if err != nil {
			return err
		}
		return nil
	}
	a.accounts.Import(string(file))
	return nil
}

func (a *Adb) listFiles(ext string, path string) ([]string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	list := make([]string, 0)
	for _, fileName := range files {
		if fns := fileName.Name(); strings.HasSuffix(fns, ext) {
			list = append(list, fns)
		}
	}
	if len(list) > 0 {
		sort.Strings(list)
	}
	return list, nil
}
