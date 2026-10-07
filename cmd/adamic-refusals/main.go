// Command adamic-refusals requires deliberate refused constructs to keep their diagnostics.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/system-inc/adamic/internal/refusalprobe"
)

func main() { os.Exit(run()) }
func run() int {
	seed := flag.Uint64("seed", 1, "deterministic writer seed")
	count := flag.Int("count", 2000, "programs to check")
	root := flag.String("root", ".", "compiler repository for catalog completeness audit")
	out := flag.String("out", "", "directory for finding programs and neighbors")
	stop := flag.Bool("stop", false, "stop at the first finding")
	only := flag.String("only", "", "comma-separated executable entries for targeted regression runs")
	catalog := flag.Bool("catalog", false, "print all catalog entries and explicit boundaries")
	flag.Parse()
	if *count < 0 {
		fmt.Fprintln(os.Stderr, "count must be nonnegative")
		return 2
	}
	if err := refusalprobe.ValidateCatalog(*root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if *catalog {
		if err := json.NewEncoder(os.Stdout).Encode(refusalprobe.Catalog()); err != nil {
			return 2
		}
		return 0
	}
	var selected []string
	if *only != "" {
		selected = strings.Split(*only, ",")
		for _, name := range selected {
			if _, err := refusalprobe.GenerateEntries(*seed, 0, []string{name}); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 2
			}
		}
	}
	started := time.Now()
	counts := map[string]int{"refused-as-expected": 0}
	checked := 0
	for index := 0; index < *count; index++ {
		program, err := refusalprobe.GenerateEntries(*seed, index, selected)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		finding := refusalprobe.Check(context.Background(), program)
		checked++
		if finding == nil {
			counts["refused-as-expected"]++
			continue
		}
		counts[finding.Kind]++
		// Timing is separate from deterministic answers, and goes to stderr.
		fmt.Fprintf(os.Stderr, "finding program=%d entry=%s seconds=%.6f kind=%s\n", checked, program.Entry, time.Since(started).Seconds(), finding.Kind)
		if *out != "" {
			if err := refusalprobe.WriteFinding(*out, finding); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 2
			}
		}
		if err := json.NewEncoder(os.Stdout).Encode(finding); err != nil {
			return 2
		}
		if *stop {
			break
		}
	}
	if err := json.NewEncoder(os.Stdout).Encode(struct {
		Programs   int
		Boundaries []refusalprobe.Entry
		Counts     map[string]int
	}{checked, boundaries(), counts}); err != nil {
		return 2
	}
	if counts["refused-as-expected"] != checked {
		return 1
	}
	return 0
}

func boundaries() []refusalprobe.Entry {
	var result []refusalprobe.Entry
	for _, entry := range refusalprobe.Catalog() {
		if entry.Boundary != "" {
			result = append(result, entry)
		}
	}
	return result
}
