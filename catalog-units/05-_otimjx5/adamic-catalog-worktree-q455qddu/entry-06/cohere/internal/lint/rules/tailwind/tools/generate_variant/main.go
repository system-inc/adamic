// Command generate_variant captures how the real Tailwind engine orders variants, so that the
// Go port in internal/lint/rules/tailwind/collapse/variant.go is tested against measurements rather than against a
// reading of the TypeScript source.
//
// Usage:
//
//	go run ./internal/lint/rules/tailwind/tools/generate_variant -package-root <tailwindcss package> [-corpus <corpus.json>] [-output <file.json>] [-check]
//
// The distinction this tool enforces is the one generate_theme enforces, and for this component
// it lands on `getVariantOrder`. Reading `variants.ts` suggests every variant holds a fixed
// position, and cohere has shipped a 145-entry table of exactly that shape. Asking the engine says
// the position is assigned per run over only the variants that were parsed, densely, with ties
// collapsed: `hover` is index 0 in a run that parsed two variants and index 1 in a run that parsed
// six. No static table can represent that, and a port built from one agrees with the engine on any
// corpus whose variant set happens to match the table's and diverges silently on every other.
//
// `-corpus` names a JSON array of `{name, path, sourceRoot}`. The stylesheet at `path` is what the
// engine loads; `sourceRoot` is the tree this tool harvests real class lists from, so the fixture
// carries the class lists two real repositories actually wrote rather than only the ones a test
// author thought to invent. Both repositories in the default corpus declare their own
// `@custom-variant dark`, which is the per-repository fact the whole port exists to respect. Either
// may be spelled inside a corpus, such as `ahra:` for ahra's root, found through the corpus's variable
// or the corpora file and recorded as spelled, so the fixture reads the same on every machine (#sycrdr6).
//
// `-check` regenerates and fails when the committed fixture disagrees, which is what makes a
// Tailwind upgrade that changes variant ordering show up as a failing gate instead of a silent
// behaviour change. Like the other internal/lint/rules/tailwind/tools/generate_* tools, nothing invokes it automatically; it is a
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
	"regexp"
	"sort"
	"strings"

	"github.com/system-inc/cohere/internal/corpus"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
)

