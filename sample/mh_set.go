package sample

import (
	"hash/crc32"

	"github.com/multiformats/go-multihash"
)

type multihashSet struct {
	hashset map[uint32]struct{}
	mhs     []multihash.Multihash
}

func newMultihashSet() *multihashSet {
	return &multihashSet{
		hashset: make(map[uint32]struct{}),
	}
}

func (ms *multihashSet) putIfAbsent(mh multihash.Multihash) bool {
	key := crc32.ChecksumIEEE(mh)
	_, seen := ms.hashset[key]
	if !seen {
		ms.hashset[key] = struct{}{}
		ms.mhs = append(ms.mhs, mh)
	}
	return !seen
}

func (ms *multihashSet) len() int {
	return len(ms.mhs)
}

func (ms *multihashSet) slice() []multihash.Multihash {
	return ms.mhs
}
