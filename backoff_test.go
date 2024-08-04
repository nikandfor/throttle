package throttle

import (
	"testing"
	"time"
)

func TestBackoff(tb *testing.T) {
	ts := int64(0)
	b := NewBackoff(ts, time.Second/10, 10*time.Second, time.Second)

	start := time.Hour
	ts += start.Nanoseconds()

	tb.Logf("%+v", b)

	for j := 0; j < 2; j++ {
		for i := 0; i < 5; i++ {
			d := b.Delay(ts)

			tb.Logf("backoff i %3d  ts %10v  %v", i, time.Duration(ts)-start, d)

			ts += 1000 * time.Millisecond.Nanoseconds()

			b.Backoff(ts)
		}

		for i := 0; i < 5; i++ {
			d := b.Delay(ts)

			tb.Logf("cooloff i %3d  ts %10v  %v", i, time.Duration(ts)-start, d)

			ts += 5000 * time.Millisecond.Nanoseconds()
		}
	}
}

func round(d time.Duration) time.Duration {
	return d.Round(10 * time.Millisecond)
}
