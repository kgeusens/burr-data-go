package solver

import (
	"log"

	dlx "github.com/Kappeh/dlx"
)

type assembly_t []*annotation_t

func (sc ProblemCache_t) assemble() (solutions []assembly_t) {
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

/*
GetAssemblies returns an array of the possible assemblies of the problem (represented by this cache)
GetAssemblies[x] returns assembly number x
*/
func (sc *ProblemCache_t) GetAssemblies() []assembly_t {
	if sc.assemblyCache == nil {
		sc.assemblyCache = sc.assemble()
	}
	return sc.assemblyCache
}
