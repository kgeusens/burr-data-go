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
	"slices"

	burrutils "github.com/kgeusens/go/burr-data/burrutils"
)

const worldOrigin = 100
const worldMax = 2*worldOrigin + 1
const worldOriginIndex = worldOrigin * (worldMax*worldMax + worldMax + 1)
const worldSize = worldMax * worldMax * worldMax

var worldSteps = [3]int{1, worldMax, worldMax * worldMax}

type Worldmap map[int]int

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

func (wm Worldmap) Has(hash int) bool {
	_, ok := wm[hash]
	return ok
}

func (wm Worldmap) Translate(x, y, z int) {
	twm := make(Worldmap)
	for key, val := range wm {
		twm[key] = val
	}
	clear(wm)
	var nkey, offset int
	for key, val := range twm {
		offset = worldSteps[0]*x + worldSteps[1]*y + worldSteps[2]*z
		nkey = key + offset
		wm[nkey] = val
	}
}

func (wm Worldmap) Rotate(rot int) {
	twm := make(Worldmap)
	for key, val := range wm {
		twm[key] = val
	}
	clear(wm)
	var nkey int
	for key, val := range twm {
		x, y, z := HashToPoint(key)
		nkey = PointToHash(burrutils.Rotate(x, y, z, rot))
		wm[nkey] = val
	}
}

func (wm Worldmap) Clone() Worldmap {
	twm := make(Worldmap)
	for key, val := range wm {
		twm[key] = val
	}
	return twm
}

func NewWorldmapFromVoxel(v *Voxel) Worldmap {
	wm := make(Worldmap)
	for x := 0; x < v.X; x++ {
		for y := 0; y < v.Y; y++ {
			for z := 0; z < v.Z; z++ {
				if s := v.GetVoxelState(x, y, z); s > 0 {
					wm[PointToHash(x, y, z)] = s
				}
			}
		}
	}
	return wm
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
	// Make sure the arrays are sorted from smallest pixel to largest pixel (based on hash value of pixel)
	var filledHashSequence, variHashSequence []int
	for hash, state := range resmap {
		if state == 1 {
			filledHashSequence = append(filledHashSequence, hash)
		} else {
			variHashSequence = append(variHashSequence, hash)
		}
	}
	slices.Sort(filledHashSequence)
	slices.Sort(variHashSequence)
	// create a map of hash -> arrayindex for performance
	filledLen := len(filledHashSequence)
	//	variLen := len(variHashSequence)
	lookupMap := make(map[int]int)
	for idx, hash := range filledHashSequence {
		lookupMap[hash] = idx
	}
	for idx, hash := range variHashSequence {
		lookupMap[hash] = idx + filledLen
	}
	// filled and vari now contain the hashes of the filled and variable pixels of the puzzle
	// We do this now for every call, we can speed things up if we do this once when we create the
	// full DLX matrix at time of "solve"
	for hash := range piecemap {
		// check if we can place the pixels of piecmap into resmap
		if !resmap.Has(hash) {
			// if we can not place a pixel, bail out and return nil (no DLXmap to create)
			return nil
		}
		// The DLX algorithm in go is different, we just need to pass the positions of the "1"s
		result = append(result, lookupMap[hash])
	}
	return
}
