package backpressure

import (
	"fmt"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
)

type AIMDConfig struct {
	// DecidePeriod defines periods when the decision on capacity is made: increase, keep same, decrease
	DecidePeriod time.Duration

	// ThresholdPercent defines a congestion threshold (tp).
	// If congestion below threshold the capacity is kept the same.
	// If congestion above threshold the capacity is decreased.
	ThresholdPercent float64

	// IncreasePercent defines an increase percent of current capacity.
	IncreasePercent float64

	// DecreasePercent defines an decrease percent of current capacity.
	DecreasePercent float64

	// MaxMax defines a maximum possible capacity. Default math.MaxInt64
	MaxMax int64
	// MInMax defines a minimum possible capacity. Default 1
	MinMax int64

	// SameLatency The capacity is kept the same if latency goes above the value at given percentile
	SameLatency           time.Duration
	SameLatencyPercentile float64

	// DecreaseLatency The capacity is decreased if latency goes above the value at given percentile
	DecreaseLatency           time.Duration
	DecreaseLatencyPercentile float64
}

type AIMDStats struct {
	Max  int64
	Used int64

	MaxMax                int64
	MaxMin                int64
	SuccessfulCounter     int64
	CongestedCounter      int64
	DeniedCounter         int64
	DecideIncreaseCounter int64
	DecideDecreaseCounter int64
	DecideSameCounter     int64
}

func DefaultAIMDConfig() AIMDConfig {
	return AIMDConfig{
		DecidePeriod:     time.Second * 5,
		ThresholdPercent: 0.01,
		IncreasePercent:  0.02,
		DecreasePercent:  0.2,
		MaxMax:           math.MaxInt,
		MinMax:           1,
	}
}

type AIMD struct {
	cfg AIMDConfig
	dt  *time.Ticker

	max        int64
	used       int64
	usedMax    int64
	denied     int64
	successful int64
	congested  int64

	stats    AIMDStats
	muxStats sync.RWMutex

	muxH sync.RWMutex
	h    *hdrhistogram.Histogram
}

func NewAIMD(cfg AIMDConfig) (*AIMD, error) {
	if err := validateAIMDConfig(cfg); err != nil {
		return nil, err
	}

	if cfg.MinMax <= 0 {
		cfg.MinMax = 1
	}
	if cfg.MaxMax == 0 {
		cfg.MaxMax = math.MaxInt64
	}

	bp := &AIMD{
		cfg: cfg,
		dt:  time.NewTicker(cfg.DecidePeriod),

		max: cfg.MaxMax,

		h: hdrhistogram.New(0, (time.Minute * 10).Nanoseconds(), 1),
	}

	return bp, nil
}

func (bp *AIMD) Acquire() (Token, bool) {
	select {
	case <-bp.dt.C:
		bp.decide()
	default:
	}

	used := atomic.AddInt64(&bp.used, 1)
	if used > atomic.LoadInt64(&bp.max) {
		atomic.AddInt64(&bp.used, -1)
		atomic.AddInt64(&bp.denied, 1)
		return Token{}, false
	}

loop:
	for {
		usedMax := atomic.LoadInt64(&bp.usedMax)
		if used <= usedMax {
			break loop
		}

		if atomic.CompareAndSwapInt64(&bp.usedMax, usedMax, used) {
			break loop
		}
	}

	return Token{
		start: time.Now().UnixNano(),
	}, true
}

func (bp *AIMD) Release(t Token) {
	atomic.AddInt64(&bp.used, -1)

	if !t.Congested {
		atomic.AddInt64(&bp.successful, 1)
	} else {
		atomic.AddInt64(&bp.congested, 1)
	}

	if bp.cfg.DecreaseLatencyPercentile > 0 || bp.cfg.SameLatencyPercentile > 0 {
		startT := time.Unix(0, t.start)
		dur := time.Now().Sub(startT).Nanoseconds()

		bp.muxH.Lock()
		defer bp.muxH.Unlock()
		if err := bp.h.RecordValue(dur); err != nil {
			log.Printf("[ERROR] backpressure: histogram: record value: %s", err)
		}
	}
}

func (bp *AIMD) Stats() AIMDStats {
	bp.muxStats.RLock()
	defer bp.muxStats.RUnlock()

	s := bp.stats
	s.Used = atomic.LoadInt64(&bp.used)

	return s
}

