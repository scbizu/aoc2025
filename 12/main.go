package main

// Optional: solution verifier to check a user-provided letter placement against present shapes.
// Additionally, when AOC12_VERIFY=1, we can cross-check whether the concrete placements implied by
// the provided layout exist in the solver's pre-enumerated placement lists. Enable with:
//   AOC12_CHECK_PLACEMENTS=1
// Enable by setting AOC12_VERIFY=1 and providing AOC12_LAYOUT with 5 lines of 12 chars ('.' for empty)
// separated by '\n'. The verifier maps letters A..F to present indices 0..5 by default.
//
// Example:
//   AOC12_VERIFY=1 \
//   AOC12_LAYOUT='....AAAFFE.E\n.BBBAAFFFEEE\nDDDBAAFFCECE\nDBBB....CCC.\nDDD.....C.C.' \
//   go run .

import (
	"context"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"time"

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

func shapeAnchor(p present) (minX, minY int) {
	first := true
	for v := range p.shape {
		if first {
			minX, minY = v.X, v.Y
			first = false
			continue
		}
		if v.X < minX {
			minX = v.X
		}
		if v.Y < minY {
			minY = v.Y
		}
	}
	return
}

// rotate 把 p.shape 以图像中心点为轴顺时针旋转 90 度
func (p present) rotate() present {
	// 以当前形状的 anchor(minX,minY) 作为 3x3 局部网格的原点
	ax, ay := shapeAnchor(p)

	const n = 3
	newShape := grid.NewVecMatrix[byte]()
	for v, val := range p.shape {
		lx := v.X - ax
		ly := v.Y - ay

		nx := ly
		ny := n - 1 - lx

		newShape.Add(grid.Vec{X: nx + ax, Y: ny + ay}, val)
	}
	return present{shape: newShape}
}

// placementKey creates a stable key for a set of absolute cells (Vecs).
func placementKey(vs []grid.Vec) string {
	if len(vs) == 0 {
		return ""
	}
	slices.SortFunc(vs, func(a, b grid.Vec) int {
		if a.X != b.X {
			return a.X - b.X
		}
		return a.Y - b.Y
	})
	var b strings.Builder
	b.Grow(len(vs) * 8)
	for _, v := range vs {
		fmt.Fprintf(&b, "%d,%d;", v.X, v.Y)
	}
	return b.String()
}

// verifyLayoutAgainstPresents checks whether a given layout places each present (by index) exactly once,
// allowing '.' as empty. Letters A..F map to present indices 0..5 by default.
func verifyLayoutAgainstPresents(layout string, presents []present, rows, cols int) error {
	lines := strings.Split(strings.TrimSpace(layout), "\n")
	if len(lines) != rows {
		return fmt.Errorf("layout row count mismatch: got %d want %d", len(lines), rows)
	}
	for i := range lines {
		if len(lines[i]) != cols {
			return fmt.Errorf("layout col count mismatch at row %d: got %d want %d", i, len(lines[i]), cols)
		}
	}

	// Collect cells by letter
	byLetter := make(map[byte][]grid.Vec)
	for x := 0; x < rows; x++ {
		for y := 0; y < cols; y++ {
			ch := lines[x][y]
			if ch == '.' {
				continue
			}
			if ch < 'A' || ch > 'Z' {
				return fmt.Errorf("invalid char %q at (%d,%d)", ch, x, y)
			}
			byLetter[ch] = append(byLetter[ch], grid.Vec{X: x, Y: y})
		}
	}

	// Helper: build a shape from absolute coords and normalize to origin for comparison
	makeNormalized := func(vs []grid.Vec) present {
		m := grid.NewVecMatrix[byte]()
		for _, v := range vs {
			m.Add(v, '#')
		}
		return normalizeToOrigin(present{shape: m})
	}

	// Helper: compare against any orientation (with flips) of a present
	matchesAnyOrientation := func(target present, base present) bool {
		targetKey := presentKey(normalizeToOrigin(target))
		for _, o := range uniqueOrientations(normalizeToOrigin(base)) {
			if presentKey(o) == targetKey {
				return true
			}
		}
		return false
	}

	// Build mapping from letters to present indices.
	// Default mapping: A->0, B->1, ... F->5
	// Override by setting AOC12_MAP like: "A=0,B=4,C=5,D=4,E=5,F=2"
	letterToIdx := make(map[byte]int, 26)
	for ch := byte('A'); ch <= byte('Z'); ch++ {
		letterToIdx[ch] = int(ch - 'A')
	}
	if m := os.Getenv("AOC12_MAP"); m != "" {
		parts := strings.Split(m, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			kv := strings.Split(part, "=")
			if len(kv) != 2 {
				return fmt.Errorf("invalid AOC12_MAP entry %q (expected like A=0)", part)
			}
			k := strings.TrimSpace(kv[0])
			v := strings.TrimSpace(kv[1])
			if len(k) != 1 || k[0] < 'A' || k[0] > 'Z' {
				return fmt.Errorf("invalid AOC12_MAP letter %q in entry %q", k, part)
			}
			var idx int
			if _, err := fmt.Sscanf(v, "%d", &idx); err != nil {
				return fmt.Errorf("invalid AOC12_MAP index %q in entry %q", v, part)
			}
			letterToIdx[k[0]] = idx
		}
	}

	for letter, cells := range byLetter {
		idx, ok := letterToIdx[letter]
		if !ok {
			return fmt.Errorf("no mapping for letter %c", letter)
		}
		if idx < 0 || idx >= len(presents) {
			return fmt.Errorf("letter %c maps to present index %d out of range (presents=%d)", letter, idx, len(presents))
		}
		if len(cells) == 0 {
			continue
		}
		target := makeNormalized(cells)
		if len(target.shape) != len(presents[idx].shape) {
			return fmt.Errorf("letter %c area mismatch: got %d want %d", letter, len(target.shape), len(presents[idx].shape))
		}
		if !matchesAnyOrientation(target, presents[idx]) {
			return fmt.Errorf("letter %c does not match present #%d under rotate/flip", letter, idx)
		}
	}

	return nil
}

// verifyPlacementsExist checks whether each letter's absolute cell-set exists inside the solver's placement lists.
// This helps diagnose "solver returns 0 but a verified layout exists" by finding which type's placement enumeration is missing entries.
func verifyPlacementsExist(layout string, rows, cols int, typePlacements [][]struct{ cells []grid.Vec }, letterToType map[byte]int) error {
	lines := strings.Split(strings.TrimSpace(layout), "\n")
	if len(lines) != rows {
		return fmt.Errorf("layout row count mismatch: got %d want %d", len(lines), rows)
	}
	for i := range lines {
		if len(lines[i]) != cols {
			return fmt.Errorf("layout col count mismatch at row %d: got %d want %d", i, len(lines[i]), cols)
		}
	}

	byLetter := make(map[byte][]grid.Vec)
	for x := 0; x < rows; x++ {
		for y := 0; y < cols; y++ {
			ch := lines[x][y]
			if ch == '.' {
				continue
			}
			byLetter[ch] = append(byLetter[ch], grid.Vec{X: x, Y: y})
		}
	}

	// Build a fast lookup set for each typePlacements list
	typeSets := make([]map[string]struct{}, len(typePlacements))
	for ti := range typePlacements {
		m := make(map[string]struct{}, len(typePlacements[ti]))
		for _, pl := range typePlacements[ti] {
			k := placementKey(append([]grid.Vec(nil), pl.cells...))
			m[k] = struct{}{}
		}
		typeSets[ti] = m
	}

	for letter, cells := range byLetter {
		ti, ok := letterToType[letter]
		if !ok {
			// ignore letters not in mapping
			continue
		}
		if ti < 0 || ti >= len(typePlacements) {
			return fmt.Errorf("letter %c maps to type %d out of range (types=%d)", letter, ti, len(typePlacements))
		}
		k := placementKey(append([]grid.Vec(nil), cells...))
		if _, ok := typeSets[ti][k]; !ok {
			return fmt.Errorf("layout placement for letter %c not found in enumerated placements for type %d", letter, ti)
		}
	}

	return nil
}

func (p present) offset(vec grid.Vec) present {
	newShape := grid.NewVecMatrix[byte]()
	for v, val := range p.shape {
		newV := grid.Vec{X: v.X + vec.X, Y: v.Y + vec.Y}
		newShape.Add(newV, val)
	}
	return present{shape: newShape}
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

func place(ava grid.VecMatrix[struct{}], p present) grid.VecMatrix[struct{}] {
	ava = maps.Clone(ava)
	for v := range p.shape {
		delete(ava, v)
	}
	return ava
}

func vecSetKey[T any](m grid.VecMatrix[T]) string {
	vs := make([]grid.Vec, 0, len(m))
	for v := range m {
		vs = append(vs, v)
	}
	slices.SortFunc(vs, func(a, b grid.Vec) int {
		if a.X != b.X {
			return a.X - b.X
		}
		return a.Y - b.Y
	})
	var b strings.Builder
	b.Grow(len(vs) * 8)
	for _, v := range vs {
		fmt.Fprintf(&b, "%d,%d;", v.X, v.Y)
	}
	return b.String()
}

func minVecInSet(m grid.VecMatrix[struct{}]) (grid.Vec, bool) {
	first := true
	var best grid.Vec
	for v := range m {
		if first {
			best = v
			first = false
			continue
		}
		if v.X < best.X || (v.X == best.X && v.Y < best.Y) {
			best = v
		}
	}
	if first {
		return grid.Vec{}, false
	}
	return best, true
}

func boundsOfShape(p present) (minX, minY, maxX, maxY int) {
	first := true
	for v := range p.shape {
		if first {
			minX, maxX = v.X, v.X
			minY, maxY = v.Y, v.Y
			first = false
			continue
		}
		if v.X < minX {
			minX = v.X
		}
		if v.X > maxX {
			maxX = v.X
		}
		if v.Y < minY {
			minY = v.Y
		}
		if v.Y > maxY {
			maxY = v.Y
		}
	}
	return
}

// normalizeToOrigin shifts shape so that its minX/minY become 0/0.
// This is used to build canonical orientations independent of absolute offsets.
func normalizeToOrigin(p present) present {
	minX, minY, _, _ := boundsOfShape(p)
	if minX == 0 && minY == 0 {
		return p
	}
	return p.offset(grid.Vec{X: -minX, Y: -minY})
}

func presentKey(p present) string {
	return vecSetKey(p.shape)
}

// flipH mirrors the shape horizontally within its current bounding box (left-right flip),
// then normalizes it back to origin so further operations can assume minX/minY == 0/0.
func flipH(p present) present {
	p = normalizeToOrigin(p)
	_, _, _, maxY := boundsOfShape(p)

	newShape := grid.NewVecMatrix[byte]()
	for v, val := range p.shape {
		newShape.Add(grid.Vec{X: v.X, Y: maxY - v.Y}, val)
	}
	return normalizeToOrigin(present{shape: newShape})
}

// flipV mirrors the shape vertically within its current bounding box (top-bottom flip),
// then normalizes it back to origin so further operations can assume minX/minY == 0/0.
func flipV(p present) present {
	p = normalizeToOrigin(p)
	_, _, maxX, _ := boundsOfShape(p)

	newShape := grid.NewVecMatrix[byte]()
	for v, val := range p.shape {
		newShape.Add(grid.Vec{X: maxX - v.X, Y: v.Y}, val)
	}
	return normalizeToOrigin(present{shape: newShape})
}

// uniqueOrientations returns unique rotations and mirrored rotations of p, normalized to (0,0).
// With flip allowed, this is up to 8 unique orientations.
func uniqueOrientations(p present) []present {
	seen := make(map[string]struct{}, 8)
	var out []present

	addRotations := func(base present) {
		cur := normalizeToOrigin(base)
		for i := 0; i < 4; i++ {
			k := presentKey(cur)
			if _, ok := seen[k]; !ok {
				seen[k] = struct{}{}
				out = append(out, cur)
			}
			cur = normalizeToOrigin(cur.rotate())
		}
	}

	addRotations(p)
	addRotations(flipH(p))
	addRotations(flipV(p))

	return out
}

func countsKey(counts []int) string {
	var b strings.Builder
	// rough capacity: up to ~3 chars per count + separator
	b.Grow(len(counts) * 4)
	for i, c := range counts {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, "%d", c)
	}
	return b.String()
}

func anyCellOutOfBounds(p present, rows, cols int) bool {
	for v := range p.shape {
		if v.X < 0 || v.Y < 0 || v.X >= rows || v.Y >= cols {
			return true
		}
	}
	return false
}

func minPresentArea(presents []present) int {
	if len(presents) == 0 {
		return 0
	}
	minA := len(presents[0].shape)
	for i := 1; i < len(presents); i++ {
		if a := len(presents[i].shape); a < minA {
			minA = a
		}
	}
	return minA
}

// prune: if any connected component in ava has size < minArea, it's impossible to fill it
func hasTooSmallComponent(ava grid.VecMatrix[struct{}], minArea int) bool {
	if minArea <= 1 {
		return false
	}
	if len(ava) == 0 {
		return false
	}

	visited := make(map[grid.Vec]struct{}, len(ava))

	neighbors := func(v grid.Vec) [4]grid.Vec {
		return [4]grid.Vec{
			{X: v.X - 1, Y: v.Y},
			{X: v.X + 1, Y: v.Y},
			{X: v.X, Y: v.Y - 1},
			{X: v.X, Y: v.Y + 1},
		}
	}

	for start := range ava {
		if _, ok := visited[start]; ok {
			continue
		}

		// iterative DFS (avoid recursion)
		stack := make([]grid.Vec, 0, 64)
		stack = append(stack, start)
		visited[start] = struct{}{}

		count := 0
		for len(stack) > 0 {
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			count++

			for _, nb := range neighbors(v) {
				if _, ok := ava[nb]; !ok {
					continue
				}
				if _, ok := visited[nb]; ok {
					continue
				}
				visited[nb] = struct{}{}
				stack = append(stack, nb)
			}
		}

		if count < minArea {
			return true
		}
	}

	return false
}

func (r *region) fill(presents []present) bool {
	if len(presents) == 0 {
		return false
	}

	// Debug (optional): enable with AOC12_DEBUG=1, tune with AOC12_DEBUG_EVERY (default 50000)
	debug := os.Getenv("AOC12_DEBUG") != ""
	debugEvery := int64(50000)
	if v := os.Getenv("AOC12_DEBUG_EVERY"); v != "" {
		var n int64
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n > 0 {
			debugEvery = n
		}
	}

	type dbgCounters struct {
		calls              int64
		seenHits           int64
		anchorMissing      int64
		pruneRemainingArea int64
		pruneComponent     int64
		placementsTried    int64
		placementsFit      int64
		lastLog            time.Time
	}
	var dbg dbgCounters
	dbg.lastLog = time.Now()

	// Build counts by present *index* (type), not by shape.
	//
	// We must preserve distinct present indices (0..N-1), even if some shapes are identical,
	// because the input semantics are "counts per present index".
	//
	// NOTE: `presents` here is already expanded by `validRegions()` (i.e., contains duplicates).
	// We de-duplicate by shapeKey to reconstruct the "type list", but we DO NOT merge different
	// indices that happen to share the same shape.
	//
	// To achieve this, we rebuild a stable list of types by scanning `f.presents` order indirectly:
	// we assign each distinct shape in `presents` a unique slot, and counts are by that slot.
	//
	// IMPORTANT: This assumes each present index has a unique shape in the input.
	// If your input can contain two different indices with identical shapes, then the caller must
	// pass the index along (not just the shape) so we can keep them distinct.
	typeKeyToIdx := make(map[string]int)
	var types []present
	var counts []int
	for _, p := range presents {
		n := normalizeToOrigin(p)
		k := presentKey(n)
		idx, ok := typeKeyToIdx[k]
		if !ok {
			idx = len(types)
			typeKeyToIdx[k] = idx
			types = append(types, n)
			counts = append(counts, 0)
		}
		counts[idx]++
	}

	// Precompute unique orientations for each type.
	typeOrients := make([][]present, len(types))
	typeAreas := make([]int, len(types))
	for i := range types {
		typeOrients[i] = uniqueOrientations(types[i])
		typeAreas[i] = len(types[i].shape)
	}

	remainingArea := 0
	for i, c := range counts {
		if c == 0 {
			continue
		}
		remainingArea += c * typeAreas[i]
	}

	// Extra debug dump at fill-start
	if debug {
		fmt.Printf("[dbg] fill: region cols=%d rows=%d ava=%d remainingArea=%d\n", r.cols, r.rows, len(r.ava), remainingArea)
		fmt.Printf("[dbg] types=%d\n", len(types))
		for i := range types {
			fmt.Printf("[dbg] type[%d] count=%d area=%d orients=%d shapeKey=%s\n",
				i, counts[i], typeAreas[i], len(typeOrients[i]), presentKey(types[i]),
			)
		}
		fmt.Printf("[dbg] countsKey=%s\n", countsKey(counts))
	}

	// Memoize failures on (avaKey, countsKey).
	seen := make(map[string]struct{})

	// Precompute a bounded list of candidate placements for each type on this region.
	// This is the "piece-first" model: we only branch on actual placements, not on skipping cells.
	//
	// Because the region does NOT need to be fully covered, we must not force covering an "anchor" cell.
	// Instead, we pick a remaining piece type and try all placements that fit in current ava.
	type placement struct {
		cells []grid.Vec
	}
	typePlacements := make([][]placement, len(types))
	for ti := range types {
		// de-duplicate placements by their cell-set key
		placementSeen := make(map[string]struct{}, 256)
		for _, orient := range typeOrients[ti] {
			// Determine the legal translation range by bounding box in the normalized orientation.
			minX, minY, maxX, maxY := boundsOfShape(orient)
			h := maxX - minX + 1
			w := maxY - minY + 1
			if h <= 0 || w <= 0 {
				continue
			}

			// For a normalized orientation, minX/minY should be 0,0, but we don't assume it.
			// Translate so that the bounding box fits within [0,rows) x [0,cols).
			for offX := 0; offX <= r.rows-h; offX++ {
				for offY := 0; offY <= r.cols-w; offY++ {
					shift := grid.Vec{X: offX - minX, Y: offY - minY}
					placed := orient.offset(shift)
					if anyCellOutOfBounds(placed, r.rows, r.cols) {
						continue
					}

					vs := make([]grid.Vec, 0, len(placed.shape))
					for v := range placed.shape {
						vs = append(vs, v)
					}

					k := placementKey(append([]grid.Vec(nil), vs...))
					if _, ok := placementSeen[k]; ok {
						continue
					}
					placementSeen[k] = struct{}{}

					// store sorted cells for later membership checks
					slices.SortFunc(vs, func(a, b grid.Vec) int {
						if a.X != b.X {
							return a.X - b.X
						}
						return a.Y - b.Y
					})
					typePlacements[ti] = append(typePlacements[ti], placement{cells: vs})
				}
			}
		}
	}

	// If requested, cross-check that the user-provided verified layout's placements exist in these placement lists.
	// This helps diagnose why the solver returns 0 despite a verified layout.
	//
	// IMPORTANT: the solver's `types` are merged by shapeKey (not by original present index),
	// so we must map layout letters -> present index -> shapeKey -> solver type index.
	if os.Getenv("AOC12_VERIFY") != "" && os.Getenv("AOC12_CHECK_PLACEMENTS") != "" {
		layout := os.Getenv("AOC12_LAYOUT")
		if layout != "" {
			// Build letter->presentIndex mapping.
			// Default: A->0, B->1, ...; override with AOC12_MAP like: "A=0,B=4,C=5,D=4,E=5,F=2"
			letterToPresent := make(map[byte]int, 26)
			for ch := byte('A'); ch <= byte('Z'); ch++ {
				letterToPresent[ch] = int(ch - 'A')
			}
			if m := os.Getenv("AOC12_MAP"); m != "" {
				parts := strings.Split(m, ",")
				for _, part := range parts {
					part = strings.TrimSpace(part)
					if part == "" {
						continue
					}
					kv := strings.Split(part, "=")
					if len(kv) != 2 {
						continue
					}
					k := strings.TrimSpace(kv[0])
					v := strings.TrimSpace(kv[1])
					if len(k) != 1 {
						continue
					}
					var idx int
					if _, err := fmt.Sscanf(v, "%d", &idx); err == nil {
						letterToPresent[k[0]] = idx
					}
				}
			}

			// shapeKey -> solver type index (these types are in `types` slice)
			shapeToType := make(map[string]int, len(types))
			for ti := range types {
				shapeToType[presentKey(normalizeToOrigin(types[ti]))] = ti
			}

			// Finally build letter -> solver type mapping.
			letterToType := make(map[byte]int, 26)
			for letter, pi := range letterToPresent {
				if pi < 0 || pi >= len(presents) {
					// out of range for the parsed presents; skip
					continue
				}
				sk := presentKey(normalizeToOrigin(presents[pi]))
				if ti, ok := shapeToType[sk]; ok {
					letterToType[letter] = ti
				}
			}

			// Adapt typePlacements to the verifier signature
			tmp := make([][]struct{ cells []grid.Vec }, len(typePlacements))
			for i := range typePlacements {
				tmp[i] = make([]struct{ cells []grid.Vec }, len(typePlacements[i]))
				for j := range typePlacements[i] {
					tmp[i][j] = struct{ cells []grid.Vec }{cells: typePlacements[i][j].cells}
				}
			}
			if err := verifyPlacementsExist(layout, r.rows, r.cols, tmp, letterToType); err != nil {
				fmt.Printf("[verify] placement-list mismatch: %v\n", err)
			} else {
				fmt.Printf("[verify] placement-list coverage: OK\n")
			}
		}
	}

	// Helper: count how many placements of a type are compatible with current ava.
	compatibleCount := func(ava grid.VecMatrix[struct{}], pls []placement) int {
		n := 0
		for _, pl := range pls {
			ok := true
			for _, v := range pl.cells {
				if _, exists := ava[v]; !exists {
					ok = false
					break
				}
			}
			if ok {
				n++
			}
		}
		return n
	}

	var dfs func(ava grid.VecMatrix[struct{}], counts []int, remainingArea int) bool
	dfs = func(ava grid.VecMatrix[struct{}], counts []int, remainingArea int) bool {
		dbg.calls++
		if debug && dbg.calls%debugEvery == 0 {
			elapsed := time.Since(dbg.lastLog)
			dbg.lastLog = time.Now()
			fmt.Printf(
				"[dbg] calls=%d seen=%d seenHits=%d ava=%d remainingArea=%d counts=%s | prune(remArea=%d cc=%d) | place(tried=%d fit=%d) | dt=%s\n",
				dbg.calls, len(seen), dbg.seenHits, len(ava), remainingArea, countsKey(counts),
				dbg.pruneRemainingArea, dbg.pruneComponent,
				dbg.placementsTried, dbg.placementsFit,
				elapsed,
			)
		}

		if remainingArea == 0 {
			// All presents placed.
			return true
		}
		if remainingArea > len(ava) {
			dbg.pruneRemainingArea++
			return false
		}

		key := "ava=" + vecSetKey(ava) + "|counts=" + countsKey(counts)
		if _, ok := seen[key]; ok {
			dbg.seenHits++
			return false
		}

		// Choose the next type to place using MRV: smallest number of compatible placements.
		chosen := -1
		chosenCompat := 0
		for ti, c := range counts {
			if c == 0 {
				continue
			}
			comp := compatibleCount(ava, typePlacements[ti])
			if comp == 0 {
				// No way to place this remaining type at all => dead end
				seen[key] = struct{}{}
				return false
			}
			if chosen == -1 || comp < chosenCompat {
				chosen = ti
				chosenCompat = comp
				if chosenCompat == 1 {
					// can't do better than 1
					break
				}
			}
		}
		if chosen == -1 {
			// no types remaining but remainingArea != 0 should be impossible
			seen[key] = struct{}{}
			return false
		}

		area := typeAreas[chosen]
		nextRemaining := remainingArea - area

		// Try all compatible placements of the chosen type.
		for _, pl := range typePlacements[chosen] {
			dbg.placementsTried++

			fits := true
			for _, v := range pl.cells {
				if _, ok := ava[v]; !ok {
					fits = false
					break
				}
			}
			if !fits {
				continue
			}
			dbg.placementsFit++

			// Apply placement (clone ava once)
			newAva := maps.Clone(ava)
			for _, v := range pl.cells {
				delete(newAva, v)
			}

			nextCounts := slices.Clone(counts)
			nextCounts[chosen]--

			if dfs(newAva, nextCounts, nextRemaining) {
				return true
			}
		}

		seen[key] = struct{}{}
		return false
	}

	return dfs(r.ava, counts, remainingArea)
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

		// Optional verifier: checks a user-provided layout against present shapes.
		// Only runs when AOC12_VERIFY is set.
		if os.Getenv("AOC12_VERIFY") != "" {
			layout := os.Getenv("AOC12_LAYOUT")
			if layout == "" {
				fmt.Printf("[verify] AOC12_VERIFY set but AOC12_LAYOUT is empty\n")
			} else {
				// Verify against the first region dimensions if present.
				if len(regions) > 0 {
					if err := verifyLayoutAgainstPresents(layout, presents, regions[0].rows, regions[0].cols); err != nil {
						fmt.Printf("[verify] FAIL: %v\n", err)
					} else {
						fmt.Printf("[verify] OK\n")
					}
				} else {
					fmt.Printf("[verify] no regions parsed\n")
				}
			}
		}

		fmt.Printf("p1: %d\n", f.validRegions())
		return nil
	})
}

func main() {
	p1()
}
