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

type nodecache struct {
	freeList []*node_t
	// stackpointer int
}

func (nc *nodecache) request() *node_t {
	cacheLen := len(nc.freeList)
	var node *node_t
	if cacheLen > 0 {
		cacheLen--
		node = nc.freeList[cacheLen]
		nc.freeList = nc.freeList[:cacheLen]
	} else {
		node = new(node_t)
		node.movingPieceList = []burrutils.Id_t{}
		node.offsetList = []burrutils.Distance_t{}
	}
	/*
		if nc.stackpointer > 0 {
			nc.stackpointer--
			return nc.freeList[nc.stackpointer]
		} else {
			n := new(node_t)
			n.movingPieceList = []burrutils.Id_t{}
			return new(node_t)
		}
	*/
	return node
}

/*
func (nc *nodecache) release(node *node_t) {
	nc.freeList[nc.stackpointer] = node
	nc.stackpointer++
}
*/

func releaseNode(node *node_t) {
	node.parent = nil
	node.root = nil
	node.isSeparation = false
	node.offsetList = node.offsetList[:0]
	node.movingPieceList = node.movingPieceList[:0]
	node.moveDirection[0] = 0
	node.moveDirection[1] = 0
	node.moveDirection[2] = 0
	node.id = ""
	node.rootDetails = nil
	theNodecache.freeList = append(theNodecache.freeList, node)
	// theNodecache.stackpointer++
}

var theNodecache nodecache = nodecache{make([]*node_t, 0)}

func NewNodeChild(parent *node_t, movingPieceList []burrutils.Id_t, translation [3]burrutils.Distance_t, separation bool) (child *node_t) {
	child = theNodecache.request()
	child.root = parent.root
	child.parent = parent
	child.isSeparation = separation
	child.offsetList = append(child.offsetList, parent.offsetList...) // copy slice of ints
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
	root := theNodecache.request()
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
			newRoot := theNodecache.request()
			newRoot.rootDetails = &rootDetails_t{[]burrutils.Id_t{}, []burrutils.Id_t{}, []burrutils.Distance_t{}}
			newRoot.parent = node
			newRoot.root = newRoot
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
			newRoot := theNodecache.request()
			newRoot.rootDetails = &rootDetails_t{[]burrutils.Id_t{}, []burrutils.Id_t{}, []burrutils.Distance_t{}}
			newRoot.parent = node
			newRoot.root = newRoot
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

// room for 100 pieces
var str []string = make([]string, 301)

func (node *node_t) GetId() string {
	if node.id == "" {
		nPieces := len(node.root.rootDetails.pieceList)
		offsetList := node.offsetList
		//		str := make([]string, nPieces*3+1)
		str[0] = "id"
		for idx := 0; idx < nPieces; idx++ {
			str[1+idx*3] = strconv.Itoa(int(offsetList[idx*3] - offsetList[0]))
			str[2+idx*3] = strconv.Itoa(int(offsetList[idx*3+1] - offsetList[1]))
			str[3+idx*3] = strconv.Itoa(int(offsetList[idx*3+2] - offsetList[2]))
		}
		node.id = strings.Join(str[:nPieces*3+1], " ")
	}
	return node.id
}
