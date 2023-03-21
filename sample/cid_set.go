package sample

import (
	"hash/crc32"

	"github.com/ipfs/go-cid"
)

type cidSet struct {
	hashset map[uint32]struct{}
	cids    []cid.Cid
}

func newCidSet() *cidSet {
	return &cidSet{
		hashset: make(map[uint32]struct{}),
	}
}

func (cs *cidSet) putIfAbsent(c cid.Cid) bool {
	key := crc32.ChecksumIEEE(c.Bytes())
	_, seen := cs.hashset[key]
	if !seen {
		cs.hashset[key] = struct{}{}
		cs.cids = append(cs.cids, c)
	}
	return !seen
}

func (cs *cidSet) len() int {
	return len(cs.cids)
}

func (cs *cidSet) slice() []cid.Cid {
	return cs.cids
}
