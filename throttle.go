package throttle

import (
	"context"
	"time"
)

type (
	Bucket struct {
		lastts int64
		Limit  int64
	}

	Throttle struct {
		Bucket
		Price int64
	}
)

func NewBucket(ts, limit int64) *Bucket {
	return &Bucket{lastts: ts, Limit: limit}
}

func (b *Bucket) Reset(ts, limit int64) {
	b.lastts = ts
	b.Limit = limit
}

func NewRateLimit(value, rate, limit float64) *Throttle {
	ts := time.Now().UnixNano()
	price := int64(1e9 / rate)
	lim := int64(1e9 * rate * limit)
	tokens := ts - int64(1e9/rate*value)

	return New(ts-tokens, price, lim)
}

func NewRateWindow(value, rate float64, window time.Duration) *Throttle {
	ts := time.Now().UnixNano()
	price := int64(1e9 / rate)
	limit := window.Nanoseconds()
	tokens := ts - int64(1e9/rate*value)

	return New(ts-tokens, price, limit)
}

func New(ts, price, limit int64) *Throttle {
	b := &Throttle{}
	b.Reset(ts, price, limit)

	return b
}

func (b *Throttle) Reset(ts, price, limit int64) {
	b.Bucket.Reset(ts, limit)
	b.Price = price
}

func (b *Throttle) ValueT(now time.Time) int {
	return b.Value(now.UnixNano())
}

func (b *Throttle) Value(ts int64) int {
	return int(b.Bucket.Value(ts) / b.Price)
}

func (b *Bucket) Value(ts int64) int64 {
	b.advance(ts)

	return ts - b.lastts
}

func (b *Throttle) HaveT(now time.Time, n int) bool {
	return b.Have(now.UnixNano(), n)
}

func (b *Throttle) Have(ts int64, n int) bool {
	return b.Bucket.Have(ts, b.Price*int64(n))
}

func (b *Bucket) Have(ts, cost int64) bool {
	b.advance(ts)

	return b.tokens(ts) >= cost
}

func (b *Throttle) TakeT(now time.Time, n int) bool {
	return b.Take(now.UnixNano(), n)
}

func (b *Throttle) Take(ts int64, n int) bool {
	return b.Bucket.Take(ts, b.Price*int64(n))
}

func (b *Bucket) Take(ts, cost int64) bool {
	if !b.Have(ts, cost) {
		return false
	}

	b.spend(cost)

	return true
}

func (b *Throttle) BorrowT(now time.Time, n int) time.Duration {
	return b.Borrow(now.UnixNano(), n)
}

func (b *Throttle) Borrow(ts int64, n int) time.Duration {
	return b.Bucket.Borrow(ts, b.Price*int64(n))
}

func (b *Bucket) Borrow(ts, cost int64) time.Duration {
	b.advance(ts)
	b.spend(cost)

	d := time.Duration(b.lastts - ts)
	if d < 0 {
		d = 0
	}

	return d
}

func (b *Throttle) WaitT(ctx context.Context, now time.Time, n int) error {
	return b.Wait(ctx, now.UnixNano(), n)
}

func (b *Throttle) Wait(ctx context.Context, ts int64, n int) error {
	return b.Bucket.Wait(ctx, ts, b.Price*int64(n))
}

func (b *Bucket) Wait(ctx context.Context, ts, cost int64) error {
	d := b.Borrow(ts, cost)
	if d == 0 {
		return nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (b *Throttle) SetValueT(now time.Time, n int) {
	b.SetValue(now.UnixNano(), n)
}

func (b *Throttle) SetValue(ts int64, n int) {
	b.Bucket.SetValue(ts, b.Price*int64(n))
}

func (b *Bucket) SetValue(ts, cost int64) {
	b.lastts = ts - cost
}

func (b *Throttle) ReturnT(now time.Time, n int) {
	b.Return(now.UnixNano(), n)
}

func (b *Throttle) Return(ts int64, n int) {
	b.Bucket.Return(ts, b.Price*int64(n))
}

func (b *Bucket) Return(ts, cost int64) {
	b.spend(-cost)
}

func (b *Bucket) advance(ts int64) {
	tokens := ts - b.lastts

	if tokens > b.Limit {
		tokens = b.Limit
	}

	b.lastts = ts - tokens
}

func (b *Bucket) tokens(ts int64) int64 {
	return ts - b.lastts
}

func (b *Bucket) spend(tokens int64) {
	b.lastts += tokens
}
