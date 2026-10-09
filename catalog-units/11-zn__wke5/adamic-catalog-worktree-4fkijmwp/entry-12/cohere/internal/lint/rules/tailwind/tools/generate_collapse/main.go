// Command generate_collapse regenerates the Tailwind collapse table from the installed Tailwind.
//
// Usage:
//
//	go run ./internal/lint/rules/tailwind/tools/generate_collapse -entry-point <theme.css> [-output <file.go>] [-check]
//	    [-resolve-root <dir>] [-verify-invariance-against <theme.css>[:<resolve-root>]]
//
// # The one invariance claim in this file, and how to re-measure it
//
// `CollapseFamilies` ships as a fact about Tailwind. Everything else this generator writes is a
// per-repository extraction, and `KnownStatics` proves it by carrying `markdown-content`, one of
// ahra's own `@utility` blocks.
//
// That claim was first verified by generating against ~/Projects/ahra and
// ~/Projects/connected/www-connected-app and diffing to zero. The instrument could not support the
// conclusion: both vendor the Structure submodule and import its `global.css`, so they share its
// `@theme` and `@utility` blocks and are one observation wearing two names. Re-measured against a
// system sharing nothing, the claim does hold — the same 44 families, identical down to the value
// each was discovered at, across a registry differing by 25 functional roots and 7,850 probed
// pairs. `-verify-invariance-against` is that measurement made repeatable, and it refuses a second
// system whose registry matches the first, since two systems that cannot differ cannot detect a
// table that varies.
//
//	go run ./internal/lint/rules/tailwind/tools/generate_collapse \
//	    -entry-point ~/Projects/ahra/app/_theme/styles/theme.css \
//	    -verify-invariance-against internal/lint/rules/tailwind/tools/generate_descriptor_base/testdata/independent_theme.css:tools/tailwind/generate_descriptor_base/testdata
//
// `-resolve-root` is what makes that second path work at all. Bare specifiers such as `tailwindcss`
// used to resolve from the entry point's own directory, which is correct for every repository and
// impossible for a design system checked in under `tools/` with no `node_modules` above it. The two
// are separate inputs now, defaulting to the old behaviour.
//
// The table it writes is what lets `enforce-canonical-classes` run in Go without a JavaScript engine
// at lint time. Tailwind's own collapse logic is a signature-equivalence search over the whole
// utility registry, eight to ten thousand lines that change every minor release. Porting the search
// was rejected twice for that reason. Porting its answers is a different proposition: 327 functional
// roots produce 53,301 pairs, of which 38 collapse, and the families are facts about roots rather
// than computations per class.
//
// The generator exists so that staleness is visible instead of silent. Today a Tailwind upgrade can
// change what collapses and nobody finds out, because the only way to notice is a finding that stops
// being reported. With a generated table, an upgrade produces a diff in a checked-in file, which a
// human reads in review. `-check` makes that a gate: it regenerates and fails if either committed
// table disagrees, so a version bump that changes the answers cannot land quietly.
//
// It writes two files, and `-check` covers both. `collapse_table.go` is what
// `enforce-canonical-classes` reads; `property_order_table.go` is what
// `enforce-consistent-class-order` reads. The check used to return before the second was rendered,
// so one generated file was gated and the other was written by every normal run and compared by
// nothing, under a message naming only the file it had actually read.
//
// Nothing invokes `-check` automatically yet. It is a manual gate, which is worth knowing before
// trusting that an upgrade would have been caught.
//
// Node is required to run it and never to use the result. The enumeration asks the real Tailwind
// engine, because a hand-maintained table is a guess and the guesses have been wrong: an earlier
// attempt enumerated from the classes this codebase happens to use and missed nine families.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// enumeration is what enumerate.mjs reports.
type enumeration struct {
	TailwindVersion string   `json:"tailwindVersion"`
	EntryPoint      string   `json:"entryPoint"`
	FunctionalRoots int      `json:"functionalRoots"`
	StaticUtilities int      `json:"staticUtilities"`
	PairsProbed     int      `json:"pairsProbed"`
	Families        []family `json:"families"`
	// PhysicalOnlyFamilies collapse only with logicalToPhysical off. The table cannot say that, so a
	// run reporting one refuses to write rather than drop it.
	PhysicalOnlyFamilies []family `json:"physicalOnlyFamilies"`
	// The declared-property fields are gone along with the tables they printed.
	//
	// RootProperties, RootColorProperties, ClassProperties and StaticProperties fed
	// `no-conflicting-classes`'s four property tables until #mz0m6k8, which computes the answer from
	// the ported handle bodies instead. Decoding a field nothing prints is how a JSON producer and
	// its consumer drift without either one failing, so they are removed here as well as in
	// enumerate.mjs.
	UnreachableRoots []string `json:"unreachableRoots"`
	PropertyOrder    []string `json:"propertyOrder"`
	VariantOrder     []string `json:"variantOrder"`

	// VerifiedAgainst is the design systems CollapseFamilies was re-measured against on this run,
	// filled in by the Go side rather than read from the enumeration. Rendered into the generated
	// file so the header states what was checked instead of what someone believed.
	VerifiedAgainst []string `json:"-"`
}

