package check

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type (
	Option  func(*options) error
	options struct {
		name           string
		httpClient     *http.Client
		checkTimeout   time.Duration
		ipniEndpoint   *url.URL
		parallelism    int
		ipfsDhtCascade bool
	}
)

func newOptions(o ...Option) (*options, error) {
	opts := options{
		httpClient:   http.DefaultClient,
		parallelism:  10,
		checkTimeout: 30 * time.Second,
	}
	for _, apply := range o {
		if err := apply(&opts); err != nil {
			return nil, err
		}
	}
	var err error
	if opts.ipniEndpoint == nil {
		opts.ipniEndpoint, err = url.Parse("https://cid.contact")
		if err != nil {
			return nil, err
		}
	}
	if opts.name == "" {
		opts.name = opts.ipniEndpoint.Host
	}
	return &opts, nil
}

func WithName(name string) Option {
	return func(o *options) error {
		o.name = name
		return nil
	}
}

func WithHttpClient(httpClient *http.Client) Option {
	return func(o *options) error {
		o.httpClient = httpClient
		return nil
	}
}

func WithCheckTimeout(checkTimeout time.Duration) Option {
	return func(o *options) error {
		o.checkTimeout = checkTimeout
		return nil
	}
}

func WithIpniEndpoint(endpoint string) Option {
	return func(o *options) error {
		var err error
		o.ipniEndpoint, err = url.Parse(endpoint)
		if err != nil {
			return err
		}
		return nil
	}
}

func WithParallelism(parallelism int) Option {
	return func(o *options) error {
		if parallelism < 1 {
			return fmt.Errorf("parallelism cannot be less than 1; got %d", parallelism)
		}
		o.parallelism = parallelism
		return nil
	}
}

func WithIpfsDhtCascade(ipfsDhtCascade bool) Option {
	return func(o *options) error {
		o.ipfsDhtCascade = ipfsDhtCascade
		return nil
	}
}
