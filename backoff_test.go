package throttle

import (
	"testing"
	"time"
)

func TestBackoff(tb *testing.T) {
	ts := int64(0)
	b := NewBackoff(ts, time.Second/10, time.Minute)

	pp := func(ts int64, n string, i int, d time.Duration) {
		tb.Logf("ts %8v %s step %3d  price %8v  dur %8v  val %8v", round(dur(ts)), n, i, round(b.Price), round(d), round(b.Value(ts)))
	}

	tb.Logf("backing off")

	for i := 0; i < 10; i++ {
		// running the task
		ts += int64(time.Second)

		b.BackOff(ts)
		d := b.Borrow(ts)

		pp(ts, "backoff", i, d)

		ts += int64(d)
	}

	tb.Logf("cooling off")

	for i := 0; i < 10; i++ {
		// running the task
		ts += int64(time.Second)

		b.CoolOff(ts)
		d := b.Borrow(ts)

		pp(ts, "cooloff", i, d)

		ts += int64(d)
	}
}

func round(d time.Duration) time.Duration {
	return d.Round(10 * time.Millisecond)
}
