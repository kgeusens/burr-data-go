package xmpuzzle

type Boundingbox struct {
	min [3]int
	max [3]int
}

func (b Boundingbox) Size() (x, y, z int) {
	x = b.max[0] - b.min[0]
	y = b.max[1] - b.min[1]
	z = b.max[2] - b.min[2]
	return
}
