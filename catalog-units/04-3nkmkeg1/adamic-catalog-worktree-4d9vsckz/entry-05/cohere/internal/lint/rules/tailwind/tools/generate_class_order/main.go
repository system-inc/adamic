// Command generate_class_order captures the Tier 2 differential fixture: `{order, count}` for
// every class in every design system, as the shipped Tailwind engine computes it, together with the
// per-repository tables the Go side needs to answer the same question.
//
// Usage:
//
//	go run ./internal/lint/rules/tailwind/tools/generate_class_order [-systems <systems.json>] [-output <fixture.json>] [-check]
//
// # Why this tool exists next to six that look like it
//
// Each landed component has a `generate_*` of its own and each captures that component's inputs
// and outputs. This one captures neither: it captures a *class name* and the engine's answer, and
// nothing about the path between them. That is the whole distinction. A component fixture supplies a
// value the JavaScript side has already decoded, which is correct for testing the component and
// steps over the seam between components. `segment` is the recorded precedent — verified only
// through `InferDataType`, which launders its errors, so mutations to it changed nothing any test
// could see. A seam reachable only through a fixture that has already crossed it has the same hole.
//
// # One fixture, three scripts, two design systems
//
// The Go side needs three things per design system and two of them already have proven extractors
// owned by other tasks:
//
//   - `generate_descriptors/extract.mjs --json` writes the descriptor table, the readings.
//   - `generate_descriptor_table/context.mjs` writes the namespaces, the namespace keys, the
//     property order and the per-declaration roots.
//   - `enumerate.mjs`, here, writes the populations and the registration tables `ParseCandidate`
//     asks about.
//
// This command runs all three per system and folds them into one file. It does not reimplement the
// first two. Copying their logic would create a second definition of the descriptor table that
// drifts from the first, and the drift would surface here as a differential finding pointing at the
// Go port, which is the most expensive possible way to discover a copy-paste.
//
// The table is per repository by construction — `keysByNamespace` is keyed on that repository's own
// `@theme` — so the fixture carries one per system rather than one shared. That is the defect the
// whole port exists to fix, and a fixture that carried a single table would reproduce it inside the
// instrument built to detect it.
//
// # -check
//
// Regenerates and fails when the committed fixture disagrees, so a Tailwind upgrade that changes any
// reading shows up as a failing gate rather than a silent behaviour change. Like every other
// `generate_*`, nothing invokes it automatically; it is a manual gate.
//
// Node is required to produce the fixture and never to use it.
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
	systems := flag.String("systems", "", "path to a JSON array of {name, path} naming the design systems to measure, each path a file or a corpus spelling such as ahra:app/_theme/styles/theme.css (default: systems.json next to this command)")
	output := flag.String("output", filepath.Join("internal", "tailwind", "testdata", "classorder_fixtures.json"), "where to write the fixture")
	check := flag.Bool("check", false, "regenerate and fail if the committed fixture disagrees")
	flag.Parse()

	toolDirectory, err := toolDirectory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "locating this tool: %v\n", err)
		os.Exit(1)
	}
	systemsPath := *systems
	if systemsPath == "" {
		systemsPath = filepath.Join(toolDirectory, "systems.json")
	}

	generated, summary, err := build(toolDirectory, systemsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "build: %v\n", err)
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
			fmt.Fprintf(os.Stderr, "%s is stale: the engine now reads at least one class differently. Re-run without -check and read the diff.\n", *output)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "%s matches the engine\n", *output)
		return
	}

	if err := os.WriteFile(*output, generated, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", *output, err)
		os.Exit(1)
	}

	// Printed per system rather than totalled. Two numbers that differ say more than one number that
	// does not, and the per-repository difference is the measurement this whole port argues from.
	fmt.Fprintf(os.Stderr, "wrote %s (%.1f MB): tailwind %s, %d design systems\n",
		*output, float64(len(generated))/(1024*1024), summary.TailwindVersion, len(summary.Systems))
	for _, system := range summary.Systems {
		fmt.Fprintf(os.Stderr,
			"  %s: %d registry classes (%d engine nulls), %d corpus classes (%d engine nulls), %d distinct readings, %d utility roots\n",
			system.Name,
			system.Counts.RegistryClasses, system.Counts.RegistryNullReadings,
			system.Counts.CorpusClasses, system.Counts.CorpusNullReadings,
			system.Counts.DistinctReadings, len(system.Registration.UtilityRoots),
		)
	}
	fmt.Fprintf(os.Stderr,
		"  design systems: %d shared registry classes, %d read differently, %d present in only one\n",
		summary.Divergence.SharedClasses, summary.Divergence.DivergentReadings, summary.Divergence.UniqueToOneSystem,
	)

	// The claim this fixture exists to keep honest, checked at generation time as well as in the
	// suite. `f7d1d8d`'s discipline: an assertion that fails if the claim ever becomes trivially
	// true. If the two systems ever agreed on membership *and* on every reading, the fixture would
	// still be green and would have stopped saying anything.
	if summary.Divergence.UniqueToOneSystem == 0 && summary.Divergence.DivergentReadings == 0 {
		fmt.Fprintln(os.Stderr,
			"  the two design systems no longer diverge at all, so this fixture no longer demonstrates that a per-repository table is required")
		os.Exit(4)
	}
}

