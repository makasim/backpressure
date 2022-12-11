package backpressure_test

import (
	"testing"
	"time"

	"github.com/makasim/backpressure"
	"github.com/stretchr/testify/require"
)

func TestNoCongestion(t *testing.T) {
	bp, err := backpressure.NewAIMD(backpressure.AIMDConfig{
		DecideInterval:   time.Millisecond * 5,
		ThresholdPercent: 0.01,
		IncreasePercent:  1.01,
		DecreasePercent:  0.8,
	})
	require.NoError(t, err)

	var totalAllowed int
	var totalForbidden int
	for i := 0; i < 1000; i++ {
		t, allowed := bp.Acquire()
		if !allowed {
			totalForbidden++
			continue
		}

		totalAllowed++
		bp.Release(t)
	}

	require.Equal(t, 1000, totalAllowed)
	require.Equal(t, 0, totalForbidden)
}
