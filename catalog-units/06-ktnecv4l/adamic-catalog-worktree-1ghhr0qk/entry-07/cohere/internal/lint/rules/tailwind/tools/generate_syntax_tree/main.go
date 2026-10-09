// Command generate_syntax_tree captures compiled utility trees from the real Tailwind engine, together
// with the {order, count} reading the engine takes from each, so that the Go AST and PropertySort
// in internal/tailwind are tested against measurements rather than against a reading of the
// TypeScript source.
//
// Usage:
//
//	go run ./internal/lint/rules/tailwind/tools/generate_syntax_tree -package-root <tailwindcss package> [-output <file.json>] [-check]
//
// The distinction this tool exists to enforce is the same one generate_data_type enforces, and
// it bites harder here. `getPropertySort` looks like a tree walk and is not one: it is a
// breadth-first queue that descends only into `rule` and `at-rule`. A port that reads the source
// quickly and reaches for the file's own `walk` helper gets a depth-first traversal that also
// descends into `at-root`, and that port agrees with the engine on the overwhelming majority of
// real input. The fixture carries the trees where it does not.
//
// `-check` regenerates and fails when the committed fixture disagrees, which is what makes a
// Tailwind upgrade that changes a tree shape or a reading show up as a failing gate instead of a
// silent behavior change. Like generate_data_type and generate_collapse, nothing invokes it
// automatically; it is a manual gate.
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
	output := flag.String("output", filepath.Join("internal", "tailwind", "testdata", "ast_fixtures.json"), "where to write the fixture")
	check := flag.Bool("check", false, "regenerate and fail if the committed fixture disagrees")
	flag.Parse()

	if *packageRoot == "" {
		fmt.Fprintln(os.Stderr, "-package-root is required")
		os.Exit(2)
	}

	generated, err := enumerate(*packageRoot)
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
		// Go whitespace: this tool comparing its own committed fixture with its own output.
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
		TailwindVersion    string `json:"tailwindVersion"`
		ClassListSize      int    `json:"classListSize"`
		DistinctSignatures int    `json:"distinctSignatures"`
		FullList           struct {
			Compiled                          int `json:"compiled"`
			BreadthFirstDiffersFromDepthFirst int `json:"breadthFirstDiffersFromDepthFirst"`
			WithAtRootSubtree                 int `json:"withAtRootSubtree"`
		} `json:"fullList"`
		Cases          []json.RawMessage `json:"cases"`
		TraversalCases []json.RawMessage `json:"traversalCases"`
	}
	if err := json.Unmarshal(generated, &summary); err != nil {
		fmt.Fprintf(os.Stderr, "summarize: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(
		os.Stderr,
		"wrote %s: tailwind %s, %d cases over %d distinct tree shapes, %d traversal-divergence cases; "+
			"of %d compiled candidates in the full class list, %d carry an at-root subtree and %d visit differently breadth-first than depth-first\n",
		*output,
		summary.TailwindVersion,
		len(summary.Cases),
		summary.DistinctSignatures,
		len(summary.TraversalCases),
		summary.FullList.Compiled,
		summary.FullList.WithAtRootSubtree,
		summary.FullList.BreadthFirstDiffersFromDepthFirst,
	)
}

// enumerate runs the Node side and returns its JSON.
//
// The engine is JavaScript and there is no faithful way to ask it these questions from Go. The
// boundary is here, at generation time, and never at lint time: the committed fixture is what the
// test reads.
func enumerate(packageRoot string) ([]byte, error) {
	directory, err := toolDirectory()
	if err != nil {
		return nil, err
	}
	script := filepath.Join(directory, "enumerate.mjs")

	command := exec.Command("node", script, packageRoot)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("node %s: %w\n%s", script, err, stderr.String())
	}
	return stdout.Bytes(), nil
}
