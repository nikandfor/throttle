[![Documentation](https://pkg.go.dev/badge/nikand.dev/go/throttle)](https://pkg.go.dev/nikand.dev/go/throttle?tab=doc)
[![Go workflow](https://github.com/nikandfor/throttle/actions/workflows/go.yml/badge.svg)](https://github.com/nikandfor/throttle/actions/workflows/go.yml)
[![CircleCI](https://circleci.com/gh/nikandfor/throttle.svg?style=svg)](https://circleci.com/gh/nikandfor/throttle)
[![codecov](https://codecov.io/gh/nikandfor/throttle/branch/master/graph/badge.svg)](https://codecov.io/gh/nikandfor/throttle)
[![Go Report Card](https://goreportcard.com/badge/nikand.dev/go/throttle)](https://goreportcard.com/report/nikand.dev/go/throttle)
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/nikandfor/throttle?sort=semver)

# throttle

`throttle` is an efficient rate limiter built on integer arithmetics.
The implementation is inspired by one from the Linux kernel.

## Usage

Create limiter
```go
// limit to 100 requests per minute, at most 10 in a burst, start from 0 tokens available

ts := time.Now().UnixNano()
price := throttle.Price(100, time.Minute)
limit := 10 * price

t := throttle.New(ts, price, limit)

// limit bandwidth to 100 KB/s, allow 1 MB as a burst, start from full burst available

ts := 0
price := throttle.Price(100 * 1024, time.Second)
limit := 1 * 1024 * 1024 / price

t := throttle.New(ts, price, limit)

t.SetValueT(time.Now().UnixNano(), 512 * 1024) // reset available value to 512 KB
```

Take or drop
```go
func (c *Conn) Write(p []byte) (int, error) {
	if !t.TakeT(time.Now(), len(p)) {
		return 0, ErrLimited
	}

	return c.Conn.Write(p)
}
```

Borrow and wait
```go
func (c *Conn) Write(p []byte) (int, error) {
	delay := l.BorrowT(time.Now(), len(p))

	if delay != 0 {
		time.Sleep(delay)
	}

	return c.Conn.Write(p)
}
```

Write as much as we can
```go
func (c *Conn) Write(p []byte) (int, error) {
	now := time.Now()

	n := l.Value(now)
	if n > len(p) {
		n = len(p)
	}

	_ = l.TakeT(now, n) // must be true

	n, err := c.Conn.Write(p[:n])
	if err != nil {
		return n, err
	}
	if n != len(p) {
		err = ErrLimited
	}

	return n, err
}
```

## How it works

There is a bucket that collects time passing.
It can be real nanoseconds, can be seconds, can be monotonic clock ticks, whatever you pass as ts.
XxxxxT methods use nanoseconds of wall time.
Each time we pass ts to any method, number of available time in the bucket is updated to the provided point in time.

Bucket has a `limit`, no more than that is kept.

There is a `price`, which is how much time each `token` costs.
It's allowed to take `n` `tokens` if we have `n * price` time in the bucket.
