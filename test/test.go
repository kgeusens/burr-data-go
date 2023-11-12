package main

import (
	"fmt"

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

	cache := solver.NewSolverCache(&puzzle, 0)

	assemblies := cache.GetAssemblies()
	fmt.Println(len(assemblies))
	fmt.Println(*assemblies[0][2])
	node := solver.NewNodeFromAssembly(&assemblies[0])
	fmt.Println(node)
	// fmt.Println(cache.CalcCutlerMatrix(node))
}
