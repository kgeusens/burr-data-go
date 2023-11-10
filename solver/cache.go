package solver

import (
	//	"fmt"

	xmpuzzle "github.com/kgeusens/go/burr-data/xmpuzzle"
)

type maxVal_t [3]int16

/*
maxValMatrix is a two dimensional (idSize X idSize) array
values are arrays with 3 positions ([3]int16), one value per axis
*/
//type maxValMatrix_t []maxVal_t

type SolverCache_t struct {
	puzzle         *xmpuzzle.Puzzle
	problemIndex   uint
	idSize         uint
	shapemap       []uint8
	resultVoxel    *xmpuzzle.Voxel
	resultInstance *VoxelInstance
	instanceCache  map[uint]*VoxelInstance
	movementCache  map[uint64]*maxVal_t
}

/*
Needed to calculate hashes
*/
const worldOrigin uint64 = 100
const worldMax uint64 = 2*worldOrigin + 1
const worldOriginIndex uint64 = worldOrigin * (worldMax*worldMax + worldMax + 1)
const worldSize uint64 = worldMax * worldMax * worldMax

/*
 */
func NewSolverCache(puzzle *xmpuzzle.Puzzle, problemIdx uint) (sc SolverCache_t) {
	psc := new(SolverCache_t)
	sc = *psc
	sc.puzzle = puzzle
	sc.problemIndex = problemIdx
	sc.shapemap = sc.GetProblem().GetShapemap()
	sc.idSize = uint(len(sc.shapemap))
	sc.resultVoxel = &puzzle.Shapes[sc.GetProblem().Result.Id]
	resi := NewVoxelinstance(sc.resultVoxel, 0)
	sc.resultInstance = &resi
	sc.instanceCache = make(map[uint]*VoxelInstance)
	sc.movementCache = make(map[uint64]*maxVal_t)
	return
}

func (sc SolverCache_t) GetProblem() (pb *xmpuzzle.Problem) {
	return &sc.puzzle.Problems[sc.problemIndex]
}

func (sc SolverCache_t) GetShapeInstance(id, rot uint) (vi *VoxelInstance) {
	// hash is based on 24 max rotations
	hash := id*24 + rot
	vi = sc.instanceCache[hash]
	if vi == nil {
		instance := NewVoxelinstance(&sc.puzzle.Shapes[sc.shapemap[id]], rot)
		vi = &instance
		sc.instanceCache[hash] = vi
	}
	return
}

func (sc SolverCache_t) GetResultInstance() (vi *VoxelInstance) {
	return sc.resultInstance
}

/*
Calculate a unique uint64 hashvalue for movements
*/
func (sc SolverCache_t) CalcMovementHash(id1, rot1, id2, rot2 uint, dx, dy, dz int) (hash uint64) {
	bigid1 := uint64(id1)
	bigrot1 := uint64(rot1)
	bigid2 := uint64(id2)
	bigrot2 := uint64(rot2)
	hash = (((bigid1*24+bigrot1)*uint64(sc.idSize)+bigid2)*24+bigrot2)*worldSize + uint64(int(worldOriginIndex)+int(worldMax)*(dz*int(worldMax)+dy)+dx)
	return
}

func (sc SolverCache_t) GetMaxValues(id1, rot1, id2, rot2 uint, dx, dy, dz int) (pmoves *maxVal_t) {
	hash := sc.CalcMovementHash(id1, rot1, id2, rot2, dx, dy, dz)
	pmoves = sc.movementCache[hash]
	if pmoves == nil {
		pmoves = new(maxVal_t)
		sc.movementCache[hash] = pmoves
		// now start calculating
		s1 := sc.GetShapeInstance(id1, rot1)
		s2 := sc.GetShapeInstance(id2, rot2)
		intersection := xmpuzzle.NewBoundingbox()
		union := xmpuzzle.NewBoundingbox()
		bb1 := s1.GetBoundingbox()
		bb2 := s2.GetBoundingbox()
		s1wm := s1.GetWorldmap()
		s2wm := s2.GetWorldmap()
		mx := int16(32000)
		my := int16(32000)
		mz := int16(32000)
		imin := &intersection.Min
		imax := &intersection.Max
		umin := &union.Min
		umax := &union.Max
		imin[0] = max(bb1.Min[0], bb2.Min[0]+dx)
		imin[1] = max(bb1.Min[1], bb2.Min[1]+dy)
		imin[2] = max(bb1.Min[2], bb2.Min[2]+dz)
		imax[0] = min(bb1.Max[0], bb2.Max[0]+dx)
		imax[1] = min(bb1.Max[1], bb2.Max[1]+dy)
		imax[2] = min(bb1.Max[2], bb2.Max[2]+dz)
		umin[0] = min(bb1.Min[0], bb2.Min[0]+dx)
		umin[1] = min(bb1.Min[1], bb2.Min[1]+dy)
		umin[2] = min(bb1.Min[2], bb2.Min[2]+dz)
		umax[0] = max(bb1.Max[0], bb2.Max[0]+dx)
		umax[1] = max(bb1.Max[1], bb2.Max[1]+dy)
		umax[2] = max(bb1.Max[2], bb2.Max[2]+dz)
		var gap int16
		yStart := imin[1]
		yStop := imax[1]
		zStart := imin[2]
		zStop := imax[2]
		xStart := umin[0]
		xStop := umax[0]
		for y := yStart; y <= yStop; y++ {
			for z := zStart; z <= zStop; z++ {
				gap = 32000
				for x := xStart; x <= xStop; x++ {
					if s1wm.Has([3]int{x, y, z}) {
						gap = 0
					} else if s2wm.Has([3]int{x - dx, y - dy, z - dz}) {
						if gap < mx {
							mx = gap
						}
					} else { // s1 is empty, s2 is empty
						gap++
					}
				}
			}
		}
		xStart = imin[0]
		xStop = imax[0]
		zStart = imin[2]
		zStop = imax[2]
		yStart = umin[1]
		yStop = umax[1]
		for x := xStart; x <= xStop; x++ {
			for z := zStart; z <= zStop; z++ {
				gap = 32000
				for y := yStart; y <= yStop; y++ {
					if s1wm.Has([3]int{x, y, z}) {
						gap = 0
					} else if s2wm.Has([3]int{x - dx, y - dy, z - dz}) {
						if gap < my {
							my = gap
						}
					} else { // s1 is empty, s2 is empty
						gap++
					}
				}
			}
		}
		xStart = imin[0]
		xStop = imax[0]
		yStart = imin[1]
		yStop = imax[1]
		zStart = umin[2]
		zStop = umax[2]
		for x := xStart; x <= xStop; x++ {
			for y := yStart; y <= yStop; y++ {
				gap = 32000
				for z := zStart; z <= zStop; z++ {
					if s1wm.Has([3]int{x, y, z}) {
						gap = 0
					} else if s2wm.Has([3]int{x - dx, y - dy, z - dz}) {
						if gap < mz {
							mz = gap
						}
					} else { // s1 is empty, s2 is empty
						gap++
					}
				}
			}
		}
		pmoves[0] = mx
		pmoves[1] = my
		pmoves[2] = mz
	}
	return
}
