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

	"github.com/pkg/profile"
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
	fXYNum      = flag.Int("n", 200, "Number of rows in file")
	fMaking     = flag.Bool("m", false, "Making a file at start?")
	fSimulate   = flag.Bool("sim", false, "Simulate function when creating file?")
	fCross      = flag.Bool("cross", false, "Is crossover using?")
	fGenStd     = flag.Int("std", 1, "Standard devotion when generating population")
	fMean       = flag.Int("mean", 0, "Mean value when generating population")
	fdnaNum     = flag.Int("c", 30, "Number of dnas in population")
	fIter       = flag.Int("i", 999999, "Number of populations")
	fTournament = flag.Int("t", 0, "Use tournament selection if t > 0")
	fTimeOut    = flag.Int("w", 20, "Timeout for main loop")

	//debug
	fDebug = flag.Bool("d", false, "Debug mode")
)

func main() {

	defer profile.Start(profile.MemProfile, profile.ProfilePath("./pro")).Stop()

	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	var generation []*dna
	var source []xy

	if *fMaking {
		if !(*fSimulate) {
			source = makeFile(*fXYNum, func(x float64) float64 {
				return m.Sin(x * 3) //m.Pow(x, 2)
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

	var averageFitness, curBestDna float64
	var bestDna *dna
	var timeBest time.Time
	var curStd = float64(*fGenStd) //curStd is current standar dev for gaussian distrubtion

	//Main loop
	for i := 1; i <= *fIter; i++ {

		fmt.Printf("\n##### Generation %d #####\n", i)

		averageFitness, curBestDna = 0, 0
		var bestDnaNum int
		for n, v := range generation {
			fit := v.fitness
			if curBestDna == 0 || curBestDna > fit {
				curBestDna = fit
				bestDnaNum = n
			}
			averageFitness += fit
			if *fDebug {
				fmt.Printf("%.2f|", fit)
			}
		}

		if i == 1 || bestDna.fitness > curBestDna {
			bestDna = generation[bestDnaNum]
			timeBest = time.Now()
		} else {
			if time.Since(timeBest) > time.Second*time.Duration(*fTimeOut) { // if program cant find better dna in 20 sec it ends
				break
			}
		}

		//every 50 iterations
		if i%50 == 0 {
			curStd /= 2
		}

		fmt.Printf("\nAverage Raw fitness: %v, Best dna in current population: %v, Best overall dna: %v, Std: %v", averageFitness/float64(len(generation)), curBestDna, bestDna.fitness, curStd)
		fmt.Printf("\nBest dna: %v", bestDna.gene)

		if *fTournament > 0 {
			generation = tournamentEvolve(source, generation, curStd, *fTournament)
		} else {
			normal := getNormal(generation)
			generation = evolve(source, generation, normal, curStd)
		}

		timeStamp(&start)

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

//Create first population
func populate(in []xy, num int) []*dna {
	g := make([]*dna, num)
	for i := 0; i < num; i++ {
		g[i] = generateDna()
		rawFitness(in, g[i])
	}
	return g
}

//Take and MODIFY population by roulette selection method
func evolve(in []xy, gen []*dna, normal float64, std float64) []*dna {
	//Roulette selection UGLY version TODO optimization
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
				newGen, isNew := mutate(gen[n], 75, std)
				next = append(next, newGen)
				if isNew {
					rawFitness(in, newGen)
				}
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

//Evolve with tournament selection, take random [:window] from generation and leave 1 best
func tournamentEvolve(in []xy, gen []*dna, std float64, window int) []*dna {
	if window < 2 {
		window = 2
	}
	l := len(gen)
	next := make([]*dna, l)

	for n := range gen {
		r := rand.Intn(l - 1)
		var exhib []*dna
		if (r + window) < l-1 {
			exhib = gen[r : r+window]
		} else {
			exhib = gen[r-window : r]
		}
		var newDna *dna
		for n, v := range exhib {
			if newDna == nil || newDna.fitness > v.fitness {
				newDna = exhib[n]
			}
		}
		var isNew bool
		newDna, isNew = mutate(newDna, 75, std)
		if isNew {
			rawFitness(in, newDna)
		}
		next[n] = newDna
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
func mutate(c *dna, chance int, std float64) (*dna, bool) {
	if rand.Intn(100) > chance {
		return c, false
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
	return mut, true
}

//dumb version of crossover
func crossover(dna1, dna2 *dna) *dna {

	if rand.Intn(2) == 0 {
		dna1, dna2 = dna2, dna1
	}

	cross := rand.Intn(len((*dna1).gene) - 1) //"line" of crossing
	crossed := new(dna)
	copy(crossed.gene[:cross], dna1.gene[:cross])
	copy(crossed.gene[cross:], dna2.gene[cross:])
	return crossed

}

//Generate single dna
func generateDna() *dna {
	c := new(dna)
	for i := range c.gene {
		c.gene[i] = rand.NormFloat64()*float64(*fGenStd) + float64(*fMean) //Gaussian dist
	}
	return c
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

//Calculate and change raw fitness for a single dna
func rawFitness(in []xy, c *dna) {
	var total float64
	for _, v := range in {
		total += m.Pow(v.y-f(v.x, c), 2) //Least squares sort of
	}
	c.fitness = total
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
		out[i].x = rand.Float64()*20 - 10
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