func main() {
	packageRoot := flag.String("package-root", "", "path to the tailwindcss package root (the directory holding package.json)")
	corpusFile := flag.String("corpus", "", "path to a JSON array of {name, path, sourceRoot} naming real repositories to include")
	output := flag.String("output", filepath.Join("internal", "tailwind", "testdata", "variant_fixtures.json"), "where to write the fixture")
	maximumClassLists := flag.Int("max-class-lists", 1200, "how many harvested class lists to keep per repository")
	check := flag.Bool("check", false, "regenerate and fail if the committed fixture disagrees")
	flag.Parse()

	if *packageRoot == "" {
		fmt.Fprintln(os.Stderr, "-package-root is required")
		os.Exit(2)
	}

	generated, err := enumerate(*packageRoot, *corpusFile, *maximumClassLists)
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
			fmt.Fprintf(os.Stderr, "%s is stale: the engine now orders variants differently. Re-run without -check and read the diff.\n", *output)
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
			Name     string `json:"name"`
			Source   string `json:"source"`
			Registry struct {
				Entries []struct {
					Name string `json:"name"`
				} `json:"entries"`
			} `json:"registry"`
			ClassLists []struct {
				Classes []string `json:"classes"`
			} `json:"classLists"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(generated, &summary); err != nil {
		fmt.Fprintf(os.Stderr, "summarize: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(
		os.Stderr,
		"wrote %s: tailwind %s, %d synthetic and %d repository cases\n",
		*output, summary.TailwindVersion, summary.SyntheticCaseCount, summary.RepositoryCaseCount,
	)
	// The per-repository line is what the divergence question argues from: two repositories'
	// registries differ, and the count of real class lists is the population any agreement
	// percentage is a percentage of.
	for _, aCase := range summary.Cases {
		if aCase.Source != "repository" {
			continue
		}
		if len(aCase.Registry.Entries) > 0 {
			fmt.Fprintf(os.Stderr, "  %s: %d registered variants\n", aCase.Name, len(aCase.Registry.Entries))
		}
		if len(aCase.ClassLists) > 0 {
			classes := 0
			for _, classList := range aCase.ClassLists {
				classes += len(classList.Classes)
			}
			fmt.Fprintf(os.Stderr, "  %s: %d class lists, %d classes\n", aCase.Name, len(aCase.ClassLists), classes)
		}
	}
}

// corpusEntry is one repository in the corpus file.
type corpusEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
	// SourceRoot is the tree real class lists are harvested from. Separate from Path because the
	// stylesheet the engine loads lives inside the tree rather than beside it.
	SourceRoot string `json:"sourceRoot"`
	// ClassLists is filled in by this tool and read by the Node side.
	ClassLists [][]string `json:"classLists,omitempty"`
}

// enumerate harvests the corpus, runs the Node side, and returns its JSON.
//
// The engine is JavaScript and there is no faithful way to ask it these questions from Go. The
// boundary is here, at generation time, and never at lint time: the committed fixture is what the
// test reads.
func enumerate(packageRoot, corpusPath string, maximumClassLists int) ([]byte, error) {
	directory, err := toolDirectory()
	if err != nil {
		return nil, err
	}
	script := filepath.Join(directory, "enumerate.mjs")

	arguments := []string{script, packageRoot}

	if corpusPath != "" {
		entries, err := readCorpus(corpusPath, maximumClassLists)
		if err != nil {
			return nil, err
		}
		harvested, err := json.Marshal(entries)
		if err != nil {
			return nil, fmt.Errorf("marshal harvested corpus: %w", err)
		}
		// Written beside the script rather than passed as an argument, because a repository's
		// harvested class lists run to megabytes and an argv that size is a portability question
		// nobody should have to think about while reading a fixture.
		harvestedPath := filepath.Join(directory, "corpus.harvested.json")
		if err := os.WriteFile(harvestedPath, harvested, 0o644); err != nil {
			return nil, fmt.Errorf("write harvested corpus: %w", err)
		}
		defer os.Remove(harvestedPath)
		arguments = append(arguments, harvestedPath)
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

// readCorpus loads the corpus file and fills each entry's class lists from its source tree.
func readCorpus(corpusPath string, maximumClassLists int) ([]corpusEntry, error) {
	raw, err := os.ReadFile(corpusPath)
	if err != nil {
		return nil, fmt.Errorf("read corpus: %w", err)
	}

	var entries []corpusEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, fmt.Errorf("parse corpus: %w", err)
	}

	for index := range entries {
		if entries[index].SourceRoot == "" {
			continue
		}
		// A corpus spelling such as "ahra:" names a private checkout, found the way the tests find it.
		sourceRoot, err := corpus.Locate(entries[index].SourceRoot)
		if err != nil {
			return nil, fmt.Errorf("harvest %s: %w", entries[index].Name, err)
		}
		classLists, err := harvestClassLists(sourceRoot, maximumClassLists)
		if err != nil {
			return nil, fmt.Errorf("harvest %s: %w", entries[index].Name, err)
		}
		entries[index].ClassLists = classLists
	}

	return entries, nil
}

// classAttributePattern finds the static class strings this corpus is built from.
//
// Deliberately narrower than internal/rules/tailwind's ClassLiteralReader, and it must stay
// narrower rather than try to catch up. That reader resolves callees, template literals and
// variable-name patterns through a real TypeScript AST, and the classes it recovers from an
// interpolated template are fragments rather than lists a user wrote. What this fixture needs is
// whole class lists exactly as written, because a list is the unit the engine sorts. A quoted
// attribute value is the one form guaranteed to be a complete list, so that is the only form
// harvested. Missing a list costs corpus size, which is measurable; harvesting a fragment costs
// correctness of the comparison, which is not.
var classAttributePattern = regexp.MustCompile(`(?:class|className)\s*=\s*"([^"\n{}$]+)"`)

// harvestClassLists reads whole class lists out of a repository's source tree.
//
// Sorted and deduplicated before truncation so the fixture is reproducible: a filesystem walk is
// ordered, but "the first 1,200 lists found" would still change whenever a file was added, and a
// fixture that churns on unrelated edits stops being read.
func harvestClassLists(sourceRoot string, maximum int) ([][]string, error) {
	seen := map[string]bool{}
	var keys []string

	err := filepath.WalkDir(sourceRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			// A tree this large has directories the walker cannot enter, and none of them hold
			// class literals. Failing the whole harvest on one of them would make the fixture
			// depend on the permissions of a directory it does not read.
			return nil //nolint:nilerr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case "node_modules", ".git", ".next", "dist", "build", "coverage", "data":
				return filepath.SkipDir
			}
			return nil
		}
		switch filepath.Ext(path) {
		case ".tsx", ".ts", ".jsx", ".js":
		default:
			return nil
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			return nil //nolint:nilerr
		}

		for _, match := range classAttributePattern.FindAllSubmatch(contents, -1) {
			classes := text.WhitespaceFields(string(match[1]))
			// A single class has no order to get wrong, and a list that repeats a class is
			// `no-duplicate-classes`'s finding rather than an ordering question. Both would pass
			// any comparator and inflate the agreement count without testing it.
			if len(classes) < 2 || hasRepeat(classes) {
				continue
			}
			key := strings.Join(classes, " ")
			if seen[key] {
				continue
			}
			seen[key] = true
			keys = append(keys, key)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(keys)
	if maximum > 0 && len(keys) > maximum {
		keys = keys[:maximum]
	}

	classLists := make([][]string, 0, len(keys))
	for _, key := range keys {
		classLists = append(classLists, text.WhitespaceFields(key))
	}
	return classLists, nil
}

// hasRepeat reports whether a class list names the same class twice.
func hasRepeat(classes []string) bool {
	seen := make(map[string]bool, len(classes))
	for _, class := range classes {
		if seen[class] {
			return true
		}
		seen[class] = true
	}
	return false
}
