package throttle

import "testing"

func BenchmarkTake(tb *testing.B) {
	t := New(100, 1, 1000)

	for i := 0; i < tb.N; i++ {
		t.Take(int64(100+i), 1)
	}
}

func BenchmarkAtomicTake(tb *testing.B) {
	t := NewAtomic(100, 1, 1000)

	for i := 0; i < tb.N; i++ {
		t.Take(int64(100+i), 1)
	}
}
