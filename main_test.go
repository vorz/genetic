package main

import (
	"fmt"
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

//Test normal fitness
func TestNormFitness(t *testing.T) {
	s := parseFile()
	g := populate(100)
	normal := getNormal(s, g)
	var sum float64
	for _, v := range g {
		sum += nFitness(s, v, normal)
	}
	number := fmt.Sprintf("%.2f", sum) //sort of rounding, LOL
	if number != "1.00" {
		t.Errorf("Sum of normal fitnesses != 1: %s", number)
	}
}

func TestMutate(t *testing.T) {
	s := parseFile()
	g := populate(1)
	normal := getNormal(s, g)
	//fmt.Println(g[0])
	g = evolve(s, g, normal)
	//fmt.Println(g[0])
}

func TestCrossover(t *testing.T) {
	s := parseFile()
	g := populate(5)
	normal := getNormal(s, g)
	fmt.Println(g[0])
	fmt.Println(g[1])
	fmt.Println(g[2])
	fmt.Println(g[3])
	fmt.Println(g[4])
	fmt.Println("=======================================================")
	g = evolve(s, g, normal)
	fmt.Println(g[0])
	fmt.Println(g[1])
	fmt.Println(g[2])
	fmt.Println(g[3])
	fmt.Println(g[4])
}
