package solver

import (
	burrutils "github.com/kgeusens/go/burr-data/burrutils"
)

type searchState int

const (
	forwardState searchState = 0
	advanceState searchState = 1
	backupState  searchState = 2
	recoverState searchState = 3
	doneState    searchState = 4
)

type Row_t struct {
	coveredColumns []int
	data           any
}

type nodeindex_t int
type columnindex_t int

type solutioncache_t struct {
	holes     int  // number of holes in the solution
	tmin      int  // total of all minimums
	tmax      int  // total of all maximums
	usesRange bool // true if any of the parts uses a range (min != max)
}

type Searchconfig_t struct {
	NumSolutions  int
	problemCache  ProblemCache_t
	rows          []Row_t
	solutionCache solutioncache_t
}

func NewSearchconfig(pc ProblemCache_t) (sc Searchconfig_t) {
	sc.problemCache = pc

	tminSize := 0
	tmaxSize := 0
	usesRange := false
	for idx, shape := range pc.GetProblem().Shapes {
		psize := pc.puzzle.Shapes[pc.GetProblem().Shapes[idx].Id].Size()
		sc.solutionCache.tmin += int(shape.GetPartMinimum())
		sc.solutionCache.tmax += int(shape.GetPartMaximum())
		tminSize += int(shape.GetPartMinimum()) * psize
		tmaxSize += int(shape.GetPartMaximum()) * psize
		usesRange = usesRange || (shape.GetPartMinimum() != shape.GetPartMaximum())
	}
	sc.solutionCache.usesRange = usesRange
	if !usesRange {
		sc.solutionCache.holes = pc.numPrimary + pc.numSecondary - tminSize
	} else {
		sc.solutionCache.holes = 0xFFFFFF
	}
	return
}

func (sc *Searchconfig_t) NumPrimary() nodeindex_t {
	return nodeindex_t(sc.problemCache.numPrimary) + nodeindex_t(sc.solutionCache.tmin)
}

func (sc *Searchconfig_t) NumSecondary() nodeindex_t {
	return nodeindex_t(sc.problemCache.numSecondary) + nodeindex_t(sc.solutionCache.tmax) - nodeindex_t(sc.solutionCache.tmin)
}

type result_t struct {
	index int
	data  any
}

func (r *result_t) GetData() any {
	return r.data
}

func (sc *Searchconfig_t) AddRow(columns []int, data any) {
	if sc.rows == nil {
		sc.rows = make([]Row_t, 0)
	}
	sc.rows = append(sc.rows, Row_t{columns, data})
}

