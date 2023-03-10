package sample

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ipfs/go-cid"
)

var _ Sampler = (*SaturnTopCidsSampler)(nil)

const saturnTopCids = "https://orchestrator.strn.pl/top-cids"

type SaturnTopCidsSampler struct {
	*options
}

func NewSaturnTopCidsSampler(o ...Option) (*SaturnTopCidsSampler, error) {
	opts, err := newOptions(o...)
	if err != nil {
		return nil, err
	}
	return &SaturnTopCidsSampler{options: opts}, nil
}

func (s *SaturnTopCidsSampler) Sample(ctx context.Context) (*Set, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, saturnTopCids, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var scids []string
	if err := json.NewDecoder(resp.Body).Decode(&scids); err != nil {
		return nil, err
	}
	mhs := newMultihashSet()
	for _, sc := range scids {
		cc := strings.SplitN(sc, "/", 2)
		if len(cc) > 0 {
			c, err := cid.Decode(cc[0])
			if err != nil {
				logger.Warnw("Invalid CID from saturn orchestrator", "cid", cc[0], "originalValue", sc, "err", err)
				continue
			}
			mhs.putIfAbsent(c.Hash())
		}
	}
	if mhs.len() == 0 {
		logger.Warn("No CIDs were found from saturn orchestrator")
	}
	return &Set{
		Multihashes: mhs.slice(),
		Name:        s.name,
	}, nil
}
