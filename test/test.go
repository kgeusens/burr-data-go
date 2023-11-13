package main

import (
	"fmt"

	solver "github.com/kgeusens/go/burr-data/solver"
	xmpuzzle "github.com/kgeusens/go/burr-data/xmpuzzle"
)

func main() {
	xmlstring, err := xmpuzzle.ReadFile("./Misused Key.xmpuzzle")
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

	cache := solver.NewSolverCache(&puzzle, 0)

	assemblies := cache.GetAssemblies()
	fmt.Println(len(assemblies))
	fmt.Println(*assemblies[0][2])
	node := solver.NewNodeFromAssembly(&assemblies[0])
	fmt.Println(node)

	cache.Solve(&assemblies[0])

	for i, a := range assemblies {
		if cache.Solve(&a) {
			fmt.Println("Solution found at ", i)
		}
	}
}
