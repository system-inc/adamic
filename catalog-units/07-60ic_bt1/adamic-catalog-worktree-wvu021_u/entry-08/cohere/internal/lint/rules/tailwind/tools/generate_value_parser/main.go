// Command generate_value_parser captures the node tree the real Tailwind value parser produces
// for a corpus of CSS value shapes, so that the Go port in internal/tailwind is tested against
// measurements rather than against a reading of the TypeScript source.
//
// Usage:
//
//	go run ./internal/lint/rules/tailwind/tools/generate_value_parser -package-root <tailwindcss package> [-output <file.json>] [-check]
//
// The distinction this tool exists to enforce is the same one generate_data_type enforces:
// reading the source and asking the engine are different claims. For this parser they disagree in
// places a careful reader would not predict, because several behaviors fall out of JavaScript
// semantics rather than intent. `a)b` parses to a single word `b`, the pending `a` discarded by an
// optional chain against an empty stack. A trailing `\` produces the literal word `\undefined`.
// `foo(bar` puts `bar` at the top level rather than inside the function it visually sits in. A
// transcription would plausibly get all three wrong, and each wrong answer is well-formed enough to
// survive review.
//
// `-check` regenerates and fails when the committed fixture disagrees, which is what makes a
// Tailwind upgrade that changes a tree show up as a failing gate instead of a silent behavior
// change. Like generate_data_type and generate_collapse, nothing invokes it automatically;
// it is a manual gate.
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
	output := flag.String("output", filepath.Join("internal", "tailwind", "testdata", "valueparser_fixtures.json"), "where to write the fixture")
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
			fmt.Fprintf(os.Stderr, "read %s: %v\n", *output, err)
			os.Exit(1)
		}
		// Go whitespace: this tool comparing its own committed fixture with its own output.
		if !bytes.Equal(bytes.TrimSpace(committed), bytes.TrimSpace(generated)) {
			fmt.Fprintf(os.Stderr, "%s is stale: the engine now parses differently. Re-run without -check and read the diff.\n", *output)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "%s matches the engine\n", *output)
		return
	}

	if err := os.WriteFile(*output, generated, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", *output, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", *output)
}

// enumerate runs the Node side and returns its JSON, verified to be well-formed and non-trivial.
//
// The size check is not decoration. A corpus that silently shrank to a handful of cases is the
// failure mode this whole approach exists to prevent, and it would otherwise look like a passing
// run.
func enumerate(packageRoot string) ([]byte, error) {
	directory, err := toolDirectory()
	if err != nil {
		return nil, err
	}
	script := filepath.Join(directory, "enumerate.mjs")

	command := exec.Command("node", script, packageRoot)
	command.Stderr = os.Stderr
	out, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("node %s: %w", script, err)
	}

	var parsed struct {
		TailwindVersion string `json:"tailwindVersion"`
		ParseSymbol     string `json:"parseSymbol"`
		Cases           []any  `json:"cases"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, fmt.Errorf("parse enumeration: %w", err)
	}
	if parsed.TailwindVersion == "" {
		return nil, fmt.Errorf("enumeration reported no Tailwind version")
	}
	if parsed.ParseSymbol == "" {
		return nil, fmt.Errorf("enumeration did not report which bundle symbol it extracted")
	}
	if len(parsed.Cases) < 250 {
		return nil, fmt.Errorf("enumeration produced only %d cases; the corpus is meant to be several hundred shapes", len(parsed.Cases))
	}
	return out, nil
}
