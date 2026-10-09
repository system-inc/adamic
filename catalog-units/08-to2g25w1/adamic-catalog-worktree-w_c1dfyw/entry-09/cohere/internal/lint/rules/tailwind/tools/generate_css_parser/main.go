// Command generate_css_parser captures parse trees from the real Tailwind CSS parser, so that
// the Go port in internal/lint/rules/tailwind/collapse/cssparser.go is tested against measurements rather than against
// a reading of the TypeScript source.
//
// Usage:
//
//	go run ./internal/lint/rules/tailwind/tools/generate_css_parser -package-root <tailwindcss package> [-stylesheets <corpus.json>] [-output <file.json>] [-check]
//
// The distinction this tool enforces is the one generate_data_type and generate_syntax_tree enforce,
// and for this component it lands on the custom-property branch. A `--foo` declaration keeps the
// literal newlines and indentation of its source, because that branch slices the input directly
// rather than routing bytes through the whitespace-collapsing buffer every other declaration uses.
// A port that treats custom properties as ordinary declarations returns the same values with
// interior whitespace collapsed to single spaces: plausible on inspection, wrong on every real
// theme entry, and invisible without a measured fixture.
//
// `-stylesheets` names a JSON array of `{name, path}` so the corpus carries a real repository's
// theme and its whole @import graph, which is what the exit criterion asks for. Without it only the
// synthetic branch-coverage cases are captured.
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
	stylesheets := flag.String("stylesheets", "", "path to a JSON array of {name, path} naming real stylesheets to include")
	output := flag.String("output", filepath.Join("internal", "tailwind", "testdata", "cssparser_fixtures.json"), "where to write the fixture")
	check := flag.Bool("check", false, "regenerate and fail if the committed fixture disagrees")
	flag.Parse()

	if *packageRoot == "" {
		fmt.Fprintln(os.Stderr, "-package-root is required")
		os.Exit(2)
	}

	generated, err := enumerate(*packageRoot, *stylesheets)
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
		TailwindVersion    string            `json:"tailwindVersion"`
		ParseExportName    string            `json:"parseExportName"`
		StylesheetCount    int               `json:"stylesheetCount"`
		ThemeEntryCount    int               `json:"themeEntryCount"`
		UtilityBlockCount  int               `json:"utilityBlockCount"`
		CustomVariantCount int               `json:"customVariantCount"`
		Cases              []json.RawMessage `json:"cases"`
		ErrorCases         []json.RawMessage `json:"errorCases"`
	}
	if err := json.Unmarshal(generated, &summary); err != nil {
		fmt.Fprintf(os.Stderr, "summarize: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(
		os.Stderr,
		"wrote %s: tailwind %s, parse() located by identity as %s, %d parse cases and %d error cases; "+
			"across %d real stylesheets the engine saw %d @theme entries, %d @utility blocks, %d @custom-variant rules\n",
		*output,
		summary.TailwindVersion,
		summary.ParseExportName,
		len(summary.Cases),
		len(summary.ErrorCases),
		summary.StylesheetCount,
		summary.ThemeEntryCount,
		summary.UtilityBlockCount,
		summary.CustomVariantCount,
	)
}

// enumerate runs the Node side and returns its JSON.
//
// The engine is JavaScript and there is no faithful way to ask it these questions from Go. The
// boundary is here, at generation time, and never at lint time: the committed fixture is what the
// test reads.
func enumerate(packageRoot, stylesheets string) ([]byte, error) {
	directory, err := toolDirectory()
	if err != nil {
		return nil, err
	}
	script := filepath.Join(directory, "enumerate.mjs")

	arguments := []string{script, packageRoot}
	if stylesheets != "" {
		arguments = append(arguments, stylesheets)
	}

	command := exec.Command("node", arguments...)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("node %s: %w\n%s", script, err, stderr.String())
	}
	return stdout.Bytes(), nil
}
