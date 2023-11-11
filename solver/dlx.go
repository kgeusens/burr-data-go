package solver

import (
	burrutils "github.com/kgeusens/go/burr-data/burrutils"
)

type row_t []int
type matrix_t []*row_t

type DLXmatrix_t struct {
	Matrix       *matrix_t
	NumPrimary   int
	NumSecondary int
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

/*
GetDLXrow returns a single DLX row for a rotated and translated shape.
The shape is identified by its id, rotation, and offset relative to the result
*/
func (sc SolverCache_t) GetDLXrow(shapeid, rotid uint, x, y, z int) (result row_t) {
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
	// Finally we need to add the optional column for the piece (regardless of rotation)
	// This is at index "(size of resultvoxel) + pieceID"
	result = append(result, resmap.Size()+int(shapeid))
	return
}

func (sc SolverCache_t) GetDLXmatrix() *matrix_t {
	matrix := make(matrix_t, 0)
	// calculate rotaionLists
	rotationLists := make([]int, sc.idSize)
	r := sc.GetResultInstance()
	rbb := r.GetBoundingbox()
	for psid := range sc.shapemap {
		voxel := sc.GetShapeInstance(uint(psid), 0).voxel
		symgroupID := voxel.CalcSelfSymmetries()
		rotlist := burrutils.RotationsToCheck[symgroupID]
		rotationLists[psid] = rotlist // need to copy??
		// need to imlplement breaker logic here
	}
	//
	// now start building the DLX matrix

	for psid := range sc.shapemap {
		rotlist := burrutils.HashToRotations(rotationLists[psid])
		for _, rotidx := range rotlist {
			rotatedInstance := sc.GetShapeInstance(uint(psid), uint(rotidx))
			pbb := rotatedInstance.GetBoundingbox()

			for x := rbb.Min[0] - pbb.Min[0]; x <= rbb.Max[0]-pbb.Max[0]; x++ {
				for y := rbb.Min[1] - pbb.Min[1]; y <= rbb.Max[1]-pbb.Max[1]; y++ {
					for z := rbb.Min[2] - pbb.Min[2]; z <= rbb.Max[2]-pbb.Max[2]; z++ {
						row := sc.GetDLXrow(uint(psid), uint(rotidx), x, y, z)
						if len(row) > 0 {
							matrix = append(matrix, &row)
							// KG: Now track the metadata for this row somewhere too
						}
					}
				}
			}

		}
	}

	return &matrix
}
