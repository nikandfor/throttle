package throttle

import (
	"context"
	"time"
)

type (
	Bucket struct {
		lastts int64
		Limit  time.Duration
	}
)

func NewBucket(ts int64, limit time.Duration) *Bucket {
	return &Bucket{lastts: ts, Limit: limit}
}

func (b *Bucket) Reset(ts int64, limit time.Duration) {
	b.lastts = ts
	b.Limit = limit
}

func (b *Bucket) Value(ts int64) time.Duration {
	b.advance(ts)

	return b.tokens(ts)
}

func (b *Bucket) Have(ts int64, cost time.Duration) bool {
	b.advance(ts)

	return b.tokens(ts) >= cost
}

func (b *Bucket) Take(ts int64, cost time.Duration) bool {
	if !b.Have(ts, cost) {
		return false
	}

	b.spend(cost)

	return true
}

func (b *Bucket) Borrow(ts int64, cost time.Duration) time.Duration {
	b.advance(ts)
	b.spend(cost)

	d := b.tokens(ts)
	if d > 0 {
		return 0
	}

	return -d
}

func (b *Bucket) Wait(ctx context.Context, ts int64, cost time.Duration) error {
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

func (b *Bucket) SetValue(ts int64, cost time.Duration) {
	b.lastts = ts - int64(cost)
}

func (b *Bucket) Return(ts int64, cost time.Duration) {
	b.spend(-cost)
}

func (b *Bucket) advance(ts int64) {
	tokens := ts - b.lastts

	if tokens > int64(b.Limit) {
		tokens = int64(b.Limit)
	}

	b.lastts = ts - tokens
}

func (b *Bucket) tokens(ts int64) time.Duration {
	return dur(ts - b.lastts)
}

func (b *Bucket) spend(tokens time.Duration) {
	b.lastts += int64(tokens)
}
