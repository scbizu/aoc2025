package main

import (
	"container/heap"
	"context"
	"fmt"
	"maps"
	"math"
	"math/big"
	"sort"
	"strings"

	"github.com/magejiCoder/magejiAoc/input"
)

type machine struct {
	current      int
	result       int
	resLen       int
	instructions []string

	min int
	// 某一位被置位了多少次
	bitMap map[int]int
}

func NewMachine(result string, ins []string) *machine {
	m := &machine{
		current:      0,
		result:       parseResult(result),
		resLen:       len(result) - 2,
		instructions: ins,
		min:          len(ins) + 1,
		bitMap:       make(map[int]int),
	}
	return m
}

func NewMachineV2(result string, ins []string, counter string) *machine {
	m := &machine{
		current:      0,
		result:       parseResult(result),
		resLen:       len(result) - 2,
		instructions: ins,
		min:          math.MaxInt,
		bitMap:       parseBitMap(counter),
	}
	return m
}

func (m *machine) reach() int {
	visited := make(map[int]int)
	m.pushWithState(make(map[string]struct{}), visited, m.current, 0)
	return m.min
}

func (m *machine) reachV2() int {
	visited := make(map[state]int)
	bitMap := make(map[int]int)
	m.pushWithJoltage(visited, bitMap, m.current, 0, []string{})
	return m.min
}

func parseResult(s string) int {
	raw := s[1 : len(s)-1]
	// fmt.Printf("raw: %s\n", raw)
	var highs []int
	for i, b := range raw {
		if b == '#' {
			highs = append(highs, len(raw)-i-1)
		}
	}
	return buildBinary(highs)
}

func parseBitMap(s string) map[int]int {
	raw := s[1 : len(s)-1]
	m := make(map[int]int)
	parts := strings.Split(raw, ",")
	for i, b := range parts {
		m[len(parts)-i-1] = input.Atoi(b)
	}
	return m
}

func parseInstruction(size int, in string) []int {
	raw := in[1 : len(in)-1]
	var highs []int
	for part := range strings.SplitSeq(raw, ",") {
		highs = append(highs, size-1-input.Atoi(part))
	}
	return highs
}

func matchBitMap(dst, src map[int]int) bool {
	if len(dst) != len(src) {
		return false
	}
	for k, v := range dst {
		if src[k] != v {
			return false
		}
	}
	return true
}

func xor(a, b int) int {
	return a ^ b
}

func buildBinary(highIndex []int) int {
	var binary int
	for _, index := range highIndex {
		binary |= 1 << index
	}
	return binary
}

func bitmapString(m map[int]int) string {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('|')
		}
		b.WriteString(fmt.Sprintf("%d:%d", k, m[k]))
	}
	return b.String()
}

func push(current int, highs []int) int {
	return xor(current, buildBinary(highs))
}

type state struct {
	cur    int
	bitMap string
}

// 启发式函数：估计还需要多少次按压
func (m *machine) estimateRemaining(bitMap map[int]int) int {
	remaining := 0
	for bit, targetCount := range m.bitMap {
		currentCount := bitMap[bit]
		if currentCount < targetCount {
			remaining += targetCount - currentCount
		}
	}
	return remaining
}

