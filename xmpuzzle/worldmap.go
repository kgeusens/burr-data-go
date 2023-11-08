package xmpuzzle

import (
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

func (wm Worldmap) Set(hash, val int) int {
	wm[hash] = val
	return val
}

func (wm Worldmap) Get(hash, val int) int {
	return wm[hash]
}

func (wm Worldmap) SetState(hash, state int) int {
	if state == 0 {
		delete(wm, hash)
	} else {
		wm[hash] = state
	}
	return state
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

func GetDLXmap(resmap, piecemap Worldmap) (result []int) {
	// The DLX algorith in go is different,
	return
}
