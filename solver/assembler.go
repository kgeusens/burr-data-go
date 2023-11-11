package solver

import (
	"log"

	dlx "github.com/Kappeh/dlx"
)

func (sc SolverCache_t) assemble() (solutions [][]*annotation_t) {
	tempMatrix := sc.getDLXmatrix()
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
		solution := []*annotation_t{}
		for _, rowid := range row {
			solution = append(solution, (*tempMatrix)[rowid].annotation)
		}
		solutions = append(solutions, solution)
	})
	return solutions
}

func (sc *SolverCache_t) GetAssemblies() [][]*annotation_t {
	if sc.assemblyCache == nil {
		sc.assemblyCache = sc.assemble()
	}
	return sc.assemblyCache
}