func (m *machine) pushWithJoltage(
	visited map[state]int,
	bitMap map[int]int,
	cur int, pushes int,
	history []string,
) {
	// fmt.Printf("state: %v\n", visited)
	if pushes >= m.min {
		return
	}

	// 启发式剪枝：估计总步数，如果超过当前最小值则剪枝
	estimate := m.estimateRemaining(bitMap)
	if pushes+estimate >= m.min {
		fmt.Printf("heuristic prune: pushes %d + estimate %d >= min %d\n", pushes, estimate, m.min)
		return
	}

	// 剪枝：如果任何位的计数超过目标，直接返回
	for bit, count := range bitMap {
		if targetCount, ok := m.bitMap[bit]; ok {
			if count > targetCount {
				fmt.Printf("prune: bit %d count %d > target %d\n", bit, count, targetCount)
				return
			}
		} else if count > 0 {
			// 如果目标中没有这个位，但我们却按了，也要剪枝
			fmt.Printf("prune: bit %d not in target but count is %d\n", bit, count)
			return
		}
	}

	s := state{cur, bitmapString(bitMap)}
	if v, ok := visited[s]; ok && pushes > v {
		fmt.Printf("skip: %d >= %d\n", pushes, v)
		return
	}
	visited[s] = pushes
	if cur == m.result {
		// fmt.Printf("bitmap:%v\n", bitMap)
		// fmt.Printf("cur: %4b\n", cur)
		if matchBitMap(bitMap, m.bitMap) {
			// fmt.Printf("bitmap :%v\n", bitMap)
			if pushes < m.min {
				m.min = pushes
			}
			return
		}
	}

	fmt.Printf("current: %04b\n", cur)
	fmt.Printf("bitmap: %v\n", bitMap)
	fmt.Printf("pushed: %d\n", pushes)
	fmt.Printf("estimate remaining: %d\n", estimate)
	fmt.Printf("history: %v\n", history)

	// 按照贡献度排序指令，优先尝试能满足更多缺失位的指令
	type instrScore struct {
		instr string
		score int
	}
	var scoredInstr []instrScore

	for _, nin := range m.instructions {
		inst := parseInstruction(m.resLen, nin)

		// 计算这个指令的贡献分数（能满足多少个缺失的位）
		score := 0
		for _, idx := range inst {
			currentCount := bitMap[idx]
			if targetCount, ok := m.bitMap[idx]; ok {
				if currentCount < targetCount {
					score++
				}
			}
		}
		scoredInstr = append(scoredInstr, instrScore{nin, score})
	}

	// 按分数降序排序
	sort.Slice(scoredInstr, func(i, j int) bool {
		return scoredInstr[i].score > scoredInstr[j].score
	})

	// 按排序后的顺序尝试指令
	for _, si := range scoredInstr {
		nin := si.instr
		fmt.Printf("press: %v (score: %d)\n", nin, si.score)
		inst := parseInstruction(m.resLen, nin)
		type delta struct{ idx, add int }
		var deltas []delta
		for _, idx := range inst {
			if _, ok := bitMap[idx]; !ok {
				bitMap[idx] = 0
			}
			bitMap[idx] += 1
			deltas = append(deltas, delta{idx: idx, add: 1})
		}
		next := push(cur, inst)
		// 为递归调用创建 bitMap 的副本
		bc := make(map[int]int)
		maps.Copy(bc, bitMap)
		// 创建新的 history 副本
		newHistory := make([]string, len(history))
		copy(newHistory, history)
		newHistory = append(newHistory, nin)
		m.pushWithJoltage(visited, bc, next, pushes+1, newHistory)
		for _, d := range deltas {
			bitMap[d.idx] -= d.add
		}
	}
}

func (m *machine) pushWithState(pressed map[string]struct{}, visited map[int]int, cur int, pushes int) {
	if pushes >= m.min {
		return
	}
	if v, ok := visited[cur]; ok && pushes >= v {
		return
	}
	visited[cur] = pushes
	if cur == m.result {
		if pushes < m.min {
			m.min = pushes
		}
		return
	}
	// 只遍历还需要置位的指令集合（贪心：根据当前状态和目标的差异位进行过滤）
	need := xor(cur, m.result)
	for _, nin := range m.instructions {
		if _, ok := pressed[nin]; ok {
			continue
		}
		inst := parseInstruction(m.resLen, nin)
		mask := buildBinary(inst)
		// 如果这条指令不影响“还需要置位”的位，跳过
		if mask&need == 0 {
			continue
		}
		pressed[nin] = struct{}{}
		next := push(cur, inst)
		m.pushWithState(pressed, visited, next, pushes+1)
		delete(pressed, nin)
	}
}

func p1() {
	txt := input.NewTXTFile("input.txt")
	ctx := context.TODO()

	var ms []*machine
	txt.ReadByLine(ctx, func(line string) error {
		parts := strings.Split(line, " ")
		ms = append(
			ms,
			NewMachine(parts[0], parts[1:len(parts)-1]),
		)
		return nil
	})
	var sum int
	for _, m := range ms {
		sum += m.reach()
	}
	fmt.Printf("p1: %d\n", sum)
}

