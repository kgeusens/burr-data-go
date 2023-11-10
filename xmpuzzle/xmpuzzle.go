package xmpuzzle

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

type GridType struct {
	Type int `xml:"type,attr"`
}

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

	xml = resB.String()

	//	fmt.Println(xml)
	return
}
