package check

import (
	"context"
	"net/http"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipni/lookout/perform"
	"github.com/ipni/lookout/sample"
)

type (
	IpniNonStreamingChecker struct {
		*options
	}
)

func NewIpniNonStreamingChecker(o ...Option) (*IpniNonStreamingChecker, error) {
	opts, err := newOptions(o...)
	if err != nil {
		return nil, err
	}
	return &IpniNonStreamingChecker{
		options: opts,
	}, nil
}

func (c *IpniNonStreamingChecker) Check(ctx context.Context, set *sample.Set) *Results {
	results := &Results{
		Results:       make([]*Result, 0, len(set.Cids)),
		SampleSetName: set.Name,
		CheckerName:   c.name,
	}
	rch := perform.InParallel(ctx, c.parallelism, set.Cids, func(ctx context.Context, mh cid.Cid) *Result {
		result := &Result{
			Multihash: mh.Hash(),
			Timeout:   c.checkTimeout,
		}
		path := c.ipniEndpoint.JoinPath("cid", mh.String())
		if len(c.cascadeLabels) != 0 {
			query := path.Query()
			for _, label := range c.cascadeLabels {
				query.Add("cascade", label)
			}
			path.RawQuery = query.Encode()
		}
		start := time.Now()
		cctx, cancel := context.WithTimeout(ctx, c.checkTimeout)
		defer cancel()
		request, err := http.NewRequestWithContext(cctx, http.MethodGet, path.String(), nil)
		if err != nil {
			logger.Errorw("Failed to instantiate HTTP request", "err", err)
			result.Err = err
			return result
		}
		request.Header.Add("Accept", "application/json")
		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			logger.Errorw("Failed to perform HTTP call", "err", err)
			result.Err = err
			return result
		}
		result.Elapsed = time.Since(start)
		result.StatusCode = resp.StatusCode
		return result
	})
	for {
		select {
		case <-ctx.Done():
			return results
		case result, ok := <-rch:
			if !ok {
				return results
			}
			results.Results = append(results.Results, result)
		}
	}
}
