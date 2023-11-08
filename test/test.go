package main

import (
	"fmt"

	xmpuzzle "github.com/kgeusens/go/burr-data/xmpuzzle"
)

func main() {
	xmlstring, err := xmpuzzle.ReadFile("./two face 3.xmpuzzle")
	if err != nil {
		fmt.Println(err)
		return
	}
	// we have an xml string now
	puzzle := xmpuzzle.ParseXML(xmlstring)

	fmt.Println(puzzle)
	for _, v := range puzzle.Shapes {
		fmt.Println(v)
	}

	//	var mapje xmpuzzle.Worldmap
	m := make(xmpuzzle.Worldmap)
	m.Set(xmpuzzle.PointToHash(1, 2, 3), 1000)

	fmt.Println(m)
	m.Rotate(1)
	fmt.Println(m)
	for h := range m {
		x, y, z := xmpuzzle.HashToPoint(h)
		fmt.Println(m, x, y, z)
	}
}