func p2() {
	txt := input.NewTXTFile("input.txt")
	ctx := context.TODO()

	var ms []*machine
	txt.ReadByLine(ctx, func(line string) error {
		parts := strings.Split(line, " ")
		ms = append(
			ms,
			NewMachineV2(parts[0], parts[1:len(parts)-1], parts[len(parts)-1]),
		)
		return nil
	})

	// LTR helpers: parse result, instructions, counter in left-to-right orientation
	parseResultLTR := func(s string) int {
		raw := s[1 : len(s)-1]
		mask := 0
		// bit 0 is leftmost
		for i, ch := range raw {
			if ch == '#' {
				mask |= 1 << i
			}
		}
		return mask
	}
	parseInstructionLTR := func(size int, in string) []int {
		raw := in[1 : len(in)-1]
		var bits []int
		for part := range strings.SplitSeq(raw, ",") {
			// indices are already LTR, bit j is j from left
			bits = append(bits, input.Atoi(part))
		}
		return bits
	}
	// parseCounterLTR := func(s string) []int {
	// 	raw := s[1 : len(s)-1]
	// 	parts := strings.Split(raw, ",")
	// 	out := make([]int, len(parts))
	// 	for i, p := range parts {
	// 		out[i] = input.Atoi(p)
	// 	}
	// 	return out
	// }

	var sum int
	for _, m := range ms {
		// rebuild inputs in unified LTR orientation
		// result mask LTR
		resMaskLTR := parseResultLTR(fmtResultMask(m.result, m.resLen))
		// instructions LTR
		var ins [][]int
		for _, nin := range m.instructions {
			ins = append(ins, parseInstructionLTR(m.resLen, nin))
		}
		// counter LTR (build directly from bitMap to LTR: index 0 is leftmost)
		counter := make([]int, m.resLen)
		for i := 0; i < m.resLen; i++ {
			counter[i] = m.bitMap[m.resLen-1-i]
		}

		// Diagnostics
		resultBits := make([]int, m.resLen)
		for j := 0; j < m.resLen; j++ {
			resultBits[j] = (resMaskLTR >> j) & 1
		}
		counterParity := make([]int, m.resLen)
		for j := 0; j < m.resLen; j++ {
			counterParity[j] = counter[j] % 2
		}
		// fmt.Printf("p2[diag] resultBits(LTR): %v\n", resultBits)
		// fmt.Printf("p2[diag] counterParity(LTR): %v\n", counterParity)
		// fmt.Printf("p2[diag] instructions(LTR): %v\n", ins)

		best := Solve(ins, counter, resMaskLTR)
		if best == math.MaxInt {
			fmt.Printf("p2[Rational]: infeasible or not found\n")
			continue
		}
		sum += best
		// fmt.Printf("p2[Rational]: presses=%d\n", best)
	}
	fmt.Printf("p2: %d\n", sum)
}

func main() {
	p1()
	p2()
}

// fmtResultMask formats an existing parsed mask into bracket string to reuse LTR parsing.
// It outputs a string like "[.#..]" with length resLen, bit 0 is leftmost.
func fmtResultMask(mask int, resLen int) string {
	var b strings.Builder
	b.WriteByte('[')
	for i := 0; i < resLen; i++ {
		if ((mask >> (resLen - 1 - i)) & 1) == 1 {
			b.WriteByte('#')
		} else {
			b.WriteByte('.')
		}
	}
	b.WriteByte(']')
	return b.String()
}

// fmtCounterMap formats bitMap into "{a,b,c,...}" in LTR order (index 0 is leftmost)
func fmtCounterMap(bitMap map[int]int, resLen int) string {
	var b strings.Builder
	b.WriteByte('{')
	for i := 0; i < resLen; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(fmt.Sprintf("%d", bitMap[i]))
	}
	b.WriteByte('}')
	return b.String()
}

// rat1 returns a new *big.Rat set to 1
func rat1() *big.Rat { return big.NewRat(1, 1) }

// rat0 returns a new *big.Rat set to 0
func rat0() *big.Rat { return big.NewRat(0, 1) }

