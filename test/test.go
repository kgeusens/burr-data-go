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
	fmt.Println(puzzle.Shapes[8].GetVoxelState(7, 5, 5))

	m := xmpuzzle.NewWorldmapFromVoxel(&puzzle.Shapes[8])
	for h, v := range m {
		x, y, z := xmpuzzle.HashToPoint(h)
		fmt.Println(h, x, y, z, v)
	}

}
