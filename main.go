package main

import (
	"bufio"
	"flag"
	"fmt"
	m "math"
	"math/rand"
	"os"
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

type chrom [NUM*2 + 1]float64

var (
	fXYNum    = flag.Int("n", 200, "Number of rows in file")
	fMaking   = flag.Bool("m", false, "Making a file at start?")
	fLimit    = flag.Int("l", 10, "Limit in random creation")
	fStart    = flag.Int("s", -10, "Starting num in random creation")
	fChromNum = flag.Int("c", 30, "Number of chroms in population")
	fIter     = flag.Int("i", 5, "Number of populations")

	//debug
	fDebug   = flag.Bool("d", false, "Debug mode")
	fLimitA0 = flag.Int("l0", 50, "Limit in random creation")
	fStartA0 = flag.Int("s0", 0, "Starting num in random creation")
	fLimitA  = flag.Int("l1", 50, "Limit in random creation")
	fStartA  = flag.Int("s1", 0, "Starting num in random creation")
	fLimitB  = flag.Int("l2", 50, "Limit in random creation")
	fStartB  = flag.Int("s2", 0, "Starting num in random creation")
)

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	var generation []*chrom
	var source []xy

	if *fMaking {
		source = makeFile(*fXYNum, func(x float64) float64 {
			return m.Cos(3*x) + m.Sin(x)
		})
	} else {
		source = parseFile()
	}

	start := time.Now()

	generation = populate(*fChromNum)
	fmt.Printf("Populated: %d chroms, %d gens", *fChromNum, NUM*2+1)

	timeStamp(&start)

	var averageFitness, bestChrom float64

	for i := 1; i <= *fIter; i++ {

		fmt.Printf("\n##### Generation %d #####\n", i)

		averageFitness, bestChrom = 0, 0
		for _, v := range generation {
			//fmt.Printf("RawFitness for chrom %d %v is %v \n\n", n, *v, rawFitness(source, v))
			fit := rawFitness(source, v)
			if bestChrom == 0 || bestChrom > fit {
				bestChrom = fit
			}
			averageFitness += fit
			if *fDebug {
				fmt.Printf("%.2f|", fit)
			}
		}
		fmt.Printf("\nAverage Raw fitness: %v, Best chrom: %v", averageFitness/float64(len(generation)), bestChrom)

		//timeStamp(&start)
		timeStamp(&start)

		normal := getNormal(source, generation)
		generation = evolve(source, generation, normal)
		fmt.Print("Next population created.")
		// for _, v := range generation {
		// 	//fmt.Printf("RawFitness for chrom %d %v is %v \n\n", n, *v, rawFitness(source, v))
		// 	fmt.Printf("%v , ", nFitness(source, v, normal))
		// }

		timeStamp(&start)
	}
}

//for debug TODO: real benchmarks
func timeStamp(t *time.Time) {
	duration := time.Since(*t)
	fmt.Printf("\n===Time elapsed: %s ===\n", duration)
	*t = time.Now()
}

//###############################################################################################

func populate(num int) []*chrom {
	g := make([]*chrom, num)
	for i := 0; i < num; i++ {
		g[i] = generateChrom()
	}
	return g
}

//Take and MODIFY population
func evolve(in []xy, gen []*chrom, normal float64) []*chrom {
	//Roulette rand UGLY version
	type MinMax struct {
		min float64
		max float64
	}
	full := make([]MinMax, len(gen))
	var current float64
	for n, v := range gen {
		full[n].min = current
		current += nFitness(in, v, normal)
		full[n].max = current
	}

	next := make([]*chrom, 0, len(gen))
	for i := 0; i < len(gen); i++ {
		rnd := rand.Float64()
		for n, v := range full {
			if rnd >= v.min && rnd < v.max {
				newGen := mutate(gen[n], 80, v.max-v.min)
				newGen = crossover(newGen, next, 80)
				next = append(next, newGen)
			}
		}
	}

	return next
}

//function mutates 1 to *bitsNum* with the chance of *chance*
// func mutateBits(c *chrom, chance int, bitsNum int) *chrom {
// 	if rand.Intn(100) > chance {
// 		return c
// 	}
//
// 	mut := new(chrom)
// 	bits := make([]int, rand.Intn(bitsNum))
// 	for n := range bits {
// 		bits[n] = rand.Intn(len(*c) * 64)
//
// 	}
// 	return mut
// }

func mutate(c *chrom, chance int, close float64) *chrom {
	if rand.Intn(100) > chance {
		return c
	}
	l := len(*c)
	mut := new(chrom)
	*mut = *c
	for i := 0; i < 3; i++ {
		mut[rand.Intn(l-1)] += (rand.Float64()*2 - 1)
	}
	return mut
}

func crossover(c *chrom, next []*chrom, chance int) *chrom {
	if rand.Intn(100) > chance {
		return c
	}
	curLength := len(next) - 1
	if curLength <= 1 {
		return c
	}
	cross := 5 //rand.Intn(len(*c) - 1) //"line" of crossing
	crossed := new(chrom)
	copy(crossed[:cross], c[:cross])
	copy(crossed[cross:], next[rand.Intn(curLength)][cross:])
	return crossed
}

func generateChrom() *chrom {
	c := new(chrom)
	for i := range c {
		//debug
		if *fDebug {
			switch {
			case i == 0:
				c[i] = rand.Float64()*float64(*fLimitA0) + float64(*fStartA0)
			case i%2 > 0:
				c[i] = rand.Float64()*float64(*fLimitA) + float64(*fStartA)
			default:
				c[i] = rand.Float64()*float64(*fLimitB) + float64(*fStartB)
			}
		} else {
			c[i] = rand.Float64()*float64(*fLimit) + float64(*fStart)
		}
	}
	return c
}

func rawFitness(in []xy, c *chrom) float64 {
	var total float64

	for _, v := range in {
		total += m.Abs(v.y - f(v.x, c))
	}

	return total / float64(len(in))
}

func offFitness(in []xy, c *chrom) float64 {
	return 1 / (1 + rawFitness(in, c))
}

func getNormal(in []xy, gen []*chrom) float64 {
	var sum float64
	for _, v := range gen {
		sum += offFitness(in, v)
	}
	return sum
}

func nFitness(in []xy, c *chrom, norm float64) float64 {
	return offFitness(in, c) / norm
}

//Our "fourier" function
func f(x float64, c *chrom) float64 {
	y := c[0] / 2
	for i := 1; i <= NUM; i++ {
		y += c[i*2-1]*m.Cos(float64(i)) + c[i*2]*m.Sin(float64(i))
	}
	return y
}

//###############################################################################################

func makeFile(n int, f func(x float64) float64) []xy {
	out := make([]xy, n)

	for i := 0; i < n; i++ {
		out[i].x = rand.Float64()*m.Pi*2 - m.Pi
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
