package solver

import (
	xmpuzzle "github.com/kgeusens/go/burr-data/xmpuzzle"
)

type VoxelInstance struct {
	voxel       *xmpuzzle.Voxel
	translation [3]int
	rotation    int
}

func (vi VoxelInstance) NewWorldmap() (wm xmpuzzle.Worldmap) {
	return
}
