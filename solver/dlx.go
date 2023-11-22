package solver

import (
	"slices"

	burrutils "github.com/kgeusens/go/burr-data/burrutils"
)

type row_t []int

type annotation_t struct {
	partID     burrutils.Id_t
	instanceID burrutils.Id_t
	shapeID    burrutils.Id_t
	rotation   burrutils.Id_t
	hotspot    [3]burrutils.Distance_t
	offset     [3]burrutils.Distance_t
}

type matrixEntry_t struct {
	row        *row_t
	annotation *annotation_t
}

type matrix_t []*matrixEntry_t

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

/*
calcDLXrow returns a single DLX row for a rotated and translated shape.
The shape is identified by its id, rotation, and offset relative to the result
*/
func (sc ProblemCache_t) calcDLXrow(shapeid, rotid burrutils.Id_t, x, y, z burrutils.Distance_t) (result row_t) {
	// Get the worldmap of the resultvoxel
	r := sc.GetResultInstance()
	resmap := *(r.GetWorldmap())
	// Get a clone of the worldmap of the shape and translate it
	piecemap := sc.GetShapeInstance(shapeid, rotid).GetWorldmap().Clone()
	piecemap.Translate(x, y, z)
	lookupMap := sc.dlxLookupmap
	// filled and vari now contain the positions of the filled and variable pixels of the puzzle
	// We do this now for every call, we can speed things up if we do this once when we create the
	// full DLX matrix at time of "solve"
	for key := range piecemap {
		// check if we can place the pixels of piecmap into resmap
		if !resmap.Has(piecemap.Position(key)) {
			// if we can not place a pixel, bail out and return nil (no DLXmap to create)
			return nil
		}
		// The DLX algorithm in go is different, we just need to pass the positions of the "1"s
		result = append(result, lookupMap[piecemap.Position(key)])
	}
	slices.Sort(result)
	return
}

/*
func (sc *ProblemCache_t) calcDLXmatrix() *matrix_t {
	matrix := make(matrix_t, 0)
	// calculate rotaionLists
	rotationLists := make([]int, sc.idSize)
	r := sc.GetResultInstance()
	rbb := r.GetBoundingbox()
	rsymgroupID := r.voxel.CalcSelfSymmetries()
	// Determine symmetry breaker
	breakerID := -1
	voxelSize := uint8(0)
	breakerReduction := 0
	reducedRotlist := 0
	shapeDefs := sc.puzzle.Problems[sc.problemIndex].Shapes
	for idx, shape := range shapeDefs {
		voxel := sc.puzzle.Shapes[shape.Id]
		voxelSize = shape.GetPartMinimum()
		symgroupID := voxel.CalcSelfSymmetries()
		rotlist := burrutils.RotationsToCheck[symgroupID]
		rotationLists[idx] = rotlist // no need to copy, this is just an integer bitmap
		reducedRotlist = burrutils.ReduceRotations(rsymgroupID, rotlist)
		rotlistLength := burrutils.BitmapSize(rotlist)
		reducedRotlistLength := burrutils.BitmapSize(reducedRotlist)

		if (rotlistLength - reducedRotlistLength) >= breakerReduction {
			if (rotlistLength - reducedRotlistLength) == breakerReduction {
				if voxelSize == 1 {
					breakerID = idx
				}
			} else {
				breakerID = idx
				breakerReduction = rotlistLength - reducedRotlistLength
			}
		}
	}

	// caclulate the reduced rotationlist
	if breakerID >= 0 {
		reducedRotlist = burrutils.ReduceRotations(rsymgroupID, rotationLists[breakerID])
		rotationLists[breakerID] = reducedRotlist
	}
	// now start building the DLX matrix, keeping track of duplicate voxels
	// First we need to calculate the rows for every voxel, because the additional
	// constraint columns will depend on the number of rows per voxel.
	// We also need to keep track if the source rotation has been reduced or not
	rowMap := make(map[burrutils.Id_t][]row_t)
	annotationMap := make(map[burrutils.Id_t][]annotation_t)
	var row row_t
	for i := range shapeDefs {
		idx := burrutils.Id_t(i)
		//
		rotlist := burrutils.HashToRotations(rotationLists[idx])
		rowMap[idx] = make([]row_t, 0)
		annotationMap[idx] = make([]annotation_t, 0)
		for _, rotidx := range rotlist {
			rotatedInstance := sc.GetShapeInstance(idx, rotidx)
			pbb := rotatedInstance.GetBoundingbox()
			for x := rbb.Min[0] - pbb.Min[0]; x <= rbb.Max[0]-pbb.Max[0]; x++ {
				for y := rbb.Min[1] - pbb.Min[1]; y <= rbb.Max[1]-pbb.Max[1]; y++ {
					for z := rbb.Min[2] - pbb.Min[2]; z <= rbb.Max[2]-pbb.Max[2]; z++ {
						row = sc.calcDLXrow(idx, rotidx, x, y, z)
						if len(row) > 0 {
							rowMap[idx] = append(rowMap[idx], row)
							annotationMap[idx] = append(annotationMap[idx], annotation_t{burrutils.Id_t(idx), rotidx, rotatedInstance.hotspot, [3]burrutils.Distance_t{x, y, z}})
						}
					}
				}
			}
		}
		// Now we have all the rows for the referenced voxels.
		// The index in rowMap is NOT the voxelid, but the id of the Shapedefinition in the problem.
	}

	// construct the matrix
	for i := range shapeDefs {
		idx := burrutils.Id_t(i)
		nRows := len(rowMap[idx])
		for n := 0; n < nRows; n++ {
			newrow := rowMap[idx][n]
			newannotation := annotationMap[idx][n]
			matrix = append(matrix, &matrixEntry_t{&newrow, &newannotation})
		}

	}

	return &matrix
}
*/

