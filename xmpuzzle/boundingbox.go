package xmpuzzle

type Boundingbox struct {
	Min [3]int
	Max [3]int
}

func (b Boundingbox) Size() (x, y, z int) {
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
