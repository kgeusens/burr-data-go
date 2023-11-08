package xmpuzzle

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type GridType struct {
	Type int `xml:"type,attr"`
}

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
	colorlessState := regexp.MustCompile(`[#_+]?`)
	statePositions := colorlessState.FindAllStringIndex(v.Text, -1)
	switch char := v.Text[statePositions[x+y*v.X+z*v.X*v.Y][0]]; char {
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
func (v Voxel) CalcSelfSymmetries() (symgroup int) {
	rotSequence := [16]int{1,4,10,2,8,16,5,7,13,15,6,9,11,14,18,22}

}
*/

type Shape struct {
	XMLName xml.Name `xml:"shape"`
	Id      int      `xml:"id,attr"`
	Count   int      `xml:"count,attr"`
	Min     int      `xml:"min,attr"`
	Max     int      `xml:"max,attr"`
	Group   int      `xml:"group,attr"`
}

type Result struct {
	XMLName xml.Name `xml:"result"`
	Id      int      `xml:"id,attr"`
}

type Solution struct {
	XMLName    xml.Name   `xml:"solution"`
	AsmNum     int        `xml:"asmNum,attr"`
	Assembly   Assembly   `xml:"assembly"`
	Separation Separation `xml:"separation"`
}

type Assembly struct {
	XMLName xml.Name `xml:"assembly"`
	Text    string   `xml:",chardata"`
}

type Separation struct {
	XMLName     xml.Name
	Pieces      Pieces       `xml:"pieces"`
	State       []State      `xml:"state"`
	Type        string       `xml:"type,attr"`
	Separations []Separation `xml:"separation"`
}

type Pieces struct {
	XMLName xml.Name `xml:"pieces"`
	Count   int      `xml:"count,attr"`
	Text    string   `xml:",chardata"`
}

type StatePositions struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
}

type State struct {
	XMLName xml.Name       `xml:"state"`
	DX      StatePositions `xml:"dx"`
	DY      StatePositions `xml:"dy"`
	DZ      StatePositions `xml:"dz"`
}

func (s *State) X() []string {
	return strings.Split(s.DX.Text, " ")
}
func (s *State) Y() []string {
	return strings.Split(s.DY.Text, " ")
}
func (s *State) Z() []string {
	return strings.Split(s.DZ.Text, " ")
}

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

type Comment struct {
	XMLName xml.Name `xml:"comment"`
	Text    string   `xml:",chardata"`
}

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

func ParseXML(xmlstring string) Puzzle {
	xmldata := []byte(xmlstring)
	var p Puzzle
	xml.Unmarshal(xmldata, &p)
	return p
}

func (p *Puzzle) NumPieces() int {
	return len(p.Shapes)
}

func (p *Puzzle) NumProblems() int {
	return len(p.Problems)
}

func ReadFile(filename string) (xml string, err error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("error")
		return
	}

	b := bytes.NewBuffer(f)

	var r io.Reader
	r, err = gzip.NewReader(b)
	if err != nil {
		return
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return
	}

	xml = string(resB.Bytes())

	//	fmt.Println(xml)
	return
}
