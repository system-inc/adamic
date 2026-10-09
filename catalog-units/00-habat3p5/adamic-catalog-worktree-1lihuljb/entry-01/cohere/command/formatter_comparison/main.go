// Command formatdiff measures typescript-go's formatter against an already-Prettier-formatted tree
// and reports the taxonomy of what differs.
//
// It writes nothing. The corpus is read-only and the tool's whole value is the classification, so
// there is no --write and no --fix: this answers whether a reconciliation layer is worth building
// before anyone builds one.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/system-inc/cohere/internal/format/comparison"
)

func main() {
	root := flag.String("root", "", "the tree to measure (must already be formatted by our Prettier)")
	fileList := flag.String("files", "", "a newline-delimited file of paths to measure, relative to -root; without it the tree is walked")
	limit := flag.Int("limit", 0, "stop after this many files (0 means all)")
	samples := flag.Int("samples", 3, "how many example lines to print per difference kind")
	flag.Parse()

	if *root == "" {
		fmt.Fprintln(os.Stderr, "formatdiff: -root is required")
		os.Exit(1)
	}

	var files []string
	var err error
	if *fileList != "" {
		files, err = readList(*root, *fileList, *limit)
	} else {
		files, err = collect(*root, *limit)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "formatdiff: %v\n", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		// An empty corpus reads exactly like a clean one, and that is the failure this whole repo
		// exists to make impossible.
		fmt.Fprintln(os.Stderr, "formatdiff: found no files to measure, refusing to report a clean run")
		os.Exit(1)
	}

	var (
		parsed, identical, differing int
		unparseable                  []string
		panicked                     []string
		kinds                        = map[string]int{}
		examples                     = map[string][]string{}
	)

	for _, file := range files {
		raw, readErr := os.ReadFile(file)
		if readErr != nil {
			continue
		}
		text := string(raw)

		formatted, ok, crashed, why := formatter_comparison.FormatFile(file, text)
		if crashed {
			panicked = append(panicked, fmt.Sprintf("%s: %s", rel(*root, file), why))
			continue
		}
		if !ok {
			unparseable = append(unparseable, file)
			continue
		}
		parsed++

		if formatted == text {
			identical++
			continue
		}
		differing++

		for kind, sample := range classify(text, formatted) {
			kinds[kind]++
			if len(examples[kind]) < *samples {
				examples[kind] = append(examples[kind], fmt.Sprintf("%s: %s", rel(*root, file), sample))
			}
		}
	}

	fmt.Printf(
		"corpus:      %d files found, %d parsed, %d unparseable, %d panicked the formatter\n",
		len(files), parsed, len(unparseable), len(panicked),
	)
	fmt.Printf("agreement:   %d identical, %d differing", identical, differing)
	if parsed > 0 {
		fmt.Printf("  (%.1f%% of parsed files differ)", 100*float64(differing)/float64(parsed))
	}
	fmt.Println()

	if len(panicked) > 0 {
		fmt.Printf("\nthe formatter panicked on %d files:\n", len(panicked))
		for _, entry := range panicked[:min(len(panicked), 10)] {
			fmt.Printf("  %s\n", entry)
		}
	}

	if len(unparseable) > 0 {
		fmt.Printf("\nunparseable (%d):\n", len(unparseable))
		for _, file := range unparseable[:min(len(unparseable), 10)] {
			fmt.Printf("  %s\n", rel(*root, file))
		}
	}

	fmt.Println("\ndifference kinds, by how many files each appears in:")
	ordered := make([]string, 0, len(kinds))
	for kind := range kinds {
		ordered = append(ordered, kind)
	}
	sort.Slice(ordered, func(a, b int) bool { return kinds[ordered[a]] > kinds[ordered[b]] })
	for _, kind := range ordered {
		fmt.Printf("\n  %-28s %5d files\n", kind, kinds[kind])
		for _, example := range examples[kind] {
			fmt.Printf("      %s\n", example)
		}
	}
}

// classify names the kinds of difference present between the two texts, with one example line each.
//
// The taxonomy is the deliverable, so this reports every kind it finds in a file rather than
// stopping at the first: a file that both loses a wrap and gains an indent is two findings, and
// collapsing it to one would understate the reconciliation cost.
func classify(before string, after string) map[string]string {
	found := map[string]string{}

	beforeLines := strings.Split(before, "\n")
	afterLines := strings.Split(after, "\n")

	if len(beforeLines) != len(afterLines) {
		found["line-count-change"] = fmt.Sprintf("%d lines -> %d lines", len(beforeLines), len(afterLines))
		return found
	}

	for index := range beforeLines {
		b, a := beforeLines[index], afterLines[index]
		if b == a {
			continue
		}

		lineNumber := index + 1
		bIndent := len(b) - len(strings.TrimLeft(b, " \t"))
		aIndent := len(a) - len(strings.TrimLeft(a, " \t"))

		switch {
		case bIndent != aIndent:
			record(found, "indentation", fmt.Sprintf("line %d: indent %d -> %d", lineNumber, bIndent, aIndent))
		case strings.TrimRight(b, " \t") == strings.TrimRight(a, " \t"):
			record(found, "trailing-whitespace", fmt.Sprintf("line %d", lineNumber))
		case collapse(b) == collapse(a):
			record(found, "interior-spacing", fmt.Sprintf("line %d: %s", lineNumber, trim(a)))
		default:
			record(found, "other", fmt.Sprintf("line %d: %s  ->  %s", lineNumber, trim(b), trim(a)))
		}
	}
	return found
}

func record(found map[string]string, kind string, example string) {
	if _, seen := found[kind]; !seen {
		found[kind] = example
	}
}

// collapse removes every space and tab, so two lines that differ only in how their tokens are
// spaced compare equal.
func collapse(line string) string {
	return strings.NewReplacer(" ", "", "\t", "").Replace(line)
}

func trim(line string) string {
	line = strings.TrimSpace(line)
	if len(line) > 72 {
		return line[:72] + "..."
	}
	return line
}

func rel(root string, file string) string {
	if relative, err := filepath.Rel(root, file); err == nil {
		return relative
	}
	return file
}

// readList takes the corpus from a file instead of walking, which is how the measurement is pinned
// to exactly the set our Prettier formats. Walking the tree cannot honor .prettierignore, .gitignore
// and Structure's defaults, and a corpus that includes ignored files measures a formatter against
// text nobody formats, producing a real-looking number about the wrong question.
func readList(root string, listPath string, limit int) ([]string, error) {
	raw, err := os.ReadFile(listPath)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if limit > 0 && len(files) >= limit {
			break
		}
		if !filepath.IsAbs(line) {
			line = filepath.Join(root, line)
		}
		files = append(files, line)
	}
	return files, nil
}

// collect walks the tree for the extensions typescript-go can parse at all.
//
// Our Prettier also formats .json, .css, .md, .graphql and .gql. Those are excluded here because
// the formatter has no parser for them, which is itself part of the finding rather than something
// this tool should paper over.
func collect(root string, limit int) ([]string, error) {
	var files []string
	skip := map[string]bool{
		"node_modules": true, ".git": true, ".next": true, ".cache": true,
		"dist": true, "build": true, "out": true,
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if skip[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if limit > 0 && len(files) >= limit {
			return filepath.SkipAll
		}
		switch filepath.Ext(path) {
		case ".ts", ".tsx", ".js", ".jsx", ".mjs":
			// Declaration files are generated, and .d.ts is not what anyone formats by hand.
			if !strings.HasSuffix(path, ".d.ts") {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}