// systemDefinition is one entry of systems.json.
type systemDefinition struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// fixture is the file this command writes.
type fixture struct {
	// TailwindVersion is the engine every system was measured against. A single field rather than
	// one per system, because a fixture spanning two Tailwind versions would report version skew as
	// a port defect, and the build refuses to produce one.
	TailwindVersion string `json:"tailwindVersion"`
	// GroundTruth names the API the readings came from, so a reader knows the answers are
	// `compileAstNodes` and not the weaker public `getClassOrder`.
	GroundTruth string          `json:"groundTruth"`
	Divergence  divergenceBlock `json:"divergence"`
	Systems     []fixtureSystem `json:"systems"`
}

type divergenceBlock struct {
	SharedClasses      int               `json:"sharedClasses"`
	DivergentReadings  int               `json:"divergentReadings"`
	UniqueToOneSystem  int               `json:"uniqueToOneSystem"`
	ReadingExamples    []json.RawMessage `json:"readingExamples"`
	MembershipExamples []json.RawMessage `json:"membershipExamples"`
}

type fixtureSystem struct {
	Name               string          `json:"name"`
	EntryPoint         string          `json:"entryPoint"`
	TailwindVersion    string          `json:"tailwindVersion"`
	CorpusRoot         string          `json:"corpusRoot"`
	Counts             fixtureCounts   `json:"counts"`
	CorpusNullExamples []string        `json:"corpusNullExamples"`
	Registration       registration    `json:"registration"`
	PublicOrder        json.RawMessage `json:"publicOrder"`
	Readings           json.RawMessage `json:"readings"`
	RegistryCases      json.RawMessage `json:"registryCases"`
	CorpusCases        json.RawMessage `json:"corpusCases"`
	// DescriptorTable and Context are the per-repository halves, produced by the two extractors this
	// command orchestrates rather than reimplements.
	DescriptorTable json.RawMessage `json:"descriptorTable"`
	Context         json.RawMessage `json:"context"`
}

type fixtureCounts struct {
	RegistryClasses      int `json:"registryClasses"`
	CorpusOccurrences    int `json:"corpusOccurrences"`
	CorpusClasses        int `json:"corpusClasses"`
	RegistryNullReadings int `json:"registryNullReadings"`
	CorpusNullReadings   int `json:"corpusNullReadings"`
	DistinctReadings     int `json:"distinctReadings"`
}

type registration struct {
	Prefix       string                     `json:"prefix"`
	UtilityRoots map[string][]string        `json:"utilityRoots"`
	VariantRoots map[string]json.RawMessage `json:"variantRoots"`
}

