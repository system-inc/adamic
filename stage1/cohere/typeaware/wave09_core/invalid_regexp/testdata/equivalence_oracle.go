package main

import (
	"bufio"
	"fmt"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"os"
	"slices"
)

func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	for _, mode := range []bool{false, true} {
		fmt.Fprintln(out, mode)
		groups := slices.Clone(esregexp.CaseEquivalenceGroups(mode))
		slices.SortFunc(groups, func(a, b []rune) int { return int(a[0] - b[0]) })
		for _, group := range groups {
			fmt.Fprint(out, "group")
			for _, r := range group {
				fmt.Fprintf(out, "\t%d", r)
			}
			fmt.Fprintln(out)
		}
		for r := rune(-1); r <= 1114112; r++ {
			members := esregexp.CaseEquivalents(r, mode)
			if len(members) == 0 {
				continue
			}
			fmt.Fprintf(out, "member\t%d", r)
			for _, member := range members {
				fmt.Fprintf(out, "\t%d", member)
			}
			fmt.Fprintln(out)
		}
	}
}
