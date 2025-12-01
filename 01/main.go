package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/magejiCoder/magejiAoc/input"
)

func main() {
	p1()
	p2()
}

func p1() {
	f := input.NewTXTFile("input.txt")
	cur := int64(50)
	var zero int
	f.ReadByLine(context.TODO(), func(line string) error {
		r := from(line)
		// fmt.Printf("cur: %d, r: %v\n", cur, r)
		cur = round(turn(cur, r))
		// slog.Info("turn", "cur", cur)
		if cur == 0 {
			zero++
		}
		return nil
	})
	fmt.Printf("p1: %d\n", zero)
}

func p2() {
	f := input.NewTXTFile("input.txt")
	cur := int64(50)
	var cr int64
	f.ReadByLine(context.TODO(), func(line string) error {
		r := from(line)
		t := turn(cur, r)
		cur, cr = round2(cur, t, cr)
		// fmt.Printf("The dial is rotated %s to point at %d; it points at 0 : %d\n", line, cur, cr)
		return nil
	})
	fmt.Printf("p2: %d\n", cr)
}

type direction byte

type rotation struct {
	d direction
	n int64
}

func (r rotation) String() string {
	return fmt.Sprintf("%c-%d", r.d, r.n)
}

func turn(cur int64, r rotation) int64 {
	switch r.d {
	case 'L':
		return cur - r.n
	case 'R':
		return cur + r.n
	}
	panic("not a valid direction")
}

func from(s string) rotation {
	if s == "" {
		panic("not a valid ratation")
	}
	ds := s[0]
	n := s[1:]
	num, err := strconv.ParseInt(n, 10, 64)
	if err != nil {
		panic("not a valid number")
	}
	return rotation{
		d: direction(ds),
		n: num,
	}
}

func round(n int64) int64 {
	if n == 100 {
		return 0
	}
	if n >= 0 && n < 100 {
		return n
	}
	if n > 100 {
		return round(n - 100)
	}
	if n < 0 {
		return round(n + 100)
	}
	panic("not reached")
}

func round2(from, to int64, rd int64) (int64, int64) {
	if to == 100 || to == 0 {
		return 0, rd + 1
	}
	if from == 0 {
		if to > 0 && to < 100 {
			return to, rd
		}
		if to < 0 && to > -100 {
			return round(to + 100), rd
		}
	} else {
		if to > 0 && to < 100 {
			return to, rd
		}
		if to < 0 && to > -100 {
			return round(to + 100), rd + 1
		}
	}
	if to > 100 {
		return round2(from, to-100, rd+1)
	}
	if to < 0 {
		return round2(from, to+100, rd+1)
	}
	panic("not reached")
}
