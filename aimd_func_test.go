package backpressure_test

import (
	"fmt"
	"math"
	"sync/atomic"
	"testing"
	"time"

	"github.com/makasim/backpressure"
	"github.com/stretchr/testify/require"
)

type req struct {
	resCh chan error
}

var tMul = time.Duration(2)

func TestNoCongestion(t *testing.T) {
	t.Parallel()

	bp, err := backpressure.NewAIMD(backpressure.AIMDConfig{
		DecideInterval:  time.Millisecond * tMul,
		IncreasePercent: 0.02,
		DecreasePercent: 0.2,
	})
	require.NoError(t, err)

	c := newClient(10)
	p := newProxy(c.outCh, 50, bp)
	o := newOrigin(func(idx, used int64, req req) {
		time.Sleep(time.Microsecond * 900 * tMul)
		req.resCh <- nil
	}, 10, p.outCh)

	defer o.run()()
	defer p.run()()
	defer c.run()()

	time.Sleep(time.Second * tMul)

	ok, failed := c.stats()
	int64InRange(t, 0, 0, failed)
	int64InRange(t, 9500, 10000, ok)

	s := bp.Stats()
	int64InRange(t, 4, 12, s.Used)
	int64InRange(t, 4, math.MaxInt64, s.Max)
	int64InRange(t, 0, 0, s.DeniedCounter)
	int64InRange(t, 0, 0, s.CongestedCounter)
	int64InRange(t, 9500, 10000, s.SuccessfulCounter)

}

func TestNoCongestionSlowHandlers(t *testing.T) {
	t.Parallel()

	bp, err := backpressure.NewAIMD(backpressure.AIMDConfig{
		DecideInterval:  time.Millisecond * tMul,
		IncreasePercent: 0.02,
		DecreasePercent: 0.2,
	})
	require.NoError(t, err)

	c := newClient(10)
	p := newProxy(c.outCh, 50, bp)
	o := newOrigin(func(idx, used int64, req req) {
		time.Sleep(time.Microsecond * 1900 * tMul)
		req.resCh <- nil
	}, 20, p.outCh)

	defer o.run()()
	defer p.run()()
	defer c.run()()

	time.Sleep(time.Second * tMul)

	ok, failed := c.stats()
	int64InRange(t, 0, 0, failed)
	int64InRange(t, 9500, 10000, ok)

	s := bp.Stats()
	int64InRange(t, 17, 23, s.Used)
	int64InRange(t, 17, math.MaxInt64, s.Max)
	int64InRange(t, 0, 0, s.DeniedCounter)
	int64InRange(t, 0, 0, s.CongestedCounter)
	int64InRange(t, 9500, 10000, s.SuccessfulCounter)
}

func TestNoCongestionFewHandlers(t *testing.T) {
	t.Parallel()

	bp, err := backpressure.NewAIMD(backpressure.AIMDConfig{
		DecideInterval:  time.Millisecond * tMul,
		IncreasePercent: 0.02,
		DecreasePercent: 0.2,
	})
	require.NoError(t, err)

	c := newClient(10)
	p := newProxy(c.outCh, 50, bp)
	o := newOrigin(func(idx, used int64, req req) {
		time.Sleep(time.Microsecond * 400 * tMul)
		req.resCh <- nil
	}, 5, p.outCh)

	defer o.run()()
	defer p.run()()
	defer c.run()()

	time.Sleep(time.Second * tMul)

	ok, failed := c.stats()
	int64InRange(t, 0, 0, failed)
	int64InRange(t, 9500, 10000, ok)

	s := bp.Stats()
	int64InRange(t, 4, 6, s.Used)
	int64InRange(t, 4, math.MaxInt64, s.Max)
	int64InRange(t, 0, 0, s.DeniedCounter)
	int64InRange(t, 0, 0, s.CongestedCounter)
	int64InRange(t, 9500, 10000, s.SuccessfulCounter)
}

