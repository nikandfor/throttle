package throttle

import (
	"context"
	"time"
)

type (
	Backoff struct {
		Bucket

		Price time.Duration

		MinPrice time.Duration
		MaxPrice time.Duration

		Increase time.Duration
		Factor   Fract
		Jitter   Fract

		Decrease   time.Duration
		CoolFactor Fract
		CoolJitter Fract
	}

	Fract struct {
		Num, Den uint16
	}

	dur = time.Duration
)

func NewBackoffNow(price, limit time.Duration) *Backoff {
	return NewBackoff(time.Now().UnixNano(), price, limit)
}

func NewBackoff(ts int64, price, limit time.Duration) *Backoff {
	b := &Backoff{
		MinPrice: price,
		MaxPrice: limit,

		Increase: price,
		Factor:   Fract{Num: 17, Den: 10},
		Jitter:   Fract{Num: 1, Den: 10},

		Decrease:   price / 6,
		CoolFactor: Fract{Num: 4, Den: 10},
		CoolJitter: Fract{Num: 1, Den: 10},
	}

	b.Reset(ts, price, limit)

	return b
}

func (b *Backoff) Reset(ts int64, price, limit time.Duration) {
	b.Bucket.Reset(ts, limit)
	b.Price = price
}

func (b *Backoff) AutoOffNow() {
	b.AutoOffT(time.Now())
}

func (b *Backoff) AutoOffT(t time.Time) {
	b.AutoOff(t.UnixNano())
}

func (b *Backoff) AutoOff(ts int64) {
	tk := b.Bucket.tokens(ts)

	switch {
	case tk <= 2*b.Price:
		b.BackOff(ts)
	case tk <= b.MaxPrice:
		// keep the same
	default:
		b.CoolOff(ts)
	}
}

func (b *Backoff) BackOffT(now time.Time) { b.BackOff(now.UnixNano()) }

func (b *Backoff) BackOff(ts int64) {
	b.advance(ts)

	p := b.Price

	if p < b.MinPrice {
		p = b.MinPrice
	}

	p += b.Increase
	p = b.Factor.Mul(p)

	j := b.Jitter.Mul(p)
	p += fastrand(ts, j)

	if p > b.MaxPrice {
		p = b.MaxPrice
	}

	b.Price = p
}

func (b *Backoff) CoolOffT(now time.Time) { b.CoolOff(now.UnixNano()) }

func (b *Backoff) CoolOff(ts int64) {
	b.advance(ts)

	p := b.Price

	//	println("ini", p)

	p -= b.Decrease
	//	println("dec", p)
	p = b.CoolFactor.Mul(p)
	//	println("mul", p)

	j := b.CoolJitter.Mul(p)
	p += fastrand(ts, j)
	//	println("jtr", p)

	if p < b.MinPrice {
		p = b.MinPrice
	}

	//	println("res", p)

	b.Price = p
}

func (b *Backoff) Recover() {
	b.Price = b.MinPrice
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
	return b.Bucket.Have(ts, b.Price)
}

func (b *Backoff) TakeT(now time.Time) bool {
	return b.Take(now.UnixNano())
}

func (b *Backoff) Take(ts int64) bool {
	return b.Bucket.Take(ts, b.Price)
}

func (b *Backoff) BorrowT(now time.Time) time.Duration {
	return b.Borrow(now.UnixNano())
}

func (b *Backoff) Borrow(ts int64) time.Duration {
	return b.Bucket.Borrow(ts, b.Price)
}

func (b *Backoff) WaitT(ctx context.Context, now time.Time) error {
	return b.Wait(ctx, now.UnixNano())
}

func (b *Backoff) Wait(ctx context.Context, ts int64) error {
	return b.Bucket.Wait(ctx, ts, b.Price)
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
	return p * time.Duration(f.Num) / time.Duration(f.Den)
}

func fastrand(ts int64, j time.Duration) time.Duration {
	r := uint64(ts) * 0x1e35a7bd >> 32
	return time.Duration(r)%(2*j) - j
}
