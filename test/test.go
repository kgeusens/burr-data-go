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
	cache := solver.NewSolverCache(&puzzle, 0)

	assemblies := cache.GetAssemblies()

	for i, a := range assemblies {
		if cache.Solve(&a) {
			fmt.Println("Solution found at ", i)
		}
	}
}