func TestCongestion20Percent(t *testing.T) {
	t.Parallel()

	bp, err := backpressure.NewAIMD(backpressure.AIMDConfig{
		DecideInterval:  time.Millisecond * tMul,
		IncreasePercent: 0.02,
		DecreasePercent: 0.2,
	})
	require.NoError(t, err)

	c := newClient(10)
	p := newProxy(c.outCh, 50, bp)
	o := newOrigin(func(idx, used int64, req req) {
		time.Sleep(time.Microsecond * 1200 * tMul)
		req.resCh <- nil
	}, 10, p.outCh)

	defer o.run()()
	defer p.run()()
	defer c.run()()

	time.Sleep(time.Second * tMul)

	ok, failed := c.stats()
	int64InRange(t, 500, 2500, failed)
	int64InRange(t, 7500, 9000, ok)

	s := bp.Stats()
	int64InRange(t, 20, 50, s.Used)
	int64InRange(t, 20, math.MaxInt64, s.Max)

	int64InRange(t, 1, math.MaxInt64, s.CongestedCounter)
	int64InRange(t, 1, math.MaxInt64, s.DeniedCounter)
	int64InRange(t, 500, 2500, s.DeniedCounter+s.CongestedCounter)

	int64InRange(t, 7500, 9000, s.SuccessfulCounter)
}

func TestCongestion50Percent(t *testing.T) {
	t.Parallel()

	bp, err := backpressure.NewAIMD(backpressure.AIMDConfig{
		DecideInterval:  time.Millisecond * tMul,
		IncreasePercent: 0.02,
		DecreasePercent: 0.2,
	})
	require.NoError(t, err)

	c := newClient(10)
	p := newProxy(c.outCh, 50, bp)
	o := newOrigin(func(idx, used int64, req req) {
		time.Sleep(time.Microsecond * 2000 * tMul)
		req.resCh <- nil
	}, 10, p.outCh)

	defer o.run()()
	defer p.run()()
	defer c.run()()

	time.Sleep(time.Second * tMul)

	ok, failed := c.stats()
	int64InRange(t, 4000, 6000, failed)
	int64InRange(t, 4000, 6000, ok)

	s := bp.Stats()
	int64InRange(t, 20, 50, s.Used)
	int64InRange(t, 20, math.MaxInt64, s.Max)

	int64InRange(t, 1, math.MaxInt64, s.CongestedCounter)
	int64InRange(t, 1, math.MaxInt64, s.DeniedCounter)
	int64InRange(t, 4000, 6000, s.DeniedCounter+s.CongestedCounter)

	int64InRange(t, 4000, 6000, s.SuccessfulCounter)
}

func TestCongestion50PercentAndRecover(t *testing.T) {
	t.Parallel()

	bp, err := backpressure.NewAIMD(backpressure.AIMDConfig{
		DecideInterval:  time.Millisecond * tMul,
		IncreasePercent: 0.02,
		DecreasePercent: 0.2,
	})
	require.NoError(t, err)

	c := newClient(10)
	p := newProxy(c.outCh, 50, bp)
	o := newOrigin(func(idx, used int64, req req) {
		if idx < 2000 {
			time.Sleep(time.Microsecond * 1900 * tMul)
		} else {
			time.Sleep(time.Microsecond * 950 * tMul)
		}
		req.resCh <- nil
	}, 10, p.outCh)

	defer o.run()()
	defer p.run()()
	defer c.run()()

	time.Sleep(time.Second * tMul)

	ok, failed := c.stats()
	int64InRange(t, 1000, 2000, failed)
	int64InRange(t, 8000, 8500, ok)

	s := bp.Stats()
	int64InRange(t, 8, 12, s.Used)
	int64InRange(t, 8, math.MaxInt64, s.Max)

	int64InRange(t, 1, math.MaxInt64, s.CongestedCounter)
	int64InRange(t, 1, math.MaxInt64, s.DeniedCounter)
	int64InRange(t, 1000, 3000, s.DeniedCounter+s.CongestedCounter)

	int64InRange(t, 7500, 8500, s.SuccessfulCounter)
}

func TestDecreaseLatency(t *testing.T) {
	t.Parallel()

	bp, err := backpressure.NewAIMD(backpressure.AIMDConfig{
		DecideInterval:            time.Millisecond * tMul,
		IncreasePercent:           0.02,
		DecreasePercent:           0.2,
		DecreaseLatencyPercentile: 0.8,
		DecreaseLatency:           time.Microsecond * 1500 * tMul,
	})
	require.NoError(t, err)

	c := newClient(30)
	p := newProxy(c.outCh, 1000, bp)
	o := newOrigin(func(idx, used int64, req req) {
		if used > 25 {
			time.Sleep(time.Microsecond * 2500 * tMul)
		} else if used > 20 {
			time.Sleep(time.Microsecond * 2000 * tMul)
		} else if used > 15 {
			time.Sleep(time.Microsecond * 1750 * tMul)
		} else if used > 12 {
			time.Sleep(time.Microsecond * 1500 * tMul)
		} else {
			time.Sleep(time.Microsecond * 950 * tMul)
		}
		req.resCh <- nil
	}, 1000, p.outCh)

	defer o.run()()
	defer p.run()()
	defer c.run()()

	time.Sleep(time.Second * tMul)

	ok, failed := c.stats()
	int64InRange(t, 18000, 23000, failed)
	int64InRange(t, 9000, 11000, ok)

	s := bp.Stats()
	int64InRange(t, 0, math.MaxInt64, s.CongestedCounter)
	int64InRange(t, 1, math.MaxInt64, s.DeniedCounter)
	int64InRange(t, 18000, 23000, s.DeniedCounter+s.CongestedCounter)

	int64InRange(t, 9000, 11000, s.SuccessfulCounter)
}

