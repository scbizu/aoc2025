package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/magejiCoder/magejiAoc/input"
)

type ingRange struct {
	start, end int
}

func (r ingRange) IsInRange(n int) bool {
	return n >= r.start && n <= r.end
}

func (r ingRange) merge(r2 ingRange) ingRange {
	nr := r
	if r2.start < r.start {
		nr.start = r2.start
	}
	if r2.end > r.end {
		nr.end = r2.end
	}
	return nr
}

func countLen(rg ingRange) int {
	return rg.end - rg.start + 1
}

func p2() {
	txt := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	var rgs []ingRange
	txt.ReadByBlock(ctx, "\n\n", func(block []string) error {
		rangeParts := block[0]
		for rp := range strings.SplitSeq(rangeParts, "\n") {
			l, r := mustParseRange(rp)
			rgs = append(rgs, ingRange{
				start: l,
				end:   r,
			})
		}
		return nil
	})
	nrangs := make(map[ingRange]struct{}, len(rgs))
	for _, rg := range rgs {
		for exg := range nrangs {
			if rg.IsInRange(exg.start) || rg.IsInRange(exg.end) || exg.IsInRange(rg.start) || exg.IsInRange(rg.end) {
				rg = rg.merge(exg)
				delete(nrangs, exg)
			}
		}
		nrangs[rg] = struct{}{}
	}
	// fmt.Printf("ranges : %v\n", nrangs)
	var all int
	for r := range nrangs {
		all += countLen(r)
	}
	fmt.Printf("p2: %d\n", all)
}

func p1() {
	txt := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	var rgs []ingRange
	var ingIds []int
	txt.ReadByBlock(ctx, "\n\n", func(block []string) error {
		rangeParts := block[0]
		for rp := range strings.SplitSeq(rangeParts, "\n") {
			l, r := mustParseRange(rp)
			rgs = append(rgs, ingRange{
				start: l,
				end:   r,
			})
		}
		ids := block[1]
		for id := range strings.SplitSeq(ids, "\n") {
			ingIds = append(ingIds, input.Atoi(id))
		}
		return nil
	})
	var i int
	for _, id := range ingIds {
		for _, r := range rgs {
			if r.IsInRange(id) {
				i++
				break
			}
		}
	}
	fmt.Printf("p1: %d\n", i)
}

func mustParseRange(raw string) (int, int) {
	parts := strings.Split(raw, "-")
	l, r := parts[0], parts[1]
	return input.Atoi(l), input.Atoi(r)
}

func main() {
	p1()
	p2()
}
