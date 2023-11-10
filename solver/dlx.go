package solver

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

func (sc SolverCache_t) GetDLXrow(shapeid, rotid uint, x, y, z int) (result []int) {
	// Baseline the resmap by creating 2 arrays:
	// one for the filled pixels, and one for the vari pixels
	r := NewVoxelinstance(sc.resultVoxel, 0)
	//	rbb := r.GetBoundingbox()
	resmap := *(r.GetWorldmap())
	piecemap := sc.GetShapeInstance(shapeid, rotid).GetWorldmap().Clone()
	piecemap.Translate(x, y, z)
	var filledHashSequence, variHashSequence [][3]int
	for key := range resmap {
		if resmap.Value(key) == 1 {
			filledHashSequence = append(filledHashSequence, resmap.Position(key))
		} else {
			variHashSequence = append(variHashSequence, resmap.Position(key))
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
		if !resmap.Has(piecemap.Position(key)) {
			// if we can not place a pixel, bail out and return nil (no DLXmap to create)
			return nil
		}
		// The DLX algorithm in go is different, we just need to pass the positions of the "1"s
		result = append(result, lookupMap[piecemap.Position(key)])
	}
	return
}
