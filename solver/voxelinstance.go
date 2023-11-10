package solver

import (
	burrutils "github.com/kgeusens/go/burr-data/burrutils"
	xmpuzzle "github.com/kgeusens/go/burr-data/xmpuzzle"
)

type VoxelInstance struct {
	//	voxel          *xmpuzzle.Voxel
	//	offset         [3]int
	hotspot        [3]int
	rotation       uint
	cachedWorldmap *xmpuzzle.Worldmap
	cachedBB       *xmpuzzle.Boundingbox
}

func NewVoxelinstance(voxel *xmpuzzle.Voxel, rot uint) (vi VoxelInstance) {
	pvi := new(VoxelInstance)
	vi = *pvi
	//	vi.voxel = voxel
	vi.rotation = rot
	/*
		vi.offset[0] = offset[0]
		vi.offset[1] = offset[1]
		vi.offset[2] = offset[2]
	*/
	// cache the worldmap
	wm := voxel.NewWorldmap()
	vi.cachedWorldmap = &wm
	// rotate
	vi.cachedWorldmap.Rotate(rot)
	// move to positive quadrant and then translate over offset
	bb := vi.cachedWorldmap.CalcBoundingbox()
	trans := [3]int{-1 * bb.Min[0], -1 * bb.Min[1], -1 * bb.Min[2]}
	vi.cachedWorldmap.Translate(trans[0], trans[1], trans[2])
	// cache the boundingbox
	// KG: instead of creating a new boundingbox, consider just translating bb (memory efficiency)
	bb = vi.cachedWorldmap.CalcBoundingbox()
	vi.cachedBB = &bb
	// hotspot
	h1, h2, h3 := burrutils.Rotate(0, 0, 0, rot)
	vi.hotspot[0], vi.hotspot[1], vi.hotspot[2] = burrutils.Translate(h1, h2, h3, trans[0], trans[1], trans[2])
	return
}

func (vi VoxelInstance) GetWorldmap() (wm *xmpuzzle.Worldmap) {
	return vi.cachedWorldmap
}

func (vi VoxelInstance) GetBoundingbox() (wm *xmpuzzle.Boundingbox) {
	return vi.cachedBB
}

/*
func (vi VoxelInstance) GetVoxel() (v *xmpuzzle.Voxel) {
	return vi.voxel
}
*/
