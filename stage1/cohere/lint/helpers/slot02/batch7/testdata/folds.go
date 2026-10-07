package main

import (
	"encoding/json"
	"os"
	"unicode"
)

func main() {
	pairs := []int{}
	for r := rune(0); r <= unicode.MaxRune; r++ {
		if r >= 0xD800 && r <= 0xDFFF {
			continue
		}
		m := r
		for q := unicode.SimpleFold(r); q != r; q = unicode.SimpleFold(q) {
			if q < m {
				m = q
			}
		}
		if m != r {
			pairs = append(pairs, int(r), int(m))
		}
	}
	data, err := json.Marshal(struct {
		Version string
		Pairs   []int
	}{unicode.Version, pairs})
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(os.Args[1], data, 0644); err != nil {
		panic(err)
	}
}