func (config *Searchconfig_t) Search() [][]result_t {
	numSolutions := config.NumSolutions
	//	numPrimary, numSecondary, rows := config.NumPrimary(), config.NumSecondary(), config.rows
	headerSize := config.NumPrimary() + config.NumSecondary() + nodeindex_t(config.solutionCache.tmax)
	root := columnindex_t(0)

	numNodes := nodeindex_t(0)
	for i := range config.rows {
		numNodes += nodeindex_t(len(config.rows[i].coveredColumns)) + 1
	}

	solutions := [][]result_t{{}}
	nleft := make([]nodeindex_t, numNodes+headerSize+1)
	nright := make([]nodeindex_t, numNodes+headerSize+1)
	nup := make([]nodeindex_t, numNodes+headerSize+1)
	ndown := make([]nodeindex_t, numNodes+headerSize+1)
	ncol := make([]columnindex_t, numNodes+headerSize+1)
	nindex := make([]int, numNodes+headerSize+1)
	ndata := make([]any, numNodes+headerSize+1)
	chead := make([]nodeindex_t, headerSize+1)
	clen := make([]nodeindex_t, headerSize+1)
	cprev := make([]columnindex_t, headerSize+1)
	cnext := make([]columnindex_t, headerSize+1)

	currentSearchState := forwardState
	running := true
	level := 0
	choice := make([]nodeindex_t, 100)
	var bestCol columnindex_t
	var currentNode nodeindex_t

	var readColumnNames = func() {
		// Skip root node
		curColIndex := columnindex_t(1)
		curNodeIndex := nodeindex_t(0)

		for i := nodeindex_t(0); i < config.NumPrimary(); i++ {
			head := curNodeIndex
			nup[head] = head
			ndown[head] = head

			column := curColIndex
			chead[column] = head
			clen[column] = 0

			cprev[column] = column - 1
			cnext[column-1] = column

			curColIndex += 1
			curNodeIndex += 1
		}
		// Link the last primary constraint to wrap back into the root
		cnext[curColIndex-1] = root
		cprev[root] = curColIndex - 1

		// The secondary columns do not wrap in a circle but are standalone
		for i := nodeindex_t(0); i < config.NumSecondary(); i++ {
			head := curNodeIndex
			nup[head] = head
			ndown[head] = head

			column := curColIndex
			chead[column] = head
			clen[column] = 0
			cprev[column] = column
			cnext[column] = column

			curColIndex += 1
			curNodeIndex += 1
		}
	}

	var readRows = func() {
		// we need to assign this row to the correct column (piecenode)
		// to do that we need to get the shapeID and the instanceID from row.data
		curNodeIndex := nodeindex_t(config.NumPrimary() + config.NumSecondary() + 1)

		for i, row := range config.rows {
			var rowStart nodeindex_t

			annot := row.data.(annotation_t)
			partID := annot.partID
			partInstance := annot.instanceID
			// create the "piecenode" that represents the piece and put it in the correct column
			// This node is the start of the nodes of the row
			{
				node := curNodeIndex
				nleft[node] = node
				nright[node] = node
				ndown[node] = node
				nup[node] = node
				nindex[node] = i
				ndata[node] = row.data
				rowStart = node
				// figure out the column
				var col columnindex_t
				if partInstance < burrutils.Id_t(config.problemCache.GetProblem().Shapes[partID].GetPartMinimum()) {
					// its a mandatory piece (primary)
					col = 1 + columnindex_t(config.problemCache.numPrimary) + columnindex_t(partInstance)
					for i := 0; i < int(partID); i++ {
						col += columnindex_t(config.problemCache.GetProblem().Shapes[i].GetPartMinimum())
					}
				}
				ncol[node] = col
				// now insert it into its column
				nup[node] = nup[chead[col]]
				ndown[nup[chead[col]]] = node
				nup[chead[col]] = node
				ndown[node] = chead[col]
				clen[col] += 1
				curNodeIndex += 1
			}
			// Now add the row entries
			for _, columnIndex := range row.coveredColumns {
				// We need to make a distinction between primary and secondary to assign to the correct columns
				// Prep the node
				node := curNodeIndex
				nleft[node] = node
				nright[node] = node
				ndown[node] = node
				nup[node] = node
				nindex[node] = i
				ndata[node] = row.data
				// we already prep'ed a piecenode so we just continue adding to the circle
				nleft[node] = node - 1
				nright[node-1] = node
				// now check if this is a primary, or secondary entry
				var col columnindex_t
				if columnIndex < config.problemCache.numPrimary {
					col = 1 + columnindex_t(columnIndex)
				} else {
					col = 1 + columnindex_t(config.solutionCache.tmax+columnIndex)
				}
				// now insert the node in the correct column
				ncol[node] = col
				nup[node] = nup[chead[col]]
				ndown[nup[chead[col]]] = node
				nup[chead[col]] = node
				ndown[node] = chead[col]
				clen[col] += 1
				curNodeIndex += 1
			}
			// I think this is no longer needed
			nleft[rowStart] = curNodeIndex - 1
			nright[curNodeIndex-1] = rowStart
		}
	}

	var cover = func(c columnindex_t) {
		// Unlink column
		cnext[cprev[c]] = cnext[c]
		cprev[cnext[c]] = cprev[c]

		// From top to bottom, left to right unlink every row node from its column
		for rr := ndown[chead[c]]; rr != chead[c]; rr = ndown[rr] {
			for nn := nright[rr]; nn != rr; nn = nright[nn] {
				ndown[nup[nn]] = ndown[nn]
				nup[ndown[nn]] = nup[nn]

				clen[ncol[nn]] -= 1
			}
		}
	}

	var uncover = func(c columnindex_t) {
		// From bottom to top, right to left relink every row node to its column
		//		var uu, dd nodeindex_t
		for rr := nup[chead[c]]; rr != chead[c]; rr = nup[rr] {
			for nn := nleft[rr]; nn != rr; nn = nleft[nn] {

				ndown[nup[nn]] = nn
				nup[ndown[nn]] = nn

				clen[ncol[nn]] += 1
			}
		}

		// Unlink column
		cnext[cprev[c]] = c
		cprev[cnext[c]] = c
	}

	var pickBestColum = func() {
		lowestLen := clen[cnext[root]]
		lowest := cnext[root]

		for curCol := cnext[root]; curCol != root; curCol = cnext[curCol] {
			length := clen[curCol]
			if length < lowestLen {
				lowestLen = length
				lowest = curCol
			}
		}

		bestCol = lowest
	}

	var recordSolution = func() {
		results := []result_t{}
		for l := 0; l <= level; l++ {
			node := choice[l]
			results = append(results, result_t{nindex[node], ndata[node]})
		}
		solutions = append(solutions, results)
	}

	//	stateMethods := []func(){forward, advance, backup, recover, done}

	readColumnNames()
	readRows()

	for running {
		switch currentSearchState {
		case forwardState:
			// pick the best column to process, and select the first node of the first row (currentNode)
			pickBestColum()
			cover(bestCol)
			currentNode = ndown[chead[bestCol]]
			choice[level] = currentNode
			currentSearchState = advanceState
		case advanceState:
			// analyze the selected row from the previous step
			// either go to:
			//   backupState (deadend, rollback because there is no row to process)
			//   doneState (solution found, but we reached the limit of numSolutions to find)
			//   recoverState (solution found, we need to move on to find more solutions)
			//   forwardState (no solution yet, no deadend yet, go to the next column)
			if currentNode == chead[bestCol] {
				// if the currentNode == the header, then this column has 0 selectable rows
				currentSearchState = backupState
				break
			}
			for pp := nright[currentNode]; pp != currentNode; pp = nright[pp] {
				// cover all the columns for the row containing currentNode
				cover(ncol[pp])
			}
			if cnext[root] == root {
				// if there are no remaining columns to process, we have a solution
				recordSolution()
				if len(solutions) == numSolutions {
					currentSearchState = doneState
				} else {
					currentSearchState = recoverState
				}
				break
			}
			level = level + 1
			currentSearchState = forwardState
		case backupState:
			// recover from a deadend, go a level back
			// either go to:
			//    doneState (if we are back at level 0, we are done)
			//    recoverState (continue the hunt for a solution)
			uncover(bestCol)
			if level == 0 {
				currentSearchState = doneState
				break
			}
			level = level - 1
			currentNode = choice[level]
			bestCol = ncol[currentNode]
			currentSearchState = recoverState
		case recoverState:
			// uncover the current row
			// move on to the next potential row for the current column
			// go to:
			//   advanceState (analyze the selected row)
			for pp := nleft[currentNode]; pp != currentNode; pp = nleft[pp] {
				uncover(ncol[pp])
			}
			currentNode = ndown[currentNode]
			choice[level] = currentNode
			currentSearchState = advanceState
		case doneState:
			// we're done, go home
			running = false
		}
	}

	return solutions
}

