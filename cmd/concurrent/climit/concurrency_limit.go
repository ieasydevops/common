package climit

//////////////////////////////////

// ConcurrencyLimiter is a go routine safe struct that can be used to
// ensure that no more than a specifid max number of goroutines are
// executing.
type ConcurrencyLimiter struct {
	inc   chan chan struct{}
	dec   chan struct{}
	max   int
	count int
}

// NewConcurrencyLimiter returns a configured limiter that will
// ensure that calls to Increment will block if the max is hit.
func NewConcurrencyLimiter(max int) *ConcurrencyLimiter {
	c := &ConcurrencyLimiter{
		inc: make(chan chan struct{}),
		dec: make(chan struct{}, max),
		max: max,
	}
	go c.handleLimits()
	return c
}

// Increment will increase the count of running goroutines by 1.
// if the number is currently at the max, the call to Increment
// will block until another goroutine decrements.
func (c *ConcurrencyLimiter) Increment() {
	r := make(chan struct{})
	c.inc <- r
	<-r
}

// Decrement will reduce the count of running goroutines by 1
func (c *ConcurrencyLimiter) Decrement() {
	c.dec <- struct{}{}
}

// handleLimits runs in a goroutine to manage the count of
// running goroutines.
func (c *ConcurrencyLimiter) handleLimits() {
	for {
		r := <-c.inc
		if c.count >= c.max {
			<-c.dec
			c.count--
		}
		c.count++
		r <- struct{}{}
	}
}
