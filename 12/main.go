package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/magejiCoder/magejiAoc/grid"
	"github.com/magejiCoder/magejiAoc/input"
)

type Farm struct {
	presents []present
	regions  []region
}

type present struct {
	// shape 为 3*3 的矩阵
	shape grid.VecMatrix[byte]
}

type region struct {
	cols, rows int
	// ava 表示可用区域
	ava          grid.VecMatrix[struct{}]
	needPresents []acquiredPresent
}

type acquiredPresent struct {
	index  int
	number int
}

func (r *region) fill(presents []present) bool {
	if len(presents) == 0 {
		return false
	}

	// 计算所有 presents 的格点总数
	totalArea := 0
	for _, p := range presents {
		totalArea += len(p.shape)
	}

	return totalArea <= len(r.ava)
}

func (f Farm) validRegions() int {
	var valid int
	for _, r := range f.regions {
		var presents []present
		for _, ap := range r.needPresents {
			for i := 0; i < ap.number; i++ {
				presents = append(presents, f.presents[ap.index])
			}
		}
		if ok := r.fill(presents); ok {
			valid++
		}
	}
	return valid
}

func p1() {
	txt := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	txt.ReadByBlock(ctx, "\n\n", func(block []string) error {
		// fist N blocks are presents (shape)
		var presents []present
		for i := 0; i < len(block)-1; i++ {
			p := present{
				shape: grid.NewVecMatrix[byte](),
			}
			// per line
			parts := strings.Split(block[i], "\n")
			// ignore presents index string
			for li, line := range parts[1:] {
				for co, col := range line {
					if col == '#' {
						p.shape.Add(grid.Vec{X: li, Y: co}, byte(col))
					}
				}
			}
			presents = append(presents, p)
		}
		// last 1 block are region
		parts := strings.Split(block[len(block)-1], "\n")
		var regions []region
		for _, line := range parts {
			r := region{
				ava: grid.NewVecMatrix[struct{}](),
			}
			lps := strings.Split(line, ":")
			// part 0 is the grid
			gridParts := strings.Split(lps[0], "x")
			for i := 0; i < input.Atoi(gridParts[0]); i++ {
				for j := 0; j < input.Atoi(gridParts[1]); j++ {
					r.ava.Add(grid.Vec{X: i, Y: j}, struct{}{})
				}
			}
			// gridParts[0] is cols (width), gridParts[1] is rows (height)
			// Our coordinate system uses X as row index and Y as col index:
			//   X in [0, rows), Y in [0, cols)
			r.cols = input.Atoi(gridParts[0])
			r.rows = input.Atoi(gridParts[1])

			// part 1 is the needed presents
			pparts := strings.Split(lps[1], " ")
			for i, p := range pparts[1:] {
				n := input.Atoi(p)
				if n == 0 {
					continue
				}
				r.needPresents = append(r.needPresents, acquiredPresent{
					index:  i,
					number: n,
				})
			}
			regions = append(regions, r)
		}
		f := Farm{
			presents: presents,
			regions:  regions,
		}

		fmt.Printf("p1: %d\n", f.validRegions())
		return nil
	})
}

func main() {
	p1()
}
