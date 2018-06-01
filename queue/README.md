The library implements the queue with a maximum speed of execution.

[![API documentation](https://godoc.org/github.com/claygod/queue?status.svg)](https://godoc.org/github.com/claygod/queue)

The queue can be used both in style LIFO and a FIFO. Also, urgent messages (with high priority) using FIFO can be put in the queue head.

# Usage

An example of using the Queue Library:
```Go
package main

import (
	"fmt"
	"github.com/claygod/queue"
)

func main() {
	q := New()
	q.PushTail(1)
	q.PushTail(2)
	if m, ok := q.PopHead() {
		fmt.Print("\n result (1) = ",m)
	}
	if m, ok := q.PopHead() {
		fmt.Print("\n result (2) = ",m)
	}
}
```

### LIFO

```Go
	q := New()
	// in
	q.PushTail(1)
	q.PushTail(2)
	// out
	m, ok := q.PopTail() // -> 2
	m, ok := q.PopTail() // -> 1
```

### FIFO

```Go
	q := New()
	// in
	q.PushTail(1)
	q.PushTail(2)
	// out
	m, ok := q.PopHead() // -> 1
	m, ok := q.PopHead() // -> 2
```

### Priority

```Go
	q := New()
	// in
	q.PushTail(1)
	q.PushTail(2)
	q.PushHead(3) // immediately in front of the queue
	// out
	m, ok := q.PopHead() // -> 3
	m, ok := q.PopHead() // -> 1
	m, ok := q.PopHead() // -> 2
```

# Perfomance

Operation added to the queue head is more resource-intensive,
so it is necessary to organize the code in the operation of
the library so that the addition was mainly in the tail queue.

Note: in parallel execution speed is decreased. If desired, you can artificially
remove the parallelism in the calling code with a code `runtime.GOMAXPROCS(1)`

```
BenchmarkPushTail-8            	50000000	        22.2 ns/op
BenchmarkPushTailParallel-8    	50000000	        27.4 ns/op
BenchmarkPushHeadLimit-8       	2000000000	         0.20 ns/op
BenchmarkPushHead-8            	 1000000	     12939 ns/op
BenchmarkPopHead-8             	2000000000	         0.01 ns/op
BenchmarkPopTail-8             	2000000000	         0.01 ns/op
BenchmarkQueueList-8           	2000000000	         0.03 ns/op
BenchmarkQueueListParallel-8   	50000000	        31.0 ns/op
```

# API

Methods:
-  *New* - create new queue.
-  *PushTai*l - insert element in the tail queue
-  *PushHead* - paste item in the queue head
-  *PopHead* - get the first element of the queue
-  *PopHeadList* - get the first X elements of the queue
-  *PopTail* - get the item from the queue tail
-  *LenQueue* - the number of elements in the queue
-  *SizeQueue* - the size reserved for queue
-  *clean* - resetting the queue (not thread-safe, is called only after the lock)
-  *lock* - block queue