// buildRationalGrid builds a rows x cols grid of *big.Rat from instructions
// rows = m bits, cols = n buttons; entry[r][c] = 1 if button c covers bit r, else 0.
func buildRationalGrid(instructions [][]int, m int) [][]*big.Rat {
	n := len(instructions)
	grid := make([][]*big.Rat, m)
	for r := 0; r < m; r++ {
		row := make([]*big.Rat, n)
		for c := 0; c < n; c++ {
			row[c] = rat0()
		}
		grid[r] = row
	}
	for c, inst := range instructions {
		for _, bit := range inst {
			grid[bit][c] = rat1()
		}
	}
	return grid
}

// buildRHS builds RHS vector (length m) from counter []int as *big.Rat
func buildRHS(counter []int) []*big.Rat {
	rhs := make([]*big.Rat, len(counter))
	for i, v := range counter {
		rhs[i] = big.NewRat(int64(v), 1)
	}
	return rhs
}

// swap rows r1, r2 in grid and rhs
func swapRows(grid [][]*big.Rat, rhs []*big.Rat, r1, r2 int) {
	grid[r1], grid[r2] = grid[r2], grid[r1]
	rhs[r1], rhs[r2] = rhs[r2], rhs[r1]
}

// swap columns c1, c2 in grid and buttons permutation (to track column reorder)
func swapCols(grid [][]*big.Rat, colPerm []int, c1, c2 int) {
	rows := len(grid)
	for r := 0; r < rows; r++ {
		grid[r][c1], grid[r][c2] = grid[r][c2], grid[r][c1]
	}
	colPerm[c1], colPerm[c2] = colPerm[c2], colPerm[c1]
}

// gaussianEliminate performs rational Gaussian elimination with row/col swaps.
// Returns truncated grid and rhs (only nonzero rows kept), and colPerm (column permutation).
func gaussianEliminate(grid [][]*big.Rat, rhs []*big.Rat) ([][]*big.Rat, []*big.Rat, []int) {
	rows := len(grid)
	if rows == 0 {
		return grid, rhs, nil
	}
	cols := len(grid[0])
	colPerm := make([]int, cols)
	for i := 0; i < cols; i++ {
		colPerm[i] = i
	}

	i := 0
	for i < rows && i < cols {
		// find pivot (row >= i, col >= i) with nonzero
		pivotFound := false
		pr, pc := -1, -1
		for c := i; c < cols && !pivotFound; c++ {
			for r := i; r < rows; r++ {
				if grid[r][c].Sign() != 0 {
					pr, pc = r, c
					pivotFound = true
					break
				}
			}
		}
		if !pivotFound {
			break
		}
		if pr != i {
			swapRows(grid, rhs, i, pr)
		}
		if pc != i {
			swapCols(grid, colPerm, i, pc)
		}
		// normalize pivot to 1
		pivot := grid[i][i]
		if pivot.Cmp(rat1()) != 0 {
			for c := i; c < cols; c++ {
				grid[i][c] = new(big.Rat).Quo(grid[i][c], pivot)
			}
			rhs[i] = new(big.Rat).Quo(rhs[i], pivot)
		}
		// eliminate other rows' column i
		for r := 0; r < rows; r++ {
			if r == i {
				continue
			}
			if grid[r][i].Sign() != 0 {
				factor := new(big.Rat).Quo(grid[r][i], grid[i][i]) // grid[i][i] is 1
				for c := i; c < cols; c++ {
					sub := new(big.Rat).Mul(factor, grid[i][c])
					grid[r][c] = new(big.Rat).Sub(grid[r][c], sub)
				}
				subR := new(big.Rat).Mul(factor, rhs[i])
				rhs[r] = new(big.Rat).Sub(rhs[r], subR)
			}
		}
		i++
	}

	// find last nonzero row index
	lastNonZero := -1
RowsScan:
	for r := rows - 1; r >= 0; r-- {
		for c := 0; c < cols; c++ {
			if grid[r][c].Sign() != 0 {
				lastNonZero = r
				break RowsScan
			}
		}
	}
	if lastNonZero == -1 {
		// all-zero grid; require rhs all zero
		for _, v := range rhs {
			if v.Sign() != 0 {
				// infeasible; keep as is
				break
			}
		}
		return grid, rhs, colPerm
	}

	// check trailing zero rows have rhs==0 (if not, infeasible)
	for r := lastNonZero + 1; r < rows; r++ {
		if rhs[r].Sign() != 0 {
			// infeasible; we still truncate and caller can detect via solving failure
		}
	}

	// truncate to nonzero rows
	newRows := lastNonZero + 1
	grid = grid[:newRows]
	rhs = rhs[:newRows]
	return grid, rhs, colPerm
}

