package lookout

import (
	"context"
	"net"
	"net/http"

	"github.com/ipfs/go-log/v2"
	"github.com/ipni/lookout/check"
	"github.com/ipni/lookout/metrics"
	"github.com/ipni/lookout/perform"
	"github.com/ipni/lookout/sample"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	logger = log.Logger("ipni/lookout")
)

type (
	Lookout struct {
		*options
		s       *http.Server
		metrics *metrics.Metrics
	}
)

func New(o ...Option) (*Lookout, error) {
	var l Lookout
	var err error
	l.options, err = newOptions(o...)
	if err != nil {
		return nil, err
	}
	l.s = &http.Server{
		Addr:      l.httpListenAddr,
		Handler:   l.serveMux(),
		TLSConfig: nil,
	}
	l.metrics = metrics.New()
	return &l, nil
}

func (l *Lookout) Start(ctx context.Context) error {
	if err := l.metrics.Start(); err != nil {
		return err
	}
	ln, err := net.Listen("tcp", l.s.Addr)
	if err != nil {
		return err
	}
	go func() { _ = l.s.Serve(ln) }()

	wctx, cancel := context.WithCancel(context.Background())
	ssch := make(chan *sample.Set)
	go l.sample(wctx, ssch)
	go l.check(wctx, ssch)
	l.s.RegisterOnShutdown(cancel)

	logger.Infow("Server started", "httpAddr", ln.Addr())
	return nil
}

func (l *Lookout) check(ctx context.Context, targets <-chan *sample.Set) {
	for {
		select {
		case <-ctx.Done():
			logger.Info("Checkers cycle stopped", "err", ctx.Err())
			return
		case ss, ok := <-targets:
			if !ok {
				logger.Info("Checkers cycle stopped; no more work")
				return
			}
			logger := logger.With("size", len(ss.Multihashes), "name", ss.Name)
			logger.Info("Running checks on sample set...")

			results := perform.InParallel(ctx, l.checkersParallelism, l.checkers, func(ctx context.Context, c check.Checker) *check.Results {
				return c.Check(ctx, ss)
			})
		ResultsLoop:
			for {
				select {
				case <-ctx.Done():
					logger.Warnw("Check cycle stopped while performing checks.", "err", ctx.Err())
					return
				case r, ok := <-results:
					if !ok {
						break ResultsLoop
					}
					l.metrics.NotifyCheckResults(ctx, r)
				}
			}
			logger.Info("Checks finished.")
		}
	}
}

func (l *Lookout) sample(ctx context.Context, check chan<- *sample.Set) {
	runCycle := func() {
		sets := perform.InParallel(ctx, l.samplersParallelism, l.samplers, func(ctx context.Context, s sample.Sampler) *sample.Set {
			ss, err := s.Sample(ctx)
			if err != nil {
				logger.Errorw("Failed to sample.", "err", err)
				return nil
			}
			return ss
		})
		for {
			select {
			case <-ctx.Done():
				return
			case set, ok := <-sets:
				if !ok {
					return
				}
				if set == nil {
					continue
				}
				logger.Infow("Selected samples", "count", len(set.Multihashes), "name", set.Name)
				l.metrics.NotifySampleSet(ctx, set)
				select {
				case <-ctx.Done():
					return
				case check <- set:
				}
			}
		}
	}
	runCycle()
	for {
		select {
		case <-ctx.Done():
			logger.Info("Monitoring stopped", "err", ctx.Err())
			return
		case <-l.checkInterval.C:
			runCycle()
		}
	}
}

func (l *Lookout) serveMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}

func (l *Lookout) Shutdown(ctx context.Context) error {
	serr := l.s.Shutdown(ctx)
	_ = l.metrics.Shutdown(ctx)
	return serr
}
