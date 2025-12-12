package main

import (
	"context"
	"fmt"
	"math"
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
	m.pushWithJoltage(make(map[string]struct{}), visited, bitMap, m.current, 0)
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

func (m *machine) pushWithJoltage(
	pressed map[string]struct{},
	visited map[state]int,
	bitMap map[int]int,
	cur int, pushes int,
) {
	if pushes >= m.min {
		return
	}
	s := state{cur, bitmapString(bitMap)}
	if v, ok := visited[s]; ok && pushes >= v {
		return
	}
	visited[s] = pushes
	if cur == m.result {
		fmt.Printf("bitmap:%v\n", bitMap)
		// fmt.Printf("cur: %4b\n", cur)
		if matchBitMap(bitMap, m.bitMap) {
			// fmt.Printf("bitmap :%v\n", bitMap)
			if pushes < m.min {
				m.min = pushes
			}
			return
		}
	}
	for _, nin := range m.instructions {
		// if _, ok := pressed[nin]; ok {
		// 	continue
		// }
		inst := parseInstruction(m.resLen, nin)
		// pressed[nin] = struct{}{}

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
		m.pushWithJoltage(pressed, visited, bitMap, next, pushes+1)
		// 回溯
		for _, d := range deltas {
			bitMap[d.idx] -= d.add
			if bitMap[d.idx] == 0 {
				delete(bitMap, d.idx)
			}
		}
		// delete(pressed, nin)
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
	var sum int
	for _, m := range ms {
		fmt.Printf("counter: %v\n", m.bitMap)
		sum += m.reachV2()
	}
	fmt.Printf("p2: %d\n", sum)
}

func main() {
	p1()
	p2()
}
