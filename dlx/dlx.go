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

type node_t struct {
	left  nodeindex_t
	right nodeindex_t
	up    nodeindex_t
	down  nodeindex_t
	col   columnindex_t
	index int
	data  any
}

type column_t struct {
	head nodeindex_t
	len  int
	prev columnindex_t
	next columnindex_t
}

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

	colArray := make([]column_t, numPrimary+numSecondary+1)
	numNodes := 0
	for i := range config.rows {
		numNodes += len(rows[i].coveredColumns)
	}
	nodeArray := make([]node_t, numNodes+numPrimary+numSecondary+1)
	solutions := [][]result_t{{}}

	currentSearchState := forwardState
	running := true
	level := 0
	choice := make([]node_t, 100)
	var bestCol columnindex_t
	var currentNode nodeindex_t

	var readColumnNames = func() {
		// Skip root node
		curColIndex := columnindex_t(1)
		curNodeIndex := nodeindex_t(0)

		for i := 0; i < numPrimary; i++ {
			head := curNodeIndex
			nodeArray[head].up = head
			nodeArray[head].down = head

			column := curColIndex
			colArray[column].head = head
			colArray[column].len = 0

			colArray[column].prev = column - 1
			colArray[column-1].next = column

			//			colArray[curColIndex] = column
			curColIndex += 1
			curNodeIndex += 1
		}

		lastCol := curColIndex - 1
		// Link the last primary constraint to wrap back into the root
		colArray[lastCol].next = root
		colArray[root].prev = lastCol

		for i := 0; i < numSecondary; i++ {
			head := curNodeIndex
			nodeArray[head].up = head
			nodeArray[head].down = head

			column := curColIndex
			colArray[column].head = head
			colArray[column].len = 0
			colArray[column].prev = column
			colArray[column].next = column

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
				nodeArray[node].left = node
				nodeArray[node].right = node
				nodeArray[node].down = node
				nodeArray[node].up = node
				nodeArray[node].index = i
				nodeArray[node].data = row.data

				//				nodeArray[curNodeIndex] = node

				if rowStart == 0 {
					rowStart = node
				} else {
					nodeArray[node].left = node - 1
					nodeArray[node-1].right = node
				}

				col := colArray[columnIndex+1]
				nodeArray[node].col = columnindex_t(columnIndex + 1)

				nodeArray[node].up = nodeArray[col.head].up
				nodeArray[nodeArray[col.head].up].down = node

				nodeArray[col.head].up = node
				nodeArray[node].down = col.head

				colArray[columnIndex+1].len += 1
				curNodeIndex += 1
			}

			nodeArray[rowStart].left = curNodeIndex - 1
			nodeArray[curNodeIndex-1].right = rowStart
		}
	}

	var cover = func(c *column_t) {
		l := c.prev
		r := c.next

		// Unlink column
		l.next = r
		r.prev = l

		// From to to bottom, left to right unlink every row node from its column
		for rr := c.head.down; rr != c.head; rr = rr.down {
			for nn := rr.right; nn != rr; nn = nn.right {
				uu := nn.up
				dd := nn.down

				uu.down = dd
				dd.up = uu

				nn.col.len -= 1
			}
		}
	}

	var uncover = func(c *column_t) {
		// From bottom to top, right to left relink every row node to its column
		for rr := c.head.up; rr != c.head; rr = rr.up {
			for nn := rr.left; nn != rr; nn = nn.left {
				uu := nn.up
				dd := nn.down

				uu.down = nn
				dd.up = nn

				nn.col.len += 1
			}
		}

		l := c.prev
		r := c.next

		// Unlink column
		l.next = c
		r.prev = c
	}

	var pickBestColum = func() {
		lowestLen := root.next.len
		lowest := root.next

		for curCol := root.next; curCol != root; curCol = curCol.next {
			length := curCol.len
			if length < lowestLen {
				lowestLen = length
				lowest = curCol
			}
		}

		bestCol = lowest
	}

	var forward = func() {
		pickBestColum()
		cover(bestCol)

		currentNode = bestCol.head.down
		choice[level] = currentNode

		currentSearchState = advanceState
	}

	var recordSolution = func() {
		results := []result_t{}
		for l := 0; l <= level; l++ {
			node := choice[l]
			results = append(results, result_t{node.index, node.data})
		}
		solutions = append(solutions, results)
	}

	var advance = func() {
		if currentNode == bestCol.head {
			currentSearchState = backupState
			return
		}

		for pp := currentNode.right; pp != currentNode; pp = pp.right {
			cover(pp.col)
		}

		if root.next == root {
			recordSolution()
			if len(solutions) == numSolutions {
				currentSearchState = doneState
			} else {
				currentSearchState = recoverState
			}
			return
		}

		level = level + 1
		currentSearchState = forwardState
	}

	var backup = func() {
		uncover(bestCol)

		if level == 0 {
			currentSearchState = doneState
			return
		}

		level = level - 1

		currentNode = choice[level]
		bestCol = currentNode.col

		currentSearchState = recoverState
	}

	var recover = func() {
		for pp := currentNode.left; pp != currentNode; pp = pp.left {
			uncover(pp.col)
		}
		currentNode = currentNode.down
		choice[level] = currentNode
		currentSearchState = advanceState
	}

	var done = func() {
		running = false
	}

	stateMethods := []func(){forward, advance, backup, recover, done}

	readColumnNames()
	readRows()

	for running {
		stateMethods[currentSearchState]()
	}

	return solutions
}
