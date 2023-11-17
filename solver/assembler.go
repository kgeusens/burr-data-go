package solver

import (
	//	dlx "github.com/Kappeh/dlx"
	"github.com/kgeusens/go/burr-data/dlx"
)

type assembly_t []*annotation_t

func (sc ProblemCache_t) assemble() (solutions []assembly_t) {
	tempMatrix := sc.getDLXmatrix()
	searchConfig := new(dlx.Searchconfig_t)
	searchConfig.NumPrimary = sc.GetNumPrimary()
	searchConfig.NumSecondary = sc.GetNumSecondary()
	searchConfig.NumSolutions = 1000000
	for _, entry := range *tempMatrix {
		searchConfig.AddRow(*entry.row, entry.annotation)
	}
	res := searchConfig.Search()
	for i := range res {
		solution := assembly_t{}
		for _, row := range res[i] {
			solution = append(solution, row.GetData().(*annotation_t))
		}
		solutions = append(solutions, solution)
	}
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