func (sc *ProblemCache_t) calcDLXmatrix() *matrix_t {
	matrix := make(matrix_t, 0)
	// calculate rotaionLists
	rotationLists := make([]int, sc.idSize)
	r := sc.GetResultInstance()
	rbb := r.GetBoundingbox()
	rsymgroupID := r.voxel.CalcSelfSymmetries()
	// Determine symmetry breaker
	breakerID := -1
	breakerSize := 30000
	voxelSize := 0
	breakerReduction := 0
	reducedRotlist := 0
	shapeDefs := sc.puzzle.Problems[sc.problemIndex].Shapes
	// sc.puzzle.Problems[sc.problemIndex].Shapes[x] -> Id, Count, Min, Max, Group
	// sc.puzzle.Shapes[Id] -> voxel corresponding to Id
	for idx, shape := range shapeDefs {
		voxel := sc.puzzle.Shapes[shape.Id]
		voxelSize = int(shape.Count)
		symgroupID := voxel.CalcSelfSymmetries()
		rotlist := burrutils.RotationsToCheck[symgroupID]
		rotationLists[idx] = rotlist // no need to copy, this is just an integer bitmap
		reducedRotlist = burrutils.ReduceRotations(rsymgroupID, rotlist)
		rotlistLength := burrutils.BitmapSize(rotlist)
		reducedRotlistLength := burrutils.BitmapSize(reducedRotlist)

		if (rotlistLength - reducedRotlistLength) >= breakerReduction {
			if (rotlistLength - reducedRotlistLength) == breakerReduction {
				if voxelSize < breakerSize {
					breakerSize = voxelSize
					breakerID = idx
				}
			} else {
				breakerID = idx
				breakerReduction = rotlistLength - reducedRotlistLength
			}
		}
	}

	// caclulate the reduced rotationlist
	if breakerID >= 0 {
		reducedRotlist = burrutils.ReduceRotations(rsymgroupID, rotationLists[breakerID])
	}
	// now start building the DLX matrix, keeping track of duplicate voxels
	// First we need to calculate the rows for every voxel, because the additional
	// constraint columns will depend on the number of rows per voxel.
	// We also need to keep track if the source rotation has been reduced or not
	rowMap := make(map[burrutils.Id_t][]row_t)
	annotationMap := make(map[burrutils.Id_t][]annotation_t)
	breakerIsReduced := make(map[int]bool)
	psid := burrutils.Id_t(0)
	var row row_t
	for i, shape := range shapeDefs {
		idx := burrutils.Id_t(i)
		rotlist := burrutils.HashToRotations(rotationLists[idx])
		rowMap[idx] = make([]row_t, 0)
		annotationMap[idx] = make([]annotation_t, 0)
		for _, rotidx := range rotlist {
			rotatedInstance := sc.GetShapeInstance(psid, rotidx)
			pbb := rotatedInstance.GetBoundingbox()
			for x := rbb.Min[0] - pbb.Min[0]; x <= rbb.Max[0]-pbb.Max[0]; x++ {
				for y := rbb.Min[1] - pbb.Min[1]; y <= rbb.Max[1]-pbb.Max[1]; y++ {
					for z := rbb.Min[2] - pbb.Min[2]; z <= rbb.Max[2]-pbb.Max[2]; z++ {
						row = sc.calcDLXrow(psid, rotidx, x, y, z)
						if len(row) > 0 {
							rowMap[idx] = append(rowMap[idx], row)
							annotationMap[idx] = append(annotationMap[idx], annotation_t{burrutils.Id_t(idx), 0, 0, rotidx, rotatedInstance.hotspot, [3]burrutils.Distance_t{x, y, z}})
							// if this is the symmetry breaker, track a mapping of rownumber -> isReduced
							if (i == breakerID) && (reducedRotlist&(1<<rotidx) > 0) {
								breakerIsReduced[len(rowMap[idx])-1] = true
							}
						}
					}
				}
			}
		}
		// Now we have all the rows for the referenced voxels.
		// The index in rowMap is NOT the voxelid, but the id of the Shapedefinition in the problem.
		psid += burrutils.Id_t(shape.Count)
	}

	// We can now calculate the additional constraints and generate the matrix entries
	curRowsize := sc.GetNumPrimary() + sc.GetNumSecondary()
	psid = 0
	for i, shape := range shapeDefs {
		idx := burrutils.Id_t(i)
		nRows := len(rowMap[idx])
		nCopies := int(shape.Count)
		for c := 0; c < nCopies; c++ {
			var delta int
			for n := 0; n < nRows; n++ {
				delta = 0
				// clone the source row
				newrow := row_t{}
				newrow = append(newrow, rowMap[idx][n]...)
				// create preamble (permutation limiter)
				if c > 0 {
					for j := 0; j < nRows-1-n; j++ {
						newrow = append(newrow, curRowsize+n+j)
					}
					delta = nRows - 1
				}
				// Add the column of 1's (piece identity)
				// For puzzles using a "min" value of occurences, this column should move to the primary columns, not the secondary
				newrow = append(newrow, curRowsize+delta)
				delta += 1
				// Add the [I] matrix (only if there are multiple copies and this is not the last one)
				if c < nCopies-1 {
					newrow = append(newrow, curRowsize+delta+n)
				}
				newannotation := annotationMap[idx][n]
				newannotation.instanceID = burrutils.Id_t(c)
				newannotation.shapeID = psid
				// we now have the row defined, and need to add it to the final matrix
				// KG : If this is the symmetry breaker, we still need to filter on the reduced rotations

				if c > 0 {
					matrix = append(matrix, &matrixEntry_t{&newrow, &newannotation})
				} else {
					if breakerID != i {
						matrix = append(matrix, &matrixEntry_t{&newrow, &newannotation})
					} else {
						if breakerIsReduced[n] {
							matrix = append(matrix, &matrixEntry_t{&newrow, &newannotation})
						} else {
							psid += 0
						}
					}
				}
			}
			psid += 1
			curRowsize += delta
		}
	}

	// KG: quick and dirty patch of the secondary column size in the cache
	sc.numSecondary = curRowsize - sc.GetNumPrimary()

	return &matrix
}

func (sc *ProblemCache_t) getDLXmatrix() *matrix_t {
	if sc.dlxMatrixCache == nil {
		sc.dlxMatrixCache = sc.calcDLXmatrix()
	}
	return sc.dlxMatrixCache
}
