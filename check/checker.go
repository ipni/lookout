package check

import (
	"context"
	"time"

	"github.com/ipfs/go-log/v2"
	"github.com/ipni/lookout/sample"
	"github.com/multiformats/go-multihash"
)

var logger = log.Logger("ipni/lookout/check")

type (
	Checker interface {
		Check(context.Context, *sample.Set) *Results
	}
	Results struct {
		Results       []*Result
		SampleSetName string
		CheckerName   string
	}
	Result struct {
		Multihash  multihash.Multihash
		Err        error
		StatusCode int
		Timeout    time.Duration
		Elapsed    time.Duration
		Streaming  bool
	}
)