type client struct {
	rpms  int
	outCh chan req

	ok     int64
	failed int64
}

func newClient(rpms int) *client {
	return &client{
		rpms:   rpms,
		outCh:  make(chan req, 20),
		ok:     0,
		failed: 0,
	}
}

func (c *client) run() func() {
	closeCh := make(chan struct{})

	go c.worker(closeCh)

	return func() {
		close(closeCh)
	}
}

func (c *client) worker(closeCh chan struct{}) {
	internalRPS := int(math.Ceil(float64(c.rpms / 10)))
	tokensCh := make(chan struct{}, internalRPS)

	t := time.NewTicker(time.Microsecond * 100 * tMul)
	defer t.Stop()

	for {
		select {
		case <-t.C:
		fill:
			for {
				select {
				case tokensCh <- struct{}{}:
				default:
					break fill
				}
			}
		case <-tokensCh:
			go func() {
				req := req{resCh: make(chan error, 1)}

				select {
				case c.outCh <- req:
					if err := <-req.resCh; err != nil {
						atomic.AddInt64(&c.failed, 1)
					} else {
						atomic.AddInt64(&c.ok, 1)
					}
				default:
					atomic.AddInt64(&c.failed, 1)
				}
			}()
		case <-closeCh:
			return
		}
	}
}

func (c *client) stats() (int64, int64) {
	ok := atomic.LoadInt64(&c.ok)
	failed := atomic.LoadInt64(&c.failed)

	return ok, failed
}

type proxy struct {
	inCh  chan req
	outCh chan req
	bp    *backpressure.AIMD
	wNum  int
}

func newProxy(inCh chan req, wNum int, bp *backpressure.AIMD) *proxy {
	return &proxy{
		inCh:  inCh,
		outCh: make(chan req, 20),
		bp:    bp,
		wNum:  wNum,
	}
}

func (p *proxy) run() func() {
	closeCh := make(chan struct{})

	for i := 0; i < p.wNum; i++ {
		go p.worker(closeCh)
	}

	return func() {
		close(closeCh)
	}
}

func (p *proxy) worker(closeCh chan struct{}) {
	for {
		select {
		case req := <-p.inCh:
			t, allowed := p.bp.Acquire()
			if !allowed {
				req.resCh <- fmt.Errorf("bp: disallowed")
				continue
			}

			clientResCh := req.resCh
			req.resCh = make(chan error, 1)

			select {
			case p.outCh <- req:
				res := <-req.resCh
				clientResCh <- res
				p.bp.Release(t)
			default:
				req.resCh <- fmt.Errorf("client: no capacity")
				t.Congested = true
				p.bp.Release(t)
			}
		case <-closeCh:
			return
		}
	}
}

type origin struct {
	inCh chan req
	wNum int
	idx  int64
	used int64
	h    func(idx, used int64, req req)
}

func newOrigin(h func(idx, used int64, req req), wNum int, inCh chan req) *origin {
	return &origin{
		inCh: inCh,
		wNum: wNum,
		h:    h,
	}
}

func (o *origin) run() func() {
	closeCh := make(chan struct{})

	for i := 0; i < o.wNum; i++ {
		go o.worker(closeCh)
	}

	return func() {
		close(closeCh)
	}
}

func (o *origin) worker(closeCh chan struct{}) {
	for {
		select {
		case req := <-o.inCh:
			used := atomic.AddInt64(&o.used, 1)
			idx := atomic.AddInt64(&o.idx, 1)
			o.h(idx, used, req)
			atomic.AddInt64(&o.used, -1)
		case <-closeCh:
			return
		}
	}
}

func intInRange(t *testing.T, from, to, act int) {
	require.GreaterOrEqual(t, act, from)
	require.LessOrEqual(t, act, to)
}

func int64InRange(t *testing.T, from, to, act int64) {
	require.GreaterOrEqual(t, act, from)
	require.LessOrEqual(t, act, to)
}
