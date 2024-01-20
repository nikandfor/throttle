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

		AutoBackOff Fract // Price multiplier
		AutoCoolOff Fract
	}

	Fract struct {
		Num, Den uint16
	}

	dur = time.Duration
)

func NewBackoffT(t time.Time, minPrice, maxPrice time.Duration) *Backoff {
	return NewBackoff(t.UnixNano(), minPrice, maxPrice)
}

func NewBackoff(ts int64, minPrice, maxPrice time.Duration) *Backoff {
	b := &Backoff{
		MinPrice: minPrice,
		MaxPrice: maxPrice,

		Increase: minPrice,
		Factor:   Fract{Num: 17, Den: 10},
		Jitter:   Fract{Num: 1, Den: 10},

		Decrease:   minPrice / 6,
		CoolFactor: Fract{Num: 4, Den: 10},
		CoolJitter: Fract{Num: 1, Den: 10},

		AutoBackOff: Fract{Num: 2, Den: 1},
		AutoCoolOff: Fract{Num: 8, Den: 1},
	}

	b.Reset(ts, 0, 0)

	return b
}

func (b *Backoff) Reset(ts int64, price, limit time.Duration) {
	b.Bucket.Reset(ts, limit)
	b.Price = price
}

func (b *Backoff) AutoWaitT(ctx context.Context, t time.Time) error {
	return b.AutoWait(ctx, t.UnixNano())
}

func (b *Backoff) AutoWait(ctx context.Context, ts int64) error {
	b.AutoOff(ts)

	return b.Wait(ctx, ts)
}

func (b *Backoff) AutoOffT(t time.Time) {
	b.AutoOff(t.UnixNano())
}

func (b *Backoff) AutoOff(ts int64) {
	tk := b.Bucket.tokens(ts)

	if tk <= b.AutoBackOff.Mul(b.Price) {
		b.BackOff(ts)
		return
	}

	for tk > b.AutoCoolOff.Mul(b.Price) {
		tk -= b.Price
		b.CoolOff(ts)
	}
}

func (b *Backoff) AutoErrWaitT(ctx context.Context, t time.Time, err error) error {
	return b.AutoErrWait(ctx, t.UnixNano(), err)
}

func (b *Backoff) AutoErrWait(ctx context.Context, ts int64, err error) error {
	b.AutoErr(ts, err)

	return b.Wait(ctx, ts)
}

func (b *Backoff) AutoErrT(t time.Time, err error) {
	b.AutoErr(t.UnixNano(), err)
}

func (b *Backoff) AutoErr(ts int64, err error) {
	if err != nil {
		b.BackOff(ts)
	} else {
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
	b.Limit = p
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
	b.Limit = p
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
	if f == (Fract{}) {
		return p
	}

	return p * time.Duration(f.Num) / time.Duration(f.Den)
}

func fastrand(ts int64, j time.Duration) time.Duration {
	r := uint64(ts) * 0x1e35a7bd >> 32
	return time.Duration(r)%(2*j) - j
}
