// Command generate_framework_variants prints the framework's variant registrations into shipped Go.
//
// Usage:
//
//	go run ./internal/lint/rules/tailwind/tools/generate_framework_variants -package-root <tailwindcss package> [-output <file.go>] [-check]
//
// The table it writes is what lets `LoadDesignSystem` build a registry that knows `hover`. Without
// it a run-scoped design system registers only the repository's own `@custom-variant` names —
// measured on this repository, one registration against the engine's 88 — and the consequence is
// not degraded ranking but refusal to parse: `ParseVariant` returns nil for an unregistered root
// and `ParseCandidate` then yields zero candidates for the whole class. Every rule reading it would
// decline on every class carrying a variant, and a differential would score that as agreement
// because both sides went silent.
//
// # Why this is a table when the variant *order* is not
//
// `variant.go` argues that no static table can hold a variant's position, and nothing here
// contradicts it. `getVariantOrder` assigns a dense index per run over only the variants that were
// parsed, so `hover` is 0 in one run and 1 in another, and that computation stays in
// `BuildVariantOrder`. What this table holds is the *registration*: the name, the order number
// `Variants.set` assigned, and the kind. Those are facts about the installed framework.
//
// The shipped `VariantOrder` in property_order_table.go is the shape being argued against and is a
// different thing again: it is keyed on composed prefixes like `group-hover:`, which have no
// position of their own. This table is keyed on registration roots, which is what the engine
// actually registers.
//
// # The invariance measurement this table rests on, and the one it used to claim
//
// A table claiming to describe Tailwind while actually holding one repository's tokens is the bug
// the whole port exists to remove, so the claim has to be measured rather than assumed. It was not.
// The header this generator wrote said the table had been verified against "each repository's own
// `theme.css` loaded through its own install", and no such load was possible: `enumerate.mjs` took
// a package root, built every design system from the literal string `@import "tailwindcss";`, and
// never opened a repository stylesheet. The four "independent generations" were one framework
// `index.css` read four times, and readings that cannot differ cannot detect a table that varies.
// The two repositories named also vendor the same Structure submodule, so even a real comparison
// between them would have measured a shared dependency rather than framework invariance.
//
// Re-measured with `-design-system`, which is the argument that had to exist first, the claim
// holds and holds more broadly than was claimed: 88 registrations with identical names, orders and
// kinds across ~/Projects/ahra, ~/Projects/connected/www-connected-app, ~/Projects/phi/www-phi-health,
// an independent design system sharing no submodule with any of them, and a bare install.
//
// The mechanism, probed rather than assumed. Every corpus repository declares
// `@custom-variant dark` and it does not appear as a difference, because `Variants.set` assigns
// kind and applyFn onto an existing record and never touches `order`: redefining a framework
// variant changes the selector it emits without moving where it sorts. Only a `@custom-variant`
// under a *new* name appends a position. A planted system declaring `dark` plus two new names put
// those two at orders 83 and 84, moved no framework registration, and is what proves the zero above
// came from an instrument that can move.
//
// `-verify-invariance-against` makes that measurement repeatable, and a run that skips it says so
// in the file it writes.
//
// # What crosses this boundary and what cannot
//
// The comparison functions do not. Each is a closure over the design system's theme, and four
// order numbers carry one in a default build. The generated file records which orders carry one and
// in which direction; `LoadDesignSystem` registers `CompareBreakpoints` against them, reading the
// *live* theme. That distinction is load-bearing and measured: redefining `--breakpoint-sm` to
// `200rem` moves `sm` from dense index 0 to index 4 on the engine, so a comparison closed over a
// hardcoded scale would be wrong on exactly the repository that customized its breakpoints.
//
// The directions are probed rather than read off `variants.ts`, so a release that flipped one
// produces a diff here instead of a silent reversal of every breakpoint-bearing class.
//
// `-check` regenerates and fails when the committed table disagrees, which is what makes a Tailwind
// upgrade that registers a new variant show up as a failing gate rather than as a rule that quietly
// stops parsing classes using it. Like the sibling internal/lint/rules/tailwind/tools/generate_* tools, nothing invokes it
// automatically; it is a manual gate.
//
// Node is required to run it and never to use the result.
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
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/system-inc/cohere/internal/lint/rules/tailwind/tools/tooldirectory"
)

