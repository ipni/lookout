package lookout

import (
	"time"

	"github.com/ipni/lookout/check"
	"github.com/ipni/lookout/sample"
)

type (
	Option  func(*options) error
	options struct {
		httpListenAddr      string
		checkInterval       *time.Ticker
		checkersParallelism int
		samplersParallelism int
		checkers            []check.Checker
		samplers            []sample.Sampler
	}
)

func newOptions(o ...Option) (*options, error) {
	opts := options{
		httpListenAddr:      "0.0.0.0:40080",
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

func WithHttpListenAddr(a string) Option {
	return func(o *options) error {
		o.httpListenAddr = a
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
