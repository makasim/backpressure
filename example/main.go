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
		DecidePeriod:    time.Second,
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
