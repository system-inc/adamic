// Command generate_utility captures what the shipped Tailwind engine does with a repository's
// own `@utility` blocks, so that the Go evaluator in internal/lint/rules/tailwind/collapse/utility.go is tested against
// measurements rather than against a reading of `createCssUtility`.
//
// Usage:
//
//	go run ./internal/lint/rules/tailwind/tools/generate_utility -theme <theme.css> [-output <file.json>] [-check]
//
// The population is chosen so that a clean result is evidence rather than a coincidence, which for
// this component is not a rhetorical concern: the mechanism it ports was predicted by the
// architecture analysis and then confirmed at scale by Phase 0, so the answer was known before the
// Go was written and a suite that only confirmed it would have proved nothing about the Go.
//
// Two halves, and both are load-bearing:
//
//   - The repository half, which is the registry classes on `@utility` roots plus a shape sweep. It
//     covers the 18 known exceptions exactly as generate_descriptors reported them, and the
//     exception set is read out of that tool's own report rather than re-derived here. A first
//     attempt did re-derive it with a proxy and got 30 instead of 18, because `<root>-full` also
//     resolves through two paths and the descriptor model answers it correctly.
//
//   - The synthetic half, small standalone design systems reaching branches the repository's blocks
//     never do: `--value(percentage)` and its integer guard, `--modifier(…)` and the post-conditions
//     that gate it, `--default(…)`, and quoted literals. Each was found by mutating the Go and
//     watching nothing fail. Without them, four ported branches have coverage in the diff and
//     nothing behind it.
//
// Running the repository half invokes generate_descriptors, which takes several minutes. That is
// the right place to pay for the exception set being defined by the tool that measured it.
//
// `-check` regenerates and fails when the committed fixture disagrees, which is what makes a
// Tailwind upgrade that changes utility evaluation show up as a failing gate instead of a silent
// behavior change. Like the other internal/lint/rules/tailwind/tools/generate_* tools, nothing invokes it automatically.
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
	theme := flag.String("theme", "", "path to the repository's theme entry point (the CSS file that imports tailwindcss)")
	output := flag.String("output", filepath.Join("internal", "tailwind", "testdata", "utility_fixtures.json"), "where to write the fixture")
	check := flag.Bool("check", false, "regenerate and fail if the committed fixture disagrees")
	flag.Parse()

	if *theme == "" {
		fmt.Fprintln(os.Stderr, "-theme is required")
		os.Exit(2)
	}

	generated, err := enumerate(*theme)
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
		TailwindVersion     string            `json:"tailwindVersion"`
		UtilityBlocks       []json.RawMessage `json:"utilityBlocks"`
		PerDeclarationRoots []string          `json:"perDeclarationRoots"`
		KnownExceptions     []json.RawMessage `json:"knownExceptions"`
		ShadowQuirkProbes   []json.RawMessage `json:"shadowQuirkProbes"`
		SyntheticCases      []json.RawMessage `json:"syntheticCases"`
		Cases               []json.RawMessage `json:"cases"`
		NullReadings        int               `json:"nullReadings"`
	}
	if err := json.Unmarshal(generated, &summary); err != nil {
		fmt.Fprintf(os.Stderr, "summarize: %v\n", err)
		os.Exit(1)
	}

	// The counts are printed rather than only the filename. The number that matters is the
	// exception count: it is this component's acceptance criterion, and a run that produced a
	// different one is a change in what the port has to answer rather than a fixture refresh.
	fmt.Fprintf(
		os.Stderr,
		"wrote %s: tailwind %s, %d @utility blocks, %d roots resolving per declaration, "+
			"%d registry classes the descriptor model declines, %d shadow-quirk probes recorded and not modelled, "+
			"%d synthetic design systems, %d measured classes of which %d the engine rejects\n",
		*output,
		summary.TailwindVersion,
		len(summary.UtilityBlocks),
		len(summary.PerDeclarationRoots),
		len(summary.KnownExceptions),
		len(summary.ShadowQuirkProbes),
		len(summary.SyntheticCases),
		len(summary.Cases),
		summary.NullReadings,
	)
}

// enumerate runs the Node side and returns its JSON.
//
// The engine is JavaScript and there is no faithful way to ask it these questions from Go. The
// boundary is here, at generation time, and never at lint time: the committed fixture is what the
// test reads.
func enumerate(theme string) ([]byte, error) {
	directory, err := toolDirectory()
	if err != nil {
		return nil, err
	}
	script := filepath.Join(directory, "enumerate.mjs")

	command := exec.Command("node", script, theme)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("node %s: %w\n%s", script, err, stderr.String())
	}
	return stdout.Bytes(), nil
}
