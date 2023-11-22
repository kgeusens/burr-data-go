package xmpuzzle

import (
	"encoding/xml"
	"fmt"
	"regexp"

	burrutils "github.com/kgeusens/go/burr-data/burrutils"
)

type Voxel struct {
	XMLName xml.Name             `xml:"voxel"`
	X       burrutils.Distance_t `xml:"x,attr"`
	Y       burrutils.Distance_t `xml:"y,attr"`
	Z       burrutils.Distance_t `xml:"z,attr"`
	Weight  uint                 `xml:"weight,attr"`
	Name    string               `xml:"name,attr"`
	Type    uint                 `xml:"type,attr"`
	Text    string               `xml:",chardata"`
}

func (v Voxel) String() string {
	return fmt.Sprintf("Piece Name:%v (X:%v Y:%v Z:%v) Value:%v", v.Name, v.X, v.Y, v.Z, v.Text)
}

func (v Voxel) GetVoxelState(x, y, z burrutils.Distance_t) (state int8) {
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

/*
CalcSelfSymmetries calculates the rotations that for which a voxel is symmetric.

# Params

none

# Result

symgroupID int: number from 0 to 29 (the id of the symmetrygroup)
*/
func (v Voxel) CalcSelfSymmetries() (symgroupID int) {
	rotSequence := [16]burrutils.Id_t{1, 4, 10, 2, 8, 16, 5, 7, 13, 15, 6, 9, 11, 14, 18, 22}
	wm := v.NewWorldmap()
	bb := wm.CalcBoundingbox()
	bbX, bbY, bbZ := bb.Size()

	symmetryMatrix := 1 // rotation 0
	rbb := NewBoundingbox()
	next := burrutils.Id_t(0)
	rotidx := burrutils.Id_t(0)
	rotlen := burrutils.Id_t(len(rotSequence))

	for next < rotlen {
		rotidx = rotSequence[next]
		bit := 1 << rotidx
		if (symmetryMatrix & bit) == 1 {
			next++
			continue
		}
		symmetric := true
		// calculate new boundingbox and the offset between the boundingboxes
		rotmin := [3]burrutils.Distance_t{}
		rotmax := [3]burrutils.Distance_t{}
		offset := [3]burrutils.Distance_t{0, 0, 0}
		rotmin[0], rotmin[1], rotmin[2] = burrutils.Rotate(bb.Min[0], bb.Min[1], bb.Min[2], rotidx)
		rotmax[0], rotmax[1], rotmax[2] = burrutils.Rotate(bb.Max[0], bb.Max[1], bb.Max[2], rotidx)
		for i := 0; i < 3; i++ {
			rbb.Min[i] = min(rotmin[i], rotmax[i])
			rbb.Max[i] = max(rotmin[i], rotmax[i])
			offset[i] = bb.Min[i] - rbb.Min[i]
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
			if !wm.Has([3]burrutils.Distance_t{rX, rY, rZ}) {
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
	// we need to return the ID of the symmetrygroup, not the group itself
	symgroupID = -1
	for i := range burrutils.SymmetryGroups {
		if burrutils.SymmetryGroups[i] == symmetryMatrix {
			symgroupID = i
			break
		}
	}
	return symgroupID
}

func (v *Voxel) NewWorldmap() Worldmap {
	wm := NewWorldmap()
	for z := burrutils.Distance_t(0); z < v.Z; z++ {
		for y := burrutils.Distance_t(0); y < v.Y; y++ {
			for x := burrutils.Distance_t(0); x < v.X; x++ {
				if s := v.GetVoxelState(x, y, z); s > 0 {
					wm = append(wm, worldmapEntry{[3]burrutils.Distance_t{x, y, z}, s})
				}
			}
		}
	}
	return wm
}

func (v *Voxel) Size() (size int) {
	size = 0
	for _, c := range v.Text {
		if c == '+' || c == '#' {
			size++
		}
	}
	return size
}

func (v *Voxel) Volume() (size int) {
	return int(v.X) * int(v.Y) * int(v.Z)
}
