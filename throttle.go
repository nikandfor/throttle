package throttle

import (
	"context"
	"time"
)

type (
	Throttle struct {
		lastts int64
		price  int64
		limit  int64
	}
)

func NewRateWindow(value, rate float64, window time.Duration) *Throttle {
	ts := time.Now().UnixNano()
	price := int64(1e9 / rate)
	limit := window.Nanoseconds()
	tokens := ts - int64(1e9/rate*value)

	return &Throttle{
		lastts: ts - tokens,
		price:  price,
		limit:  limit,
	}
}

func New(ts, price, limit int64) *Throttle {
	return &Throttle{
		lastts: ts,
		price:  price,
		limit:  limit,
	}
}

func (t *Throttle) Reset(ts, price, limit int64) {
	t.lastts = ts
	t.price = price
	t.limit = limit
}

func (t *Throttle) Have(now time.Time, n int) bool {
	return t.HaveTs(now.UnixNano(), n)
}

func (t *Throttle) HaveTs(ts int64, n int) bool {
	return t.HaveTsCost(ts, t.price*int64(n))
}

func (t *Throttle) HaveTsCost(ts, cost int64) bool {
	t.advance(ts)

	return t.tokens(ts) >= cost
}

func (t *Throttle) Take(now time.Time, n int) bool {
	return t.TakeTs(now.UnixNano(), n)
}

func (t *Throttle) TakeTs(ts int64, n int) bool {
	return t.TakeTsCost(ts, t.price*int64(n))
}

func (t *Throttle) TakeTsCost(ts, cost int64) bool {
	if !t.HaveTsCost(ts, cost) {
		return false
	}

	t.spend(cost)

	return true
}

func (t *Throttle) Borrow(now time.Time, n int) time.Duration {
	return t.BorrowTs(now.UnixNano(), n)
}

func (t *Throttle) BorrowTs(ts int64, n int) time.Duration {
	return t.BorrowTsCost(ts, t.price*int64(n))
}

func (t *Throttle) BorrowTsCost(ts, cost int64) time.Duration {
	t.advance(ts)
	t.spend(cost)

	d := time.Duration(t.lastts - ts)
	if d < 0 {
		d = 0
	}

	return d
}

func (t *Throttle) Wait(ctx context.Context, now time.Time, n int) error {
	return t.WaitTs(ctx, now.UnixNano(), n)
}

func (t *Throttle) WaitTs(ctx context.Context, ts int64, n int) error {
	return t.WaitTsCost(ctx, ts, t.price*int64(n))
}

func (t *Throttle) WaitTsCost(ctx context.Context, ts, cost int64) error {
	d := t.BorrowTsCost(ts, cost)
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

func (t *Throttle) advance(ts int64) {
	tokens := ts - t.lastts

	if tokens > t.limit {
		tokens = t.limit
	}

	t.lastts = ts - tokens
}

func (t *Throttle) tokens(ts int64) int64 {
	return ts - t.lastts
}

func (t *Throttle) spend(tokens int64) {
	t.lastts += tokens
}
