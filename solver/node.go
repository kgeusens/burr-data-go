package solver

import (
	"strconv"
	"strings"

	burrutils "github.com/kgeusens/go/burr-data/burrutils"
	xmpuzzle "github.com/kgeusens/go/burr-data/xmpuzzle"
)

const maxShapes = 30

/*
Limitation:
The problem we are solving can not have more than 30 pieces.
That is because we needed a super fast "GetID" solution for the nodes that is "comparable"
and an array of fixed length seemed to be the fastest. However, the length of the array
has a considerable impace on the performance.
*/
type id_t [3 * maxShapes]burrutils.Distance_t

type node_t struct {
	parent          *node_t
	root            *node_t
	isSeparation    bool
	offsetList      []burrutils.Distance_t
	movingPieceList []burrutils.Id_t
	moveDirection   [3]burrutils.Distance_t
	id              id_t
	idValid         bool
	rootDetails     *rootDetails_t
}

type rootDetails_t struct {
	pieceList    []burrutils.Id_t
	rotationList []burrutils.Id_t
	hotspotList  []burrutils.Distance_t
	separation   *xmpuzzle.Separation
}

func (node *node_t) GetId() id_t {
	if !node.idValid {
		nPieces := len(node.root.rootDetails.pieceList)
		offsetList := node.offsetList
		for idx := 0; idx < nPieces; idx++ {
			node.id[idx*3] = offsetList[idx*3] - offsetList[0]
			node.id[1+idx*3] = offsetList[idx*3+1] - offsetList[1]
			node.id[2+idx*3] = offsetList[idx*3+2] - offsetList[2]
		}
		node.idValid = true
	}
	return node.id
}

func (node *node_t) RecordSeparationInRoot() {
	// set the values for pieces
	sep := node.root.rootDetails.separation
	sep.Pieces.Count = len(node.root.rootDetails.pieceList)
	str := make([]string, 0)
	for _, piece := range node.root.rootDetails.pieceList {
		str = append(str, strconv.Itoa(int(piece)))
	}
	sep.Pieces.Text = strings.Join(str, " ")
	// add the states by walking back up to the root
	keepWalking := true
	n := node
	for keepWalking {
		state := xmpuzzle.State{}
		dx := []string{}
		dy := []string{}
		dz := []string{}
		for i := 0; i < sep.Pieces.Count; i++ {
			dx = append(dx, strconv.Itoa(int(n.offsetList[i*3])))
			dy = append(dy, strconv.Itoa(int(n.offsetList[i*3+1])))
			dz = append(dz, strconv.Itoa(int(n.offsetList[i*3+2])))
		}
		state.DX.Text = strings.Join(dx, " ")
		state.DY.Text = strings.Join(dy, " ")
		state.DZ.Text = strings.Join(dz, " ")
		sep.State = append(sep.State, state)
		if n != n.root {
			n = n.root
		} else {
			keepWalking = false
		}
	}
	// the states are in reverse order now, we should correct this
	for i := 0; i < len(sep.State)/2; i++ {
		sep.State[i], sep.State[len(sep.State)-1-i] = sep.State[len(sep.State)-1-i], sep.State[i]
	}
	// now record this in the rootdetails of the root
}