// utilityProperties is one utility and the CSS property names it declares.
type utilityProperties struct {
	Root       string   `json:"root"`
	Name       string   `json:"name"`
	Properties []string `json:"properties"`
}

// key returns whichever identifier the producer supplied.
func (u utilityProperties) key() string {
	if u.Name != "" {
		return u.Name
	}
	return u.Root
}

// family is one collapse: two roots that merge into a third when everything else about them agrees.
type family struct {
	Inputs            []string `json:"inputs"`
	Output            string   `json:"output"`
	DiscoveredAtValue string   `json:"discoveredAtValue"`
	// RequiresLogicalToPhysical is true when the pair does not collapse to Output with the
	// canonicalizer's logicalToPhysical off, which is `enforce-canonical-classes`'s `logical: false`.
	RequiresLogicalToPhysical bool `json:"requiresLogicalToPhysical"`
}

// signature is one family as the invariance check and the difference report compare it.
func (f family) signature() string {
	logical := ""
	if f.RequiresLogicalToPhysical {
		logical = " (logical)"
	}
	return fmt.Sprintf("%s => %s @ %s%s", strings.Join(f.Inputs, " + "), f.Output, f.DiscoveredAtValue, logical)
}

// pathList collects a repeatable flag.
type pathList []string

func (p *pathList) String() string { return strings.Join(*p, ", ") }

func (p *pathList) Set(value string) error {
	*p = append(*p, value)
	return nil
}

// designSystemInput is one design system to enumerate: a CSS entry point, and where its bare
// specifiers resolve from.
//
// The two are separate because they stop coinciding for the file that matters. Written
// `<theme.css>` when the CSS sits inside the tree that installed Tailwind, and
// `<theme.css>:<resolve-root>` when it does not.
type designSystemInput struct {
	entryPoint  string
	resolveRoot string
}

// parseDesignSystemInput splits `<theme.css>[:<resolve-root>]`.
//
// Split on the last colon rather than the first, so an absolute Windows-style path or a directory
// with a colon in its name still names a file rather than a truncated one. A value with no colon is
// an entry point whose resolve root defaults to its own directory, which is every repository.
func parseDesignSystemInput(value string) designSystemInput {
	if index := strings.LastIndex(value, ":"); index > 0 {
		candidate := value[:index]
		if _, err := os.Stat(candidate); err == nil {
			return designSystemInput{entryPoint: candidate, resolveRoot: value[index+1:]}
		}
	}
	return designSystemInput{entryPoint: value}
}

