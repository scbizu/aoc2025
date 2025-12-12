package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/magejiCoder/magejiAoc/grid"
	"github.com/magejiCoder/magejiAoc/input"
	"github.com/magejiCoder/magejiAoc/math"
)

type interval struct {
	start int
	end   int
}

type tile struct {
	vecs         []grid.Vec
	rowIntervals map[int][]interval // y -> sorted, merged intervals of x
	colIntervals map[int][]interval // x -> sorted, merged intervals of y
}

var bestV1, bestV2 grid.Vec

func recArea(v1, v2 grid.Vec) int {
	return int((math.Abs(v1.X-v2.X) + 1) * (math.Abs(v1.Y-v2.Y) + 1))
}

func mergeIntervals(xs []int) []interval {
	if len(xs) == 0 {
		return nil
	}
	sort.Ints(xs)
	// Pair xs as [x0,x1], [x2,x3], ..., and build closed intervals
	var raw []interval
	for i := 0; i+1 < len(xs); i += 2 {
		a, b := xs[i], xs[i+1]
		if a > b {
			a, b = b, a
		}
		raw = append(raw, interval{start: a, end: b})
	}
	if len(raw) == 0 {
		return nil
	}
	// Merge overlapping/adjacent intervals
	sort.Slice(raw, func(i, j int) bool {
		if raw[i].start == raw[j].start {
			return raw[i].end < raw[j].end
		}
		return raw[i].start < raw[j].start
	})
	res := []interval{raw[0]}
	for i := 1; i < len(raw); i++ {
		last := &res[len(res)-1]
		cur := raw[i]
		if cur.start <= last.end+0 { // closed intervals, adjacency allowed if sharing endpoints
			if cur.end > last.end {
				last.end = cur.end
			}
		} else {
			res = append(res, cur)
		}
	}
	return res
}

func (t *tile) buildIntervals() {
	t.rowIntervals = make(map[int][]interval)
	t.colIntervals = make(map[int][]interval)
	n := len(t.vecs)
	if n < 2 {
		return
	}
	// Collect vertical edges hits for rows, horizontal edges hits for columns
	rowHits := make(map[int][]int) // y -> x hits
	colHits := make(map[int][]int) // x -> y hits
	for i := 0; i < n; i++ {
		a := t.vecs[i]
		b := t.vecs[(i+1)%n]
		if a.X == b.X {
			// Vertical edge at x = a.X, y in [minY, maxY)
			ymin, ymax := a.Y, b.Y
			if ymin > ymax {
				ymin, ymax = ymax, ymin
			}
			for y := ymin; y < ymax; y++ { // half-open to avoid double counting at vertices
				rowHits[y] = append(rowHits[y], a.X)
			}
			// Also contribute to column intervals as a solid segment on this x
			colHits[a.X] = append(colHits[a.X], a.Y, b.Y)
		} else if a.Y == b.Y {
			// Horizontal edge at y = a.Y, x in [minX, maxX)
			xmin, xmax := a.X, b.X
			if xmin > xmax {
				xmin, xmax = xmax, xmin
			}
			for x := xmin; x < xmax; x++ { // half-open
				colHits[x] = append(colHits[x], a.Y)
			}
			// Add to row intervals as a solid segment on this y
			rowHits[a.Y] = append(rowHits[a.Y], a.X, b.X)
		} else {
			// Should not happen per problem constraints (axis-aligned)
			continue
		}
	}
	// Build row intervals
	for y, xs := range rowHits {
		t.rowIntervals[y] = mergeIntervals(xs)
	}
	// Build column intervals
	for x, ys := range colHits {
		t.colIntervals[x] = mergeIntervals(ys)
	}
}

func (t tile) intervalCoversRow(y, L, R int) bool {
	ints, ok := t.rowIntervals[y]
	if !ok {
		return false
	}
	// Find interval with start <= L and end >= R
	// Binary search by start
	i := sort.Search(len(ints), func(i int) bool { return ints[i].start > L })
	if i == 0 {
		// Check first interval
		if len(ints) > 0 && ints[0].start <= L && ints[0].end >= R {
			return true
		}
		return false
	}
	// Candidate is i-1
	idx := i - 1
	if ints[idx].start <= L && ints[idx].end >= R {
		return true
	}
	return false
}

func (t tile) intervalCoversCol(x, L, R int) bool {
	ints, ok := t.colIntervals[x]
	if !ok {
		return false
	}
	i := sort.Search(len(ints), func(i int) bool { return ints[i].start > L })
	if i == 0 {
		if len(ints) > 0 && ints[0].start <= L && ints[0].end >= R {
			return true
		}
		return false
	}
	idx := i - 1
	if ints[idx].start <= L && ints[idx].end >= R {
		return true
	}
	return false
}

func (t tile) pointInByIntervals(v grid.Vec) bool {
	// Prefer row intervals for quick check
	ints, ok := t.rowIntervals[v.Y]
	if !ok || len(ints) == 0 {
		return false
	}
	// Find interval with start <= v.X
	i := sort.Search(len(ints), func(i int) bool { return ints[i].start > v.X })
	if i == 0 {
		if ints[0].start <= v.X && v.X <= ints[0].end {
			return true
		}
		return false
	}
	idx := i - 1
	return ints[idx].start <= v.X && v.X <= ints[idx].end
}