// enumeration is what enumerate.mjs reports.
type enumeration struct {
	TailwindVersion string              `json:"tailwindVersion"`
	Entries         []registryEntry     `json:"entries"`
	CompareFnOrders []int               `json:"compareFnOrders"`
	Comparisons     []groupedComparison `json:"comparisons"`

	// VerifiedAgainst is the design systems the registrations were re-measured against on this run,
	// filled in by the Go side. Rendered into the generated header so a reader can tell a measured
	// claim from an asserted one, which is the distinction the previous header lost.
	VerifiedAgainst []string `json:"-"`
}

// pathList collects a repeatable flag.
type pathList []string

func (p *pathList) String() string { return strings.Join(*p, ", ") }

func (p *pathList) Set(value string) error {
	*p = append(*p, value)
	return nil
}

// invarianceProvenance renders what the registrations were re-measured against, or says nothing was.
//
// In the file rather than in a commit message, because the file is what a reader has when they
// decide whether to trust the table, and the previous header's asserted claim was indistinguishable
// from a measured one at exactly that moment.
func invarianceProvenance(verifiedAgainst []string) string {
	if len(verifiedAgainst) == 0 {
		return `//
// This table was generated WITHOUT the invariance check. Pass -verify-invariance-against with a
// design system that shares no submodule with the one generated from.
`
	}

	var buffer strings.Builder
	buffer.WriteString("//\n// Re-measured on this run against:\n")
	for _, system := range verifiedAgainst {
		fmt.Fprintf(&buffer, "//\t%s\n", portablePath(system))
	}
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

// registrationDifference names the first way two registries disagree, or returns empty.
//
// Name, order and kind all count. Order especially: a registration that keeps its name and moves
// its number changes where every class carrying it sorts, and a comparison on names alone would
// call that agreement.
func registrationDifference(left, right []registryEntry) string {
	index := func(entries []registryEntry) map[string]registryEntry {
		byName := make(map[string]registryEntry, len(entries))
		for _, entry := range entries {
			byName[entry.Name] = entry
		}
		return byName
	}
	inLeft, inRight := index(left), index(right)

	names := make([]string, 0, len(inLeft)+len(inRight))
	for name := range inLeft {
		names = append(names, name)
	}
	for name := range inRight {
		if _, seen := inLeft[name]; !seen {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	for _, name := range names {
		leftEntry, inLeftPresent := inLeft[name]
		rightEntry, inRightPresent := inRight[name]
		switch {
		case !inRightPresent:
			return fmt.Sprintf("%q is registered at order %d in the first system and absent from the second", name, leftEntry.Order)
		case !inLeftPresent:
			return fmt.Sprintf("%q is registered at order %d in the second system and absent from the first", name, rightEntry.Order)
		case leftEntry.Order != rightEntry.Order:
			return fmt.Sprintf("%q is order %d in the first system and %d in the second", name, leftEntry.Order, rightEntry.Order)
		case leftEntry.Kind != rightEntry.Kind:
			return fmt.Sprintf("%q is kind %q in the first system and %q in the second", name, leftEntry.Kind, rightEntry.Kind)
		}
	}
	return ""
}

// registryEntry is one registered variant root, as the engine holds it.
type registryEntry struct {
	Name  string `json:"name"`
	Order int    `json:"order"`
	Kind  string `json:"kind"`
}

// groupedComparison is one order number that carries a comparison function, and its direction.
//
// Ascending is a pointer because "the probe could not decide" and "descending" are different
// answers and must not collapse into the same `false`. A nil direction is a hard failure rather
// than a default: a guessed direction reverses every class using that group.
type groupedComparison struct {
	Order     int      `json:"order"`
	Ascending *bool    `json:"ascending"`
	Members   []string `json:"members"`
	Probe     *struct {
		Small      string `json:"small"`
		SmallIndex *int   `json:"smallIndex"`
		Large      string `json:"large"`
		LargeIndex *int   `json:"largeIndex"`
	} `json:"probe"`
}

func main() {
	packageRoot := flag.String("package-root", "", "path to the tailwindcss package root (the directory holding package.json)")
	designSystem := flag.String("design-system", "", "CSS entry point to read the registry from; defaults to a bare `@import \"tailwindcss\"`")
	output := flag.String("output", filepath.Join("internal", "tailwind", "framework_variant_table.go"), "where to write the generated table")
	check := flag.Bool("check", false, "regenerate and fail if the committed table disagrees, instead of writing it")
	var invarianceInputs pathList
	flag.Var(&invarianceInputs, "verify-invariance-against", "a design system entry point whose registrations must match; repeatable")
	flag.Parse()

	if *packageRoot == "" {
		fmt.Fprintln(os.Stderr, "generate_framework_variants: -package-root is required")
		fmt.Fprintln(os.Stderr, "  example: go run ./internal/lint/rules/tailwind/tools/generate_framework_variants -package-root ~/Projects/ahra/node_modules/tailwindcss")
		os.Exit(2)
	}

	result, err := enumerate(*packageRoot, *designSystem)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate_framework_variants: %v\n", err)
		os.Exit(1)
	}

	if err := validate(result); err != nil {
		fmt.Fprintf(os.Stderr, "generate_framework_variants: %v\n", err)
		os.Exit(1)
	}

	// The registrations ship as framework facts, so the claim is re-measured here.
	//
	// This is the check whose absence let a false provenance into the generated header. That header
	// said the table had been verified against "each of the two corpus repositories' own theme.css
	// loaded through its own install", and no such load existed: the script took only a package
	// root and built every design system from a bare `@import "tailwindcss"`, so the four
	// "independent generations" were one framework stylesheet read four times.
	//
	// Re-measured properly the table does hold, and more broadly than was claimed: identical names,
	// orders and kinds across ahra, www-connected-app, www-phi-health, and an independent design
	// system that shares no submodule with any of them. What makes that a real zero rather than an
	// instrument that cannot move is the planted control below, which registers two new variants and
	// is caught.
	verifiedAgainst := make([]string, 0, len(invarianceInputs))
	for _, entryPoint := range invarianceInputs {
		otherResult, err := enumerate(*packageRoot, entryPoint)
		if err != nil {
			fmt.Fprintf(os.Stderr, "generate_framework_variants: enumerating %s for the invariance check: %v\n", entryPoint, err)
			os.Exit(1)
		}
		if err := validate(otherResult); err != nil {
			fmt.Fprintf(os.Stderr, "generate_framework_variants: %s: %v\n", entryPoint, err)
			os.Exit(1)
		}

		if difference := registrationDifference(result.Entries, otherResult.Entries); difference != "" {
			fmt.Fprintln(os.Stderr, "generate_framework_variants: the registrations are NOT repository-invariant.")
			fmt.Fprintf(os.Stderr, "  %s\n", difference)
			fmt.Fprintln(os.Stderr, "  A repository that moves a framework registration cannot be linted from a baked table.")
			fmt.Fprintln(os.Stderr, "  This finding is worth more than the confirmation.")
			os.Exit(1)
		}
		verifiedAgainst = append(verifiedAgainst, entryPoint)
	}
	result.VerifiedAgainst = verifiedAgainst

	rendered, err := render(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate_framework_variants: %v\n", err)
		os.Exit(1)
	}

	if *check {
		committed, err := os.ReadFile(*output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "generate_framework_variants: cannot read %s to check it: %v\n", *output, err)
			os.Exit(1)
		}
		if !bytes.Equal(committed, rendered) {
			fmt.Fprintf(os.Stderr, "generate_framework_variants: %s is stale for Tailwind %s.\n", *output, result.TailwindVersion)
			fmt.Fprintln(os.Stderr, "  The installed Tailwind registers a different set of variants, or registers them in a different order.")
			fmt.Fprintln(os.Stderr, "  Regenerate it and read the diff: what changed is what this upgrade changed.")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "generate_framework_variants: %s is current for Tailwind %s (%d registrations).\n",
			*output, result.TailwindVersion, len(result.Entries))
		return
	}

	if err := os.WriteFile(*output, rendered, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "generate_framework_variants: %v\n", err)
		os.Exit(1)
	}

	distinct := map[int]bool{}
	for _, entry := range result.Entries {
		distinct[entry.Order] = true
	}
	fmt.Fprintf(
		os.Stderr,
		"generate_framework_variants: wrote %s — Tailwind %s, %d registrations over %d distinct orders, %d carrying a comparison.\n",
		*output, result.TailwindVersion, len(result.Entries), len(distinct), len(result.Comparisons),
	)
	for _, comparison := range result.Comparisons {
		direction := "descending"
		if comparison.Ascending != nil && *comparison.Ascending {
			direction = "ascending"
		}
		fmt.Fprintf(os.Stderr, "  order %d %s: %s\n", comparison.Order, direction, strings.Join(comparison.Members, ", "))
	}
}

