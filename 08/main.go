package main

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/magejiCoder/magejiAoc/grid"
	"github.com/magejiCoder/magejiAoc/input"
	"github.com/magejiCoder/magejiAoc/set"
)

type junctionBox struct {
	all []grid.Vector3D[int]
}

func euclideanDistance(v1, v2 grid.Vector3D[int]) float64 {
	return math.Sqrt(math.Pow(float64(v1.X-v2.X), 2) + math.Pow(float64(v1.Y-v2.Y), 2) + math.Pow(float64(v1.Z-v2.Z), 2))
}

type conn struct {
	j1, j2 grid.Vector3D[int]
	dist   float64
}

func (jb junctionBox) connect(times int) int {
	all := jb.all
	var conns []conn
	for i := range all {
		for j := i + 1; j < len(all); j++ {
			dist := euclideanDistance(all[i], all[j])
			conns = append(conns, conn{
				j1:   all[i],
				j2:   all[j],
				dist: dist,
			})
		}
	}
	sort.Slice(conns, func(i, j int) bool {
		return conns[i].dist < conns[j].dist
	})
	// fmt.Printf("conn: %v\n", conns)
	jbs := []*set.Set[grid.Vector3D[int]]{}
	for i, conn := range conns[:times] {
		fmt.Printf("time: %d\n", i)
		var merge []*set.Set[grid.Vector3D[int]]
		for index, b := range jbs {
			if b.Has(conn.j1) || b.Has(conn.j2) {
				b.Add(conn.j1)
				b.Add(conn.j2)
				merge = append(merge, b)
				jbs = append(jbs[:index], jbs[index+1:]...)
			}
		}
		if len(merge) == 0 {
			jbs = append(jbs, set.New(conn.j1, conn.j2))
		} else {
			jbs = append(jbs, set.Union(merge...))
		}
		// fmt.Printf("jbs: %v\n", jbs)
	}
	var bcons []int
	for _, b := range jbs {
		bcons = append(bcons, b.Size())
	}
	sort.Slice(bcons, func(i, j int) bool {
		return bcons[i] > bcons[j]
	})
	return bcons[0] * bcons[1] * bcons[2]
}

func (jb junctionBox) combine() int {
	all := jb.all
	var conns []conn
	for i := range all {
		for j := i + 1; j < len(all); j++ {
			dist := euclideanDistance(all[i], all[j])
			conns = append(conns, conn{
				j1:   all[i],
				j2:   all[j],
				dist: dist,
			})
		}
	}
	sort.Slice(conns, func(i, j int) bool {
		return conns[i].dist < conns[j].dist
	})
	// fmt.Printf("conn: %v\n", conns)
	jbs := []*set.Set[grid.Vector3D[int]]{}
	for _, conn := range conns {
		// fmt.Printf("time: %d\n", i)
		// fmt.Printf("conn: %v,%v\n", conn.j1, conn.j2)
		s := set.New(
			conn.j1, conn.j2,
		)
		var njb []*set.Set[grid.Vector3D[int]]
		for _, b := range jbs {
			if b.HasAny(conn.j1, conn.j2) {
				s = set.Union(s, b)
			} else {
				njb = append(njb, b)
			}
			// fmt.Printf("len: %d,index: %d\n", len(jbs), index)
		}
		njb = append(njb, s)
		jbs = njb
		if len(jbs) == 1 && jbs[0].Size() == len(jb.all) {
			// fmt.Printf("jbs[%d:%d]: %v\n", jbs[0].Size(), len(jb.all), jbs)
			return conn.j1.X * conn.j2.X
		}
	}
	panic("can not reach")
}

func p1() {
	txt := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	bx := junctionBox{}
	txt.ReadByLine(ctx, func(line string) error {
		parts := strings.Split(line, ",")
		bx.all = append(bx.all, grid.Vector3D[int]{
			X: input.Atoi(parts[0]),
			Y: input.Atoi(parts[1]),
			Z: input.Atoi(parts[2]),
		})
		return nil
	})
	fmt.Printf("p1: %d\n", bx.connect(10))
}

func p2() {
	txt := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	bx := junctionBox{}
	txt.ReadByLine(ctx, func(line string) error {
		parts := strings.Split(line, ",")
		bx.all = append(bx.all, grid.Vector3D[int]{
			X: input.Atoi(parts[0]),
			Y: input.Atoi(parts[1]),
			Z: input.Atoi(parts[2]),
		})
		return nil
	})
	fmt.Printf("p2: %d\n", bx.combine())
}

func main() {
	p1()
	p2()
}
