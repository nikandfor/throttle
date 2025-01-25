package throttle

import (
	"context"
	"time"
)

type (
	AtomicThrottle struct {
		AtomicBucket
		Delay time.Duration
	}
)

func NewAtomic(ts int64, delay, limit time.Duration) *AtomicThrottle {
	b := &AtomicThrottle{}
	b.Reset(ts, delay, limit)

	return b
}

func (b *AtomicThrottle) Reset(ts int64, delay, limit time.Duration) {
	b.AtomicBucket.Reset(ts, limit)
	b.Delay = delay
}

func (b *AtomicThrottle) ValueT(now time.Time) int {
	return b.Value(now.UnixNano())
}

// Value returns the number of tokens we have so far.
func (b *AtomicThrottle) Value(ts int64) int {
	return int(b.AtomicBucket.Value(ts) / b.Delay)
}

func (b *AtomicThrottle) HaveT(now time.Time, n int) bool {
	return b.Have(now.UnixNano(), n)
}

// Have checks if we have n tokens to take.
func (b *AtomicThrottle) Have(ts int64, n int) bool {
	return b.AtomicBucket.Have(ts, b.Delay*dur(n))
}

func (b *AtomicThrottle) TakeT(now time.Time, n int) bool {
	return b.Take(now.UnixNano(), n)
}

// Take takes n tokens.
// It returns if it was enough tokens to take.
// If operaion wasn't successful, tokens are not taken.
func (b *AtomicThrottle) Take(ts int64, n int) bool {
	return b.AtomicBucket.Take(ts, b.Delay*dur(n))
}

func (b *AtomicThrottle) BorrowT(now time.Time, n int) time.Duration {
	return b.Borrow(now.UnixNano(), n)
}

// Borrow borrows n tokens event if there is not enough right now.
// It returns duration needed to wait before we can use them.
// If returned duration is 0, it means we've had enough tokens
// and it basically worked the same as Take.
//
// Borrow even works if you ask more tokens than fit into the limit.
func (b *AtomicThrottle) Borrow(ts int64, n int) time.Duration {
	return b.AtomicBucket.Borrow(ts, b.Delay*dur(n))
}

func (b *AtomicThrottle) WaitT(ctx context.Context, now time.Time, n int) error {
	return b.Wait(ctx, now.UnixNano(), n)
}

// Wait borrows n tokens and waits until we can use it.
// It returns with ctx.Err() if context was canceled.
func (b *AtomicThrottle) Wait(ctx context.Context, ts int64, n int) error {
	return b.AtomicBucket.Wait(ctx, ts, b.Delay*dur(n))
}

func (b *AtomicThrottle) SetValueT(now time.Time, n int) {
	b.SetValue(now.UnixNano(), n)
}

// SetValue sets the number of tokens available at the time ts.
func (b *AtomicThrottle) SetValue(ts int64, n int) {
	b.AtomicBucket.SetValue(ts, b.Delay*dur(n))
}

func (b *AtomicThrottle) ReturnT(now time.Time, n int) {
	b.Return(now.UnixNano(), n)
}

// Return puts n tokens back. You can add them even if you didn't take it.
func (b *AtomicThrottle) Return(ts int64, n int) {
	b.AtomicBucket.Return(ts, b.Delay*dur(n))
}