func (t tile) maxArea() int {
	var ma int
	for i := 0; i < len(t.vecs); i++ {
		for j := i + 1; j < len(t.vecs); j++ {
			area := recArea(t.vecs[i], t.vecs[j])
			if area > ma {
				ma = area
			}
		}
	}
	return ma
}

func (t tile) insideMax() int {
	var ma int
	for i := 0; i < len(t.vecs); i++ {
		for j := i + 1; j < len(t.vecs); j++ {
			area := recArea(t.vecs[i], t.vecs[j])
			if t.isRecInTile(t.vecs[i], t.vecs[j]) {
				if area > ma {
					ma = area
					bestV1 = t.vecs[i]
					bestV2 = t.vecs[j]
				}
			}
		}
	}
	return ma
}

func (t tile) isRecInTile(v1, v2 grid.Vec) bool {
	// Axis-aligned rectangle corners
	v3 := grid.Vec{X: v1.X, Y: v2.Y}
	v4 := grid.Vec{X: v2.X, Y: v1.Y}

	// Corner check via intervals
	if !t.pointInByIntervals(v1) || !t.pointInByIntervals(v2) || !t.pointInByIntervals(v3) || !t.pointInByIntervals(v4) {
		return false
	}

	// Edge coverage check via row/col intervals
	minX, maxX := v1.X, v2.X
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	minY, maxY := v1.Y, v2.Y
	if minY > maxY {
		minY, maxY = maxY, minY
	}

	// Top and bottom edges coverage in rows
	if !t.intervalCoversRow(minY, minX, maxX) {
		return false
	}
	if !t.intervalCoversRow(maxY, minX, maxX) {
		return false
	}
	// Left and right edges coverage in cols
	if !t.intervalCoversCol(minX, minY, maxY) {
		return false
	}
	if !t.intervalCoversCol(maxX, minY, maxY) {
		return false
	}

	// Optional: crossing checks are redundant if intervals are correct,
	// but keep a lightweight version to guard against construction issues.
	edges := [][2]grid.Vec{
		{v1, v3},
		{v3, v2},
		{v2, v4},
		{v4, v1},
	}
	for i := 0; i < len(t.vecs); i++ {
		a := t.vecs[i]
		b := t.vecs[(i+1)%len(t.vecs)]
		for _, rg := range edges {
			if segProperCross(a, b, rg[0], rg[1]) {
				return false
			}
		}
	}

	return true
}

// Geometry helpers retained for crossing protection

func cross(u, v grid.Vec) int {
	return grid.Multiply(u, v)
}

func sub(a, b grid.Vec) grid.Vec {
	return grid.Vec{X: a.X - b.X, Y: a.Y - b.Y}
}

func orient(a, b, c grid.Vec) int {
	return grid.Multiply(sub(b, a), sub(c, a))
}

func onSegment(p, q, r grid.Vec) bool {
	minX, maxX := p.X, r.X
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	minY, maxY := p.Y, r.Y
	if minY > maxY {
		minY, maxY = maxY, minY
	}
	return q.X >= minX && q.X <= maxX && q.Y >= minY && q.Y <= maxY && orient(p, r, q) == 0
}

func segProperCross(p1, q1, p2, q2 grid.Vec) bool {
	o1 := orient(p1, q1, p2)
	o2 := orient(p1, q1, q2)
	o3 := orient(p2, q2, p1)
	o4 := orient(p2, q2, q1)

	if o1*o2 < 0 && o3*o4 < 0 {
		return true
	}
	if o1 == 0 && onSegment(p1, p2, q1) {
		return false
	}
	if o2 == 0 && onSegment(p1, q2, q1) {
		return false
	}
	if o3 == 0 && onSegment(p2, p1, q2) {
		return false
	}
	if o4 == 0 && onSegment(p2, q1, q2) {
		return false
	}
	if (p1 == p2) || (p1 == q2) || (q1 == p2) || (q1 == q2) {
		if p1 == p2 {
			if orient(p1, q1, q2) == 0 && orient(p2, q2, q1) == 0 {
				return false
			}
			return true
		}
		if p1 == q2 {
			if orient(p1, q1, p2) == 0 && orient(q2, p2, q1) == 0 {
				return false
			}
			return true
		}
		if q1 == p2 {
			if orient(q1, p1, q2) == 0 && orient(p2, q2, p1) == 0 {
				return false
			}
			return true
		}
		if orient(q1, p1, p2) == 0 && orient(q2, p2, p1) == 0 {
			return false
		}
		return true
	}
	return false
}

func p1() {
	txt := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	t := tile{}
	txt.ReadByLine(ctx, func(line string) error {
		parts := strings.Split(line, ",")
		t.vecs = append(t.vecs, grid.Vec{
			X: input.Atoi(parts[0]),
			Y: input.Atoi(parts[1]),
		})
		return nil
	})
	a := t.maxArea()
	fmt.Printf("p1: %d\n", a)
}

