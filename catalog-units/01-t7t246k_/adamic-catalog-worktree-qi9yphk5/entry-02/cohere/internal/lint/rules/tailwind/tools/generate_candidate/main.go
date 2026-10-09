// Command generate_candidate captures candidate readings from the real Tailwind candidate
// parser, so that the Go port in internal/lint/rules/tailwind/collapse/candidate.go is tested against measurements
// rather than against a reading of the TypeScript source.
//
// Usage:
//
//	go run ./internal/lint/rules/tailwind/tools/generate_candidate -package-root <tailwindcss package> [-corpus-root <repo>] [-theme-entry <entry.css>] [-output <file.json>] [-check]
//
// The distinction this tool enforces is order. `parseCandidate` is a generator, roughly a fifth of
// the classes a real repository writes yield more than one candidate, and the first that compiles
// wins. A port that yields the same set in a different order produces a different final answer with
// no error raised anywhere. `border-b` is the case that names it: `border-b` as an exact functional
// root comes first, `border` with the named value `b` second. A fixture that recorded candidates as
// an unordered set would agree with a port that had them backwards.
//
// The fixture also carries the design system's own tables, because `parseCandidate` is not a pure
// function of its input: it asks `utilities.has`, `variants.has`, `variants.kind` and
// `variants.compoundsWith`, and those answers come from this repository's `@theme` and `@utility`
// blocks as much as from the framework. Capturing them is what makes the Go test a test of the
// parser rather than of a hand-typed stub.
//
// `-check` regenerates and fails when the committed fixture disagrees, which is what makes a
// Tailwind upgrade that changes parsing show up as a failing gate instead of a silent behavior
// change. Like the other internal/lint/rules/tailwind/tools/generate_* tools, nothing invokes it automatically; it is a manual
// gate.
//
// Node is required to run it and never to use the result.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	packageRoot := flag.String("package-root", "", "path to the tailwindcss package root (the directory holding package.json)")
	corpusRoot := flag.String("corpus-root", "", "path to a repository to scan for the classes it actually writes")
	themeEntry := flag.String("theme-entry", "", "path to the repository's CSS entry point, so its own @utility and @theme blocks are in the tables")
	output := flag.String("output", filepath.Join("internal", "tailwind", "testdata", "candidate_fixtures.json"), "where to write the fixture")
	check := flag.Bool("check", false, "regenerate and fail if the committed fixture disagrees")
	flag.Parse()

	if *packageRoot == "" {
		fmt.Fprintln(os.Stderr, "-package-root is required")
		os.Exit(2)
	}

	generated, err := enumerate(*packageRoot, *corpusRoot, *themeEntry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "enumerate: %v\n", err)
		os.Exit(1)
	}

	if *check {
		committed, err := os.ReadFile(*output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read committed fixture: %v\n", err)
			os.Exit(1)
		}
		// Go whitespace: the tool's own fixture file against its own output, a trailing newline at most.
		if !bytes.Equal(bytes.TrimSpace(committed), bytes.TrimSpace(generated)) {
			fmt.Fprintf(os.Stderr, "%s is out of date; rerun without -check\n", *output)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "%s is up to date\n", *output)
		return
	}

	if err := os.WriteFile(*output, generated, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", *output, err)
		os.Exit(1)
	}

	var summary struct {
		TailwindVersion          string            `json:"tailwindVersion"`
		ThemeEntry               *string           `json:"themeEntry"`
		UtilityRootCount         int               `json:"utilityRootCount"`
		VariantRootCount         int               `json:"variantRootCount"`
		ClassNameOccurrences     int               `json:"classNameOccurrences"`
		RepositoryClassCount     int               `json:"repositoryClassCount"`
		AmbiguousCount           int               `json:"ambiguousCount"`
		AmbiguousRepositoryCount int               `json:"ambiguousRepositoryCount"`
		Cases                    []json.RawMessage `json:"cases"`
	}
	if err := json.Unmarshal(generated, &summary); err != nil {
		fmt.Fprintf(os.Stderr, "summarize: %v\n", err)
		os.Exit(1)
	}

	entry := "framework defaults only"
	if summary.ThemeEntry != nil {
		entry = *summary.ThemeEntry
	}
	fmt.Fprintf(
		os.Stderr,
		"wrote %s: tailwind %s via %s, %d utility roots and %d variant roots; "+
			"%d classes read, of which %d yield more than one candidate; "+
			"the repository contributed %d distinct classes from %d occurrences, %d of them ambiguous\n",
		*output,
		summary.TailwindVersion,
		entry,
		summary.UtilityRootCount,
		summary.VariantRootCount,
		len(summary.Cases),
		summary.AmbiguousCount,
		summary.RepositoryClassCount,
		summary.ClassNameOccurrences,
		summary.AmbiguousRepositoryCount,
	)
}

// enumerate runs the Node side and returns its JSON.
//
// The engine is JavaScript and there is no faithful way to ask it these questions from Go. The
// boundary is here, at generation time, and never at lint time: the committed fixture is what the
// test reads.
func enumerate(packageRoot, corpusRoot, themeEntry string) ([]byte, error) {
	directory, err := toolDirectory()
	if err != nil {
		return nil, err
	}
	script := filepath.Join(directory, "enumerate.mjs")

	// The Node side reads its arguments positionally, so an empty corpus root still has to occupy
	// its slot when a theme entry follows it.
	arguments := []string{script, packageRoot, corpusRoot, themeEntry}

	command := exec.Command("node", arguments...)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("node %s: %w\n%s", script, err, stderr.String())
	}
	return stdout.Bytes(), nil
}
