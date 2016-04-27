package main

import (
	"fmt"
	m "math"
	"testing"
)

//Testing makeFile and parseFile and their returning vals
func TestMakeParseFile(t *testing.T) {
	var out1, out2 []xy
	out1 = makeFile(500, func(x float64) float64 {
		return x * 3
	})
	out2 = parseFile()
	if len(out1) != len(out2) {
		t.Errorf("Length of xy from making and parsing differs (%v vs %v)", len(out1), len(out2))
	}
	for n, v := range out1 {
		if v.x != out2[n].x || v.y != out2[n].y {
			t.Errorf("Values of xy differs at %d element: %v vs %v", n, v, out2[n])
		}
	}
}

//Test raw fitness
func TestRawFitness(t *testing.T) {
	source := makeFile(100, func(x float64) float64 {
		return 1*m.Cos(1*x) + 2*m.Cos(2*x) + 3*m.Cos(3*x) + 4*m.Cos(4*x) + 5*m.Cos(5*x)
	})

	var chromosome dna
	chromosome.gene = [11]float64{0.0, 1.0, 0.0, 2.0, 0.0, 3.0, 0.0, 4.0, 0.0, 5.0, 0.0} //NUM should be 5
	rawFitness(source, &chromosome)
	Y := f(source[0].x, &chromosome)
	if Y != source[0].y {
		t.Errorf("f(x) != y from data, %v != %v", Y, source[0].y)
	}
	if chromosome.fitness != 0 {
		t.Errorf("Raw fitness isnt working, %v != 0", chromosome.fitness)
	}
}

//Test normal fitness
func TestNormFitness(t *testing.T) {
	s := parseFile()
	g := populate(s, 100)
	normal := getNormal(g)
	var sum float64
	for _, v := range g {
		sum += nFitness(v, normal)
	}
	number := fmt.Sprintf("%.2f", sum) //sort of rounding, LOL
	if number != "1.00" {
		t.Errorf("Sum of normal fitnesses != 1: %s", number)
	}
}

// func TestMutate(t *testing.T) {
// 	s := parseFile()
// 	g := populate(s, 1)
// 	normal := getNormal(g)
// 	//fmt.Println(g[0])
// 	g = evolve(s, g, normal)
// 	//fmt.Println(g[0])
// }
