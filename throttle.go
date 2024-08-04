// throttle is a rate limiter inspired by one from the Linux kernel.
//
// I didn't save the reference to the original code from the very beginning,
// but this code seems similar.
// https://github.com/torvalds/linux/blob/795c58e4c7fc6163d8fb9f2baa86cfe898fa4b19/net/netfilter/xt_limit.c#L63
package throttle

import (
	"context"
	"time"
)

type (
	Throttle struct {
		Bucket
		Delay time.Duration
	}
)

func New(ts int64, delay, limit time.Duration) *Throttle {
	b := &Throttle{}
	b.Reset(ts, delay, limit)

	return b
}

func (b *Throttle) Reset(ts int64, delay, limit time.Duration) {
	b.Bucket.Reset(ts, limit)
	b.Delay = delay
}

func (b *Throttle) ValueT(now time.Time) int {
	return b.Value(now.UnixNano())
}

// Value returns the number of tokens we have so far.
func (b *Throttle) Value(ts int64) int {
	return int(b.Bucket.Value(ts) / b.Delay)
}

func (b *Throttle) HaveT(now time.Time, n int) bool {
	return b.Have(now.UnixNano(), n)
}

// Have checks if we have n tokens to take.
func (b *Throttle) Have(ts int64, n int) bool {
	return b.Bucket.Have(ts, b.Delay*dur(n))
}

func (b *Throttle) TakeT(now time.Time, n int) bool {
	return b.Take(now.UnixNano(), n)
}

// Take takes n tokens.
// It returns if it was enough tokens to take.
// If operaion wasn't successful, tokens are not taken.
func (b *Throttle) Take(ts int64, n int) bool {
	return b.Bucket.Take(ts, b.Delay*dur(n))
}

func (b *Throttle) BorrowT(now time.Time, n int) time.Duration {
	return b.Borrow(now.UnixNano(), n)
}

// Borrow borrows n tokens event if there is not enough right now.
// It returns duration needed to wait before we can use them.
// If returned duration is 0, it means we've had enough tokens
// and it basically worked the same as Take.
//
// Borrow even works if you ask more tokens than fit into the limit.
func (b *Throttle) Borrow(ts int64, n int) time.Duration {
	return b.Bucket.Borrow(ts, b.Delay*dur(n))
}

func (b *Throttle) WaitT(ctx context.Context, now time.Time, n int) error {
	return b.Wait(ctx, now.UnixNano(), n)
}

// Wait borrows n tokens and waits until we can use it.
// It returns with ctx.Err() if context was canceled.
func (b *Throttle) Wait(ctx context.Context, ts int64, n int) error {
	return b.Bucket.Wait(ctx, ts, b.Delay*dur(n))
}

func (b *Throttle) SetValueT(now time.Time, n int) {
	b.SetValue(now.UnixNano(), n)
}

// SetValue sets the number of tokens available at the time ts.
func (b *Throttle) SetValue(ts int64, n int) {
	b.Bucket.SetValue(ts, b.Delay*dur(n))
}

func (b *Throttle) ReturnT(now time.Time, n int) {
	b.Return(now.UnixNano(), n)
}

// Return puts n tokens back. You can add them even if you didn't take it.
func (b *Throttle) Return(ts int64, n int) {
	b.Bucket.Return(ts, b.Delay*dur(n))
}

// Delay calculate a delay of one token to achive limit of tokens/per.
// It's used for Throttle creation.
func Delay(tokens int64, per time.Duration) time.Duration {
	return per / time.Duration(tokens)
}
