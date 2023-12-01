[![Documentation](https://pkg.go.dev/badge/nikand.dev/go/throttle)](https://pkg.go.dev/nikand.dev/go/throttle?tab=doc)
[![Go workflow](https://github.com/nikandfor/throttle/actions/workflows/go.yml/badge.svg)](https://github.com/nikandfor/throttle/actions/workflows/go.yml)
[![CircleCI](https://circleci.com/gh/nikandfor/throttle.svg?style=svg)](https://circleci.com/gh/nikandfor/throttle)
[![codecov](https://codecov.io/gh/nikandfor/throttle/branch/master/graph/badge.svg)](https://codecov.io/gh/nikandfor/throttle)
[![Go Report Card](https://goreportcard.com/badge/nikand.dev/go/throttle)](https://goreportcard.com/report/nikand.dev/go/throttle)
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/nikandfor/throttle?sort=semver)

# throttle

`throttle` is an efficient rate limiter built on integer arithmetics.

## Usage

Create limiter
```go
t := throttle.NewRateWindow(
		2000,                         // initial tokens
		1000 / time.Second.Seconds(), // 1000 tokens per second
		2 * time.Second,              // window size
	)

// smooth 1KB per second with at most 128 bytes at a time
t = throttle.NewRateLimit(
		128,
		1000 / time.Second.Seconds(),
		128,
	)

// 3 MB per minute allowing to spend it all at once
// but start with only 1 MB available
t = throttle.NewRateLimit(
		1_000_000,
		3_000_000 / time.Minute.Seconds(),
		3_000_000,
	)
```

Take or drop
```go
func (c *Conn) Write(p []byte) (int, error) {
	if !t.Take(time.Now(), len(p)) {
		return 0, ErrLimited
	}

	return c.Conn.Write(p)
}
```

Borrow and wait
```go
func (c *Conn) Write(p []byte) (int, error) {
	delay := l.Borrow(time.Now(), len(p))

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

	val := l.Value(now)

	n := int(val)
	if n > len(p) {
		n = len(p)
	}

	_ = l.Take(now, float64(n)) // must be true

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