// validate refuses an enumeration that would render a plausible-looking wrong table.
//
// Each check exists because its failure is silent at the far end. A short registry renders a table
// that compiles and makes `ParseVariant` return nil for whatever it is missing, which is the exact
// symptom this generator was written to remove and would arrive looking like it had been fixed. An
// undecided direction renders a comparison that reverses every class in its group. An unknown kind
// renders a Go identifier that does not exist, which at least fails loudly, but naming it here says
// what happened rather than leaving a compile error about a missing constant.
func validate(result *enumeration) error {
	if len(result.Entries) == 0 {
		return fmt.Errorf(
			"the engine reported no variant registrations at all, which is a broken enumeration rather " +
				"than a Tailwind with no variants",
		)
	}
	// The count is a floor rather than an equality, so a release that adds a variant regenerates
	// cleanly and only a collapse trips it. 88 was measured on 4.3.3 across two installs.
	if len(result.Entries) < 60 {
		return fmt.Errorf(
			"only %d variant registrations discovered, which is too few to be the real registry: a partial "+
				"registry renders a table that silently refuses to parse whatever it is missing",
			len(result.Entries),
		)
	}

	seen := map[string]bool{}
	for _, entry := range result.Entries {
		if entry.Name == "" {
			return fmt.Errorf("a registration has an empty name, which cannot be keyed on")
		}
		if seen[entry.Name] {
			return fmt.Errorf("%q is registered twice, so one of the two orders would be lost", entry.Name)
		}
		seen[entry.Name] = true
		if _, isKnown := variantKindConstants[entry.Kind]; !isKnown {
			return fmt.Errorf("%q has kind %q, which the Go port has no constant for", entry.Name, entry.Kind)
		}
	}

	byOrder := map[int][]string{}
	for _, entry := range result.Entries {
		byOrder[entry.Order] = append(byOrder[entry.Order], entry.Name)
	}
	for _, comparison := range result.Comparisons {
		if comparison.Ascending == nil {
			return fmt.Errorf(
				"order %d carries a comparison function and the direction probe could not decide it (members: %s): "+
					"rendering a guess would reverse every class in that group",
				comparison.Order, strings.Join(comparison.Members, ", "),
			)
		}
		if len(byOrder[comparison.Order]) == 0 {
			return fmt.Errorf("order %d carries a comparison function but no registration holds that order", comparison.Order)
		}
	}

	// Every order carrying a comparison must have been probed, or the rendered table would register
	// comparisons for a subset and leave the rest ordering by root name. `sm` through `2xl` order
	// by resolved breakpoint and alphabetically `2xl` sorts first, so a missing comparison is not a
	// small difference.
	probed := map[int]bool{}
	for _, comparison := range result.Comparisons {
		probed[comparison.Order] = true
	}
	for _, order := range result.CompareFnOrders {
		if !probed[order] {
			return fmt.Errorf("order %d carries a comparison function and was never probed for a direction", order)
		}
	}
	return nil
}

