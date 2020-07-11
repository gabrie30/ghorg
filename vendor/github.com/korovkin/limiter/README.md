# go lang goroutine concurrency limiter

## builds

[![Build Status](https://travis-ci.org/korovkin/limiter.svg)](https://travis-ci.org/korovkin/limiter)

## example

limit the number of concurrent go routines to 10:

```
  limit := NewConcurrencyLimiter(10)
  for i := 0; i < 1000; i++ {
  	limit.Execute(func() {
  		// do some work
  	})
  }
  limit.Wait()
```

