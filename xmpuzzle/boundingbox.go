package xmpuzzle

import (
	//	"slices"

	burrutils "github.com/kgeusens/go/burr-data/burrutils"
)

type Boundingbox struct {
	Min [3]burrutils.Distance_t
	Max [3]burrutils.Distance_t
}

func (b Boundingbox) Size() (x, y, z burrutils.Distance_t) {
	x = b.Max[0] - b.Min[0]
	y = b.Max[1] - b.Min[1]
	z = b.Max[2] - b.Min[2]
	return
}

func NewBoundingbox() (bb Boundingbox) {
	pbb := new(Boundingbox)
	bb = *pbb
	return
}
