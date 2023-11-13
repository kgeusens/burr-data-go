package solver

import (
	"slices"
	"strconv"
	"strings"

	burrutils "github.com/kgeusens/go/burr-data/burrutils"
)

type node_t struct {
	parent          *node_t
	root            *node_t
	isSeparation    bool
	offsetList      []burrutils.Distance_t
	movingPieceList []burrutils.Id_t
	moveDirection   [3]burrutils.Distance_t
	id              string
	rootDetails     *rootDetails_t
}

type rootDetails_t struct {
	pieceList    []burrutils.Id_t
	rotationList []burrutils.Id_t
	hotspotList  []burrutils.Distance_t
}

func NewNodeChild(parent *node_t, movingPieceList []burrutils.Id_t, translation [3]burrutils.Distance_t, separation bool) (child *node_t) {
	child = new(node_t)
	child.root = parent.root
	child.parent = parent
	child.isSeparation = separation
	child.offsetList = []burrutils.Distance_t{}
	child.offsetList = append(child.offsetList, parent.offsetList...) // copy slice of ints
	child.movingPieceList = []burrutils.Id_t{}
	child.movingPieceList = append(child.movingPieceList, movingPieceList...)
	child.moveDirection[0] = translation[0]
	child.moveDirection[1] = translation[1]
	child.moveDirection[2] = translation[2]
	v := burrutils.Id_t(0)
	// update the offsetlist based on the provided translation
	for i := 0; i < len(movingPieceList); i++ {
		v = movingPieceList[i] * 3
		child.offsetList[v] += translation[0]
		child.offsetList[v+1] += translation[1]
		child.offsetList[v+2] += translation[2]
	}
	return
}

func NewNodeFromAssembly(passembly *assembly_t) *node_t {
	assembly := *passembly
	root := new(node_t)
	root.root = root
	root.rootDetails = &rootDetails_t{[]burrutils.Id_t{}, []burrutils.Id_t{}, []burrutils.Distance_t{}}
	// loop over the shape annotations
	for _, v := range assembly {
		root.rootDetails.pieceList = append(root.rootDetails.pieceList, v.shapeID)
		root.rootDetails.rotationList = append(root.rootDetails.rotationList, v.rotation)
		root.rootDetails.hotspotList = append(root.rootDetails.hotspotList, v.hotspot[0], v.hotspot[1], v.hotspot[2])
		root.offsetList = append(root.offsetList, v.offset[0], v.offset[1], v.offset[2])
	}
	return root
}

func (node *node_t) Separate() []*node_t {
	newNodes := []*node_t{}
	if node.isSeparation {
		// only add a new rootNode if it will contain more than 1 piece
		nPieces := len(node.root.rootDetails.pieceList)
		if nPieces-len(node.movingPieceList) > 1 {
			// so at this point, we know we are a separation
			// movingPieceList and movingDirection tells us what to work with
			newRoot := new(node_t)
			newRoot.rootDetails = &rootDetails_t{[]burrutils.Id_t{}, []burrutils.Id_t{}, []burrutils.Distance_t{}}
			newRoot.parent = node
			newRoot.root = newRoot
			newRoot.offsetList = []burrutils.Distance_t{}
			// only keep the pieces that are not moving. Filter out the moving pieces
			for i, v := range node.root.rootDetails.pieceList {
				if !slices.Contains(node.movingPieceList, burrutils.Id_t(i)) {
					newRoot.rootDetails.pieceList = append(newRoot.rootDetails.pieceList, v)
				}
			}
			for i, v := range node.root.rootDetails.rotationList {
				if !slices.Contains(node.movingPieceList, burrutils.Id_t(i)) {
					newRoot.rootDetails.rotationList = append(newRoot.rootDetails.rotationList, v)
				}
			}
			for idx := 0; idx < nPieces; idx++ {
				if !slices.Contains(node.movingPieceList, burrutils.Id_t(idx)) {
					newRoot.rootDetails.hotspotList = append(newRoot.rootDetails.hotspotList, node.root.rootDetails.hotspotList[idx*3], node.root.rootDetails.hotspotList[idx*3+1], node.root.rootDetails.hotspotList[idx*3+2])
					newRoot.offsetList = append(newRoot.offsetList, node.offsetList[idx*3], node.offsetList[idx*3+1], node.offsetList[idx*3+2])
				}
			}
			newNodes = append(newNodes, newRoot)
		}
		if len(node.movingPieceList) > 1 {
			// This is normally the smallest partition
			newRoot := new(node_t)
			newRoot.rootDetails = &rootDetails_t{[]burrutils.Id_t{}, []burrutils.Id_t{}, []burrutils.Distance_t{}}
			newRoot.parent = node
			newRoot.root = newRoot
			newRoot.offsetList = []burrutils.Distance_t{}
			// only keep the pieces that are not moving. Filter out the moving pieces
			for i, v := range node.root.rootDetails.pieceList {
				if slices.Contains(node.movingPieceList, burrutils.Id_t(i)) {
					newRoot.rootDetails.pieceList = append(newRoot.rootDetails.pieceList, v)
				}
			}
			for i, v := range node.root.rootDetails.rotationList {
				if slices.Contains(node.movingPieceList, burrutils.Id_t(i)) {
					newRoot.rootDetails.rotationList = append(newRoot.rootDetails.rotationList, v)
				}
			}
			for idx := 0; idx < nPieces; idx++ {
				if slices.Contains(node.movingPieceList, burrutils.Id_t(idx)) {
					newRoot.rootDetails.hotspotList = append(newRoot.rootDetails.hotspotList, node.root.rootDetails.hotspotList[idx*3], node.root.rootDetails.hotspotList[idx*3+1], node.root.rootDetails.hotspotList[idx*3+2])
					newRoot.offsetList = append(newRoot.offsetList, node.offsetList[idx*3], node.offsetList[idx*3+1], node.offsetList[idx*3+2])
				}
			}
			newNodes = append(newNodes, newRoot)
		}
	}
	return newNodes
}

func (node *node_t) GetId() string {
	if node.id == "" {
		nPieces := len(node.root.rootDetails.pieceList)
		offsetList := node.offsetList
		str := []string{"id"}

		for idx := 0; idx < nPieces; idx++ {
			str = append(str, strconv.Itoa(int(offsetList[idx*3]-offsetList[0])), strconv.Itoa(int(offsetList[idx*3+1]-offsetList[1])), strconv.Itoa(int(offsetList[idx*3+2]-offsetList[2])))
		}
		node.id = strings.Join(str, " ")
	}
	return node.id
}