// variantKindConstants maps the engine's kind strings onto the Go port's constants.
//
// Written out rather than derived by casing, because the mapping is a claim about two independent
// vocabularies and a silent mismatch would render a table naming a constant that happens to exist
// and mean something else.
var variantKindConstants = map[string]string{
	"static":     "ParsedVariantKindStatic",
	"functional": "ParsedVariantKindFunctional",
	"compound":   "ParsedVariantKindCompound",
	"arbitrary":  "ParsedVariantKindArbitrary",
}

// render produces the generated Go file.
func render(result *enumeration) ([]byte, error) {
	entries := make([]registryEntry, len(result.Entries))
	copy(entries, result.Entries)
	// Sorted by order, then by name within a shared order, so the file reads as the registration
	// sequence the engine performed and a diff after an upgrade shows the insertion point rather
	// than a reshuffle. Ties broken by name because a Map's iteration order is insertion order and
	// two runs of the same build agree, but relying on that would make the file churn if it ever
	// stopped being true.
	sort.SliceStable(entries, func(left int, right int) bool {
		if entries[left].Order != entries[right].Order {
			return entries[left].Order < entries[right].Order
		}
		return entries[left].Name < entries[right].Name
	})

	comparisons := make([]groupedComparison, len(result.Comparisons))
	copy(comparisons, result.Comparisons)
	sort.SliceStable(comparisons, func(left int, right int) bool {
		return comparisons[left].Order < comparisons[right].Order
	})

	distinctOrders := map[int]bool{}
	sharedOrders := map[int][]string{}
	for _, entry := range entries {
		distinctOrders[entry.Order] = true
		sharedOrders[entry.Order] = append(sharedOrders[entry.Order], entry.Name)
	}
	sharedCount := 0
	for _, names := range sharedOrders {
		if len(names) > 1 {
			sharedCount++
		}
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, `// Code generated by internal/lint/rules/tailwind/tools/generate_framework_variants. DO NOT EDIT.
//
// Regenerate with:
//
//	go run ./internal/lint/rules/tailwind/tools/generate_framework_variants -package-root <tailwindcss package>
//
// Source: Tailwind %s, read from the engine's own variant registry.
// %d registrations over %d distinct order numbers, %d of which are shared by more than one root,
// and %d of which carry a comparison function.
//
// Repository-invariant, and re-measured rather than asserted. An earlier version of this header
// claimed verification against "each of the two corpus repositories' own theme.css loaded through
// its own install", and that load did not exist: enumerate.mjs took only a package root and built
// every design system from a bare `+"`@import \"tailwindcss\"`"+`, so the four independent generations were
// one framework stylesheet read four times, and a set of readings that cannot differ cannot detect
// a table that varies. Those two repositories also vendor the same Structure submodule, so even a
// real comparison between them would have measured a shared dependency.
//
// The claim survives being measured properly. Identical names, orders and kinds across the corpus
// repositories, an independent design system sharing no submodule, and a bare install. The
// mechanism is Variants.set assigning kind onto an existing record without touching order, so a
// repository redefining `+"`dark`"+` changes the selector it emits without moving where it sorts, while a
// @custom-variant under a NEW name appends a position and stays live-loaded rather than being baked
// in here. Both halves were probed: two new names appended at 83 and 84 and moved no framework
// registration, which is also the control proving the comparison can fail.
%s//
// Keyed on registration roots rather than on composed prefixes. The other variant table in this
// package, VariantOrder in property_order_table.go, is keyed on shapes like `+"`group-hover:`"+`, which
// have no registration of their own; the two sitting side by side is the difference being argued.

package tailwind

`, result.TailwindVersion, len(entries), len(distinctOrders), sharedCount, len(comparisons), invarianceProvenance(result.VerifiedAgainst))

	builder.WriteString(`// FrameworkVariantRegistration is one variant the framework registers, with the order number the
// engine assigned it.
//
// Order is carried rather than re-derived, and that is the whole reason this type exists alongside
// FrameworkVariant. Replaying these through Register in sequence would give every root its own
// number, and the engine shares numbers: six roots hold one order in a default build. A shared
// order is where a comparison function gets to speak, and comparison is how the five breakpoints
// order by their resolved widths instead of alphabetically.
type FrameworkVariantRegistration struct {
	// Name is the registered root, as written before the colon: ` + "`hover`" + `, ` + "`group`" + `, ` + "`@max`" + `.
	Name string
	// Order is the registration number Variants.set assigned. Shared across every root registered
	// inside one Variants.group.
	Order int
	// Kind is the registration kind, which decides which branch of Compare applies.
	Kind ParsedVariantKind
}

// FrameworkVariantComparisonGroup is one order number whose members are separated by a comparison
// function rather than by registration.
//
// The function itself cannot be generated: upstream's is a closure over the design system's theme,
// and reading a hardcoded breakpoint scale here would be the per-repository bug this port exists to
// remove. Measured on the engine, redefining ` + "`--breakpoint-sm`" + ` to ` + "`200rem`" + ` moves ` + "`sm`" + ` from dense
// index 0 to index 4. So what is generated is the order number and the direction, and
// LoadDesignSystem attaches CompareBreakpoints over the live theme.
//
// Ascending is probed against the engine rather than read from variants.ts, so a release that
// flipped one produces a diff here instead of silently reversing every class in the group.
type FrameworkVariantComparisonGroup struct {
	// Order is the shared registration number the comparison is keyed on.
	Order int
	// Ascending is the direction: the breakpoints and the container queries ascend, ` + "`max`" + ` and
	// ` + "`@max`" + ` descend.
	Ascending bool
	// Members are the roots holding this order, for diagnostics and for the test that asserts the
	// generated groups still match the generated registrations.
	Members []string
}

`)

	fmt.Fprintf(&builder, `// FrameworkVariantRegistrations is every variant the installed Tailwind registers, in registration
// order.
//
// In order rather than sorted by name, so a reader sees the sequence the engine performed and an
// upgrade's diff shows where a new variant was inserted. Registrations() sorts by name when a
// caller needs that.
var FrameworkVariantRegistrations = []FrameworkVariantRegistration{
`)
	for _, entry := range entries {
		fmt.Fprintf(
			&builder,
			"\t{Name: %s, Order: %d, Kind: %s},\n",
			strconv.Quote(entry.Name), entry.Order, variantKindConstants[entry.Kind],
		)
	}
	builder.WriteString("}\n\n")

	builder.WriteString(`// FrameworkVariantComparisonGroups is every order number that carries a comparison function.
//
// All four are breakpoint comparisons in a default build. Without them the six roots sharing order
// 64 fall through to a root-name comparison, and ` + "`2xl`" + ` sorts before ` + "`sm`" + ` alphabetically, which
// reorders every responsive class in the tree.
var FrameworkVariantComparisonGroups = []FrameworkVariantComparisonGroup{
`)
	for _, comparison := range comparisons {
		members := make([]string, 0, len(comparison.Members))
		for _, member := range comparison.Members {
			members = append(members, strconv.Quote(member))
		}
		fmt.Fprintf(
			&builder,
			"\t{Order: %d, Ascending: %t, Members: []string{%s}},\n",
			comparison.Order, comparison.Ascending != nil && *comparison.Ascending, strings.Join(members, ", "),
		)
	}
	builder.WriteString("}\n")

	formatted, err := format.Source([]byte(builder.String()))
	if err != nil {
		return nil, fmt.Errorf("formatting the generated table: %w", err)
	}
	return formatted, nil
}

