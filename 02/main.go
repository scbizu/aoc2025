package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/magejiCoder/magejiAoc/input"
	"github.com/magejiCoder/magejiAoc/math"
)

type IDRange struct {
	left  string
	right string
}

func (idr IDRange) countInvalid() int64 {
	var invaild int
	for i := input.Atoi(idr.left); i <= input.Atoi(idr.right); i++ {
		if digitCount(i)%2 == 1 {
			continue
		}
		p := math.Power(10, digitCount(i)/2)
		if i%p == i/p {
			invaild += i
		}
	}
	return int64(invaild)
}

func (idr IDRange) countEveryInvalid() int64 {
	var invaild int
	seen := make(map[int]struct{})
	for i := input.Atoi(idr.left); i <= input.Atoi(idr.right); i++ {
		// fmt.Printf("on num: %d\n", i)
		for e := 1; e <= digitCount(i); e++ {
			if digitCount(i)%e != 0 {
				continue
			}
			// fmt.Printf("every: %d\n", e)
			next := i
			var part int
			for {
				nx := next / math.Power(10, e)
				if nx == 0 {
					break
				}
				pt := next % math.Power(10, e)
				if pt == 0 || (pt != part && part != 0) {
					// fmt.Printf("pt not eq %d <> %d\n", pt, part)
					break
				}
				if pt == nx && (pt == part || part == 0) {
					if _, ok := seen[i]; !ok {
						// fmt.Printf("Every %d,Part:%d,Invalid ID: %d\n", e, part, i)
						invaild += i
					}
					seen[i] = struct{}{}
				}
				next = nx
				part = pt
			}
		}
	}
	return int64(invaild)
}

func digitCount(i int) int {
	if i == 0 {
		return 0
	}
	return digitCount(i/10) + 1
}

func p1() {
	t := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	var invalid int
	t.ReadByBlock(ctx, ",", func(block []string) error {
		for _, ids := range block {
			parts := strings.Split(ids, "-")
			l, r := parts[0], parts[1]
			idr := IDRange{
				left:  l,
				right: r,
			}
			// fmt.Printf("%s-%s: %d\n", l, r, idr.countInvalid())
			invalid += int(idr.countInvalid())
		}
		return nil
	})
	fmt.Printf("p1: %d\n", invalid)
}

func p2() {
	t := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	var invalid int
	t.ReadByBlock(ctx, ",", func(block []string) error {
		for _, ids := range block {
			parts := strings.Split(ids, "-")
			l, r := parts[0], parts[1]
			idr := IDRange{
				left:  l,
				right: r,
			}
			// fmt.Printf("%s-%s: %d\n", l, r, idr.countInvalid())
			invalid += int(idr.countEveryInvalid())
		}
		return nil
	})
	fmt.Printf("p2: %d\n", invalid)
}

func main() {
	p1()
	p2()
}
