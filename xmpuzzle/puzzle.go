package xmpuzzle

import (
	"encoding/xml"
	"fmt"
)

type Puzzle struct {
	XMLName  xml.Name  `xml:"puzzle"`
	Version  string    `xml:"version,attr"`
	GridType GridType  `xml:"gridType"`
	Shapes   []Voxel   `xml:"shapes>voxel"`
	Problems []Problem `xml:"problems>problem"`
}

func (p Puzzle) String() string {
	return fmt.Sprintf("Puzzle NumPieces:%v NumProblems:%v", p.NumPieces(), p.NumProblems())
}

func (p *Puzzle) GetPiece(idx int) Voxel {
	return p.Shapes[idx]
}
