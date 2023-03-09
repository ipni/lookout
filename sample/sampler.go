package sample

import (
	"context"

	"github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multihash"
)

var logger = log.Logger("ipni/lookout/sample")

type (
	Sampler interface {
		Sample(context.Context) (*Set, error)
	}
	Set struct {
		Name        string
		Multihashes []multihash.Multihash
	}
)
