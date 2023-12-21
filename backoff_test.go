package throttle

import "testing"

func TestBackoff(tb *testing.T) {
	ts := int64(10000)
	b := NewBackoff(ts, 1000, 100_000)

	for i := 0; i < 10; i++ {
		p := b.Price()
		tb.Logf("backoff step %3d price %5v", i, p)

		ts += 10000
		b.BackOff(ts)
	}

	for i := 0; i < 10; i++ {
		p := b.Price()
		tb.Logf("cooloff step %3d price %5v", i, p)

		ts += 10000
		b.CoolOff(ts)
	}
}
