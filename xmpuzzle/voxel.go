package xmpuzzle

import (
	"encoding/xml"
	"fmt"
	"regexp"

	burrutils "github.com/kgeusens/go/burr-data/burrutils"
)

type Voxel struct {
	XMLName xml.Name `xml:"voxel"`
	X       int      `xml:"x,attr"`
	Y       int      `xml:"y,attr"`
	Z       int      `xml:"z,attr"`
	Weight  int      `xml:"weight,attr"`
	Name    string   `xml:"name,attr"`
	Type    int      `xml:"type,attr"`
	Text    string   `xml:",chardata"`
}

func (v Voxel) String() string {
	return fmt.Sprintf("Piece Name:%v (X:%v Y:%v Z:%v) Value:%v", v.Name, v.X, v.Y, v.Z, v.Text)
}

func (v Voxel) GetVoxelState(x, y, z int) (state int) {
	if x >= v.X || y >= v.Y || z >= v.Z {
		return 0
	}
	colorlessState := regexp.MustCompile(`/\d+/g`)
	statePositions := colorlessState.ReplaceAllString(v.Text, "")
	switch char := statePositions[x+y*v.X+z*v.X*v.Y]; char {
	case '#':
		state = 1
	case '+':
		state = 2
	default:
		state = 0
	}
	return
}

func (v Voxel) CalcSelfSymmetries() (symmetryMatrix int) {
	rotSequence := [16]int{1, 4, 10, 2, 8, 16, 5, 7, 13, 15, 6, 9, 11, 14, 18, 22}
	wm := v.NewWorldmap()
	bb := wm.CalcBoundingbox()
	bbX, bbY, bbZ := bb.Size()

	symmetryMatrix = 1 // rotation 0
	rbb := NewBoundingbox()
	next := 1
	rotidx := 0

	for next < len(rotSequence) {
		rotidx = rotSequence[next]
		bit := 1 << rotidx
		if (symmetryMatrix & bit) == 1 {
			next++
			continue
		}
		symmetric := true
		// calculate new boundingbox and the offset between the boundingboxes
		rotmin := [3]int{}
		rotmax := [3]int{}
		offset := [3]int{0, 0, 0}
		rotmin[0], rotmin[1], rotmin[2] = burrutils.Rotate(bb.min[0], bb.min[1], bb.min[2], rotidx)
		rotmax[0], rotmax[1], rotmax[2] = burrutils.Rotate(bb.max[0], bb.max[1], bb.max[2], rotidx)
		for i := 0; i < 3; i++ {
			rbb.min[i] = min(rotmin[i], rotmax[i])
			rbb.max[i] = max(rotmin[i], rotmax[i])
			offset[i] = bb.min[i] - rbb.min[i]
		}

		// only continue if boxes have the same size
		rbbX, rbbY, rbbZ := rbb.Size()
		if rbbX != bbX || rbbY != bbY || rbbZ != bbZ {
			symmetric = false
			next++
			continue
		}
		// now check rotations
		wm := v.NewWorldmap()
		for idx := range wm {
			p := wm[idx].position
			rX, rY, rZ := burrutils.Rotate(p[0], p[1], p[2], rotidx)
			rX += offset[0]
			rY += offset[1]
			rZ += offset[2]
			if !wm.Has([3]int{rX, rY, rZ}) {
				symmetric = false
				break
			}
		}
		// when we get here, symmetric determines symmetry in rotidx
		if symmetric {
			newSymmetrygroup := burrutils.RotationToSymmetrygroup[rotidx]
			// add all double rotations to symmetryMatrix
			symmetryMembers := burrutils.HashToRotations(symmetryMatrix)
			newSymmetryMembers := burrutils.HashToRotations(newSymmetrygroup)
			for _, n := range newSymmetryMembers {
				for _, r := range symmetryMembers {
					rn := burrutils.DoubleRotate(r, n)
					if rbit := 1 << rn; (symmetryMatrix & rbit) == 0 {
						symmetryMatrix = symmetryMatrix | rbit
					}
				}
			}
		}
		// continue to the next
		next++
	}

	return
}

func (v *Voxel) NewWorldmap() Worldmap {
	wm := NewWorldmap()
	for z := 0; z < v.Z; z++ {
		for y := 0; y < v.Y; y++ {
			for x := 0; x < v.X; x++ {
				if s := v.GetVoxelState(x, y, z); s > 0 {
					wm = append(wm, worldmapEntry{[3]int{x, y, z}, s})
				}
			}
		}
	}
	return wm
}
