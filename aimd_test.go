package backpressure

import (
	"math"
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

	main.Run("MinMaxDefault", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:   time.Second,
			DecreasePercent:  0.04,
			IncreasePercent:  0.02,
			ThresholdPercent: 0.01,

			MinMax: 0,
		})
		require.NoError(t, err)
		require.Equal(t, int64(1), bp.cfg.MinMax)
	})

	main.Run("MaxMaxDefault", func(t *testing.T) {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:   time.Second,
			DecreasePercent:  0.04,
			IncreasePercent:  0.02,
			ThresholdPercent: 0.01,

			MaxMax: 0,
		})
		require.NoError(t, err)
		require.Equal(t, int64(math.MaxInt64), bp.cfg.MaxMax)
	})
}

func TestDecide(main *testing.T) {
	setUp := func(t *testing.T) *AIMD {
		bp, err := NewAIMD(AIMDConfig{
			DecideInterval:   time.Second,
			DecreasePercent:  0.20,
			IncreasePercent:  0.10,
			ThresholdPercent: 0.10,

			MinMax: 1,
			MaxMax: 100,
		})
		require.NoError(t, err)

		return bp
	}

	main.Run("NoTraffic", func(t *testing.T) {
		bp := setUp(t)

		bp.cfg.IncreasePercent = 0.0001
		bp.max = 80

		bp.successful = 0
		bp.congested = 0

		bp.decide()
		require.Equal(t, int64(80), bp.max)
	})

	main.Run("MaxMax", func(t *testing.T) {
		bp := setUp(t)

		bp.cfg.MaxMax = math.MaxInt64

		bp.cfg.IncreasePercent = 0.2
		bp.max = math.MaxInt64 - 1

		bp.successful = 100
		bp.decide()

		bp.successful = 100
		bp.decide()

		bp.successful = 100
		bp.decide()

		bp.successful = 100
		bp.decide()
		require.Equal(t, int64(math.MaxInt64), bp.max)
	})

	main.Run("IncreaseSmall", func(t *testing.T) {
		bp := setUp(t)

		bp.cfg.IncreasePercent = 0.0001
		bp.max = 80

		bp.successful = 100
		bp.decide()
		require.Equal(t, int64(81), bp.max)

		bp.successful = 200
		bp.decide()
		require.Equal(t, int64(82), bp.max)
	})

	main.Run("IncreaseNormal", func(t *testing.T) {
		bp := setUp(t)

		bp.max = 80

		bp.successful = 100
		bp.decide()
		require.Equal(t, int64(89), bp.max)

		bp.successful = 100
		bp.decide()
		require.Equal(t, int64(98), bp.max)

		bp.successful = 100
		bp.decide()
		require.Equal(t, int64(100), bp.max)
	})

	main.Run("ModerateCongestion", func(t *testing.T) {
		bp := setUp(t)

		bp.max = 80

		bp.successful = 1000
		bp.congested = 99
		bp.decide()
		require.Equal(t, int64(80), bp.max)

		bp.successful = 2000
		bp.congested = 199
		bp.decide()
		require.Equal(t, int64(80), bp.max)
	})

	main.Run("HighCongestion", func(t *testing.T) {
		bp := setUp(t)

		bp.max = 80

		bp.successful = 1000
		bp.congested = 115
		bp.decide()
		require.Equal(t, int64(64), bp.max)

		bp.successful = 2000
		bp.congested = 225
		bp.decide()
		require.Equal(t, int64(52), bp.max)
	})

	main.Run("HighCongestionTooBigMaxNoUsedMax", func(t *testing.T) {
		bp := setUp(t)

		bp.max = 10000000000

		bp.successful = 1000
		bp.congested = 115
		bp.decide()
		require.Equal(t, int64(8000000000), bp.max)
	})

	main.Run("HighCongestionTooBigMaxWithUsedMax", func(t *testing.T) {
		bp := setUp(t)

		bp.max = 10000000000
		bp.usedMax = 1000

		bp.successful = 1000
		bp.congested = 115
		bp.decide()
		require.Equal(t, int64(800), bp.max)
	})

	main.Run("HighCongestionMinMax", func(t *testing.T) {
		bp := setUp(t)

		bp.cfg.DecreasePercent = 0.99
		bp.cfg.ThresholdPercent = 0

		bp.max = 100

		bp.successful = 10
		bp.congested = 1
		bp.decide()
		require.Equal(t, int64(2), bp.max)

		bp.successful = 10
		bp.congested = 1
		bp.decide()
		require.Equal(t, int64(1), bp.max)

		bp.successful = 10
		bp.congested = 1
		bp.decide()
		require.Equal(t, int64(1), bp.max)
	})

	main.Run("NoLatency", func(t *testing.T) {
		bp := setUp(t)

		bp.cfg.SameLatencyPercentile = 0.5
		bp.cfg.SameLatency = time.Second
		bp.cfg.DecreaseLatencyPercentile = 0.5
		bp.cfg.DecreaseLatency = time.Second * 2

		bp.max = 80

		require.NoError(t, bp.h.RecordValue((time.Millisecond * 900).Nanoseconds()))
		require.NoError(t, bp.h.RecordValue((time.Millisecond * 900).Nanoseconds()))
		require.NoError(t, bp.h.RecordValue((time.Millisecond * 1100).Nanoseconds()))

		bp.successful = 100
		bp.decide()
		require.Equal(t, int64(89), bp.max)
	})

	main.Run("ModerateLatency", func(t *testing.T) {
		bp := setUp(t)

		bp.cfg.SameLatencyPercentile = 0.5
		bp.cfg.SameLatency = time.Second
		bp.cfg.DecreaseLatencyPercentile = 0.5
		bp.cfg.DecreaseLatency = time.Second * 2

		bp.max = 80

		require.NoError(t, bp.h.RecordValue((time.Millisecond * 900).Nanoseconds()))
		require.NoError(t, bp.h.RecordValue((time.Millisecond * 1100).Nanoseconds()))
		require.NoError(t, bp.h.RecordValue((time.Millisecond * 1100).Nanoseconds()))

		bp.successful = 100
		bp.decide()
		require.Equal(t, int64(80), bp.max)
	})

	main.Run("HighLatency", func(t *testing.T) {
		bp := setUp(t)

		bp.cfg.SameLatencyPercentile = 0.5
		bp.cfg.SameLatency = time.Second
		bp.cfg.DecreaseLatencyPercentile = 0.5
		bp.cfg.DecreaseLatency = time.Second * 2

		bp.max = 80

		require.NoError(t, bp.h.RecordValue((time.Millisecond * 1900).Nanoseconds()))
		require.NoError(t, bp.h.RecordValue((time.Millisecond * 2100).Nanoseconds()))
		require.NoError(t, bp.h.RecordValue((time.Millisecond * 2100).Nanoseconds()))

		bp.successful = 100
		bp.decide()
		require.Equal(t, int64(64), bp.max)
	})

	main.Run("IncrSameDecr", func(t *testing.T) {
		bp := setUp(t)

		bp.max = 80

		bp.successful = 1000
		bp.congested = 0
		bp.decide()
		require.Equal(t, int64(89), bp.max)

		bp.successful = 1000
		bp.congested = 100
		bp.decide()
		require.Equal(t, int64(89), bp.max)

		bp.successful = 1000
		bp.congested = 300
		bp.decide()
		require.Equal(t, int64(72), bp.max)
	})

}
