package solver

import (
	"log"

	dlx "github.com/Kappeh/dlx"
)

func (sc SolverCache_t) assemble() (solutions [][]int) {
	tempMatrix := sc.GetDLXmatrix()
	dlxMatrix, err := dlx.New(sc.GetNumPrimary(), sc.GetNumSecondary())
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range *tempMatrix {
		err := dlx.AddRow(dlxMatrix, *entry.row...)
		if err != nil {
			log.Fatal(err)
		}
	}
	count := 0
	dlx.ForEachSolution(dlxMatrix, func(row []int) {
		count++
		solutions = append(solutions, row)
	})
	return solutions
}

func (sc *SolverCache_t) GetAssemblies() [][]int {
	if sc.assemblyCache == nil {
		sc.assemblyCache = sc.assemble()
	}
	return sc.assemblyCache
}
