package xmpuzzle

const Author = "Koen Geusens"
const worldOrigin = 100
const worldMax = 2*worldOrigin + 1
const worldOriginIndex = worldOrigin * (worldMax*worldMax + worldMax + 1)
const worldSize = worldMax * worldMax * worldMax

var worldSteps = [3]int{1, worldMax, worldMax * worldMax}

type Worldmap map[int]int

func HashToPoint(hash int) (x, y, z int) {
	var h int
	x = h % worldMax
	h = (h - x) / worldMax
	y = h % worldMax
	h = (h - y) / worldMax
	z = h
	return
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
