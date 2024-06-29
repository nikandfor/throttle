package throttle

import (
	"context"
	"time"
)

type (
	Throttle struct {
		Bucket
		Price time.Duration
	}
)

func New(ts int64, price, limit time.Duration) *Throttle {
	b := &Throttle{}
	b.Reset(ts, price, limit)

	return b
}

func (b *Throttle) Reset(ts int64, price, limit time.Duration) {
	b.Bucket.Reset(ts, limit)
	b.Price = price
}

func (b *Throttle) ValueT(now time.Time) int {
	return b.Value(now.UnixNano())
}

func (b *Throttle) Value(ts int64) int {
	return int(b.Bucket.Value(ts) / b.Price)
}

func (b *Throttle) HaveT(now time.Time, n int) bool {
	return b.Have(now.UnixNano(), n)
}

func (b *Throttle) Have(ts int64, n int) bool {
	return b.Bucket.Have(ts, b.Price*dur(n))
}

func (b *Throttle) TakeT(now time.Time, n int) bool {
	return b.Take(now.UnixNano(), n)
}

func (b *Throttle) Take(ts int64, n int) bool {
	return b.Bucket.Take(ts, b.Price*dur(n))
}

func (b *Throttle) BorrowT(now time.Time, n int) time.Duration {
	return b.Borrow(now.UnixNano(), n)
}

func (b *Throttle) Borrow(ts int64, n int) time.Duration {
	return b.Bucket.Borrow(ts, b.Price*dur(n))
}

func (b *Throttle) WaitT(ctx context.Context, now time.Time, n int) error {
	return b.Wait(ctx, now.UnixNano(), n)
}

func (b *Throttle) Wait(ctx context.Context, ts int64, n int) error {
	return b.Bucket.Wait(ctx, ts, b.Price*dur(n))
}

func (b *Throttle) SetValueT(now time.Time, n int) {
	b.SetValue(now.UnixNano(), n)
}

func (b *Throttle) SetValue(ts int64, n int) {
	b.Bucket.SetValue(ts, b.Price*dur(n))
}

func (b *Throttle) ReturnT(now time.Time, n int) {
	b.Return(now.UnixNano(), n)
}

func (b *Throttle) Return(ts int64, n int) {
	b.Bucket.Return(ts, b.Price*dur(n))
}

func Price(tokens int64, per time.Duration) time.Duration {
	return per / time.Duration(tokens)
}
