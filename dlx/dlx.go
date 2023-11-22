/*
Port of the dlx implementation of Tim Beyer (https://github.com/TimBeyer/node-dlx)
Tim created a dlx implementation that does NOT use recursion, but uses a statemachine.
Go is not very good at deep recursive calls, and hopefully this implementation will be more performant.
*/

package dlx

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

type Searchconfig_t struct {
	NumPrimary   int
	NumSecondary int
	NumSolutions int
	rows         []Row_t
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
	numPrimary, numSecondary, rows := config.NumPrimary, config.NumSecondary, config.rows
	root := columnindex_t(0)

	numNodes := 0
	for i := range config.rows {
		numNodes += len(rows[i].coveredColumns)
	}

	solutions := [][]result_t{{}}
	nleft := make([]nodeindex_t, numNodes+numPrimary+numSecondary+1)
	nright := make([]nodeindex_t, numNodes+numPrimary+numSecondary+1)
	nup := make([]nodeindex_t, numNodes+numPrimary+numSecondary+1)
	ndown := make([]nodeindex_t, numNodes+numPrimary+numSecondary+1)
	ncol := make([]columnindex_t, numNodes+numPrimary+numSecondary+1)
	nindex := make([]int, numNodes+numPrimary+numSecondary+1)
	ndata := make([]any, numNodes+numPrimary+numSecondary+1)
	chead := make([]nodeindex_t, numPrimary+numSecondary+1)
	clen := make([]nodeindex_t, numPrimary+numSecondary+1)
	cprev := make([]columnindex_t, numPrimary+numSecondary+1)
	cnext := make([]columnindex_t, numPrimary+numSecondary+1)

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

		for i := 0; i < numPrimary; i++ {
			head := curNodeIndex
			nup[head] = head
			ndown[head] = head

			column := curColIndex
			chead[column] = head
			clen[column] = 0

			cprev[column] = column - 1
			cnext[column-1] = column

			//			colArray[curColIndex] = column
			curColIndex += 1
			curNodeIndex += 1
		}

		lastCol := curColIndex - 1
		// Link the last primary constraint to wrap back into the root
		cnext[lastCol] = root
		cprev[root] = lastCol

		for i := 0; i < numSecondary; i++ {
			head := curNodeIndex
			nup[head] = head
			ndown[head] = head

			column := curColIndex
			chead[column] = head
			clen[column] = 0
			cprev[column] = column
			cnext[column] = column

			//			colArray[curColIndex] = column
			curColIndex += 1
			curNodeIndex += 1
		}
	}

	var readRows = func() {
		curNodeIndex := nodeindex_t(numPrimary + numSecondary + 1)

		for i := 0; i < len(rows); i++ {
			row := rows[i]
			var rowStart nodeindex_t

			for _, columnIndex := range row.coveredColumns {
				node := curNodeIndex
				nleft[node] = node
				nright[node] = node
				ndown[node] = node
				nup[node] = node
				nindex[node] = i
				ndata[node] = row.data

				//				nodeArray[curNodeIndex] = node

				if rowStart == 0 {
					rowStart = node
				} else {
					nleft[node] = node - 1
					nright[node-1] = node
				}

				//				col := colArray[columnIndex+1]
				ncol[node] = columnindex_t(columnIndex + 1)

				nup[node] = nup[chead[columnIndex+1]]
				ndown[nup[chead[columnIndex+1]]] = node

				nup[chead[columnIndex+1]] = node
				ndown[node] = chead[columnIndex+1]

				clen[columnIndex+1] += 1
				curNodeIndex += 1
			}

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