func (bp *AIMD) decide() {
	successful := atomic.SwapInt64(&bp.successful, 0)
	congested := atomic.SwapInt64(&bp.congested, 0)
	denied := atomic.SwapInt64(&bp.denied, 0)
	max := atomic.LoadInt64(&bp.max)

	if successful+congested == 0 {
		return
	}

	highLatency := false
	moderateLatency := false
	if bp.cfg.DecreaseLatencyPercentile > 0 || bp.cfg.SameLatencyPercentile > 0 {
		bp.muxH.Lock()
		highLatency = bp.h.ValueAtPercentile(bp.cfg.DecreaseLatencyPercentile*100) > bp.cfg.DecreaseLatency.Nanoseconds()
		moderateLatency = bp.cfg.SameLatency > 0 && bp.h.ValueAtPercentile(bp.cfg.SameLatencyPercentile*100) > bp.cfg.SameLatency.Nanoseconds()
		bp.muxH.Unlock()
	}

	congestedPercent := float64(congested) / float64(successful+congested)
	highCongestion := congestedPercent != 0 && congestedPercent >= bp.cfg.ThresholdPercent
	moderateCongestion := congestedPercent > 0 && congestedPercent < bp.cfg.ThresholdPercent

	var incr, decr, same int64
	switch {
	case highCongestion || highLatency:
		bp.decr(max)
		decr++
	case moderateCongestion || moderateLatency:
		same++
		// keep current max
	default:
		bp.incr(max)
		incr++
	}

	bp.muxStats.Lock()
	bp.stats = AIMDStats{
		Max:                   atomic.LoadInt64(&bp.max),
		MaxMax:                bp.cfg.MaxMax,
		MaxMin:                bp.cfg.MinMax,
		SuccessfulCounter:     bp.stats.SuccessfulCounter + successful,
		CongestedCounter:      bp.stats.CongestedCounter + congested,
		DeniedCounter:         bp.stats.DeniedCounter + denied,
		DecideIncreaseCounter: bp.stats.DecideIncreaseCounter + incr,
		DecideDecreaseCounter: bp.stats.DecideDecreaseCounter + decr,
		DecideSameCounter:     bp.stats.DecideSameCounter + same,
	}
	bp.muxStats.Unlock()

	if bp.cfg.DecreaseLatencyPercentile > 0 || bp.cfg.SameLatencyPercentile > 0 {
		bp.muxH.Lock()
		bp.h.Reset()
		bp.muxH.Unlock()
	}
}

func (bp *AIMD) incr(max int64) {
	newMax := int64(math.Floor(float64(max)*(1+bp.cfg.IncreasePercent)) + 1)
	if newMax < 0 {
		newMax = math.MaxInt64
	}
	if newMax > bp.cfg.MaxMax {
		newMax = bp.cfg.MaxMax
	}

	atomic.StoreInt64(&bp.max, newMax)
}

func (bp *AIMD) decr(max int64) {
	usedMax := atomic.LoadInt64(&bp.usedMax)
	if usedMax != 0 && max > usedMax {
		max = usedMax
	}

	max = int64(math.Ceil(float64(max) * (1 - bp.cfg.DecreasePercent)))
	if max < bp.cfg.MinMax {
		max = bp.cfg.MinMax
	}
	atomic.StoreInt64(&bp.max, max)
}

type Token struct {
	Congested bool
	start     int64
}

func validateAIMDConfig(cfg AIMDConfig) error {
	if cfg.DecidePeriod == 0 {
		return fmt.Errorf("DecidePeriod: required")
	} else if cfg.DecidePeriod < 0 {
		return fmt.Errorf("DecidePeriod: negative")
	}

	if cfg.MinMax != 0 && cfg.MaxMax != 0 && cfg.MinMax >= cfg.MaxMax {
		return fmt.Errorf("MinMax: must be less than MaxMax")
	}

	if err := validatePercent(cfg.DecreasePercent); err != nil {
		return fmt.Errorf("DecreasePercent: %s", err)
	}
	if cfg.DecreasePercent == 0 {
		return fmt.Errorf("DecreasePercent: required")
	}

	if err := validatePercent(cfg.IncreasePercent); err != nil {
		return fmt.Errorf("IncreasePercent: %s", err)
	}
	if cfg.IncreasePercent == 0 {
		return fmt.Errorf("IncreasePercent: required")
	}

	if cfg.DecreasePercent <= cfg.IncreasePercent {
		return fmt.Errorf("IncreasePercent: must be less than DecreasePercent")
	}

	if err := validatePercent(cfg.ThresholdPercent); err != nil {
		return fmt.Errorf("ThresholdPercent: %s", err)
	}

	if err := validatePercent(cfg.DecreaseLatencyPercentile); err != nil {
		return fmt.Errorf("DecreaseLatencyPercentile: %s", err)
	}
	if cfg.DecreaseLatencyPercentile != 0 && cfg.DecreaseLatency == 0 {
		return fmt.Errorf("DecreaseLatency: required")
	}

	if err := validatePercent(cfg.SameLatencyPercentile); err != nil {
		return fmt.Errorf("SameLatencyPercentile: %s", err)
	}
	if cfg.SameLatencyPercentile != 0 && cfg.SameLatency == 0 {
		return fmt.Errorf("SameLatency: required")
	}
	if cfg.SameLatencyPercentile != 0 && cfg.SameLatency == 0 {
		return fmt.Errorf("SameLatency: required")
	}

	return nil
}

func validatePercent(p float64) error {
	if p < 0 {
		return fmt.Errorf("less than zero")
	}
	if p > 1 {
		return fmt.Errorf("more than one")
	}

	return nil
}
