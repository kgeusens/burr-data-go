package solver

import (
	//	"fmt"

	xmpuzzle "github.com/kgeusens/go/burr-data/xmpuzzle"
)

type SolverCache_t struct {
	puzzle        *xmpuzzle.Puzzle
	problemIndex  int
	idSize        int
	shapemap      []int
	resultVoxel   *xmpuzzle.Voxel
	instanceCache map[int]*VoxelInstance
	// movementCache map[uint64]matrix
}

/*
Needed to calculate hashes
*/
const worldOrigin int = 100
const worldMax int = (2*worldOrigin + 1) * (2*worldOrigin + 1) * (2*worldOrigin + 1)

/*
 */
func NewSolverCache(puzzle *xmpuzzle.Puzzle, problemIdx int) (sc SolverCache_t) {
	psc := new(SolverCache_t)
	sc = *psc
	sc.puzzle = puzzle
	sc.problemIndex = problemIdx
	sc.shapemap = sc.GetProblem().GetShapemap()
	sc.idSize = len(sc.shapemap)
	sc.resultVoxel = &puzzle.Shapes[sc.GetProblem().Result.Id]
	sc.instanceCache = make(map[int]*VoxelInstance)
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
