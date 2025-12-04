package main

import (
	"context"
	"fmt"

	"github.com/magejiCoder/magejiAoc/grid"
	"github.com/magejiCoder/magejiAoc/input"
)

type paperGrid struct {
	m      grid.VecMatrix[byte]
	papers map[grid.Vec]struct{}
}

func NewGrid() paperGrid {
	return paperGrid{
		m:      grid.NewVecMatrix[byte](),
		papers: make(map[grid.Vec]struct{}),
	}
}

func (pg *paperGrid) clean() (int, map[grid.Vec]struct{}) {
	if len(pg.papers) == 0 {
		return 0, nil
	}

	paperSet := pg.papers
	deltas := [8]grid.Vec{
		{X: -1, Y: -1},
		{X: 0, Y: -1},
		{X: 1, Y: -1},
		{X: -1, Y: 0},
		{X: 1, Y: 0},
		{X: -1, Y: 1},
		{X: 0, Y: 1},
		{X: 1, Y: 1},
	}

	pp := 0
	removed := make(map[grid.Vec]struct{})
	for v := range pg.papers {
		cnt := 0
		for _, d := range deltas {
			n := grid.Vec{X: v.X + d.X, Y: v.Y + d.Y}
			if _, ok := paperSet[n]; ok {
				cnt++
				if cnt >= 4 {
					break
				}
			}
		}
		if cnt < 4 {
			removed[v] = struct{}{}
			pp++
		}
	}
	return pp, removed
}

func p1() {
	t := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	g := NewGrid()
	var y int
	t.ReadByLine(ctx, func(line string) error {
		for x := 0; x < len(line); x++ {
			g.m.Add(grid.Vec{
				X: x,
				Y: y,
			}, line[x])
			if line[x] == '@' {
				g.papers[grid.Vec{
					X: x,
					Y: y,
				}] = struct{}{}
			}
		}
		y++
		return nil
	})
	pp, _ := g.clean()
	fmt.Printf("p1: %d\n", pp)
}

func p2() {
	t := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	g := NewGrid()
	var y int
	t.ReadByLine(ctx, func(line string) error {
		for x := 0; x < len(line); x++ {
			g.m.Add(grid.Vec{
				X: x,
				Y: y,
			}, line[x])
			if line[x] == '@' {
				g.papers[grid.Vec{
					X: x,
					Y: y,
				}] = struct{}{}
			}
		}
		y++
		return nil
	})
	var total int
	for {
		rc, removed := g.clean()
		if rc == 0 {
			break
		}
		// fmt.Printf("removed: %d\n", rc)
		total += rc
		for v := range removed {
			delete(g.papers, v)
		}
	}
	fmt.Printf("p2: %d\n", total)
}

func main() {
	p1()
	p2()
}
