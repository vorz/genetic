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
	fLimit    = flag.Int("l", 50, "Limit in random creation")
	fStart    = flag.Int("s", 0, "Starting num in random creation")
	fChromNum = flag.Int("c", 15, "Number of chroms in population")

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
			return x * 3
		})
	} else {
		source = parseFile()
	}
	fmt.Println(*fChromNum)
	generation = populate(*fChromNum)

	for n, v := range generation {
		fmt.Printf("RawFitness for chrom %d %v is %v \n\n", n, *v, rawFitness(source, v))
	}

}

func populate(num int) []*chrom {
	g := make([]*chrom, num)
	for i := 0; i < num; i++ {
		g[i] = generateChrom()
	}
	return g
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
