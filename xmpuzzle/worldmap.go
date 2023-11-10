package xmpuzzle

// Worldmap is used to work on an instance of a voxel.
// It transforms (x,y,z) coordinates into a numeric hash to store information in a map
// Helper functions:
//   HashToPoint
//   PointToHash
// You can rotate and translate a Worldmap
// You can compare Worldmaps (needed to create the DLXmap)
// You can clone a Worldmap
// You can "instantiate" a Voxel into a Worldmap
// I started with the idea that you need to use Worldmap to track State
// but that is not the case. You need to track state on the Voxel, not the instance.

import (
	//	"slices"

	burrutils "github.com/kgeusens/go/burr-data/burrutils"
)

//const worldSize = worldMax * worldMax * worldMax

type worldmapEntry struct {
	position [3]int
	value    int8
}

// type Worldmap map[int]int
type Worldmap []worldmapEntry

/*
func HashToPoint(hash int) (x, y, z int) {
	var h int
	h = hash
	x = h % worldMax
	h = (h - x) / worldMax
	y = h % worldMax
	h = (h - y) / worldMax
	z = h
	return x - worldOrigin, y - worldOrigin, z - worldOrigin
}

func PointToHash(x, y, z int) (hash int) {
	hash = worldOriginIndex + worldMax*(z*worldMax+y) + x
	return
}
*/

func (wm Worldmap) Value(idx int) int8 {
	return wm[idx].value
}

func (wm Worldmap) Position(idx int) [3]int {
	return wm[idx].position
}

func (wm Worldmap) Has(p [3]int) (ok bool) {
	for key := range wm {
		if wm[key].position == p {
			return true
		}
	}
	return false
}

func (wm Worldmap) Translate(x, y, z int) {
	for key := range wm {
		wm[key].position[0] += x
		wm[key].position[1] += y
		wm[key].position[2] += z
	}
}

func (wm Worldmap) Rotate(rot uint) {
	for key := range wm {
		rx, ry, rz := burrutils.Rotate(wm[key].position[0], wm[key].position[1], wm[key].position[2], rot)
		wm[key].position[0] = rx
		wm[key].position[1] = ry
		wm[key].position[2] = rz
	}
}

func (wm Worldmap) Clone() Worldmap {
	twm := NewWorldmap()
	for key := range wm {
		twm = append(twm, worldmapEntry{[3]int{wm[key].position[0], wm[key].position[1], wm[key].position[2]}, wm[key].value})
	}
	return twm
}

func (wm Worldmap) CalcBoundingbox() (bb Boundingbox) {
	bb.Max[0] = wm[0].position[0]
	bb.Max[1] = wm[0].position[1]
	bb.Max[2] = wm[0].position[2]
	bb.Min[0] = wm[0].position[0]
	bb.Min[1] = wm[0].position[1]
	bb.Min[2] = wm[0].position[2]
	for idx := range wm {
		bb.Min[0] = min(wm[idx].position[0], bb.Min[0])
		bb.Min[1] = min(wm[idx].position[1], bb.Min[1])
		bb.Min[2] = min(wm[idx].position[2], bb.Min[2])
		bb.Max[0] = max(wm[idx].position[0], bb.Max[0])
		bb.Max[1] = max(wm[idx].position[1], bb.Max[1])
		bb.Max[2] = max(wm[idx].position[2], bb.Max[2])
	}
	return
}

/*
func NewWorldmapFromVoxel(v *Voxel) Worldmap {
	wm := NewWorldmap()
	for z := 0; z < v.Z; z++ {
		for y := 0; y < v.Y; y++ {
			for x := 0; x < v.X; x++ {
				if s := v.GetVoxelState(x, y, z); s > 0 {
					wm = append(wm, worldmapEntry{[3]int{x, y, z}, s})
				}
			}
		}
	}
	return wm
}
*/

func NewWorldmap() Worldmap {
	pwm := new(Worldmap)
	return *pwm
}