// familySignature is the comparable form of the invariant table.
//
// Only the families, and deliberately not the rest of the enumeration. `collapse_table.go` also
// carries KnownStatics, KnownRoots and ColorNames, which are repository facts by construction:
// KnownStatics holds `markdown-content`, one of ahra's own `@utility` blocks, in a file whose
// header says "Source: Tailwind 4.3.3". Comparing whole files across two design systems would
// therefore always fail and would say nothing about the families. What is claimed invariant is what
// gets checked.
func familySignature(families []family) string {
	lines := make([]string, 0, len(families))
	for _, entry := range families {
		lines = append(lines, entry.signature())
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

// invarianceProvenance renders what the invariance claim was checked against, or says it was not.
//
// Written into the generated header rather than into a commit message, because the header is what a
// reader has in front of them when they decide whether to trust the table. The previous version of
// this file asserted invariance in prose that no run had produced, and a reader had no way to tell
// an asserted claim from a measured one. A run that skipped the check now says so in the file it
// wrote.
func invarianceProvenance(verifiedAgainst []string) string {
	if len(verifiedAgainst) == 0 {
		return `//
// This table was generated WITHOUT the invariance check. Pass -verify-invariance-against with a
// design system that shares no submodule with the entry point, and the header will record it.
`
	}

	var buffer strings.Builder
	buffer.WriteString("//\n// CollapseFamilies re-measured on this run against:\n")
	for _, system := range verifiedAgainst {
		fmt.Fprintf(&buffer, "//\t%s\n", system)
	}
	buffer.WriteString("//\n// Identical families, down to the value each was discovered at and whether it needs logicalToPhysical.\n")
	return buffer.String()
}

// portablePath rewrites a home-relative path so the generated file does not name one machine.
//
// The paths reach here as the operator typed them, and an absolute one checked into a generated
// header is a diff every other machine produces on regeneration, which makes `-check` fail for a
// reason that has nothing to do with Tailwind.
func portablePath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if relative, err := filepath.Rel(home, path); err == nil && !strings.HasPrefix(relative, "..") {
		return filepath.Join("~", relative)
	}
	return path
}

// differingFamilies names what the two systems disagree about, in both directions.
//
// Both directions, because a family present in one and absent in the other is the interesting case
// and a count alone cannot distinguish "two systems, one family each" from "one system, both".
func differingFamilies(left, right []family) []string {
	lines := func(families []family) map[string]bool {
		present := make(map[string]bool, len(families))
		for _, entry := range families {
			present[entry.signature()] = true
		}
		return present
	}
	inLeft, inRight := lines(left), lines(right)

	differences := make([]string, 0)
	for line := range inLeft {
		if !inRight[line] {
			differences = append(differences, "only in the first system:  "+line)
		}
	}
	for line := range inRight {
		if !inLeft[line] {
			differences = append(differences, "only in the second system: "+line)
		}
	}
	sort.Strings(differences)
	return differences
}

func main() {
	entryPoint := flag.String("entry-point", "", "path to the Tailwind CSS entry point, such as app/_theme/styles/theme.css")
	resolveRoot := flag.String("resolve-root", "", "directory bare specifiers such as `tailwindcss` resolve from; defaults to the entry point's directory")
	output := flag.String("output", "internal/lint/rules/tailwind/collapse/collapse_table.go", "where to write the generated table")
	check := flag.Bool("check", false, "regenerate and fail if the committed table disagrees, instead of writing it")
	var invarianceInputs pathList
	flag.Var(&invarianceInputs, "verify-invariance-against", "a second design system, as `<theme.css>[:<resolve-root>]`, whose CollapseFamilies must match; repeatable")
	flag.Parse()

	if *entryPoint == "" {
		fmt.Fprintln(os.Stderr, "generate_collapse: -entry-point is required")
		fmt.Fprintln(os.Stderr, "  example: go run ./internal/lint/rules/tailwind/tools/generate_collapse -entry-point ~/Projects/ahra/app/_theme/styles/theme.css")
		os.Exit(2)
	}

	result, err := enumerate(*entryPoint, *resolveRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate_collapse: %v\n", err)
		os.Exit(1)
	}

	// A run that discovered nothing is the failure this generator is most likely to have, and it
	// produces a table that looks clean. `canonicalizeCandidates` returns its input unchanged when
	// called without `collapse: true`, so an options mistake enumerates zero families and writes an
	// empty table that no test would obviously reject.
	if len(result.Families) == 0 {
		fmt.Fprintln(os.Stderr, "generate_collapse: the engine reported no collapse families at all.")
		fmt.Fprintln(os.Stderr, "  That is almost certainly a broken enumeration rather than a Tailwind with no shorthands.")
		fmt.Fprintln(os.Stderr, "  Check that canonicalizeCandidates is being called with {collapse: true}.")
		os.Exit(1)
	}
	if len(result.PhysicalOnlyFamilies) > 0 {
		fmt.Fprintln(os.Stderr, "generate_collapse: these pairs collapse only with logicalToPhysical off, which the table cannot record:")
		for _, entry := range result.PhysicalOnlyFamilies {
			fmt.Fprintf(os.Stderr, "  %s\n", entry.signature())
		}
		fmt.Fprintln(os.Stderr, "  CollapseFamily needs a way to say so before this Tailwind can be enumerated.")
		os.Exit(1)
	}
	if result.FunctionalRoots < 100 {
		fmt.Fprintf(os.Stderr, "generate_collapse: only %d functional roots discovered, which is too few to be the real registry.\n", result.FunctionalRoots)
		fmt.Fprintln(os.Stderr, "  A table built from a partial registry is silently missing families.")
		os.Exit(1)
	}

	// CollapseFamilies is shipped as a framework fact, so the claim is re-measured here rather than
	// at review time.
	//
	// The claim was once verified by generating against ~/Projects/ahra and
	// ~/Projects/connected/www-connected-app and diffing to zero. Both vendor the Structure
	// submodule and both import its global.css, so they share its `@theme` and its `@utility`
	// blocks: that diff measured a shared dependency, and two repositories that cannot differ
	// cannot detect a table that varies. Re-measured against
	// internal/lint/rules/tailwind/tools/generate_descriptor_base/testdata/independent_theme.css, which shares nothing, the
	// families do hold — the same 44, identical down to the value each was discovered at, across a
	// registry that differs by 25 functional roots and 7,850 probed pairs. The claim is true. The
	// instrument that had been asserting it was not measuring it.
	//
	// It is opt-in rather than required, which is where this differs from generate_descriptor_base.
	// That generator's entire job is to separate framework from repository, so one system leaves it
	// nothing to do and it refuses. This one's job is to print families, and the families are only
	// one of several tables it writes: the rest are per-repository by construction and a second
	// system cannot speak to them. Refusing one system here would make the ordinary regeneration
	// impossible in service of a check that covers a fraction of the output. So the flag is a gate
	// an upgrade runs, and the file records whether it was run.
	verifiedAgainst := make([]string, 0, len(invarianceInputs))
	for _, raw := range invarianceInputs {
		other := parseDesignSystemInput(raw)
		otherResult, err := enumerate(other.entryPoint, other.resolveRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "generate_collapse: enumerating %s for the invariance check: %v\n", other.entryPoint, err)
			os.Exit(1)
		}
		if otherResult.TailwindVersion != result.TailwindVersion {
			fmt.Fprintf(os.Stderr, "generate_collapse: %s is Tailwind %s and %s is Tailwind %s; refusing to compare engines.\n",
				*entryPoint, result.TailwindVersion, other.entryPoint, otherResult.TailwindVersion)
			os.Exit(1)
		}

		// A second system that produced the same registry is the same observation wearing two
		// names, which is the exact failure this check exists to have caught. Refused rather than
		// counted, because a zero from it looks identical to a real one.
		if otherResult.FunctionalRoots == result.FunctionalRoots && otherResult.PairsProbed == result.PairsProbed {
			fmt.Fprintf(os.Stderr, "generate_collapse: %s and %s produce the same registry (%d roots, %d pairs).\n",
				*entryPoint, other.entryPoint, result.FunctionalRoots, result.PairsProbed)
			fmt.Fprintln(os.Stderr, "  Two design systems that cannot differ cannot cohere invariance; they are one observation under two names.")
			fmt.Fprintln(os.Stderr, "  Every repository here vendors the Structure submodule and imports its global.css, so any two of them share a @theme.")
			fmt.Fprintln(os.Stderr, "  Use internal/lint/rules/tailwind/tools/generate_descriptor_base/testdata/independent_theme.css, which shares nothing.")
			os.Exit(1)
		}

		if familySignature(otherResult.Families) != familySignature(result.Families) {
			fmt.Fprintf(os.Stderr, "generate_collapse: CollapseFamilies is NOT repository-invariant.\n")
			fmt.Fprintf(os.Stderr, "  %s reports %d families; %s reports %d.\n",
				*entryPoint, len(result.Families), other.entryPoint, len(otherResult.Families))
			for _, line := range differingFamilies(result.Families, otherResult.Families) {
				fmt.Fprintf(os.Stderr, "  %s\n", line)
			}
			fmt.Fprintln(os.Stderr, "  The table cannot ship as a framework fact. This finding is worth more than the confirmation.")
			os.Exit(1)
		}
		verifiedAgainst = append(verifiedAgainst, fmt.Sprintf("%s (%d roots, %d pairs)", portablePath(other.entryPoint), otherResult.FunctionalRoots, otherResult.PairsProbed))
	}
	result.VerifiedAgainst = verifiedAgainst

	rendered, err := render(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate_collapse: %v\n", err)
		os.Exit(1)
	}

	orderRendered, err := renderOrder(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate_collapse: %v\n", err)
		os.Exit(1)
	}
	orderPath := filepath.Join(filepath.Dir(*output), "property_order_table.go")

	// Both generated files are checked, because this generator writes two and used to check one.
	//
	// The check branch returned before `renderOrder` was ever called, so `property_order_table.go`
	// was written by a normal run and compared by nothing. A stale property table then reported
	// "current" in the same breath as the collapse table it does check, which is worse than an
	// unchecked file: the run says a name that sounds like both.
	//
	// Found by a class the ordering table could not place at all. `--enter-opacity` is absent from
	// the 359 properties while `-check` reported current, and those two facts cannot both be true
	// of one measured object.
	if *check {
		type generatedFile struct {
			path     string
			rendered []byte
			stale    string
		}
		for _, file := range []generatedFile{
			{*output, rendered, "The installed Tailwind collapses a different set of families than the committed table records."},
			{orderPath, orderRendered, "The installed Tailwind sorts by a different property order than the committed table records."},
		} {
			committed, err := os.ReadFile(file.path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "generate_collapse: cannot read %s to check it: %v\n", file.path, err)
				os.Exit(1)
			}
			if !bytes.Equal(committed, file.rendered) {
				fmt.Fprintf(os.Stderr, "generate_collapse: %s is stale for Tailwind %s.\n", file.path, result.TailwindVersion)
				fmt.Fprintf(os.Stderr, "  %s\n", file.stale)
				fmt.Fprintln(os.Stderr, "  Regenerate it and read the diff: what changed is what this upgrade changed.")
				os.Exit(1)
			}
		}
		fmt.Fprintf(os.Stderr, "generate_collapse: %s and %s are current for Tailwind %s (%d families, %d ordered properties).\n",
			*output, orderPath, result.TailwindVersion, len(result.Families), len(result.PropertyOrder))
		return
	}
	if err := os.WriteFile(orderPath, orderRendered, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "generate_collapse: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "generate_collapse: wrote %s — %d ordered properties.\n", orderPath, len(result.PropertyOrder))

	if err := os.WriteFile(*output, rendered, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "generate_collapse: %v\n", err)
		os.Exit(1)
	}

	// A root no probe could reach is unmeasured rather than empty, and every table is missing it for
	// a reason nobody could infer from its absence. Four bugs in this generator were exactly that,
	// so the count is printed rather than left for someone to notice.
	if len(result.UnreachableRoots) > 0 {
		fmt.Fprintf(os.Stderr, "generate_collapse: %d roots were unreachable by every probe value and are absent from the tables:\n", len(result.UnreachableRoots))
		for _, root := range result.UnreachableRoots {
			fmt.Fprintf(os.Stderr, "  %s\n", root)
		}
		fmt.Fprintln(os.Stderr, "  Each needs a probe value it accepts, or it is silently missing rather than known to have nothing.")
	}

	fmt.Fprintf(os.Stderr, "generate_collapse: wrote %s — Tailwind %s, %d families from %d pairs over %d roots.\n",
		*output, result.TailwindVersion, len(result.Families), result.PairsProbed, result.FunctionalRoots)
}

