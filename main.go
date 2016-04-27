package main

import (
	"bufio"
	"flag"
	"fmt"
	m "math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

//NUM total number of coefficents = NUM*2+1
const NUM = 5

type xy struct {
	x float64
	y float64
}

type dna struct {
	gene    [NUM*2 + 1]float64
	fitness float64
}

//Just for sorting purpose
type dnaCode []*dna

func (d dnaCode) Len() int           { return len(d) }
func (d dnaCode) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d dnaCode) Less(i, j int) bool { return d[i].fitness < d[j].fitness }

var (
	fXYNum    = flag.Int("n", 200, "Number of rows in file")
	fMaking   = flag.Bool("m", false, "Making a file at start?")
	fSimulate = flag.Bool("sim", false, "Simulate function when creating file?")
	fCross    = flag.Bool("cross", false, "Is crossover using?")
	fGenStd   = flag.Int("std", 1, "Standard devotion when generating population")
	fMean     = flag.Int("mean", 0, "Mean value when generating population")
	fdnaNum   = flag.Int("c", 30, "Number of dnas in population")
	fIter     = flag.Int("i", 999999, "Number of populations")

	//debug
	fDebug = flag.Bool("d", false, "Debug mode")
)

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	var generation []*dna
	var source []xy

	if *fMaking {
		if !(*fSimulate) {
			source = makeFile(*fXYNum, func(x float64) float64 {
				return x * 5
			})
		} else {
			source = makeFile(*fXYNum, func(x float64) float64 {
				return 1*m.Cos(1*x) + 2*m.Cos(2*x) + 3*m.Cos(3*x) + 4*m.Cos(4*x) + 5*m.Cos(5*x)
			})
		}
	} else {
		source = parseFile()
	}

	start := time.Now()
	totalTime := time.Now()

	generation = populate(source, *fdnaNum)
	fmt.Printf("Populated: %d dnas, %d gens", *fdnaNum, NUM*2+1)

	timeStamp(&start)

	var averageFitness, lastAvFitness, bestdna, curBest float64
	var timeBest time.Time
	var curStd float64

	for i := 1; i <= *fIter; i++ {

		fmt.Printf("\n##### Generation %d #####\n", i)

		lastAvFitness = averageFitness
		averageFitness, bestdna = 0, 0
		var bestDnaNum int
		for n, v := range generation {
			//fmt.Printf("RawFitness for dna %d %v is %v \n\n", n, *v, rawFitness(source, v))
			fit := v.fitness
			if bestdna == 0 || bestdna > fit {
				bestdna = fit
				bestDnaNum = n
			}
			averageFitness += fit
			if *fDebug {
				fmt.Printf("%.2f|", fit)
			}
		}
		fmt.Printf("\nAverage Raw fitness: %v, Best dna in current population: %v, Best overall dna: %v, Std: %v", averageFitness/float64(len(generation)), bestdna, curBest, curStd)
		fmt.Printf("\nBest dna: %v", generation[bestDnaNum].gene)

		//timeStamp(&start)

		//curStd is current standar dev for gaussian distrubtion
		if i == 1 {
			curStd = 1.0
		} else {
			if averageFitness/lastAvFitness < 1 && curStd > 0.05 {
				if (curStd - averageFitness/lastAvFitness) > 0.05 {
					curStd -= 0.05
				} else {
					curStd -= averageFitness / lastAvFitness
				}
			}
		}

		normal := getNormal(generation)
		generation = evolve(source, generation, normal, 1.0)
		fmt.Print("Next population created.")

		timeStamp(&start)
		if curBest > bestdna || i == 1 {
			curBest = bestdna
			timeBest = time.Now()
		} else {
			if time.Since(timeBest) > time.Second*10 { // if program cant find better dna in 10 sec it ends
				break
			}
		}
	}

	duration := time.Since(totalTime)
	fmt.Printf("\n--------------------------\nTotal time: %s\n", duration)

}

//for debug TODO: real benchmarks
func timeStamp(t *time.Time) {
	duration := time.Since(*t)
	fmt.Printf("\n===Time elapsed: %s ===\n", duration)
	*t = time.Now()
}

//###############################################################################################

//Crate first population
func populate(in []xy, num int) []*dna {
	g := make([]*dna, num)
	for i := 0; i < num; i++ {
		g[i] = generatedna()
		rawFitness(in, g[i])
	}
	return g
}