// build runs the three scripts per design system and folds the results into one fixture.
func build(toolDirectory, systemsPath string) ([]byte, fixture, error) {
	definitionBytes, err := os.ReadFile(systemsPath)
	if err != nil {
		return nil, fixture{}, fmt.Errorf("read %s: %w", systemsPath, err)
	}
	var definitions []systemDefinition
	if err := json.Unmarshal(definitionBytes, &definitions); err != nil {
		return nil, fixture{}, fmt.Errorf("parse %s: %w", systemsPath, err)
	}
	if len(definitions) < 2 {
		// The floor is two and it is not a preference. One design system is not a population, and
		// per-repository divergence is unobservable from inside a single repository.
		return nil, fixture{}, fmt.Errorf("%s names %d design systems; at least two are required, since one is not a population", systemsPath, len(definitions))
	}

	// The populations, the registration tables and the cross-system divergence, in one pass so that
	// both systems are read from the same process and cannot drift between two.
	enumerated, err := runNode(filepath.Join(toolDirectory, "enumerate.mjs"), systemsPath)
	if err != nil {
		return nil, fixture{}, err
	}
	var built fixture
	if err := json.Unmarshal(enumerated, &built); err != nil {
		return nil, fixture{}, fmt.Errorf("parse enumerate.mjs output: %w", err)
	}
	if len(built.Systems) != len(definitions) {
		return nil, fixture{}, fmt.Errorf("enumerate.mjs returned %d systems for %d definitions", len(built.Systems), len(definitions))
	}

	// The sibling extractors live beside this command under `tools/`, so they are reached from this
	// command's own directory rather than from the working directory the build was invoked from.
	toolsDirectory := filepath.Dir(toolDirectory)
	for index := range built.Systems {
		system := &built.Systems[index]

		if system.TailwindVersion != built.TailwindVersion {
			// Version skew makes every disagreement ambiguous: a class reading differently across
			// two Tailwind versions is not a finding about this repository. Refused rather than
			// recorded, because a fixture that carries the skew makes every later run argue about it.
			return nil, fixture{}, fmt.Errorf(
				"%s is on tailwind %s and %s is on %s; a differential across two engine versions reports skew as a port defect",
				system.Name, system.TailwindVersion, built.Systems[0].Name, built.TailwindVersion,
			)
		}

		descriptorTablePath := filepath.Join(os.TempDir(), "classorder-table-"+system.Name+".json")
		if _, err := runNode(
			filepath.Join(toolsDirectory, "generate_descriptors", "extract.mjs"),
			system.EntryPoint, "--json", descriptorTablePath,
		); err != nil {
			return nil, fixture{}, fmt.Errorf("descriptor table for %s: %w", system.Name, err)
		}
		descriptorTable, err := os.ReadFile(descriptorTablePath)
		if err != nil {
			return nil, fixture{}, fmt.Errorf("read descriptor table for %s: %w", system.Name, err)
		}
		if err := os.Remove(descriptorTablePath); err != nil {
			return nil, fixture{}, fmt.Errorf("remove %s: %w", descriptorTablePath, err)
		}
		system.DescriptorTable = json.RawMessage(descriptorTable)

		context, err := runNode(
			filepath.Join(toolsDirectory, "generate_descriptor_table", "context.mjs"),
			system.EntryPoint,
		)
		if err != nil {
			return nil, fixture{}, fmt.Errorf("context for %s: %w", system.Name, err)
		}
		system.Context = json.RawMessage(context)
	}

	// Re-marshalled rather than passed through, so the committed fixture has one canonical shape and
	// `-check` compares bytes that only change when an answer changed.
	encoded, err := json.Marshal(built)
	if err != nil {
		return nil, fixture{}, fmt.Errorf("encode fixture: %w", err)
	}
	return append(encoded, '\n'), built, nil
}

// runNode executes one script and returns its stdout, with stderr carried into the error.
//
// stderr is deliberately not forwarded to this process's stderr on success: the extractors are
// chatty, three of them run per system, and burying this command's own per-system summary under
// their progress is how a summary stops being read.
func runNode(script string, arguments ...string) ([]byte, error) {
	command := exec.Command("node", append([]string{script}, arguments...)...)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("node %s: %w\n%s", filepath.Base(script), err, stderr.String())
	}
	return stdout.Bytes(), nil
}
