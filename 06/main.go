package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/magejiCoder/magejiAoc/input"
)

type calculator struct {
	op      byte
	numbers []int
}

func (c calculator) calc() int {
	// fmt.Printf("numbers: %v,op: %s\n", c.numbers, string(c.op))
	var sum int
	switch c.op {
	case '+':
		for _, n := range c.numbers {
			sum += n
		}
	case '*':
		sum = 1
		for _, n := range c.numbers {
			sum *= n
		}
	}
	return sum
}

func resolveRow(r string) []string {
	parts := strings.SplitSeq(r, " ")
	var row []string
	for part := range parts {
		part = strings.ReplaceAll(part, " ", "")
		if part == "" {
			continue
		}
		row = append(row, part)
	}
	return row
}

func p1() {
	t := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	var rows []string
	t.ReadByLine(ctx, func(line string) error {
		rows = append(rows, line)
		return nil
	})
	problems := make(map[int][]string)
	lastRow := len(rows) - 1
	for _, r := range rows {
		parts := resolveRow(r)
		for i, p := range parts {
			problems[i] = append(problems[i], p)
		}
	}
	var sum int
	for _, problem := range problems {
		c := calculator{}
		var numbers []int
		for i, p := range problem {
			if i == lastRow {
				continue
			}
			numbers = append(numbers, input.Atoi(p))
		}
		c.numbers = numbers
		c.op = problem[lastRow][0]
		sum += c.calc()
	}
	fmt.Printf("p1: %d\n", sum)
}

func p2() {
	t := input.NewTXTFile("input.txt")
	ctx := context.TODO()
	var rows []string
	t.ReadByLine(ctx, func(line string) error {
		rows = append(rows, line)
		return nil
	})
	lastRow := len(rows) - 1
	var indexLocs []int
	for i, c := range rows[lastRow] {
		if c != ' ' {
			indexLocs = append(indexLocs, i)
		}
	}
	rowMap := make(map[int][]string)
	for rindex, r := range rows {
		if rindex == lastRow {
			continue
		}
		var next int
		var parts []string
		for i := 1; i < len(indexLocs); i++ {
			parts = append(parts, r[next:indexLocs[i]])
			rowMap[i-1] = append(rowMap[i-1], r[next:indexLocs[i]])
			next = indexLocs[i]
		}
		parts = append(parts, r[next:])
		rowMap[len(parts)-1] = append(rowMap[len(parts)-1], parts[len(parts)-1])
	}
	var sum int
	for i, r := range rowMap {
		vr := make(map[int][]string)
		for _, e := range r {
			for i := 0; i < len(e); i++ {
				vr[i] = append(vr[i], string(e[i]))
			}
		}
		cal := calculator{
			op: byte(rows[lastRow][indexLocs[i]]),
		}
		var numbers []int
		for _, vs := range vr {
			s := strings.Join(vs, "")
			s = strings.ReplaceAll(s, " ", "")
			if s != "" {
				numbers = append(numbers, input.Atoi(s))
			}
		}
		// fmt.Printf("number: %v,op: %v\n", numbers, string(rows[lastRow][indexLocs[i]]))
		cal.numbers = numbers
		sum += cal.calc()
	}
	fmt.Printf("p2: %d\n", sum)
}

func main() {
	p1()
	p2()
}