func p2() {
	txt := input.NewTXTFile("input.txt")
	ctx := context.TODO()

	// Read points
	points := []grid.Vec{}
	txt.ReadByLine(ctx, func(line string) error {
		parts := strings.Split(line, ",")
		points = append(points, grid.Vec{
			X: input.Atoi(parts[0]),
			Y: input.Atoi(parts[1]),
		})
		return nil
	})

	// Coordinate compression
	xSet := map[int]struct{}{}
	ySet := map[int]struct{}{}
	for _, p := range points {
		xSet[p.X] = struct{}{}
		ySet[p.Y] = struct{}{}
	}
	uniqX := make([]int, 0, len(xSet))
	uniqY := make([]int, 0, len(ySet))
	for x := range xSet {
		uniqX = append(uniqX, x)
	}
	for y := range ySet {
		uniqY = append(uniqY, y)
	}
	sort.Ints(uniqX)
	sort.Ints(uniqY)
	xMap := make(map[int]int, len(uniqX))
	yMap := make(map[int]int, len(uniqY))
	for i, x := range uniqX {
		xMap[x] = i
	}
	for i, y := range uniqY {
		yMap[y] = i
	}

	// Build compressed grid
	h := len(uniqY)
	w := len(uniqX)
	gridComp := make([][]byte, h)
	for i := 0; i < h; i++ {
		gridComp[i] = make([]byte, w)
		for j := 0; j < w; j++ {
			gridComp[i][j] = '.'
		}
	}

	// Rasterize polygon edges on compressed grid (include endpoints)
	for i := 0; i < len(points); i++ {
		a := points[i]
		b := points[(i+1)%len(points)]
		ax, ay := xMap[a.X], yMap[a.Y]
		bx, by := xMap[b.X], yMap[b.Y]
		if ax == bx {
			y0, y1 := ay, by
			if y0 > y1 {
				y0, y1 = y1, y0
			}
			for y := y0; y <= y1; y++ {
				gridComp[y][ax] = '#'
			}
		} else if ay == by {
			x0, x1 := ax, bx
			if x0 > x1 {
				x0, x1 = x1, x0
			}
			for x := x0; x <= x1; x++ {
				gridComp[ay][x] = '#'
			}
		}
	}

	// Flood fill interior: find an inside starting point via ray parity
	getInsidePoint := func(grid [][]byte) (int, int) {
		for y := 0; y < len(grid); y++ {
			for x := 0; x < len(grid[0]); x++ {
				if grid[y][x] != '.' {
					continue
				}
				hits := 0
				prev := byte('.')
				for i := x; i >= 0; i-- {
					cur := grid[y][i]
					if cur != prev {
						hits++
						prev = cur
					}
				}
				if hits%2 == 1 {
					return x, y
				}
			}
		}
		return -1, -1
	}
	sx, sy := getInsidePoint(gridComp)
	if sx != -1 {
		stack := []grid.Vec{{X: sx, Y: sy}}
		dirs := []grid.Vec{{X: 0, Y: 1}, {X: 0, Y: -1}, {X: 1, Y: 0}, {X: -1, Y: 0}}
		for len(stack) > 0 {
			p := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if p.Y < 0 || p.Y >= h || p.X < 0 || p.X >= w {
				continue
			}
			if gridComp[p.Y][p.X] != '.' {
				continue
			}
			gridComp[p.Y][p.X] = 'X'
			for _, d := range dirs {
				nx := p.X + d.X
				ny := p.Y + d.Y
				if ny >= 0 && ny < h && nx >= 0 && nx < w && gridComp[ny][nx] == '.' {
					stack = append(stack, grid.Vec{X: nx, Y: ny})
				}
			}
		}
	}

	// Helper: check rectangle edges enclosure on compressed grid
	isEnclosed := func(a, b grid.Vec) bool {
		x1 := xMap[a.X]
		x2 := xMap[b.X]
		y1 := yMap[a.Y]
		y2 := yMap[b.Y]
		if x1 > x2 {
			x1, x2 = x2, x1
		}
		if y1 > y2 {
			y1, y2 = y2, y1
		}
		// Top and bottom edges
		for x := x1; x <= x2; x++ {
			if gridComp[y1][x] == '.' || gridComp[y2][x] == '.' {
				return false
			}
		}
		// Left and right edges
		for y := y1; y <= y2; y++ {
			if gridComp[y][x1] == '.' || gridComp[y][x2] == '.' {
				return false
			}
		}
		return true
	}

	// Iterate all pairs
	maxArea := 0
	var maxA, maxB grid.Vec
	for i := 0; i < len(points); i++ {
		for j := i + 1; j < len(points); j++ {
			a := points[i]
			b := points[j]
			if isEnclosed(a, b) {
				area := (math.Abs(a.X-b.X) + 1) * (math.Abs(a.Y-b.Y) + 1)
				if int(area) > maxArea {
					maxArea = int(area)
					maxA, maxB = a, b
				}
			}
		}
	}

	fmt.Printf("p2: %d between (%d,%d) and (%d,%d)\n", maxArea, maxA.X, maxA.Y, maxB.X, maxB.Y)
}

func main() {
	p1()
	p2()
}
