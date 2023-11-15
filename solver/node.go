package solver

import (
	burrutils "github.com/kgeusens/go/burr-data/burrutils"
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
	id              *id_t
	rootDetails     *rootDetails_t
}

type rootDetails_t struct {
	pieceList    []burrutils.Id_t
	rotationList []burrutils.Id_t
	hotspotList  []burrutils.Distance_t
}

func (node *node_t) GetId() id_t {
	if node.id == nil {
		nPieces := len(node.root.rootDetails.pieceList)
		offsetList := node.offsetList
		node.id = new(id_t)
		//		str := make([]string, nPieces*3+1)
		for idx := 0; idx < nPieces; idx++ {
			node.id[idx*3] = offsetList[idx*3] - offsetList[0]
			node.id[1+idx*3] = offsetList[idx*3+1] - offsetList[1]
			node.id[2+idx*3] = offsetList[idx*3+2] - offsetList[2]
		}

	}
	return *node.id
}
