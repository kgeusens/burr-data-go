package xmpuzzle

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	burrutils "github.com/kgeusens/go/burr-data/burrutils"
)

type GridType struct {
	Type int `xml:"type,attr"`
}

type Shape struct {
	XMLName xml.Name       `xml:"shape"`
	Id      burrutils.Id_t `xml:"id,attr"`
	Count   uint8          `xml:"count,attr"`
	Min     uint8          `xml:"min,attr"`
	Max     uint8          `xml:"max,attr"`
	Group   uint8          `xml:"group,attr"`
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

func ParseXML(xmlstring string) Puzzle {
	xmldata := []byte(xmlstring)
	var p Puzzle
	xml.Unmarshal(xmldata, &p)
	return p
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
