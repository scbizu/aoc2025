package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/magejiCoder/magejiAoc/input"
)

type bank struct {
	batteries []byte
	max       int
	maxs      map[int]int
}

func fromString(s string) bank {
	var bs []byte
	for _, b := range s {
		bs = append(bs, byte(b))
	}
	return bank{
		batteries: bs,
		maxs:      make(map[int]int),
	}
}

func (b bank) pick2() int {
	max := buildInt(b.batteries[0], b.batteries[1])
	for i := 0; i < len(b.batteries); i++ {
		for j := i + 1; j < len(b.batteries); j++ {
			n := buildInt(b.batteries[i], b.batteries[j])
			if n > max {
				max = n
			}
		}
	}
	return max
}

func (b *bank) pick(raw []byte, cur []byte) {
	// fmt.Printf("raw: %s,cur: %s\n", string(raw), string(cur))
	if len(cur) == 12 {
		if input.Atoi(string(cur)) > b.max {
			// fmt.Printf("current: %s\n", string(cur))
			b.max = input.Atoi(string(cur))
		}
		return
	}
	if len(raw) == 0 {
		return
	}
	if v, ok := b.maxs[len(cur)]; ok && v > input.Atoi(string(cur)) {
		// drop
		b.pick(raw[1:], cur)
		return
	}
	c := append(cur, raw[0])
	if v, ok := b.maxs[len(c)]; !ok || v <= input.Atoi(string(c)) {
		// accept
		b.pick(raw[1:], c)
		b.maxs[len(c)] = input.Atoi(string(c))
	}
	// drop
	b.pick(raw[1:], cur)
}

func buildInt(r1, r2 byte) int {
	return input.Atoi(string(r1) + string(r2))
}

func p1() {
	txt := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	var sum int
	txt.ReadByLine(ctx, func(line string) error {
		s := fromString(line)
		max := s.pick2()
		sum += max
		return nil
	})

	fmt.Printf("p1: %d\n", sum)
}

func p2() {
	txt := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	var sum int
	txt.ReadByLine(ctx, func(line string) error {
		s := fromString(line)
		btys := slices.Clone(s.batteries)
		s.pick(btys, []byte{})
		sum += s.max
		return nil
	})

	fmt.Printf("p2: %d\n", sum)
}

func main() {
	p1()
	p2()
}