// parity check: counter[j]%2 == (result >> j) & 1
func parityConsistent(counter []int, resMask int) bool {
	for j := 0; j < len(counter); j++ {
		if (counter[j] % 2) != ((resMask >> j) & 1) {
			return false
		}
	}
	return true
}

// compute max presses per trailing (free) button column index (>= rows)
// buttonMasks are uint32 bitmasks of buttons in current column order.
func maxPressesPerTrailing(buttonMasks []uint32, rows int, counter []int) []int {
	var res []int
	for c := rows; c < len(buttonMasks); c++ {
		mask := buttonMasks[c]
		minNeed := math.MaxInt
		for bit := 0; bit < rows; bit++ {
			if (mask & (1 << uint(bit))) != 0 {
				if counter[bit] < minNeed {
					minNeed = counter[bit]
				}
			}
		}
		if minNeed == math.MaxInt {
			minNeed = 0
		}
		res = append(res, minNeed)
	}
	return res
}

// pressDifferencePerPress: for trailing col c, diff = 1 - sum_r grid[r][c]
func pressDifferencePerPress(grid [][]*big.Rat, rows int) []*big.Rat {
	cols := len(grid[0])
	var diffs []*big.Rat
	one := rat1()
	for c := rows; c < cols; c++ {
		sum := rat0()
		for r := 0; r < rows; r++ {
			sum = new(big.Rat).Add(sum, grid[r][c])
		}
		diff := new(big.Rat).Sub(one, sum)
		diffs = append(diffs, diff)
	}
	return diffs
}

// mixradix factors for trailing variables: base[i] = max[i]+1; factor[i] = product of previous bases
func makeFactors(maxes []int) []int {
	factors := make([]int, len(maxes))
	f := 1
	for i := range maxes {
		factors[i] = f
		f *= (maxes[i] + 1)
	}
	return factors
}

// priority queue item
type pqItem struct {
	added *big.Rat // lower is better
	at    int      // mixradix state
	index int
}

type ratPQ []*pqItem

func (h ratPQ) Len() int           { return len(h) }
func (h ratPQ) Less(i, j int) bool { return h[i].added.Cmp(h[j].added) < 0 }
func (h ratPQ) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }
func (h *ratPQ) Push(x any)        { *h = append(*h, x.(*pqItem)) }
func (h *ratPQ) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// decode trailing presses from mixradix at
func decodeTrailing(at int, factors []int, maxes []int) []int {
	k := len(factors)
	out := make([]int, k)
	for i := 0; i < k; i++ {
		base := maxes[i] + 1
		out[i] = (at / factors[i]) % base
	}
	return out
}

// compute main variable rational values given trailing presses
// presses[r] = rhs[r] - sum_i grid[r][rows+i] * trailing[i]
func computeMainPresses(grid [][]*big.Rat, rhs []*big.Rat, rows int, trailing []int) []*big.Rat {
	presses := make([]*big.Rat, rows)
	for r := 0; r < rows; r++ {
		sum := rat0()
		for i := 0; i < len(trailing); i++ {
			col := rows + i
			if grid[r][col].Sign() != 0 && trailing[i] != 0 {
				t := big.NewRat(int64(trailing[i]), 1)
				sum = new(big.Rat).Add(sum, new(big.Rat).Mul(grid[r][col], t))
			}
		}
		presses[r] = new(big.Rat).Sub(rhs[r], sum)
	}
	return presses
}

// all main presses are integers and >= 0? Return ints if feasible.
func mainPressesFeasibleInt(presses []*big.Rat) (bool, []int) {
	ints := make([]int, len(presses))
	for i, p := range presses {
		if p.Sign() < 0 {
			return false, nil
		}
		// integer check: denominator must be 1
		den := p.Denom()
		if den.Cmp(big.NewInt(1)) != 0 {
			return false, nil
		}
		num := p.Num()
		if !num.IsInt64() {
			// Should not happen for our sizes, but guard anyway
			return false, nil
		}
		ints[i] = int(num.Int64())
	}
	return true, ints
}

