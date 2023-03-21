package sample

import (
	"context"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-log/v2"
)

var logger = log.Logger("ipni/lookout/sample")

type (
	Sampler interface {
		Sample(context.Context) (*Set, error)
	}
	Set struct {
		Name string
		Cids []cid.Cid
	}
)
