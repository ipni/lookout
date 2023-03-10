package lookout

import (
	"time"

	"github.com/ipni/lookout/check"
	"github.com/ipni/lookout/sample"
)

type (
	Option  func(*options) error
	options struct {
		metricsListenAddr   string
		checkInterval       *time.Ticker
		checkersParallelism int
		samplersParallelism int
		checkers            []check.Checker
		samplers            []sample.Sampler
	}
)

func newOptions(o ...Option) (*options, error) {
	opts := options{
		metricsListenAddr:   "0.0.0.0:40080",
		checkInterval:       time.NewTicker(5 * time.Minute),
		checkersParallelism: 10,
		samplersParallelism: 10,
	}
	for _, apply := range o {
		if err := apply(&opts); err != nil {
			return nil, err
		}
	}

	return &opts, nil
}

func WithMetricsListenAddr(a string) Option {
	return func(o *options) error {
		o.metricsListenAddr = a
		return nil
	}
}

func WithCheckers(c ...check.Checker) Option {
	return func(o *options) error {
		o.checkers = c
		return nil
	}
}

func WithSamplers(s ...sample.Sampler) Option {
	return func(o *options) error {
		o.samplers = s
		return nil
	}
}

func WithCheckInterval(i time.Duration) Option {
	return func(o *options) error {
		o.checkInterval = time.NewTicker(i)
		return nil
	}
}

func WithCheckersParallelism(p int) Option {
	return func(o *options) error {
		o.checkersParallelism = p
		return nil
	}
}

func WithSamplersParallelism(p int) Option {
	return func(o *options) error {
		o.samplersParallelism = p
		return nil
	}
}
