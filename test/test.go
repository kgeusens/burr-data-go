package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/kgeusens/go/burr-data/solver"
	"github.com/kgeusens/go/burr-data/xmpuzzle"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	/*
		dlxMatrix, _ := dlx.New(3, 7)
		dlx.AddRow(dlxMatrix, 0, 3, 4)
		dlx.AddRow(dlxMatrix, 1, 3, 5)
		dlx.AddRow(dlxMatrix, 2, 3, 6)
		dlx.AddRow(dlxMatrix, 0, 4, 5, 6, 7)
		dlx.AddRow(dlxMatrix, 1, 5, 6, 8)
		dlx.AddRow(dlxMatrix, 2, 6, 9)
		dlx.AddRow(dlxMatrix, 0, 7, 8, 9)
		dlx.AddRow(dlxMatrix, 1, 8, 9)
		dlx.AddRow(dlxMatrix, 2, 9)
		dlx.ForEachSolution(dlxMatrix, func(row []int) {
			fmt.Println(row)
		})
	*/

	xmlstring, err := xmpuzzle.ReadFile("./3D Onat.xmpuzzle")
	if err != nil {
		fmt.Println(err)
		return
	}
	// we have an xml string now
	puzzle := xmpuzzle.ParseXML(xmlstring)
	cache := solver.NewProblemCache(&puzzle, 0)
	assemblies := cache.GetAssemblies()
	fmt.Println(len(assemblies), "assemblies to test")
	for i, a := range assemblies {
		res := cache.Solve(a, i)
		if res {
			fmt.Println("Solution at", i)
			//			for _, v := range a {
			//				fmt.Println(*v)
			//			}
		}
	}
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}
