package sample

import (
	"context"
	"encoding/csv"
	"io"
	"net/http"

	"github.com/ipfs/go-cid"
)

var _ Sampler = (*SaturnTopCidsSampler)(nil)

const archiveTopCids = `https://archive.org/advancedsearch.php?q=_exists_:identifier-cid&rows=*&output=csv&fl=identifier-cid&rows=300&sort=downloads:desc`

type InternetArchiveTopCidsSampler struct {
	*options
}

func NewInternetArchiveTopCidsSampler(o ...Option) (*InternetArchiveTopCidsSampler, error) {
	opts, err := newOptions(o...)
	if err != nil {
		return nil, err
	}
	return &InternetArchiveTopCidsSampler{options: opts}, nil
}

func (s *InternetArchiveTopCidsSampler) Sample(ctx context.Context) (*Set, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, archiveTopCids, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	r := csv.NewReader(resp.Body)
	r.FieldsPerRecord = 1
	cids := newCidSet()

RowsLoop:
	for {
		row, err := r.Read()
		switch {
		case err == io.EOF:
			break RowsLoop
		case err == csv.ErrFieldCount:
			logger.Warnw("Inconsistent field count in Internet Archive records; skipping", "record", row)
			continue
		case err != nil:
			logger.Errorw("Unexpected error while sampling Internet Archive records", "err", err)
		default:
			if len(row) > 0 {
				v := row[0]
				c, err := cid.Decode(v)
				if err != nil {
					logger.Warnw("Invalid CID from Internet Archive", "value", v, "err", err)
					continue
				}
				cids.putIfAbsent(c)
			}
		}
	}
	if cids.len() == 0 {
		logger.Warn("No CIDs were found from Internet Archive")
	}
	return &Set{
		Cids: cids.slice(),
		Name: s.name,
	}, nil
}
