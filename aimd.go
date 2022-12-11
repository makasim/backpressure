package backpressure

import (
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
)

type AIMDConfig struct {
	DecideInterval time.Duration

	ThresholdPercent float64
	IncreasePercent  float64
	DecreasePercent  float64

	MaxMax int64
	MinMax int64

	Latency           time.Duration
	LatencyPercentile float64
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
		DecideInterval:   time.Second * 5,
		ThresholdPercent: 0.01,
		IncreasePercent:  1.02,
		DecreasePercent:  0.8,
		MaxMax:           math.MaxInt,
		MinMax:           1,
	}
}

type AIMD struct {
	cfg     AIMDConfig
	dt      *time.Ticker
	closeCh chan struct{}

	max        int64
	used       int64
	denied     int64
	successful int64
	congested  int64

	stats    AIMDStats
	muxStats sync.RWMutex

	h *hdrhistogram.Histogram
}

func NewAIMD(cfg AIMDConfig) (*AIMD, error) {

	if cfg.MaxMax <= 0 {
		cfg.MaxMax = math.MaxInt
	}

	bp := &AIMD{
		cfg:     cfg,
		dt:      time.NewTicker(cfg.DecideInterval),
		closeCh: make(chan struct{}),

		max: cfg.MaxMax,

		h: hdrhistogram.New(0, (time.Minute * 10).Nanoseconds(), 1),
	}

	go bp.decideLoop()

	return bp, nil
}

func (bp *AIMD) Acquire() (Token, bool) {
	select {
	case <-bp.closeCh:
		return Token{}, false
	default:
	}

	used := atomic.AddInt64(&bp.used, 1)
	if used > atomic.LoadInt64(&bp.max) {
		atomic.AddInt64(&bp.used, -1)
		atomic.AddInt64(&bp.denied, 1)
		return Token{}, false
	}

	return Token{
		start: time.Now().UnixNano(),
	}, true
}

func (bp *AIMD) Release(t Token) {
	atomic.AddInt64(&bp.used, -1)

	startT := time.Unix(0, t.start)
	dur := time.Now().Sub(startT).Nanoseconds()
	if err := bp.h.RecordValue(dur); err != nil {
		log.Printf("[ERROR] backpressure: histogram: record value: %s", err)
	}

	if !t.Congested {
		atomic.AddInt64(&bp.successful, 1)
	} else {
		atomic.AddInt64(&bp.congested, 1)
	}
}

func (bp *AIMD) Stats() AIMDStats {
	bp.muxStats.RLock()
	defer bp.muxStats.RUnlock()

	s := bp.stats
	s.Used = atomic.LoadInt64(&bp.used)

	return s
}

func (bp *AIMD) decideLoop() {
	for {
		select {
		case <-bp.dt.C:
			successful := atomic.SwapInt64(&bp.successful, 0)
			congested := atomic.SwapInt64(&bp.congested, 0)
			denied := atomic.SwapInt64(&bp.denied, 0)
			max := atomic.LoadInt64(&bp.max)

			congestedPercent := float64(successful+congested) / 100 * float64(congested)

			var incr, decr, same int64
			switch {
			case bp.cfg.Latency > 0 && bp.h.ValueAtPercentile(bp.cfg.LatencyPercentile) > bp.cfg.Latency.Nanoseconds():
				bp.decr(max)
				decr++
			case congestedPercent >= bp.cfg.ThresholdPercent:
				bp.decr(max)
				decr++
			case congestedPercent < bp.cfg.ThresholdPercent && congestedPercent > 0:
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

			bp.h.Reset()
		case <-bp.closeCh:
			bp.dt.Stop()
			select {
			case <-bp.dt.C:
			default:
			}

			return
		}
	}
}

func (bp *AIMD) incr(max int64) {
	newMax := int64(math.Floor(float64(max)*bp.cfg.IncreasePercent) + 1)
	if newMax > bp.cfg.MaxMax {
		newMax = bp.cfg.MaxMax
	}

	atomic.StoreInt64(&bp.max, newMax)
}

func (bp *AIMD) decr(max int64) {
	max = int64(math.Ceil(float64(max) * bp.cfg.DecreasePercent))
	if max < bp.cfg.MinMax {
		max = bp.cfg.MinMax
	}
	atomic.StoreInt64(&bp.max, max)
}

func (bp *AIMD) Close() {
	close(bp.closeCh)

}

type Token struct {
	Congested bool
	start     int64
}