// enumerate runs the Node script that asks the real engine.
//
// Node's stderr is passed through rather than captured, so a resolution failure inside the script
// reaches the operator instead of being swallowed into a parse error about empty input.
func enumerate(entryPoint string, resolveRoot string) (*enumeration, error) {
	absoluteEntryPoint, err := filepath.Abs(entryPoint)
	if err != nil {
		return nil, fmt.Errorf("resolving %s: %w", entryPoint, err)
	}
	if _, err := os.Stat(absoluteEntryPoint); err != nil {
		return nil, fmt.Errorf("entry point %s is not readable: %w", absoluteEntryPoint, err)
	}

	scriptPath, err := scriptBesideThisTool()
	if err != nil {
		return nil, err
	}

	// The resolve root defaults to the entry point's directory, which is what every repository
	// invocation gets and what the script did unconditionally before. It is named explicitly only
	// for a design system that lives outside the tree holding the `node_modules` it needs, which is
	// the checked-in independent theme and nothing else so far.
	arguments := []string{scriptPath, absoluteEntryPoint}
	if resolveRoot != "" {
		absoluteResolveRoot, err := filepath.Abs(resolveRoot)
		if err != nil {
			return nil, fmt.Errorf("resolving %s: %w", resolveRoot, err)
		}
		if _, err := os.Stat(filepath.Join(absoluteResolveRoot, "node_modules")); err != nil {
			return nil, fmt.Errorf("resolve root %s has no node_modules, so `tailwindcss` cannot resolve from it: %w", absoluteResolveRoot, err)
		}
		arguments = append(arguments, absoluteResolveRoot)
	}

	command := exec.Command("node", arguments...)
	command.Stderr = os.Stderr

	stdout, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("running the enumeration: %w", err)
	}

	var result enumeration
	if err := json.Unmarshal(stdout, &result); err != nil {
		return nil, fmt.Errorf("parsing the enumeration: %w", err)
	}
	return &result, nil
}

