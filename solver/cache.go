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
type maxValMatrix_t []maxVal_t

type SolverCache_t struct {
	puzzle        *xmpuzzle.Puzzle
	problemIndex  uint
	idSize        uint
	shapemap      []uint8
	resultVoxel   *xmpuzzle.Voxel
	instanceCache map[int]*VoxelInstance
	movementCache map[uint64]*maxValMatrix_t
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
	sc.instanceCache = make(map[int]*VoxelInstance)
	sc.movementCache = make(map[uint64]*maxValMatrix_t)
	return
}

func (sc SolverCache_t) GetProblem() (pb *xmpuzzle.Problem) {
	return &sc.puzzle.Problems[sc.problemIndex]
}

func (sc SolverCache_t) GetShapeInstance(id, rot int) (vi *VoxelInstance) {
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
