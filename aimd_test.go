package backpressure

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNew(main *testing.T) {
	main.Run("DecideIntervalZero", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval: 0,
		})
		require.EqualError(t, err, `DecideInterval: required`)
		require.Nil(t, bp)
	})

	main.Run("DecideIntervalNegative", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval: -1,
		})
		require.EqualError(t, err, `DecideInterval: negative`)
		require.Nil(t, bp)
	})

	main.Run("DecreasePercentZero", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:  time.Second,
			DecreasePercent: 0,
		})
		require.EqualError(t, err, `DecreasePercent: required`)
		require.Nil(t, bp)
	})

	main.Run("DecreasePercentLessThanZero", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:  time.Second,
			DecreasePercent: -0.01,
		})
		require.EqualError(t, err, `DecreasePercent: less than zero`)
		require.Nil(t, bp)
	})

	main.Run("DecreasePercentMoreThanOne", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:  time.Second,
			DecreasePercent: 1.01,
		})
		require.EqualError(t, err, `DecreasePercent: more than one`)
		require.Nil(t, bp)
	})

	main.Run("IncreasePercentZero", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:  time.Second,
			DecreasePercent: 0.02,
			IncreasePercent: 0,
		})
		require.EqualError(t, err, `IncreasePercent: required`)
		require.Nil(t, bp)
	})

	main.Run("IncreasePercentLessThanZero", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:  time.Second,
			DecreasePercent: 0.02,
			IncreasePercent: -0.01,
		})
		require.EqualError(t, err, `IncreasePercent: less than zero`)
		require.Nil(t, bp)
	})

	main.Run("IncreasePercentMoreThanOne", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:  time.Second,
			DecreasePercent: 0.02,
			IncreasePercent: 1.01,
		})
		require.EqualError(t, err, `IncreasePercent: more than one`)
		require.Nil(t, bp)
	})

	main.Run("IncreasePercentLessThanDecreasePercent", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:  time.Second,
			DecreasePercent: 0.02,
			IncreasePercent: 0.02,
		})
		require.EqualError(t, err, `IncreasePercent: must be less than DecreasePercent`)
		require.Nil(t, bp)
	})

	main.Run("ThresholdPercentLessThanZero", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:   time.Second,
			DecreasePercent:  0.04,
			IncreasePercent:  0.02,
			ThresholdPercent: -0.01,
		})
		require.EqualError(t, err, `ThresholdPercent: less than zero`)
		require.Nil(t, bp)
	})

	main.Run("ThresholdPercentMoreThanOne", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:   time.Second,
			DecreasePercent:  0.04,
			IncreasePercent:  0.02,
			ThresholdPercent: 1.01,
		})
		require.EqualError(t, err, `ThresholdPercent: more than one`)
		require.Nil(t, bp)
	})

	main.Run("DecreaseLatencyPercentileLessThanZero", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:            time.Second,
			DecreasePercent:           0.04,
			IncreasePercent:           0.02,
			ThresholdPercent:          0.01,
			DecreaseLatencyPercentile: -0.01,
		})
		require.EqualError(t, err, `DecreaseLatencyPercentile: less than zero`)
		require.Nil(t, bp)
	})

	main.Run("DecreaseLatencyPercentileMoreThanOne", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:            time.Second,
			DecreasePercent:           0.04,
			IncreasePercent:           0.02,
			ThresholdPercent:          0.01,
			DecreaseLatencyPercentile: 1.01,
		})
		require.EqualError(t, err, `DecreaseLatencyPercentile: more than one`)
		require.Nil(t, bp)
	})

	main.Run("DecreaseLatencyZero", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:            time.Second,
			DecreasePercent:           0.04,
			IncreasePercent:           0.02,
			ThresholdPercent:          0.01,
			DecreaseLatencyPercentile: 0.8,
		})
		require.EqualError(t, err, `DecreaseLatency: required`)
		require.Nil(t, bp)
	})

	main.Run("SameLatencyPercentileLessThanZero", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:        time.Second,
			DecreasePercent:       0.04,
			IncreasePercent:       0.02,
			ThresholdPercent:      0.01,
			SameLatencyPercentile: -0.01,
		})
		require.EqualError(t, err, `SameLatencyPercentile: less than zero`)
		require.Nil(t, bp)
	})

	main.Run("SameLatencyPercentileMoreThanOne", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:        time.Second,
			DecreasePercent:       0.04,
			IncreasePercent:       0.02,
			ThresholdPercent:      0.01,
			SameLatencyPercentile: 1.01,
		})
		require.EqualError(t, err, `SameLatencyPercentile: more than one`)
		require.Nil(t, bp)
	})

	main.Run("SameLatencyZero", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:        time.Second,
			DecreasePercent:       0.04,
			IncreasePercent:       0.02,
			ThresholdPercent:      0.01,
			SameLatencyPercentile: 0.8,
		})
		require.EqualError(t, err, `SameLatency: required`)
		require.Nil(t, bp)
	})

}

//func TestDecide(t *testing.T) {
//	bp, err := NewAIMD(AIMDConfig{
//		DecideInterval:   time.Millisecond * 5,
//		ThresholdPercent: 0.01,
//		IncreasePercent:  1.01,
//		DecreasePercent:  0.8,
//	})
//	require.NoError(t, err)
//
//}
