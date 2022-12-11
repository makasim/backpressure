package backpressure_test

import (
	"log"
	"testing"
	"time"

	"github.com/makasim/backpressure"
)

func BenchmarkAIMD_OK(b *testing.B) {
	bp, err := backpressure.NewAIMD(backpressure.AIMDConfig{
		DecideInterval:   time.Microsecond * 100,
		ThresholdPercent: 0.01,
		IncreasePercent:  1.01,
		DecreasePercent:  0.8,
	})
	if err != nil {
		log.Fatalln(err)
	}

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		t, allowed := bp.Acquire()
		if !allowed {
			log.Fatalln("not allowed")
		}
		bp.Release(t)
	}
}

func BenchmarkAIMD_Congested(b *testing.B) {
	bp, err := backpressure.NewAIMD(backpressure.AIMDConfig{
		DecideInterval:   time.Microsecond * 100,
		ThresholdPercent: 0.01,
		IncreasePercent:  1.01,
		DecreasePercent:  0.8,
	})
	if err != nil {
		log.Fatalln(err)
	}

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		t, _ := bp.Acquire()

		if n%3 == 0 {
			t.Congested = true
		}

		bp.Release(t)
	}
}
