package throttle

import (
	"context"
	"time"
)

type (
	Backoff struct {
		Bucket

		price int64

		MinPrice int64
		MaxPrice int64

		Increase int64
		Factor   Fract
		Jitter   Fract

		Decrease   int64
		CoolFactor Fract
		CoolJitter Fract
	}

	Fract struct {
		Num, Den uint16
	}
)

func NewBackoff(ts, price, limit int64) *Backoff {
	b := &Backoff{
		MinPrice: price,
		MaxPrice: limit,

		Increase: price / 10,
		Factor:   Fract{Num: 15, Den: 10},
		Jitter:   Fract{Num: 1, Den: 10},

		Decrease:   price / 6,
		CoolFactor: Fract{Num: 4, Den: 10},
		CoolJitter: Fract{Num: 1, Den: 10},
	}

	b.Reset(ts, price, limit)

	return b
}

func (b *Backoff) Reset(ts, price, limit int64) {
	b.Bucket.Reset(ts, limit)
	b.price = price
}

func (b *Backoff) Price() int64 {
	return b.price
}

func (b *Backoff) BackOffT(now time.Time) { b.BackOff(now.UnixNano()) }

func (b *Backoff) BackOff(ts int64) {
	b.advance(ts)

	p := b.price

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

	b.price = p
}

func (b *Backoff) CoolOffT(now time.Time) { b.CoolOff(now.UnixNano()) }

func (b *Backoff) CoolOff(ts int64) {
	b.advance(ts)

	p := b.price

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

	b.price = p
}

func (b *Backoff) Recover() {
	b.price = b.MinPrice
}

func (b *Backoff) ValueT(now time.Time) int {
	return b.Value(now.UnixNano())
}

func (b *Backoff) Value(ts int64) int {
	return int(b.Bucket.Value(ts) / b.price)
}

func (b *Backoff) HaveT(now time.Time, n int) bool {
	return b.Have(now.UnixNano(), n)
}

func (b *Backoff) Have(ts int64, n int) bool {
	return b.Bucket.Have(ts, b.price*int64(n))
}

func (b *Backoff) TakeT(now time.Time, n int) bool {
	return b.Take(now.UnixNano(), n)
}

func (b *Backoff) Take(ts int64, n int) bool {
	return b.Bucket.Take(ts, b.price*int64(n))
}

func (b *Backoff) BorrowT(now time.Time, n int) time.Duration {
	return b.Borrow(now.UnixNano(), n)
}

func (b *Backoff) Borrow(ts int64, n int) time.Duration {
	return b.Bucket.Borrow(ts, b.price*int64(n))
}

func (b *Backoff) WaitT(ctx context.Context, now time.Time, n int) error {
	return b.Wait(ctx, now.UnixNano(), n)
}

func (b *Backoff) Wait(ctx context.Context, ts int64, n int) error {
	return b.Bucket.Wait(ctx, ts, b.price*int64(n))
}

func (b *Backoff) SetValueT(now time.Time, n int) {
	b.SetValue(now.UnixNano(), n)
}

func (b *Backoff) SetValue(ts int64, n int) {
	b.Bucket.SetValue(ts, b.price*int64(n))
}

func (b *Backoff) ReturnT(now time.Time, n int) {
	b.Return(now.UnixNano(), n)
}

func (b *Backoff) Return(ts int64, n int) {
	b.Bucket.Return(ts, b.price*int64(n))
}

func (f Fract) Mul(p int64) int64 {
	return p * int64(f.Num) / int64(f.Den)
}

func fastrand(ts, j int64) int64 {
	r := uint64(ts) * 0x1e35a7bd >> 32
	return int64(r)%(2*j) - j
}