// scriptBesideThisTool finds enumerate.mjs relative to the repository rather than the caller.
//
// The tool is run with `go run ./internal/lint/rules/tailwind/tools/generate_collapse` from the repository root, so the
// script sits at a known path from there. Locating it by walking up from the working directory
// would silently pick up a different checkout when run from a subdirectory.
func scriptBesideThisTool() (string, error) {
	candidate := filepath.Join("tools", "generate_collapse", "enumerate.mjs")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}
	return "", fmt.Errorf("cannot find %s; run this from the repository root", candidate)
}

// render writes the Go source for the table.
//
// Gofmt is applied here rather than left to the caller so that `-check` compares bytes that a
// formatter cannot later change, which is what makes the check a real gate rather than a source of
// spurious failures.
func render(result *enumeration) ([]byte, error) {
	var buffer bytes.Buffer

	fmt.Fprintf(&buffer, `// Code generated by internal/lint/rules/tailwind/tools/generate_collapse. DO NOT EDIT.
//
// Regenerate with:
//
//	go run ./internal/lint/rules/tailwind/tools/generate_collapse -entry-point <theme.css>
//
// Source: Tailwind %s, enumerated from %s.
// %d functional roots produced %d candidate pairs, of which %d collapse.
//
// Each entry is a fact about roots rather than about classes: %s
// holds for every value the two roots share, so the table is small and does not grow with the
// codebase. Enumerated from the design system's own class list rather than from any codebase's
// usage, because a table built from observed classes is silently missing every family nobody has
// written yet.
//
// Only CollapseFamilies is claimed invariant. The other tables in this file are per-repository by
// construction, and KnownStatics is the proof: it carries markdown-content, one of ahra's own
// @utility blocks, under a header naming only a Tailwind version. Read this file as one framework
// fact sitting beside several repository extractions, not as framework data throughout.
%s
package tailwind

// CollapseFamily is two utility roots that merge into a third.
//
// The merge only happens when everything except the root already agrees between the two classes:
// their variants, their importance, and their value. That precondition is the gate's job; this
// table answers only "if those match, do these two roots have a shorthand".
type CollapseFamily struct {
	First  string
	Second string
	Output string
	// RequiresLogicalToPhysical is true when the canonicalizer collapses this pair only while it
	// reads logical properties as their physical longhands, its logicalToPhysical option. With
	// `+"`logical: false`"+`, enforce-canonical-classes skips the family.
	RequiresLogicalToPhysical bool
}

// CollapseFamilies is every family the installed Tailwind knows, sorted by input roots.
//
// Upstream ships shorthand data and it does not derive this. Two maps keyed on CSS property,
// `+"`"+`padding-inline: [padding-left, padding-right]`+"`"+` and its kin, cover inset, margin, padding,
// scroll-margin, scroll-padding, border width/style/colour, gap, overflow, overscroll-behavior and
// ten logical block/inline pairs. This is keyed on utility roots, `+"`"+`pl + pr => px`+"`"+`, and bridging the
// two key spaces is `+"`"+`canonicalizeCandidates`+"`"+`, the signature-equivalence search `+"`"+`enumerate.mjs`+"`"+`
// prices at eight to ten thousand lines that changes every Tailwind minor.
//
// It also holds families no CSS shorthand covers, which is the half that makes the two genuinely
// different questions rather than one restated: `+"`"+`h + w => size`+"`"+`, `+"`"+`skew`+"`"+`, `+"`"+`translate`+"`"+`, the `+"`"+`rounded-*`+"`"+`
// corners, the `+"`"+`mask-*`+"`"+` edges and `+"`"+`border-spacing`+"`"+` collapse through a utility relationship rather
// than a property one.
var CollapseFamilies = []CollapseFamily{
`,
		result.TailwindVersion,
		result.EntryPoint,
		result.FunctionalRoots,
		result.PairsProbed,
		len(result.Families),
		exampleFamily(result.Families),
		invarianceProvenance(result.VerifiedAgainst),
	)

	for _, entry := range result.Families {
		if len(entry.Inputs) != 2 {
			return nil, fmt.Errorf("family %v does not have exactly two inputs", entry.Inputs)
		}
		fmt.Fprintf(&buffer, "\t{First: %q, Second: %q, Output: %q, RequiresLogicalToPhysical: %t},\n",
			entry.Inputs[0], entry.Inputs[1], entry.Output, entry.RequiresLogicalToPhysical)
	}

	fmt.Fprintf(&buffer, `}

// RootSelectorShapes used to be printed here and is not any more.
//
// internal/tailwind's SelectorShapeForRoot reads the nested-rule wrapper off the ported handle
// bodies, which is where upstream keeps it. The eight shapes this printed are reproduced exactly,
// and a repository block that declares a nested selector is answered rather than missed.

// ColorNames used to be printed here and is not any more.
//
// internal/rules/tailwind's valueIsColorIn asks the live theme's `+"`--color`"+` namespace, which is what
// upstream does, with IsColorKeyword beside it for the three keywords that sit in no namespace.
// Measured before the deletion: 558 of its 561 names answered by the theme and the other 3 by
// colorKeywords, 0 uncovered, and the table knew nothing of the three colours www-connected-app
// declares because it was generated from ahra.

// TailwindVersion is the version these families were enumerated from.
//
// Recorded so a mismatch between this and the installed Tailwind is greppable, and so a reader of a
// finding can tell which engine's opinion produced it.
const TailwindVersion = %q
`, result.TailwindVersion)

	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("formatting the generated table: %w", err)
	}
	return formatted, nil
}

