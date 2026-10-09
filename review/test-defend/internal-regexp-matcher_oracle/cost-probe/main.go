package main

import (
	"fmt"
	rx "github.com/system-inc/adamic/internal/regexp"
	"strings"
)

func main() {
	p, err := rx.Compile("a", "")
	if err != nil {
		panic(err)
	}
	for _, limit := range []uint64{0, 1000} {
		r := p.New()
		r.StepLimit = limit
		m, err := r.ExecString(strings.Repeat("b", 2000) + "a")
		fmt.Printf("limit=%d match=%+v error=%v\n", limit, m, err)
	}
}