/*
import (
	burrutils "github.com/kgeusens/go/burr-data/burrutils"
)

type Row_t struct {
	coveredColumns []int
	data           any
}

type piecePosition_t struct {
	x, y, z burrutils.Distance_t
	rot     burrutils.Id_t
	row     nodeindex_t
	piece   nodeindex_t
}

type nodeindex_t int

//type columnindex_t int

type Searchconfig_t struct {
	NumSolutions int
	problemCache ProblemCache_t
	rows         []Row_t
}

func NewSearchconfig(pc ProblemCache_t) (sc Searchconfig_t) {
	sc.problemCache = pc

	return
}

func (sc *Searchconfig_t) NumPrimary() nodeindex_t {
	return nodeindex_t(sc.problemCache.numPrimary)
}

func (sc *Searchconfig_t) NumSecondary() nodeindex_t {
	return nodeindex_t(sc.problemCache.numSecondary)
}

type result_t struct {
	index int
	data  any
}

func (r *result_t) GetData() any {
	return r.data
}

func (sc *Searchconfig_t) AddRow(columns []int, data annotation_t) {
	if sc.rows == nil {
		sc.rows = make([]Row_t, 0)
	}
	sc.rows = append(sc.rows, Row_t{columns, data})
}

func (config *Searchconfig_t) Search() [][]result_t {
	//	numSolutions := config.NumSolutions
	numPrimary, numSecondary, configrows := config.NumPrimary(), config.NumSecondary(), config.rows
	root := nodeindex_t(0)
	headerNodes := nodeindex_t(0)

	numParts := nodeindex_t(config.problemCache.GetProblem().NumShapes())
	maxSize := numParts + numPrimary + numSecondary + 2

	left := make([]nodeindex_t, 0, maxSize)
	right := make([]nodeindex_t, 0, maxSize)
	up := make([]nodeindex_t, 0, maxSize)
	down := make([]nodeindex_t, 0, maxSize)
	colCount := make([]nodeindex_t, 0, maxSize)
	weight := make([]nodeindex_t, 0, maxSize)
	min := make([]nodeindex_t, 0, maxSize)
	max := make([]nodeindex_t, 0, maxSize)
	piecePositions := make([]piecePosition_t, 0, maxSize)
	columns := make([]nodeindex_t, config.problemCache.resultVoxel.Volume())
	holeColumns := make([]nodeindex_t, 0, numSecondary)

	hidden_rows := make([]nodeindex_t, 0)
	holes := nodeindex_t(0)
	rangeColumn := nodeindex_t(0)

	var generateFirstRow = func() {
		for i := nodeindex_t(0); i < maxSize; i++ {
			right = append(right, i+1)
			left = append(left, i-1)
			up = append(up, i)
			down = append(down, i)
			colCount = append(colCount, 0)
			min = append(min, 1)
			max = append(max, 1)
			weight = append(weight, 0)
		}
		left[root] = maxSize - 1
		right[maxSize-1] = root

		// initialize min and max for the piece headers
		tmin := nodeindex_t(0)
		tmax := nodeindex_t(0)
		res_filled := numPrimary + numSecondary
		rangeMin := numPrimary
		rangeMax := res_filled
		for pc := range config.problemCache.GetProblem().Shapes {
			max[1+pc] = nodeindex_t(config.problemCache.GetProblem().GetPartMaximum(burrutils.Id_t(pc)))
			min[1+pc] = nodeindex_t(config.problemCache.GetProblem().GetPartMinimum(burrutils.Id_t(pc)))
			psize := nodeindex_t(config.problemCache.puzzle.Shapes[config.problemCache.GetProblem().Shapes[pc].Id].Size())
			tmax += max[1+pc] * psize
			tmin += min[1+pc] * psize
			if min[1+pc] == max[1+pc] {
				rangeMin -= min[1+pc] * psize
				rangeMax -= max[1+pc] * psize
			}
		}
		if rangeMin < 0 {
			rangeMin = 0
		}
		hasRange := (tmin != tmax)
		// calculate the holes
		rangeColumn = maxSize - 1
		if !hasRange {
			holes = res_filled - tmin
			// patch the header ring and remove the last entry (the range column) because there is no need for it
			right[left[rangeColumn]] = right[rangeColumn]
			left[right[rangeColumn]] = left[rangeColumn]
			rangeColumn = 0
			maxSize -= 1
		} else {
			rangeColumn = maxSize - 1
			min[rangeColumn] = rangeMin
			max[rangeColumn] = rangeMax
			if config.problemCache.GetProblem().MaxHolesDefined() {
				holes = nodeindex_t(config.problemCache.GetProblem().GetMaxHoles())
			} else {
				holes = 0xFFFFFF
			}
		}
		headerNodes = maxSize
		// Correct the min values for secondary columns and build the columns array
		// since we already sorted the columns with "filled" at the front and "vari" at the back, this logic becomes easier
		// KG: maybe the columns array can be elimintated because the position in the header determines if you are the header
		// for the root, a part, a filled voxel, a vary voxel, or the range column
		c := 1 + numParts
		for i := nodeindex_t(0); i < numPrimary; i++ {
			columns[i] = c + i
		}
		for i := numPrimary; i < numPrimary+numSecondary; i++ {
			columns[i] = c + i
			min[c+i] = 0
			holeColumns = append(holeColumns, c+i)
		}
	}

	var addPieceNode = func(piece nodeindex_t, rot burrutils.Id_t, x, y, z burrutils.Distance_t) (piecenode nodeindex_t) {
		piecenode = nodeindex_t(len(left))

		left = append(left, piecenode)
		right = append(right, piecenode)
		up = append(up, up[piece+1])
		down = append(down, piece+1)
		weight = append(weight, 1)

		down[up[piece+1]] = piecenode
		up[piece+1] = piecenode

		colCount = append(colCount, piece+1)
		colCount[piece+1]++

		piecePositions = append(piecePositions, piecePosition_t{piece: piece, x: x, y: y, z: z, rot: rot, row: piecenode})
		return piecenode
	}

	var addVoxelNode = func(col, piecenode nodeindex_t) {
		newnode := nodeindex_t(len(left))

		right = append(right, piecenode)
		left = append(left, left[piecenode])
		right[left[piecenode]] = newnode
		left[piecenode] = newnode

		up = append(up, up[col])
		down = append(down, col)
		down[up[col]] = newnode
		up[col] = newnode

		colCount = append(colCount, col)

		weight = append(weight, 1)
		colCount[col]++
	}

	var readRows = func() {
		for i := 0; i < len(configrows); i++ {
			row := configrows[i]
			annot := row.data.(annotation_t)
			piecenode := addPieceNode(nodeindex_t(annot.shapeID), annot.rotation, annot.hotspot[0]+annot.offset[0], annot.hotspot[1]+annot.offset[1], annot.hotspot[2]+annot.offset[2])
			for _, columnIndex := range row.coveredColumns {
				addVoxelNode(nodeindex_t(columnIndex)+nodeindex_t(annot.shapeID)+1, piecenode)
			}
		}
	}
	var open_column_conditions_fulfillable = func() bool {
		for col := right[0]; col > 0; col = right[col] {
			if weight[col] > max[col] {
				return false
			}
			if weight[col]+colCount[col] < min[col] {
				return false
			}
		}
		return true
	}

	var betterParams = func(n_sum, n_min, n_max, o_sum, o_min, o_max nodeindex_t) bool {
		return n_sum*n_max < o_sum*o_max
	}

	var find_best_unclosed_column = func() nodeindex_t {
		col := right[0]
		// if we have no column -> return no column
		if col == 0 {
			return nodeindex_t(-1)
		}
		// first column is best column for the beginning
		bestcol := col
		col = right[col]
		for col > 0 {
			if betterParams(colCount[col], min[col]-weight[col], max[col]-weight[col], colCount[bestcol], min[bestcol]-weight[bestcol], max[bestcol]-weight[bestcol]) {
				bestcol = col
			}
			if colCount[col] == 0 {
				return col
			}
			col = right[col]
		}
		return bestcol
	}

	var cover_column_only = func(col nodeindex_t) {
		right[left[col]] = right[col]
		left[right[col]] = left[col]
	}

	var uncover_column_only = func(col nodeindex_t) {
		right[left[col]] = col
		left[right[col]] = col
	}

	var cover_column_rows = func(col nodeindex_t) {
		for r := down[col]; r != col; r = down[r] {
			colCount[col] -= weight[r]
			for c := right[r]; c != r; c = right[c] {
				up[down[c]] = up[c]
				down[up[c]] = down[c]
				colCount[colCount[c]] -= weight[c]
			}
		}
	}

	var uncover_column_rows = func(col nodeindex_t) {
		for r := up[col]; r != col; r = up[r] {
			for c := left[r]; c != r; c = left[c] {
				colCount[colCount[c]] += weight[c]
				up[down[c]] = c
				down[up[c]] = c
			}
			colCount[col] += weight[r]
		}
	}

	var hiderow = func(r nodeindex_t) {
		for rr := right[r]; rr != r; rr = right[rr] {
			up[down[rr]] = up[rr]
			down[up[rr]] = down[rr]

			colCount[colCount[rr]] -= weight[rr]
		}
		up[down[r]] = up[r]
		down[up[r]] = down[r]
		colCount[colCount[r]] -= weight[r]
	}

	var unhiderow = func(r nodeindex_t) {
		up[down[r]] = r
		down[up[r]] = r
		colCount[colCount[r]] += weight[r]
		for rr := left[r]; rr != r; rr = left[rr] {
			up[down[rr]] = rr
			down[up[rr]] = rr
			colCount[colCount[rr]] += weight[rr]
		}
	}

	var hiderows = func(row nodeindex_t) {
		// put in the separator
		hidden_rows = append(hidden_rows, 0)

		for r := right[row]; r != row; r = right[r] {
			col := colCount[r]
			// now check all rows of this column for too big weights
			for rr := down[col]; rr != col; rr = down[rr] {
				if weight[rr]+weight[col] > max[col] {
					hiderow(rr)
					hidden_rows = append(hidden_rows, rr)
				}
			}
		}
	}

	var unhiderows = func() {
		for hidden_rows[len(hidden_rows)-1] > 0 {
			unhiderow(hidden_rows[len(hidden_rows)-1])
			hidden_rows = hidden_rows[:len(hidden_rows)-1]
		}
		// Remove separator
		hidden_rows = hidden_rows[:len(hidden_rows)-1]
	}

	var column_condition_fulfilled = func(col nodeindex_t) bool {
		return (weight[col] >= min[col]) && (weight[col] <= max[col])
	}

	var column_condition_fulfillable = func(col nodeindex_t) bool {
		if weight[col] > max[col] {
			return false
		}
		if weight[col]+colCount[col] < min[col] {
			return false
		}
		return true
	}

	var solution = func() {
		// KG: not implemented yet
	}

	var row, col nodeindex_t

	task_stack := []int{0}
	next_row_stack := []nodeindex_t{0}
	column_stack := []nodeindex_t{}
	finished_a := []nodeindex_t{}
	finished_b := []nodeindex_t{}
	rows := []nodeindex_t{}
	iterations := 0
	abort := false

	generateFirstRow()
	readRows()

	for len(task_stack) > 0 {
		iterations++
		if abort {
			if task_stack[len(task_stack)-1] == 1 || task_stack[len(task_stack)-1] == 2 || task_stack[len(task_stack)-1] == 5 {
				break
			}
		}
		switch task_stack[len(task_stack)-1] {
		case 0:
			if holes < numSecondary {
				cnt := holes
				ret := false
				var i nodeindex_t
				offset := 1 + numParts + numPrimary
				for i = 0; i < numSecondary; i++ {
					if colCount[offset+i] == 0 && weight[offset+i] == 0 {
						if cnt == 0 {
							next_row_stack = next_row_stack[:len(next_row_stack)-1]
							task_stack = task_stack[:len(task_stack)-1]
							ret = true
							break
						} else {
							cnt--
						}
					}
				}
				if ret {
					break
				}
			}
			if next_row_stack[len(next_row_stack)-1] < headerNodes {
				// when no column is left we have found a solution
				if right[root] == root {
					solution()
					next_row_stack = next_row_stack[:len(next_row_stack)-1]
					task_stack = task_stack[:len(task_stack)-1]
					break
				}
				col := find_best_unclosed_column()
				if col == -1 {
					next_row_stack = next_row_stack[:len(next_row_stack)-1]
					task_stack = task_stack[:len(task_stack)-1]
					break
				}
				// when there are no rows in the selected column, we don't need to find
				// any row set and can continue right on with a new column
				if colCount[col] == 0 {

					if column_condition_fulfilled(col) {
						// remove this column from the column list
						// we don not need to remove the rows, as there are no
						// and we start a new column
						cover_column_only(col)
						column_stack = append(column_stack, col)
						task_stack[len(task_stack)-1] = 1
						task_stack = append(task_stack, 0)
						next_row_stack = append(next_row_stack, 0)
						break
					}
				} else {
					// we can assume here that the columns condition is fulfillable
					// because whenever we call this function all columns that are left
					// must be fulfillable
					// remove this column from the column list
					// do not yet remove the rows of this column, this will be done
					// shortly before we recursively call this function again
					cover_column_only(col)
					column_stack = append(column_stack, col)
					task_stack[len(task_stack)-1] = 1
					task_stack = append(task_stack, 0)
					next_row_stack = append(next_row_stack, down[col])
					break
				}
				next_row_stack = next_row_stack[:len(next_row_stack)-1]
				task_stack = task_stack[:len(task_stack)-1]
				break
			}
			col = colCount[next_row_stack[len(next_row_stack)-1]]

			if column_condition_fulfilled(col) {

				finished_b = append(finished_b, colCount[col]+1)
				finished_a = append(finished_a, 0)

				// remove all rows that are left within this column
				// this way we make sure we are _not_ changing this columns value any more
				cover_column_rows(col)

				if open_column_conditions_fulfillable() {
					task_stack[len(task_stack)-1] = 2
					task_stack = append(task_stack, 0)
					next_row_stack = append(next_row_stack, 0)
					break
				}

				task_stack[len(task_stack)-1] = 2
				break

			} else {
				finished_b = append(finished_b, colCount[colCount[next_row_stack[len(next_row_stack)-1]]])
				finished_a = append(finished_a, 0)
			}

			task_stack[len(task_stack)-1] = 3
			//			break
		case 1:
			// reinsert this column
			uncover_column_only(column_stack[len(column_stack)-1])
			column_stack = column_stack[:len(column_stack)-1]
			next_row_stack = next_row_stack[:len(next_row_stack)-1]
			task_stack = task_stack[:len(task_stack)-1]
			//			break
		case 2:
			// reinsert rows of this column
			uncover_column_rows(colCount[next_row_stack[len(next_row_stack)-1]])
			finished_a[len(finished_a)-1]++
			// fall through
			fallthrough
		case 3:

			// add a unhiderows marker, so that the rows hidden in the loop
			// below can be unhidden properly
			hidden_rows = append(hidden_rows, 0)
			row = next_row_stack[len(next_row_stack)-1]
			if up[row] < row {
				rows = append(rows, row)
			} else {
				task_stack[len(task_stack)-1] = 7
				break
			}
			// fall through
			fallthrough
		case 4:

			row = rows[len(rows)-1]
			col = colCount[next_row_stack[len(next_row_stack)-1]]

			// add row to rowset
			weight[colCount[row]] += weight[row]
			for r := right[row]; r != row; r = right[r] {
				weight[colCount[r]] += weight[r]
			}
			// if there are unfulfillable columns we don't even need to check any further
			if open_column_conditions_fulfillable() {
				// remove useless rows (that are rows that have too much weight
				// in one of their nodes that would overflow the expected weight
				hiderows(row)
				if open_column_conditions_fulfillable() {
					if colCount[col] == 0 {
						// when there are no more rows in the current column
						// we can immediately start a new column
						// if the current column condition is really fulfilled
						if column_condition_fulfilled(col) {
							task_stack[len(task_stack)-1] = 5
							task_stack = append(task_stack, 0)
							next_row_stack = append(next_row_stack, 0)
							break
						}
					} else {
						// we need to recurse, if there are rows left and the current
						// column condition is still fulfillable, we need to check
						// the current column again because this column is no longer open,
						// is was removed on selection
						if column_condition_fulfillable(col) {
							newrow := row
							// do gown until we hit a row that is still inside the matrix
							// this works because rows are hidden one by one and so the double link
							// to the row above or below is no longer intact, when the row is gone, the down
							// pointer still points to the row that is was below before the row was hidden, but
							// the pointer from the row below doesn't point up to us, so we do down until
							// the link down-up points back to us
							for (down[newrow] >= headerNodes) && up[down[newrow]] != newrow {
								newrow = down[newrow]
							}
							task_stack[len(task_stack)-1] = 5
							task_stack = append(task_stack, 0)
							next_row_stack = append(next_row_stack, newrow)
							break
						}
					}
				}
			} else {
				task_stack[len(task_stack)-1] = 6
				break
			}
			// fall through
			fallthrough
		case 5:
			unhiderows()
			fallthrough
		case 6:

			// remove row from rowset
			row = rows[len(rows)-1]

			for r := left[row]; r != row; r = left[r] {
				weight[colCount[r]] -= weight[r]
			}
			weight[colCount[row]] -= weight[row]

			rows = rows[:len(rows)-1]

			(finished_a[len(finished_a)-1])++

			// after we finished with this row, we will never use it again, so
			// remove it from the matrix
			hiderow(row)
			hidden_rows = append(hidden_rows, row)

			row = down[row]

			if up[row] < row {
				rows = append(rows, row)
				task_stack[len(task_stack)-1] = 4
				break
			}
			fallthrough
		case 7:

			// reinsert all the rows that were remove over the course of the
			// row by row inspection
			unhiderows()

			finished_a = finished_a[:len(finished_a)-1]
			finished_b = finished_b[:len(finished_b)-1]

			next_row_stack = next_row_stack[:len(next_row_stack)-1]
			task_stack = task_stack[:len(task_stack)-1]
			//			break

		default:
			//			break

		}

	}

	return nil
}
*/
