package throttle

import (
	"context"
	"sync/atomic"
	"time"
)

type (
	AtomicBucket struct {
		lastts int64
		Limit  time.Duration
	}
)

func NewAtomicBucket(ts int64, limit time.Duration) *AtomicBucket {
	return &AtomicBucket{lastts: ts, Limit: limit}
}

func (b *AtomicBucket) Reset(ts int64, limit time.Duration) {
	atomic.StoreInt64(&b.lastts, ts)
	b.Limit = limit
}

func (b *AtomicBucket) Value(ts int64) time.Duration {
	tokens := ts - atomic.LoadInt64(&b.lastts)

	if tokens > int64(b.Limit) {
		return b.Limit
	}

	return dur(tokens)
}

func (b *AtomicBucket) Have(ts int64, cost time.Duration) bool {
	return b.Value(ts) >= cost
}

func (b *AtomicBucket) Take(ts int64, cost time.Duration) bool {
	for {
		lastts := atomic.LoadInt64(&b.lastts)
		tokens := ts - lastts
		if tokens > int64(b.Limit) {
			tokens = int64(b.Limit)
		}

		tokens -= int64(cost)
		if tokens < 0 {
			return false
		}

		if atomic.CompareAndSwapInt64(&b.lastts, lastts, ts-tokens) {
			return true
		}
	}
}

func (b *AtomicBucket) Borrow(ts int64, cost time.Duration) time.Duration {
	for {
		lastts := atomic.LoadInt64(&b.lastts)
		tokens := ts - lastts
		if tokens > int64(b.Limit) {
			tokens = int64(b.Limit)
		}

		tokens -= int64(cost)

		if atomic.CompareAndSwapInt64(&b.lastts, lastts, ts-tokens) {
			if tokens > 0 {
				return 0
			}

			return dur(-tokens)
		}
	}
}

func (b *AtomicBucket) Wait(ctx context.Context, ts int64, cost time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	d := b.Borrow(ts, cost)
	if d == 0 {
		return nil
	}

	return Wait(ctx, d)
}

func (b *AtomicBucket) SetValue(ts int64, cost time.Duration) {
	atomic.StoreInt64(&b.lastts, ts-int64(cost))
}

func (b *AtomicBucket) Return(ts int64, cost time.Duration) {
	atomic.AddInt64(&b.lastts, -int64(cost))
}