//Take and MODIFY population
func evolve(in []xy, gen []*dna, normal float64, std float64) []*dna {
	//Roulette rand UGLY version TODO optimization
	type MinMax struct {
		min float64
		max float64
	}
	full := make([]MinMax, len(gen))
	var current float64
	for n, v := range gen {
		full[n].min = current
		current += nFitness(v, normal)
		full[n].max = current
	}

	next := make([]*dna, 0, len(gen))
	for i := 0; i < len(gen); i++ {
		rnd := rand.Float64()
		for n, v := range full {
			if rnd >= v.min && rnd < v.max {
				newGen := mutate(gen[n], 75, std)
				//newGen = crossover(newGen, next, 80)
				next = append(next, newGen)
				rawFitness(in, newGen)
			}
		}
	}

	//Crossover is off by default coz its lame
	if *fCross {
		var cross = dnaCode(next)
		sort.Sort(cross)
		crossedDna := crossover(cross[0], cross[1])
		rawFitness(in, crossedDna)
		next[rand.Intn(len(next)-1)] = crossedDna
	}

	return next
}

//function mutates 1 to *bitsNum* with the chance of *chance*
// func mutateBits(c *dna, chance int, bitsNum int) *dna {
// 	if rand.Intn(100) > chance {
// 		return c
// 	}
//
// 	mut := new(dna)
// 	bits := make([]int, rand.Intn(bitsNum))
// 	for n := range bits {
// 		bits[n] = rand.Intn(len(*c) * 64)
//
// 	}
// 	return mut
// }

//Standard mutation with gaussian distrubtion of genes
func mutate(c *dna, chance int, std float64) *dna {
	if rand.Intn(100) > chance {
		return c
	}
	if std == 0 {
		std = 1.0
	}
	mut := new(dna)
	*mut = *c
	for n := range mut.gene {
		if rand.Intn(100) <= 30 { //so for each gene its 30% to mutate
			mut.gene[n] = rand.NormFloat64()*std + mut.gene[n] //mean is current gene val, std dev is 1 by def
		}
	}
	return mut
}

//most dumb version of crossover
func crossover(dna1, dna2 *dna) *dna {

	if rand.Intn(1) == 0 {
		dna1, dna2 = dna2, dna1
	}

	cross := rand.Intn(len((*dna1).gene) - 1) //"line" of crossing
	crossed := new(dna)
	copy(crossed.gene[:cross], dna1.gene[:cross])
	copy(crossed.gene[cross:], dna2.gene[cross:])
	return crossed

}

//Generate single dna
func generatedna() *dna {
	c := new(dna)
	for i := range c.gene {
		c.gene[i] = rand.NormFloat64()*float64(*fGenStd) + float64(*fMean) //Gaussian dist
	}
	return c
}

//Calculate and change raw fitness for a single dna
func rawFitness(in []xy, c *dna) {
	var total float64

	for _, v := range in {
		total += m.Pow(v.y-f(v.x, c), 2) //Least squares sort of
	}

	c.fitness = total
}

func offFitness(c *dna) float64 {
	return 1 / (1 + c.fitness)
}

func getNormal(gen []*dna) float64 {
	var sum float64
	for _, v := range gen {
		sum += offFitness(v)
	}
	return sum
}

func nFitness(c *dna, norm float64) float64 {
	return offFitness(c) / norm
}

//Our "fourier" function
func f(x float64, c *dna) float64 {
	y := c.gene[0] / 2
	for i := 1; i <= NUM; i++ {
		y += c.gene[i*2-1]*m.Cos(float64(i)*x) + c.gene[i*2]*m.Sin(float64(i)*x)
	}
	return y
}

//###############################################################################################

func makeFile(n int, f func(x float64) float64) []xy {
	out := make([]xy, n)

	for i := 0; i < n; i++ {
		out[i].x = rand.Float64()*20 - 10 //*m.Pi*2 - m.Pi
		out[i].y = f(out[i].x)
	}

	file, err := os.Create("file.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fmt.Fprintln(file, n)

	for _, val := range out {
		fmt.Fprint(file, val.x)
		fmt.Fprint(file, " ")
		fmt.Fprintln(file, val.y)
	}
	return out
}

func parseFile() []xy {
	f, err := os.Open("file.txt")
	if err != nil {
		panic(err)
	}
	scan := bufio.NewScanner(f)
	scan.Scan()
	n, _ := strconv.Atoi(scan.Text())

	out := make([]xy, n)

	i := 0
	for scan.Scan() {
		vals := strings.Split(scan.Text(), " ")
		if i <= n {
			out[i].x, _ = strconv.ParseFloat(vals[0], 64)
			out[i].y, _ = strconv.ParseFloat(vals[1], 64)
			i++
		} else {
			break
		}
	}
	return out
}
