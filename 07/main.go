package main

import (
	"context"
	"fmt"

	"github.com/magejiCoder/magejiAoc/grid"
	"github.com/magejiCoder/magejiAoc/input"
)

type manifold struct {
	startAt  grid.Vec
	splitter map[grid.Vec]bool
	maxY     int
	maxX     int

	pathFrom map[grid.Vec]int
}

func (m *manifold) move(
	startAt grid.Vec,
) {
	next := startAt.Add(grid.Vec{
		X: 0,
		Y: 1,
	})
	if next.Y > m.maxY {
		return
	}

	visit, ok := m.splitter[next]
	if ok && !visit {
		var sp1, sp2 bool
		if next.X-1 >= 0 {
			m.move(next.Add(grid.Vec{X: -1, Y: 0}))
			sp1 = true
		}
		if next.X+1 <= m.maxX {
			m.move(next.Add(grid.Vec{X: 1, Y: 0}))
			sp2 = true
		}
		if sp1 && sp2 {
			m.splitter[next] = true
		}
	}
	if !ok {
		m.move(next)
	}
}

func (m *manifold) move2(
	startAt grid.Vec,
) int {
	next := startAt.Add(grid.Vec{
		X: 0,
		Y: 1,
	})
	if next.Y > m.maxY {
		return 1
	}

	if v, ok := m.pathFrom[next]; ok {
		return v
	}

	var total int
	_, ok := m.splitter[next]
	if ok {
		if next.X-1 >= 0 {
			left := next.Add(grid.Vec{X: -1, Y: 0})
			total += m.move2(left)
		}
		if next.X+1 <= m.maxX {
			right := next.Add(grid.Vec{X: 1, Y: 0})
			total += m.move2(right)
		}
	}
	if !ok {
		total += m.move2(next)
	}
	m.pathFrom[next] = total
	return total
}

func p1() {
	txt := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	m := &manifold{
		splitter: make(map[grid.Vec]bool),
	}
	var y int
	var maxX int
	txt.ReadByLine(ctx, func(line string) error {
		maxX = len(line) - 1
		for x, c := range line {
			if c == 'S' {
				m.startAt = grid.Vec{X: x, Y: y}
			}
			if c == '^' {
				m.splitter[grid.Vec{
					X: x,
					Y: y,
				}] = false
			}
		}
		y++
		return nil
	})
	m.maxY = y - 1
	m.maxX = maxX
	m.move(m.startAt)
	var splitTimes int
	for _, visit := range m.splitter {
		if visit {
			splitTimes++
		}
	}
	fmt.Printf("p1: %d\n", splitTimes)
}

func p2() {
	txt := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	m := &manifold{
		splitter: make(map[grid.Vec]bool),
		pathFrom: make(map[grid.Vec]int),
	}
	var y int
	var maxX int
	txt.ReadByLine(ctx, func(line string) error {
		maxX = len(line) - 1
		for x, c := range line {
			if c == 'S' {
				m.startAt = grid.Vec{X: x, Y: y}
			}
			if c == '^' {
				m.splitter[grid.Vec{
					X: x,
					Y: y,
				}] = false
			}
		}
		y++
		return nil
	})
	m.maxY = y - 1
	m.maxX = maxX
	total := m.move2(m.startAt)
	fmt.Printf("p2: %d\n", total)
}

func main() {
	p1()
	p2()
}
