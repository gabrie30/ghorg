package limiter

import (
	"sync/atomic"
)

const (
	DefaultConcurrencyLimitIO = 4
)

// Concurrently - execute tasks (IO) concurrently, keep track of the first error atomically
type Concurrently struct {
	conc       *ConcurrencyLimiter
	firstError atomic.Value // type: error
	// TODO:: we can also add an atomic list of errors here if needed.
}

func NewConcurrencyLimiterForIO(limit int) *Concurrently {
	c := &Concurrently{
		conc:       NewConcurrencyLimiter(limit),
		firstError: atomic.Value{},
	}
	if c.conc == nil {
		c = nil
	}
	return c
}

func (c *Concurrently) Execute(job func()) (int, error) {
	return c.conc.Execute(job)
}

func (c *Concurrently) WaitAndClose() error {
	return c.conc.WaitAndClose()
}

func (c *Concurrently) FirstErrorStore(e error) (bool, error) {
	stored := false
	if e != nil {
		stored = c.firstError.CompareAndSwap(nil, e)
	}
	return stored, e
}

func (c *Concurrently) FirstErrorGet() error {
	e := c.firstError.Load()
	if e == nil {
		return nil
	}
	err := e.(error)
	return err
}
