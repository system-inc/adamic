// Command generate_theme captures the theme the real Tailwind engine resolved for a corpus of
// stylesheets, so that the Go port in internal/lint/rules/tailwind/collapse/theme.go is tested against measurements
// rather than against a reading of the TypeScript source.
//
// Usage:
//
//	go run ./internal/lint/rules/tailwind/tools/generate_theme -package-root <tailwindcss package> [-stylesheets <corpus.json>] [-output <file.json>] [-check]
//
// The distinction this tool enforces is the one generate_data_type and generate_css_parser
// enforce, and for this component it lands on the ignored-key map. Reading the source suggests
// `--font` names every key starting `--font-`; asking the engine says `--font-weight-*` and
// `--font-size-*` are excluded, so a repository's `--color` namespace holds 567 keys while
// `keysInNamespaces(['--color'])` returns 558. A port that missed the map returns nine extra keys
// under a namespace, which is plausible on inspection and wrong on every lookup that walks them.
//
// `-stylesheets` names a JSON array of `{name, path}` so the corpus carries real repositories'
// themes and their whole `@import` graphs, which is what the exit criterion asks for: the argument
// for this port is that two repositories on the same Tailwind resolve different themes, and that is
// only checkable against two of them. Without it only the synthetic branch-coverage cases are
// captured, and the Go test skips its repository half rather than passing vacuously. A `path` spelled
// inside a corpus, such as `ahra:app/_theme/styles/theme.css`, is found through the corpus's variable or
// the corpora file and recorded as spelled, so the fixture reads the same on every machine (#sycrdr6).
//
// `-check` regenerates and fails when the committed fixture disagrees, which is what makes a
// Tailwind upgrade that changes theme resolution show up as a failing gate instead of a silent
// behavior change. Like the other internal/lint/rules/tailwind/tools/generate_* tools, nothing invokes it automatically; it is a
// manual gate.
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
	stylesheets := flag.String("stylesheets", "", "path to a JSON array of {name, path} naming real repository stylesheets to include")
	output := flag.String("output", filepath.Join("internal", "tailwind", "testdata", "theme_fixtures.json"), "where to write the fixture")
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
			fmt.Fprintf(os.Stderr, "%s is stale: the engine now resolves a different theme. Re-run without -check and read the diff.\n", *output)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "%s matches the engine\n", *output)
		return
	}

	if err := os.WriteFile(*output, generated, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", *output, err)
		os.Exit(1)
	}

	var summary struct {
		TailwindVersion     string `json:"tailwindVersion"`
		SyntheticCaseCount  int    `json:"syntheticCaseCount"`
		RepositoryCaseCount int    `json:"repositoryCaseCount"`
		Cases               []struct {
			Name   string `json:"name"`
			Source string `json:"source"`
			Theme  struct {
				Size             int `json:"size"`
				NamespaceResults []struct {
					Namespace string `json:"namespace"`
				} `json:"namespaceResults"`
				Resolutions        []json.RawMessage `json:"resolutions"`
				ResolveWithResults []json.RawMessage `json:"resolveWithResults"`
			} `json:"theme"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(generated, &summary); err != nil {
		fmt.Fprintf(os.Stderr, "summarize: %v\n", err)
		os.Exit(1)
	}

	// The per-repository line is the measurement the whole port argues from, so it is printed
	// rather than folded into a total: two numbers that differ say more than one number that does
	// not.
	fmt.Fprintf(
		os.Stderr,
		"wrote %s: tailwind %s, %d synthetic and %d repository cases\n",
		*output, summary.TailwindVersion, summary.SyntheticCaseCount, summary.RepositoryCaseCount,
	)
	for _, aCase := range summary.Cases {
		if aCase.Source != "repository" {
			continue
		}
		fmt.Fprintf(
			os.Stderr,
			"  %s: %d theme entries, %d namespaces, %d resolutions, %d resolveWith answers\n",
			aCase.Name,
			aCase.Theme.Size,
			len(aCase.Theme.NamespaceResults),
			len(aCase.Theme.Resolutions),
			len(aCase.Theme.ResolveWithResults),
		)
	}
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
