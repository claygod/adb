package logname

// Account database
// Logs or snapshots namer
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

// limitSleep - can be changed
const limitSleep uint = 16 // uint64(1) << 16 = 65536

/*
LogName - name generator for logs or snaphot.
shift (sleep in second):
1 -> 2s
2 -> 4s
3 -> 8s
...
8 -> 256s
...
16 -> 65536s
*/
type LogName struct {
	curNum   uint64
	time     time.Time
	duration time.Duration
	shift    uint
}

func New(shift uint) *LogName {
	if shift > 16 || shift == 0 {
		return nil
	}

	l := &LogName{
		time:  time.Now(),
		shift: shift,
	}

	l.duration = time.Duration(uint64(1) << shift)
	l.curNum = l.genName()
	go l.worker()

	return l
}

func (l *LogName) GetName() string {
	return strconv.FormatUint(atomic.LoadUint64(&l.curNum), 10)
}

func (l *LogName) worker() {
	for {
		cur := l.genName()
		old := atomic.LoadUint64(&l.curNum)
		if cur != old {
			atomic.StoreUint64(&l.curNum, cur)
			time.Sleep((l.duration - time.Duration(cur-old)) * time.Second)
		}
		runtime.Gosched()
	}
}

func (l *LogName) genName() uint64 {
	return (uint64(l.time.Unix()) >> l.shift) << l.shift
}