// renderOrder writes the class-ordering data to its own file.
//
// Separate from the main table only because it is generated from a different question. It is small:
// Tailwind's sort is a comparator over a property-order list, not a position per class, so this is
// a few hundred rows rather than the 37,643 a per-class table would need.
func renderOrder(result *enumeration) ([]byte, error) {
	var buffer bytes.Buffer

	fmt.Fprintf(&buffer, `// Code generated by internal/lint/rules/tailwind/tools/generate_collapse. DO NOT EDIT.
//
// Regenerate with:
//
//	go run ./internal/lint/rules/tailwind/tools/generate_collapse -entry-point <theme.css>
//
// Source: Tailwind %s. %d ordered properties.
//
// Tailwind sorts utilities into a stable total order, and enforce-consistent-class-order reports
// class lists written in a different one. Its own sort is about twenty-five lines: compare by
// variant, then by the first differing index into an ordered list of CSS property names, then by
// declaration count descending, then alphabetically.
//
// Porting that comparator needs this list rather than a position per class. A per-class table was
// built first and measured at 37,643 entries and 2.0 MB; this is the same answer in a few hundred
// rows, verified exact against the engine over 2,465 real class literals and 296 shuffled ones.

package tailwind

// PropertyOrder is the position of each CSS property in Tailwind's sort order.
//
// Custom properties belong here too. --tw-ring-color has a position, and excluding dash-prefixed
// names left ring-(--color) with an empty sort key, falling through to the alphabetical tiebreak.
var PropertyOrder = map[string]int{
`, result.TailwindVersion, len(result.PropertyOrder))

	for position, property := range result.PropertyOrder {
		fmt.Fprintf(&buffer, "\t%q: %d,\n", property, position)
	}

	fmt.Fprintf(&buffer, `}

// SortOverrideProperties used to be printed here and is not any more.
//
// It held the twelve values a `+"`--tw-sort`"+` declaration can carry, and nothing read it. The latch it
// describes is real and is implemented in walk.go: `+"`PropertySort`"+` reads a `+"`--tw-sort`"+` declaration's
// value and looks it up in `+"`PropertyOrder`"+`, so a value that names a known property latches the order
// there and a value that does not falls through. Neither branch consults a list of which values are
// legal, because the lookup answers that by hitting or missing.

`)

	// The three ordering tables used to be printed between here and VariantOrder.
	//
	// OrderingPropertiesByRoot, OrderingPropertiesByStatic and OrderingPropertiesByClass held 312,
	// 893 and 65 entries and nothing read any of them. The class-order comparator moved onto the live
	// design system in cccaed1 and left them behind, and printing a table nobody reads is how 1,270
	// dead entries survive a passing build.
	//
	// If a consumer needs a root's emitted declarations again, ask an emitter in
	// internal/lint/rules/tailwind/collapse/frameworkhandlers.go rather than reviving these: an emitter branches on the
	// resolved value, which a table keyed on a root cannot.
	fmt.Fprintf(&buffer, `// VariantOrder is the position of each variant prefix in Tailwind's sort order.
//
// Asked of the engine rather than guessed. Ranking variants by prefix length put focus: before
// hover: because it is shorter, which is not the order Tailwind emits them in.
var VariantOrder = map[string]int{
`)
	for position, variant := range result.VariantOrder {
		fmt.Fprintf(&buffer, "\t%q: %d,\n", variant, position)
	}

	fmt.Fprintln(&buffer, "}")

	return format.Source(buffer.Bytes())
}

// quotedList renders a string slice as Go literal elements.
func quotedList(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, fmt.Sprintf("%q", value))
	}
	return strings.Join(quoted, ", ")
}

// exampleFamily names a real entry in the doc comment, so the file explains itself with something
// from this Tailwind rather than with a hardcoded example that may no longer be true.
func exampleFamily(families []family) string {
	for _, entry := range families {
		if len(entry.Inputs) == 2 && entry.Inputs[0] == "px" && entry.Inputs[1] == "py" {
			return "`px + py => p`"
		}
	}
	if len(families) > 0 && len(families[0].Inputs) == 2 {
		return "`" + strings.Join(families[0].Inputs, " + ") + " => " + families[0].Output + "`"
	}
	return "a family"
}
