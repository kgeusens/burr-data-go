package solver

import (
	"fmt"

	swiss "github.com/dolthub/swiss"
	burrutils "github.com/kgeusens/go/burr-data/burrutils"
	xmpuzzle "github.com/kgeusens/go/burr-data/xmpuzzle"
)

type maxVal_t [3]burrutils.Distance_t

const maxDistance = burrutils.Distance_t(10000)

/*
ProblemCache_t

Dynamic cache for information that is frequently needed during the
assembly and solution of a problem

Informations is either calculated and cached at time of creation,
or dynamically at time of consultation (and then cached for future).
*/
type ProblemCache_t struct {
	// these are unique per problem
	puzzle         *xmpuzzle.Puzzle
	problemIndex   uint
	idSize         int
	numPrimary     int
	numSecondary   int
	shapemap       []burrutils.Id_t
	resultVoxel    *xmpuzzle.Voxel
	resultInstance *VoxelInstance
	instanceCache  map[uint]*VoxelInstance
	//	movementCache  map[uint64]*maxVal_t // cache for getMaxValue
	movementCache  swiss.Map[uint64, *maxVal_t]
	dlxMatrixCache *matrix_t        // used by the DLX algorithm in assemble phase, contains the full DLX matrix
	assemblyCache  []assembly_t     // result of the assemble phase
	dlxLookupmap   map[maxVal_t]int // used to calculate a row in the DLX matrix. Static throughout the cache lifecycle
}

