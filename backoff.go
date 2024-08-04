package throttle

import (
	"context"
	"time"
)

type (
	Backoff struct {
		Bucket

		backts int64
		delay  time.Duration

		MinDelay time.Duration
		MaxDelay time.Duration

		Increase time.Duration
		Factor   Fract
		Jitter   Fract

		Recovery time.Duration
	}

	Fract struct {
		Num, Den uint16
	}

	dur = time.Duration
)

func NewBackoffT(t time.Time, minDelay, maxDelay, limit time.Duration) *Backoff {
	return NewBackoff(t.UnixNano(), minDelay, maxDelay, limit)
}

func NewBackoff(ts int64, minDelay, maxDelay, limit time.Duration) *Backoff {
	b := &Backoff{
		MinDelay: minDelay,
		MaxDelay: maxDelay,

		Increase: minDelay,
		Factor:   Fract{Num: 17, Den: 10},
		Jitter:   Fract{Num: 1, Den: 10},

		Recovery: time.Minute,
	}

	b.Reset(ts, 0, limit)

	return b
}

func (b *Backoff) Reset(ts int64, delay, limit time.Duration) {
	b.Bucket.Reset(ts, limit)
	b.delay = delay
}

func (b *Backoff) BackoffT(now time.Time) { b.Backoff(now.UnixNano()) }

func (b *Backoff) Backoff(ts int64) {
	b.advance(ts)

	d := b.Delay(ts)

	if d < b.MinDelay {
		d = b.MinDelay
	}

	d += b.Increase
	d = b.Factor.Mul(d)

	j := b.Jitter.Mul(d)
	d += fastrand(ts, j)

	if d > b.MaxDelay {
		d = b.MaxDelay
	}

	b.delay = d
	//	b.Limit = d

	b.backts = ts
}

func (b *Backoff) DelayT(t time.Time) time.Duration {
	return b.Delay(t.UnixNano())
}

func (b *Backoff) Delay(ts int64) time.Duration {
	const eps = time.Millisecond

	//	defer func() { println("delay", ts, delay) }()
	passed := time.Duration(ts - b.backts)
	if passed > b.Recovery {
		return b.MinDelay
	}
	if passed < eps {
		return b.delay
	}

	lt := b.backts
	rt := b.backts + b.Recovery.Nanoseconds()

	l := b.delay
	r := b.MinDelay

	var delay time.Duration

	for rt-lt > eps.Nanoseconds() {
		mt := (lt + rt) >> 1
		delay = (l + r) >> 1

		if ts <= mt {
			rt = mt
			r = delay
		} else {
			lt = mt
			l = delay
		}
	}

	return delay
}

func (b *Backoff) Recover() {
	b.delay = b.MinDelay
}

func (b *Backoff) ValueT(now time.Time) time.Duration {
	return b.Value(now.UnixNano())
}

func (b *Backoff) Value(ts int64) time.Duration {
	return b.Bucket.Value(ts)
}

func (b *Backoff) HaveT(now time.Time) bool {
	return b.Have(now.UnixNano())
}

func (b *Backoff) Have(ts int64) bool {
	return b.Bucket.Have(ts, b.Delay(ts))
}

func (b *Backoff) TakeT(now time.Time) bool {
	return b.Take(now.UnixNano())
}

func (b *Backoff) Take(ts int64) bool {
	return b.Bucket.Take(ts, b.Delay(ts))
}

func (b *Backoff) BorrowT(now time.Time) time.Duration {
	return b.Borrow(now.UnixNano())
}

func (b *Backoff) Borrow(ts int64) time.Duration {
	return b.Bucket.Borrow(ts, b.Delay(ts))
}

func (b *Backoff) WaitT(ctx context.Context, now time.Time) error {
	return b.Wait(ctx, now.UnixNano())
}

func (b *Backoff) Wait(ctx context.Context, ts int64) error {
	return b.Bucket.Wait(ctx, ts, b.Delay(ts))
}

func (b *Backoff) SetValueT(now time.Time, cost time.Duration) {
	b.SetValue(now.UnixNano(), cost)
}

func (b *Backoff) SetValue(ts int64, cost time.Duration) {
	b.Bucket.SetValue(ts, cost)
}

func (b *Backoff) ReturnT(now time.Time, cost time.Duration) {
	b.Return(now.UnixNano(), cost)
}

func (b *Backoff) Return(ts int64, cost time.Duration) {
	b.Bucket.Return(ts, cost)
}

func (f Fract) Mul(p time.Duration) time.Duration {
	if f == (Fract{}) {
		return p
	}

	return p * time.Duration(f.Num) / time.Duration(f.Den)
}

func fastrand(ts int64, j time.Duration) time.Duration {
	r := uint64(ts) * 0x1e35a7bd >> 32
	return time.Duration(r)%(2*j) - j
}
