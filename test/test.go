package main

import (
	"fmt"
	"slices"

	burrutils "github.com/kgeusens/go/burr-data/burrutils"
	solver "github.com/kgeusens/go/burr-data/solver"
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

	r := puzzle.Shapes[8].NewWorldmap()
	p := puzzle.Shapes[4].NewWorldmap()
	fmt.Println(len(r))
	fmt.Println(r)
	fmt.Println()
	fmt.Println(len(p))
	fmt.Println(p)
	fmt.Println()

	syms := puzzle.Shapes[8].CalcSelfSymmetries()
	fmt.Println(burrutils.HashToRotations(syms))

	//fmt.Println(solver.NewVoxelinstance(&puzzle.Shapes[4], 0).GetWorldmap())

	cache := solver.NewSolverCache(&puzzle, 0)
	dlx := cache.GetDLXrow(4, 0, 0, 0, 0)
	slices.Sort(dlx)
	fmt.Println(len(dlx))
	fmt.Println(dlx)
	pinstance := cache.GetShapeInstance(4, 8)
	fmt.Println(*pinstance.GetWorldmap())

	moves := cache.GetMaxValues(0, 0, 1, 0, 0, 1, 1)
	fmt.Println(*moves)

	dlxMatrix := cache.GetDLXmatrix()
	fmt.Println(len(*dlxMatrix))
}
