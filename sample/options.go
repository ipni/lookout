package sample

import "errors"

type (
	Option  func(*options) error
	options struct {
		name string
	}
)

func newOptions(o ...Option) (*options, error) {
	var opts options
	for _, apply := range o {
		if err := apply(&opts); err != nil {
			return nil, err
		}
	}
	if opts.name == "" {
		return nil, errors.New("sample name must be specified")
	}
	return &opts, nil
}

func WithName(name string) Option {
	return func(o *options) error {
		o.name = name
		return nil
	}
}