// sum of trailing + main ints
func totalPresses(mainInts []int, trailing []int) int {
	sum := 0
	for _, v := range mainInts {
		sum += v
	}
	for _, v := range trailing {
		sum += v
	}
	return sum
}

// start state: for each trailing variable, if diff<0 set to max, else 0
func initialAt(diffs []*big.Rat, maxes []int, factors []int) int {
	at := 0
	for i, diff := range diffs {
		p := 0
		if diff.Cmp(rat0()) < 0 {
			p = maxes[i]
		}
		at += p * factors[i]
	}
	return at
}

// Solve end-to-end solver returning minimal presses or math.MaxInt
// - instructions: [][]int, each inner slice lists bit indices the button toggles/affects.
// - counter: []int per-bit target counts.
// - resMask: target parity mask (must be consistent with counter parity).
func Solve(instructions [][]int, counter []int, resMask int) int {
	m := len(counter)
	n := len(instructions)

	// parity check disabled to allow solver to attempt finding feasible integer solutions
	// if !parityConsistent(counter, resMask) {
	// 	return math.MaxInt
	// }

	// build masks in column order (buttons)
	buttonMasks := make([]uint32, n)
	for i, inst := range instructions {
		var mask uint32
		for _, bit := range inst {
			mask |= 1 << uint(bit)
		}
		buttonMasks[i] = mask
	}

	// grid and rhs
	grid := buildRationalGrid(instructions, m)
	rhs := buildRHS(counter)

	// elimination with row/col swaps
	grid, rhs, colPerm := gaussianEliminate(grid, rhs)
	rows := len(grid)
	if rows == 0 {
		// no constraints; only zero counter is feasible
		allZero := true
		for _, c := range counter {
			if c != 0 {
				allZero = false
				break
			}
		}
		if allZero {
			return 0
		}
		return math.MaxInt
	}

	// reorder buttonMasks by colPerm to match current columns
	reordered := make([]uint32, len(buttonMasks))
	for c := 0; c < len(buttonMasks); c++ {
		reordered[c] = buttonMasks[colPerm[c]]
	}
	buttonMasks = reordered

	// compute trailing bounds and diffs
	maxes := maxPressesPerTrailing(buttonMasks, rows, counter)
	diffs := pressDifferencePerPress(grid, rows)
	factors := makeFactors(maxes)

	// PQ search
	start := initialAt(diffs, maxes, factors)
	seen := make(map[int]struct{})
	pq := &ratPQ{}
	heap.Init(pq)
	heap.Push(pq, &pqItem{added: rat0(), at: start})

	for pq.Len() > 0 {
		item := heap.Pop(pq).(*pqItem)
		at := item.at
		added := item.added

		if _, ok := seen[at]; ok {
			continue
		}
		seen[at] = struct{}{}

		trailing := decodeTrailing(at, factors, maxes)
		mainPress := computeMainPresses(grid, rhs, rows, trailing)
		if ok, mainInts := mainPressesFeasibleInt(mainPress); ok {
			return totalPresses(mainInts, trailing)
		}

		// neighbors
		for i := 0; i < len(maxes); i++ {
			curr := trailing[i]
			diff := diffs[i]
			factor := factors[i]
			// try decrement if diff<0 and curr>0
			if diff.Cmp(rat0()) < 0 && curr > 0 {
				newAt := at - factor
				newAdded := new(big.Rat).Sub(added, diff)
				if _, ok := seen[newAt]; !ok {
					heap.Push(pq, &pqItem{added: newAdded, at: newAt})
				}
			}
			// try increment if curr<max
			if curr < maxes[i] {
				newAt := at + factor
				newAdded := new(big.Rat).Add(added, diff)
				if _, ok := seen[newAt]; !ok {
					heap.Push(pq, &pqItem{added: newAdded, at: newAt})
				}
			}
		}
	}
	return math.MaxInt
}
