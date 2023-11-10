package xmpuzzle

import (
	"encoding/xml"
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

func (p *Problem) GetShapemap() (result []int) {
	for _, shape := range p.Shapes {
		count := shape.Count
		if count == 0 {
			count = shape.Max
		}
		for i := 0; i < count; i++ {
			result = append(result, shape.Id)
		}
	}
	return
}