type SolverCache_t struct {
	// These are used every time we analyse a potential move in the solver
	// If we want to operate in parallel, it needs to move to a cache that is unique per Assembly we are solving
	pc           *ProblemCache_t                                 // pointer to the parent ProblemCache. Can not be nil.
	cutlerMatrix [3 * maxShapes * maxShapes]burrutils.Distance_t // contains the cutlermatrix for a state in the tree
	movesList    []*node_t                                       // used in getMovementList to calculate moveable pieces from a state in the tree
	nodecache    NodeCache_t
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
func NewSolverCache(pc *ProblemCache_t) (sc SolverCache_t) {
	psc := new(SolverCache_t)
	sc = *psc
	sc.pc = pc
	sc.movesList = make([]*node_t, 0, 3*maxShapes)
	sc.nodecache = NodeCache_t{make([]*node_t, 0)}
	return
}

/*
 */
func NewProblemCache(puzzle *xmpuzzle.Puzzle, problemIdx uint) (pc ProblemCache_t) {
	ppc := new(ProblemCache_t)
	pc = *ppc
	pc.puzzle = puzzle
	pc.problemIndex = problemIdx
	pc.shapemap = pc.GetProblem().GetShapemap()
	pc.idSize = len(pc.shapemap)
	pc.resultVoxel = &puzzle.Shapes[pc.GetProblem().Result.Id]
	resi := NewVoxelinstance(pc.resultVoxel, 0)
	pc.resultInstance = &resi
	pc.instanceCache = make(map[uint]*VoxelInstance)
	pc.movementCache = *swiss.NewMap[uint64, *maxVal_t](0)
	//	pc.movementCache = make(map[uint64]*maxVal_t)
	//	pc.movesList = make([]*node_t, 3*maxShapes)

	resmap := *(resi.GetWorldmap())
	// Baseline the resmap by creating 2 arrays:
	// one for the filled pixels, and one for the vari pixels
	var filledHashSequence, variHashSequence []maxVal_t
	for key := range resmap {
		if resmap.Value(key) == 1 {
			filledHashSequence = append(filledHashSequence, resmap.Position(key))
		} else {
			variHashSequence = append(variHashSequence, resmap.Position(key))
		}
	}
	// create a lookupmap for performance
	filledLen := len(filledHashSequence)
	lookupMap := make(map[maxVal_t]int)
	for idx, pos := range filledHashSequence {
		lookupMap[pos] = idx
	}
	for idx, pos := range variHashSequence {
		lookupMap[pos] = idx + filledLen
	}
	//Now cache
	pc.numPrimary = filledLen
	pc.numSecondary = len(variHashSequence)
	pc.dlxLookupmap = lookupMap

	return
}

func (pc ProblemCache_t) GetProblem() (pb *xmpuzzle.Problem) {
	return &pc.puzzle.Problems[pc.problemIndex]
}

func (pc ProblemCache_t) GetNumPrimary() int {
	return pc.numPrimary
}

func (pc ProblemCache_t) GetNumSecondary() int {
	return pc.numSecondary
}

func (pc *ProblemCache_t) GetShapeInstance(id, rot burrutils.Id_t) (vi *VoxelInstance) {
	// hash is based on 24 max rotations
	hash := uint(id)*24 + uint(rot)
	vi = pc.instanceCache[hash]
	if vi == nil {
		instance := NewVoxelinstance(&pc.puzzle.Shapes[pc.shapemap[id]], rot)
		vi = &instance
		pc.instanceCache[hash] = vi
	}
	return
}

func (pc ProblemCache_t) GetResultInstance() (vi *VoxelInstance) {
	return pc.resultInstance
}

/*
Calculate a unique uint64 hashvalue for movements
*/
func (pc *ProblemCache_t) getMaxValues(id1, rot1, id2, rot2 burrutils.Id_t, dx, dy, dz burrutils.Distance_t) (mx, my, mz burrutils.Distance_t) {
	hash := (((uint64(id1)*24+uint64(rot1))*uint64(pc.idSize)+uint64(id2))*24+uint64(rot2))*worldSize + uint64(int(worldOriginIndex)+int(worldMax)*(int(dz)*int(worldMax)+int(dy))+int(dx))
	pmoves, ok := pc.movementCache.Get(hash)
	if !ok {
		pmoves = new(maxVal_t)
		pc.movementCache.Put(hash, pmoves)
		// now start calculating
		s1 := pc.GetShapeInstance(id1, rot1)
		s2 := pc.GetShapeInstance(id2, rot2)
		intersection := xmpuzzle.NewBoundingbox()
		union := xmpuzzle.NewBoundingbox()
		bb1 := s1.GetBoundingbox()
		bb2 := s2.GetBoundingbox()
		s1wm := s1.GetWorldmap()
		s2wm := s2.GetWorldmap()
		mx = burrutils.Distance_t(maxDistance)
		my = burrutils.Distance_t(maxDistance)
		mz = burrutils.Distance_t(maxDistance)
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
		var gap burrutils.Distance_t
		yStart := imin[1]
		yStop := imax[1]
		zStart := imin[2]
		zStop := imax[2]
		xStart := umin[0]
		xStop := umax[0]
		for y := yStart; y <= yStop; y++ {
			for z := zStart; z <= zStop; z++ {
				gap = maxDistance
				for x := xStart; x <= xStop; x++ {
					if s1wm.Has([3]burrutils.Distance_t{x, y, z}) {
						gap = 0
					} else if s2wm.Has([3]burrutils.Distance_t{x - dx, y - dy, z - dz}) {
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
				gap = maxDistance
				for y := yStart; y <= yStop; y++ {
					if s1wm.Has([3]burrutils.Distance_t{x, y, z}) {
						gap = 0
					} else if s2wm.Has([3]burrutils.Distance_t{x - dx, y - dy, z - dz}) {
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
				gap = maxDistance
				for z := zStart; z <= zStop; z++ {
					if s1wm.Has([3]burrutils.Distance_t{x, y, z}) {
						gap = 0
					} else if s2wm.Has([3]burrutils.Distance_t{x - dx, y - dy, z - dz}) {
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
	return pmoves[0], pmoves[1], pmoves[2]
}

func (sc *SolverCache_t) getMovementList(node *node_t) []*node_t {
	// pRow, pCol can only contain max nPieces, so better preallocate
	// and reuse instead of doing a lot of append calls.
	// movelist is a different beast and lenght is hard to predict.

	nPieces := len(node.root.rootDetails.pieceList)
	nDims := 3 * nPieces
	// KG: storing and reusing matrix from the cache can probably save a lot of GC effort
	//	numRow := nPieces * 3
	//	var o1, o2 int
	//	var i, j, k, dim, a int
	m := &sc.cutlerMatrix
	for j := 0; j < nPieces; j++ {
		for i := 0; i < nPieces; i++ {
			// diagonal is 0
			if i == j {
				m[j*nDims+i*3] = 0
				m[j*nDims+i*3+1] = 0
				m[j*nDims+i*3+2] = 0
			} else {
				o1 := i * 3
				o2 := j * 3
				m[j*nDims+i*3], m[j*nDims+i*3+1], m[j*nDims+i*3+2] = sc.pc.getMaxValues(node.root.rootDetails.pieceList[i], node.root.rootDetails.rotationList[i], node.root.rootDetails.pieceList[j], node.root.rootDetails.rotationList[j], node.offsetList[o2]-node.offsetList[o1], node.offsetList[o2+1]-node.offsetList[o1+1], node.offsetList[o2+2]-node.offsetList[o1+2])
			}
		}
	}
	// Phase 2: algorithm from Bill Cutler
	again := true
	var minval burrutils.Distance_t
	for again {
		again = false
		for j := 0; j < nPieces; j++ {
			for i := 0; i < nPieces; i++ {
				if i == j {
					continue
				}
				ijStart := j*nDims + i*3
				kjStart := j*nDims - 3
				ikStart := i*3 - nDims
				for k := 0; k < nPieces; k++ {
					kjStart += 3
					ikStart += nDims
					if k == j {
						continue
					}
					for dim := 0; dim < 3; dim++ {
						minval = m[ikStart+dim] + m[kjStart+dim]
						if minval < m[ijStart+dim] {
							m[ijStart+dim] = minval
							// optimize: check if this update impacts already updated values
							if !again {
								for a := 0; a < i; a++ {
									if m[j*nDims+a*3+dim] > m[i*nDims+a*3+dim]+minval {
										again = true
										break
									}
								}
							}
							if !again {
								for a := 0; a < j; a++ {
									if m[a*nDims+i*3+dim] > m[a*nDims+j*3+dim]+minval {
										again = true
										break
									}
								}
							}
						}
					}
				}
			}
		}
	}

	//	movelist := sc.movesList
	sc.movesList = sc.movesList[:0]

	nPieces = len(node.root.rootDetails.pieceList)
	pRow := make([]burrutils.Id_t, nPieces)
	pCol := make([]burrutils.Id_t, nPieces)
	var pRowLen, pColLen int
	vMoveRow := maxDistance + 1
	vMoveCol := maxDistance + 1
	var vCol, vRow burrutils.Distance_t

	for dim := 0; dim < 3; dim++ {
		for k := 0; k < nPieces; k++ {
			pRowLen = 0
			pColLen = 0
			vMoveRow = maxDistance + 1
			vMoveCol = maxDistance + 1
			for i := 0; i < nPieces; i++ {
				vCol = sc.cutlerMatrix[i*nPieces*3+k*3+dim]
				vRow = sc.cutlerMatrix[k*nPieces*3+i*3+dim]
				if vRow == 0 {
					pRow[pRowLen] = burrutils.Id_t(i)
					pRowLen++
				} else {
					vMoveRow = min(vRow, vMoveRow, maxDistance)
				}
				if vCol == 0 {
					pCol[pColLen] = burrutils.Id_t(i)
					pColLen++
				} else {
					vMoveCol = min(vCol, vMoveCol, maxDistance)
				}
			}
			offset := maxVal_t{0, 0, 0}
			if vMoveRow <= maxDistance {
				// we have a partition
				if pRowLen <= nPieces/2 {
					// process separation
					if vMoveRow >= maxDistance {
						offset[dim] = -1 * maxDistance
						// We should be returning an array of new nodes
						sc.movesList = append(sc.movesList, sc.nodecache.NewNodeChild(node, pRow[:pRowLen], offset, true))
						return sc.movesList
					}
					for step := burrutils.Distance_t(1); step <= vMoveRow; step++ {
						offset[dim] = -1 * step
						sc.movesList = append(sc.movesList, sc.nodecache.NewNodeChild(node, pRow[:pRowLen], offset, false))
					}
				}
			}
			offsetCol := maxVal_t{0, 0, 0}
			if vMoveCol <= maxDistance {
				// we have a partition
				if pColLen <= nPieces/2 {
					// process separation
					if vMoveCol >= maxDistance {
						offsetCol[dim] = maxDistance
						// We should be returning an array of new nodes
						sc.movesList = append(sc.movesList, sc.nodecache.NewNodeChild(node, pCol[:pColLen], offsetCol, true))
						return sc.movesList
					}
					for step := burrutils.Distance_t(1); step <= vMoveCol; step++ {
						offsetCol[dim] = step
						sc.movesList = append(sc.movesList, sc.nodecache.NewNodeChild(node, pCol[:pColLen], offsetCol, false))
					}
				}
			}
		}
	}
	return sc.movesList
}

func (pc *ProblemCache_t) Solve(assembly assembly_t, asmid int) bool {
	DEBUG := false
	if DEBUG {
		fmt.Println(asmid, " start solving")
	}
	sc := NewSolverCache(pc)
	var startNode *node_t
	// parking is an array.
	// push is the same as parking=append(parking, newnode)
	// pop is the same as parking=parking[:len(parking)-1]
	parking := []*node_t{sc.nodecache.NewNodeFromAssembly(assembly)}
	var node *node_t
	var level int
	closedCache := make(map[id_t]bool)
	// adding an entry to closedCache : closedCache[id]=true
	// checking if entry exists: closedCache[id]
	separated := false
	//	movesList := make([]*node_t, 0, maxShapes)
	var movesListLength int
	for len(parking) > 0 {
		// pop from parking
		if startNode != nil {
			sc.nodecache.release(startNode)
		}
		startNode = parking[len(parking)-1]
		parking = parking[:len(parking)-1]
		curListFront := 0
		newListFront := 1
		openlist := [2][]*node_t{{}, {}}
		separated = false

		closedCache[startNode.GetId()] = true
		openlist[curListFront] = append(openlist[curListFront], startNode)

		level = 1
		curLength := len(openlist[curListFront])
		for !(curLength == 0) && !separated {
			// pop
			curLength -= 1
			node = openlist[curListFront][curLength]
			openlist[curListFront] = openlist[curListFront][:curLength]
			movesList := sc.getMovementList(node)
			if DEBUG {
				fmt.Println(asmid, "node", node.GetId())
			}
			var st *node_t
			movesListLength = len(movesList)
			for movesListLength != 0 && !separated {
				// pop
				movesListLength -= 1
				st = movesList[movesListLength]
				movesList = movesList[:movesListLength]
				if DEBUG {
					fmt.Println(asmid, st.movingPieceList, st.moveDirection, st.isSeparation, st.GetId())
				}
				if closedCache[st.GetId()] {
					sc.nodecache.release(st)
					continue
				}
				// never seen this node before, add it to cache
				closedCache[st.GetId()] = true
				// check for separation
				if !st.isSeparation {
					openlist[newListFront] = append(openlist[newListFront], st)
					continue
				} else {
					// this is a separation, put the sub problems on the parking lot and continue to the next one on the parking
					// Need to record this in the partial solution
					separated = true // FLAG STOP TO GO TO NEXT ON PARKING
					parking = append(parking, sc.nodecache.Separate(st)...)
					// record this separation by walking back up to the root
					st.RecordSeparationInRoot()
					// now cleanup and release all nodes, except for the root, and this node itself (the parent of the separation)
					sc.nodecache.Purge()
					if DEBUG {
						fmt.Println(asmid, "SEPARATION FOUND level", level)
					}
				}
			}
			//
			if len(openlist[curListFront]) == 0 && !separated {
				if DEBUG {
					fmt.Println(asmid, "Next Level", level)
					level++
				}
				curListFront = 1 - curListFront
				newListFront = 1 - newListFront
				curLength = len(openlist[curListFront])
			}
		}
		// if we get here, we can check the separated flag to see if it is a dead end, or a separation
		// if it is a separation, continue to the next on the parking, else return false
		if !separated {
			return false
		}
	}
	// SUCCESS
	if DEBUG {
		fmt.Println(asmid, "Solution found")
	}
	return true
}
