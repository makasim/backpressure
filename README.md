# Backpressure

```go
package main

import (
	"log"
	"net/http"
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

	// later, while sending request to origin server

	t, allowed := bp.Acquire()
	if !allowed {
		log.Println("backpressure activated")
		return
	}

	req, _ := http.NewRequest("GET", "https://example.com", http.NoBody)

	resp, err := c.Do(req)
	t.Congested = backpressure.IsResponseCongested(resp, err)
	bp.Release(t)
}

```
