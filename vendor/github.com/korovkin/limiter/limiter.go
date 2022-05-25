package limiter

import (
	"errors"
	"sync/atomic"
)

const (
	// DefaultLimit is the default concurrency limit
	DefaultLimit = 100
)

var (
	// appending a callback to a closed limiter
	ErrorClosed = errors.New("limiter closed")
)

// ConcurrencyLimiter object
type ConcurrencyLimiter struct {
	limit         int
	tickets       chan int
	numInProgress int32
}

// NewConcurrencyLimiter allocates a new ConcurrencyLimiter
func NewConcurrencyLimiter(limit int) *ConcurrencyLimiter {
	if limit <= 0 {
		limit = DefaultLimit
	}

	c := &ConcurrencyLimiter{
		limit:   limit,
		tickets: make(chan int, limit),
	}

	// allocate the tickets:
	for i := 0; i < c.limit; i++ {
		c.tickets <- i
	}

	return c
}

// Execute adds a function to the execution queue.
// if num of go routines allocated by this instance is < limit
// launch a new go routine to execute job
// else wait until a go routine becomes available
func (c *ConcurrencyLimiter) Execute(job func()) (int, error) {
	ticket, opened := <-c.tickets
	if !opened {
		return -1, ErrorClosed
	}
	atomic.AddInt32(&c.numInProgress, 1)
	go func() {
		defer func() {
			c.tickets <- ticket
			atomic.AddInt32(&c.numInProgress, -1)
		}()

		// run the job
		job()
	}()
	return ticket, nil
}

// ExecuteWithTicket adds a job into an execution queue and returns a ticket id.
// if num of go routines allocated by this instance is < limit
// launch a new go routine to execute job
// else wait until a go routine becomes available
func (c *ConcurrencyLimiter) ExecuteWithTicket(job func(ticket int)) (int, error) {
	ticket, opened := <-c.tickets
	if !opened {
		return -1, ErrorClosed
	}
	atomic.AddInt32(&c.numInProgress, 1)
	go func() {
		defer func() {
			c.tickets <- ticket
			atomic.AddInt32(&c.numInProgress, -1)
		}()

		// run the job
		job(ticket)
	}()
	return ticket, nil
}

// WaitAndClose will block until all the previously Executed jobs completed running.
// New tasks won't be allow
//
// IMPORTANT: calling the Wait function while keep calling Execute leads to
//            un-desired race conditions
func (c *ConcurrencyLimiter) WaitAndClose() error {
	for i := 0; i < c.limit; i++ {
		<-c.tickets
	}
	return c.close()
}

// GetNumInProgress returns a (racy) counter of how many go routines are active right now
func (c *ConcurrencyLimiter) GetNumInProgress() int32 {
	return atomic.LoadInt32(&c.numInProgress)
}

// close the limiter and free the tickets channel
func (c *ConcurrencyLimiter) close() error {
	close(c.tickets)
	return nil
}
