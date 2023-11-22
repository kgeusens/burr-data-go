package main_test

import (
	"testing"
)

const theSize = 50000

type theType int32

var gs = make([]theType, theSize) // Global slice
var ga [theSize]theType           // Global array

func BenchmarkSliceGlobal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j, v := range gs {
			gs[j]++
			gs[j] = gs[j] + v + 10
			gs[j] += v
		}
	}
}

func BenchmarkArrayGlobal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j, v := range ga {
			ga[j]++
			ga[j] = ga[j] + v + 10
			ga[j] += v
		}
	}
}

func BenchmarkSliceLocal(b *testing.B) {
	var s = make([]theType, theSize)
	for i := 0; i < b.N; i++ {
		for j, v := range s {
			s[j]++
			s[j] = s[j] + v + 10
			s[j] += v
		}
	}
}

func BenchmarkArrayLocal(b *testing.B) {
	var a [theSize]theType
	for i := 0; i < b.N; i++ {
		for j, v := range a {
			a[j]++
			a[j] = a[j] + v + 10
			a[j] += v
		}
	}
}