// enumerate runs the Node script that asks the real engine.
//
// Node's stderr is passed through rather than captured, so a resolution failure inside the script
// reaches the operator instead of being swallowed into a parse error about empty input.
func enumerate(packageRoot string, designSystem string) (*enumeration, error) {
	absolutePackageRoot, err := filepath.Abs(packageRoot)
	if err != nil {
		return nil, fmt.Errorf("resolving %s: %w", packageRoot, err)
	}
	if _, err := os.Stat(filepath.Join(absolutePackageRoot, "package.json")); err != nil {
		return nil, fmt.Errorf("%s does not look like a tailwindcss package root: %w", absolutePackageRoot, err)
	}

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("cannot locate this tool's directory, so enumerate.mjs cannot be found")
	}
	directory, err := tooldirectory.Of(thisFile)
	if err != nil {
		return nil, err
	}
	script := filepath.Join(directory, "enumerate.mjs")

	// The design system is optional and defaults to a bare framework import, which is what every
	// invocation did before it could be named. It is the argument that makes the invariance claim
	// measurable rather than asserted.
	arguments := []string{script, absolutePackageRoot}
	if designSystem != "" {
		absoluteDesignSystem, err := filepath.Abs(designSystem)
		if err != nil {
			return nil, fmt.Errorf("resolving %s: %w", designSystem, err)
		}
		if _, err := os.Stat(absoluteDesignSystem); err != nil {
			return nil, fmt.Errorf("design system %s is not readable: %w", absoluteDesignSystem, err)
		}
		arguments = append(arguments, absoluteDesignSystem)
	}

	command := exec.Command("node", arguments...)
	var stdout bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("node %s: %w", script, err)
	}

	var result enumeration
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("parsing the enumeration: %w", err)
	}
	return &result, nil
}
