package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/magejiCoder/magejiAoc/input"
	"github.com/magejiCoder/magejiAoc/set"
)

const (
	START  string = "you"
	END    string = "out"
	SERVER string = "svr"
	FFT    string = "fft"
	DAC    string = "dac"
)

type rack struct {
	servers     map[string]*set.Set[string]
	totalRoutes int
}

func newRack() *rack {
	return &rack{
		servers: make(map[string]*set.Set[string]),
	}
}

type state struct {
	node       string
	vFFT, vDAC bool
}

func (r *rack) FFTAndDACRoutes(from string) int {
	memo := make(map[state]int)

	var dfs func(node string, fft, dac bool) int
	dfs = func(node string, fft, dac bool) int {
		if node == END {
			if fft && dac {
				return 1
			}
			return 0
		}

		nextFFT := fft || node == FFT
		nextDAC := dac || node == DAC

		nexts, ok := r.servers[node]
		if !ok || nexts.Size() == 0 {
			return 0
		}

		total := 0
		nexts.Each(func(next string) bool {
			st := state{node: next, vFFT: nextFFT, vDAC: nextDAC}
			if v, ok := memo[st]; ok {
				total += v
				return true
			}
			v := dfs(next, nextFFT, nextDAC)
			memo[st] = v
			total += v
			return true
		})

		return total
	}

	return dfs(from, false, false)
}

func (r *rack) routes(from string) {
	if from == END {
		r.totalRoutes += 1
		return
	}
	nexts, ok := r.servers[from]
	if !ok || nexts.Size() == 0 {
		return
	}
	nexts.Each(func(item string) bool {
		r.routes(item)
		return true
	})
}

func p1() {
	txt := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	r := newRack()
	txt.ReadByLine(ctx, func(line string) error {
		parts := strings.Split(line, ":")
		left := parts[0]
		rights := strings.Split(parts[1], " ")
		var nrs []string
		for _, r := range rights {
			if r == "" {
				continue
			}
			r = strings.TrimSpace(r)
			nrs = append(nrs, r)
		}
		r.servers[left] = set.New(nrs...)
		return nil
	})
	r.routes(START)
	fmt.Printf("p1: %d\n", r.totalRoutes)
}

func p2() {
	txt := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	r := newRack()
	txt.ReadByLine(ctx, func(line string) error {
		parts := strings.Split(line, ":")
		left := parts[0]
		rights := strings.Split(parts[1], " ")
		var nrs []string
		for _, r := range rights[1:] {
			nrs = append(nrs, strings.TrimSpace(r))
		}
		r.servers[left] = set.New(nrs...)
		return nil
	})
	// fmt.Printf("servers:  %v\n", r.servers)
	total := r.FFTAndDACRoutes(SERVER)
	fmt.Printf("p2: %d\n", total)
}

func main() {
	p1()
	p2()
}
