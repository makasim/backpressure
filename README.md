# Backpressure

```go
package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/makasim/backpressure"
)

func main() {
	bp, _ := backpressure.NewAIMD(backpressure.AIMDConfig{
		DecideInterval:  time.Second,
		IncreasePercent: 0.02,
		DecreasePercent: 0.2,
	})

	c := &http.Client{}

	wg := sync.WaitGroup{}

	// later, while sending request to origin server

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			t, allowed := bp.Acquire()
			if !allowed {
				log.Println("backpressure activated")
				return
			}

			req, _ := http.NewRequest("GET", "https://example.com", http.NoBody)

			resp, err := c.Do(req)
			t.Congested = backpressure.IsResponseCongested(resp, err)
			bp.Release(t)
		}()
	}

	wg.Done()
}
```


## Benchmark

```shell
$go test -run=XXX -v -bench=.
goos: darwin
goarch: arm64
pkg: github.com/makasim/backpressure
BenchmarkAIMD_OK
BenchmarkAIMD_OK-10           	15986818	        66.95 ns/op	       0 B/op	       0 allocs/op
BenchmarkAIMD_Congested
BenchmarkAIMD_Congested-10    	18400120	        65.22 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/makasim/backpressure	2.758s
```

## References 

* https://www.youtube.com/watch?v=m64SWl9bfvk
* https://en.wikipedia.org/wiki/TCP_congestion_control
* https://www.youtube.com/watch?v=UdT0xVacEUg
* https://github.com/marselester/capacity
