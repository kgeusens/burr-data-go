package xmpuzzle

import (
	"encoding/xml"

	burrutils "github.com/kgeusens/go/burr-data/burrutils"
)

type Problem struct {
	XMLName       xml.Name   `xml:"problem"`
	Name          string     `xml:"name,attr"`
	State         int        `xml:"state,attr"`
	Assemblies    int        `xml:"assemblies,attr"`
	SolutionCount int        `xml:"solutions,attr"`
	Time          int        `xml:"time,attr"`
	Shapes        []Shape    `xml:"shapes>shape"`
	Result        Result     `xml:"result"`
	Solutions     []Solution `xml:"solutions>solution"`
	Comment       Comment    `xml:"comment"`
}

func (p *Problem) NumShapes() int {
	return len(p.Shapes)
}

func (p *Problem) GetShapemap() (result []burrutils.Id_t) {
	for _, shape := range p.Shapes {
		count := shape.Count
		if count == 0 {
			count = shape.Max
		}
		for i := uint8(0); i < count; i++ {
			result = append(result, shape.Id)
		}
	}
	return
}

func (p *Problem) GetPartMinimum(partid burrutils.Id_t) (min uint8) {
	if p.Shapes[partid].Count > 0 {
		min = p.Shapes[partid].Count
	} else {
		min = p.Shapes[partid].Min
	}
	return min
}

func (p *Problem) GetPartMaximum(partid burrutils.Id_t) (max uint8) {
	if p.Shapes[partid].Count > 0 {
		max = p.Shapes[partid].Count
	} else {
		max = p.Shapes[partid].Max
	}
	return max
}

// KG: not implemented yet, returns false
func (p *Problem) MaxHolesDefined() bool {
	return false
}

// KG: not implemented yet, returns 0
func (p *Problem) GetMaxHoles() int {
	return 0
}
