package main

import (
	"fmt"

	burrutils "github.com/kgeusens/go/burr-data/burrutils"
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
	/*
		r := xmpuzzle.NewWorldmapFromVoxel(&puzzle.Shapes[8])
		p := xmpuzzle.NewWorldmapFromVoxel(&puzzle.Shapes[0])
		fmt.Println(r)
		fmt.Println(p)
		dlx := xmpuzzle.GetDLXmap(r, p)
		slices.Sort(dlx)
		fmt.Println(dlx)
	*/
	syms := puzzle.Shapes[8].CalcSelfSymmetries()
	fmt.Println(burrutils.HashToRotations(syms))
}
