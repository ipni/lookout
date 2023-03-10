package sample

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/ipfs/go-cid"
)

var (
	_ Sampler = (*AwesomeIpfsDatasets)(nil)

	cidHrefMatcher = regexp.MustCompile(`href="?https://ipfs.io/ipfs/(\w+)"?`)
)

const awesomeIpfsDatasets = "https://awesome.ipfs.io/datasets/"

type AwesomeIpfsDatasets struct {
}

func (s *AwesomeIpfsDatasets) Sample(ctx context.Context) (*Set, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, awesomeIpfsDatasets, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unsuccessful response from %s: %d", awesomeIpfsDatasets, resp.StatusCode)
	}
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	matches := cidHrefMatcher.FindAllSubmatch(all, -1)
	mhs := newMultihashSet()
	for _, match := range matches {
		if len(match) > 1 {
			cidMatch := string(match[1])
			c, err := cid.Decode(cidMatch)
			if err != nil {
				logger.Warnw("Failed to decode match as CID", "match", cidMatch, "err", err)
				continue
			}
			mhs.putIfAbsent(c.Hash())
		}
	}
	if mhs.len() == 0 {
		logger.Warn("No CIDs were found from IPFS Awesome Datasets")
	}
	return &Set{
		Multihashes: mhs.slice(),
		Name:        "awesome.ipfs.io/datasets",
	}, nil
}
