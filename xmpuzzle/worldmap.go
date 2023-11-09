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
	value    int
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

func (wm Worldmap) Value(idx int) int {
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

func (wm Worldmap) Rotate(rot int) {
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
		twm[key].position = wm[key].position
		twm[key].value = wm[key].value
	}
	return twm
}

func (wm Worldmap) CalcBoundingbox() (bb Boundingbox) {
	bb.max[0] = wm[0].position[0]
	bb.max[1] = wm[0].position[1]
	bb.max[2] = wm[0].position[2]
	bb.min[0] = wm[0].position[0]
	bb.min[1] = wm[0].position[1]
	bb.min[2] = wm[0].position[2]
	for idx := range wm {
		bb.min[0] = min(wm[idx].position[0], bb.min[0])
		bb.min[1] = min(wm[idx].position[1], bb.min[1])
		bb.min[2] = min(wm[idx].position[2], bb.min[2])
		bb.max[0] = max(wm[idx].position[0], bb.max[0])
		bb.max[1] = max(wm[idx].position[1], bb.max[1])
		bb.max[2] = max(wm[idx].position[2], bb.max[2])
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

// This is not a method, just a function taking 2 params
// https://arxiv.org/pdf/cs/0011047v1.pdf
// Multiple golang implementations exist.
// Most are exact implementations, quite complex to use without good knowledge of the algorithm
// github.com/Kappeh/dlx seems to be a dummy proof version, let's give it a spin
//
// Matrix is a type, when you initialize with New you specify the number of primary and optional columns
// AddRow then passes the indices of the positions with a 1, and returns its index in the Matrix
//    - I do not like this signature of the function. Souds easier to pass an array with the indices
//	  - Fortunately, go supports the "explode..." operation that converts an array into the set of params
// There is no way to annotate rows when you add them to the matrix, but we can track the annotations in a separate array
//    - Keep track of the rowId returned by AddRow and use an array Annotation[roxID]=structOfAnnotations
//
// The challenge with a map in Golang is that the sequence of iteration is unpredictable

func GetDLXmap(resmap, piecemap Worldmap) (result []int) {
	// Baseline the resmap by creating 2 arrays:
	// one for the filled pixels, and one for the vari pixels
	var filledHashSequence, variHashSequence [][3]int
	for key := range resmap {
		if resmap[key].value == 1 {
			filledHashSequence = append(filledHashSequence, resmap[key].position)
		} else {
			variHashSequence = append(variHashSequence, resmap[key].position)
		}
	}
	// create a map of  -> arrayindex for performance
	filledLen := len(filledHashSequence)
	lookupMap := make(map[[3]int]int)
	for idx, pos := range filledHashSequence {
		lookupMap[pos] = idx
	}
	for idx, pos := range variHashSequence {
		lookupMap[pos] = idx + filledLen
	}
	// filled and vari now contain the positions of the filled and variable pixels of the puzzle
	// We do this now for every call, we can speed things up if we do this once when we create the
	// full DLX matrix at time of "solve"
	for key := range piecemap {
		// check if we can place the pixels of piecmap into resmap
		if !resmap.Has(piecemap[key].position) {
			// if we can not place a pixel, bail out and return nil (no DLXmap to create)
			return nil
		}
		// The DLX algorithm in go is different, we just need to pass the positions of the "1"s
		result = append(result, lookupMap[piecemap[key].position])
	}
	return
}
