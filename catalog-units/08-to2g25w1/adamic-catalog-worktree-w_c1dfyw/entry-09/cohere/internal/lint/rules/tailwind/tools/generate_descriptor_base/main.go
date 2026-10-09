// Command generate_descriptor_base prints the framework-invariant half of the descriptor table
// into shipped Go, so a design system can be given a descriptor Table without a checked-in fixture.
//
// Usage:
//
//	go run ./internal/lint/rules/tailwind/tools/generate_descriptor_base \
//	    -input <table.json> [-input <table.json> ...] \
//	    -context <context.json> [-context <context.json> ...] \
//	    -output <base_table.go>
//
// The inputs are what `internal/lint/rules/tailwind/tools/generate_descriptors/extract.mjs --json` and
// `internal/lint/rules/tailwind/tools/generate_descriptor_table/context.mjs` write, one pair per design system. Several pairs
// are accepted and at least two are required, because this command's entire job is to separate what
// is a fact about Tailwind from what is a fact about one repository, and a single extraction cannot
// tell those apart. With one input every per-repository token looks invariant.
//
// # Why this command exists, when generate_descriptor_table already prints a whole table
//
// That command prints the table for one repository, namespaces included, and its own doc comment
// says the result must not be checked in: `Namespaces` and `KeysByNamespace` come from that
// repository's `@theme`, so a committed copy reports one repository's tokens as facts about
// Tailwind. That is correct, and it left the shipped code with no way to build a Table at all. The
// only two constructions in the tree were in tests, both assembled from untracked fixtures, so
// `Table.Lookup` answered the `{order, count}` half of the class-order sort and nothing outside a
// test could build the thing that answers it.
//
// So the table is split along the seam the measurement found rather than committed whole.
//
// # The seam, measured across three design systems
//
// Generating against ~/Projects/ahra and ~/Projects/connected/www-connected-app produces
// byte-identical `roots` arrays, which is the observation already on the record. It is not evidence
// of framework invariance: those two repositories share the Structure submodule, so they share its
// `@theme` and its `@utility` blocks, and the agreement is a shared dependency rather than a
// property of Tailwind. Measured against a third, synthetic design system that imports `tailwindcss`
// and declares only `--color-weird-*` and `--frobnicate-*`, 258 of 301 shared roots differ.
//
// Every one of those 258 differs **only** in the three namespace-keyed maps. With those removed,
// all three systems agree on every shared root: TypeList, ByType, ByLiteral, the fallbacks and the
// empties are identical, 0 differences over 327 and 301 shared roots respectively. So the descriptor
// model's shape is invariant and the namespace axis is not, which is what this command's split
// encodes.
//
// That third system is `testdata/independent_theme.css`, checked in beside this file so the
// measurement is reproducible rather than a claim about a directory that no longer exists. Its own
// comment says how to generate the pair from it.
//
// The namespace axis then splits again, and this is the part that makes composing it live exact
// rather than approximate. Over 1,425 namespace buckets across the three systems, every bucket is
// one of exactly two things and there are no others:
//
//   - **399 buckets consumed by the root**, keyed on one of 14 namespaces — `--color`, `--font`,
//     `--text`, `--leading` and their kin — which read identically in all three systems. A fact
//     about Tailwind, printed here.
//   - **1,026 buckets the root does not consume**, every one of which reads *exactly* its axis's
//     `@none` bucket. A namespace a root ignores does not produce a reading of its own; the value
//     falls through the bare-value precedence to the `@none` step. Derivable at run time from the
//     theme's own namespace list, so it is not printed.
//
// Zero violations. That is why a repository declaring `--color-weird` gets `border-weird-1` right
// without this file knowing the token exists: `border` does not consume `--color-weird`, so the
// bucket equals `@none`, and `@none` is already here. It is also why the split is not merely a size
// optimisation — the 1,026 are not omitted because they compress well, they are omitted because
// storing them would be storing one repository's token names.
//
// # What is deliberately not printed
//
// `Namespaces`, `KeysByNamespace` and the repository's own `@utility` statics. All three are read
// off the live design system by `NewTable` in descriptor_live.go. `PropertyOrder` is not printed
// either: `property_order_table.go` already ships it, measured identical across all three systems,
// and a second copy is a second thing to drift.
//
// `PerDeclaration` is not printed. It is derived from the repository's own `@utility` blocks, and
// while both corpus repositories report the same 18 roots, they report them because they share the
// Structure submodule that declares them. Deriving it live is the only answer that is right for a
// repository this command never saw.
//
// Node is required to produce the inputs and never to use the result.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"os"
	"sort"
	"strings"
)

// reading is the extractor's `{order, count}`, and a JSON null is a reading that does not exist.
type reading struct {
	Order []int `json:"order"`
	Count int   `json:"count"`
}

type rootEntry struct {
	Root                                 string              `json:"root"`
	TypeList                             []string            `json:"typeList"`
	ReadingByType                        map[string]*reading `json:"readingByType"`
	Fallback                             *reading            `json:"fallback"`
	ReadingByNamespace                   map[string]*reading `json:"readingByNamespace"`
	ReadingByLiteral                     map[string]*reading `json:"readingByLiteral"`
	Empty                                *reading            `json:"empty"`
	EmptyWithModifier                    *reading            `json:"emptyWithModifier"`
	EmptyWithThemedModifier              *reading            `json:"emptyWithThemedModifier"`
	ReadingByTypeWithModifier            map[string]*reading `json:"readingByTypeWithModifier"`
	FallbackWithModifier                 *reading            `json:"fallbackWithModifier"`
	ReadingByNamespaceWithModifier       map[string]*reading `json:"readingByNamespaceWithModifier"`
	ReadingByTypeWithThemedModifier      map[string]*reading `json:"readingByTypeWithThemedModifier"`
	FallbackWithThemedModifier           *reading            `json:"fallbackWithThemedModifier"`
	ReadingByNamespaceWithThemedModifier map[string]*reading `json:"readingByNamespaceWithThemedModifier"`
}

type extractedTable struct {
	TailwindVersion string              `json:"tailwindVersion"`
	EntryPoint      string              `json:"entryPoint"`
	Roots           []rootEntry         `json:"roots"`
	Statics         map[string]*reading `json:"statics"`
}

type extractedContext struct {
	TailwindVersion string              `json:"tailwindVersion"`
	Namespaces      []string            `json:"namespaces"`
	KeysByNamespace map[string][]string `json:"keysByNamespace"`
	PropertyOrder   []string            `json:"propertyOrder"`
	UnlistedStatics map[string]*reading `json:"unlistedStatics"`
}

// pathList collects a repeatable flag.
type pathList []string

func (list *pathList) String() string { return strings.Join(*list, ",") }

func (list *pathList) Set(value string) error {
	*list = append(*list, value)
	return nil
}

// system is one design system's pair of extractions.
type system struct {
	table   extractedTable
	context extractedContext
}

func main() {
	var inputPaths pathList
	var contextPaths pathList
	flag.Var(&inputPaths, "input", "path to a JSON file written by generate_descriptors/extract.mjs --json; repeatable")
	flag.Var(&contextPaths, "context", "path to a JSON file written by context.mjs, in the same order as -input; repeatable")
	outputPath := flag.String("output", "", "path to write the generated Go file to; - writes to stdout")
	packageName := flag.String("package", "tailwind", "package clause of the generated file")
	flag.Parse()

	if len(inputPaths) == 0 || *outputPath == "" {
		fmt.Fprintln(os.Stderr, "generate_descriptor_base: -input and -output are required")
		os.Exit(2)
	}
	if len(inputPaths) != len(contextPaths) {
		fmt.Fprintf(os.Stderr, "generate_descriptor_base: %d -input and %d -context; they pair up one to one\n",
			len(inputPaths), len(contextPaths))
		os.Exit(2)
	}
	// Two is the minimum that can tell a fact about Tailwind from a fact about one repository. With
	// one extraction every per-repository token looks invariant, and this command would print them
	// as framework facts, which is precisely the bug the port exists to remove.
	if len(inputPaths) < 2 {
		fmt.Fprintln(os.Stderr, "generate_descriptor_base: at least two design systems are required; one cannot separate the framework from a repository")
		os.Exit(2)
	}

	systems := make([]system, 0, len(inputPaths))
	for index := range inputPaths {
		var loaded system
		readJSON(inputPaths[index], &loaded.table)
		readJSON(contextPaths[index], &loaded.context)
		if len(loaded.table.Roots) == 0 || len(loaded.table.Statics) == 0 || loaded.table.TailwindVersion == "" {
			fmt.Fprintf(os.Stderr, "generate_descriptor_base: %s looks empty (%d roots, %d statics, version %q); refusing to generate\n",
				inputPaths[index], len(loaded.table.Roots), len(loaded.table.Statics), loaded.table.TailwindVersion)
			os.Exit(3)
		}
		if loaded.context.TailwindVersion != loaded.table.TailwindVersion {
			fmt.Fprintf(os.Stderr, "generate_descriptor_base: %s is Tailwind %s and %s is Tailwind %s; refusing to generate\n",
				inputPaths[index], loaded.table.TailwindVersion, contextPaths[index], loaded.context.TailwindVersion)
			os.Exit(3)
		}
		if loaded.table.TailwindVersion != systems0Version(systems) && len(systems) > 0 {
			fmt.Fprintf(os.Stderr, "generate_descriptor_base: %s is Tailwind %s and an earlier input is Tailwind %s; refusing to mix engines\n",
				inputPaths[index], loaded.table.TailwindVersion, systems0Version(systems))
			os.Exit(3)
		}
		systems = append(systems, loaded)
	}

	generated, stats := generate(systems, *packageName)

	formatted, err := format.Source(generated)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate_descriptor_base: generated file does not parse: %v\n", err)
		os.Exit(1)
	}

	if *outputPath == "-" {
		os.Stdout.Write(formatted)
	} else if err := os.WriteFile(*outputPath, formatted, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "generate_descriptor_base: writing output: %v\n", err)
		os.Exit(1)
	}

	// Printed on every run. The counts below are the measurement this file's shape rests on, so a
	// future extraction that moves them says so out loud rather than silently printing a table with
	// a different meaning.
	fmt.Printf("generate_descriptor_base: %d design systems, %d roots invariant across all of them (%d dropped as not shared)\n",
		len(systems), stats.roots, stats.rootsDropped)
	fmt.Printf("  namespace buckets: %d consumed identically by every system, %d equal to their axis @none and derived live, %d violations\n",
		stats.consumedBuckets, stats.noneBuckets, stats.violations)
	fmt.Printf("  %d framework namespaces, %d statics invariant (%d dropped as repository-specific)\n",
		stats.frameworkNamespaces, stats.statics, stats.staticsDropped)
	fmt.Printf("  readings %d interned, output %s\n", stats.distinctReadings, humanSize(len(formatted)))
}

func systems0Version(systems []system) string {
	if len(systems) == 0 {
		return ""
	}
	return systems[0].table.TailwindVersion
}

func readJSON(path string, into any) {
	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate_descriptor_base: reading %s: %v\n", path, err)
		os.Exit(1)
	}
	if err := json.Unmarshal(raw, into); err != nil {
		fmt.Fprintf(os.Stderr, "generate_descriptor_base: parsing %s: %v\n", path, err)
		os.Exit(1)
	}
}

type generationStats struct {
	roots               int
	rootsDropped        int
	consumedBuckets     int
	noneBuckets         int
	violations          int
	frameworkNamespaces int
	statics             int
	staticsDropped      int
	distinctReadings    int
}

// readingPool interns readings so that many buckets share one.
//
// Keyed on content rather than on pointer identity, because the extractor writes an independent
// object per bucket and the duplicates are only discoverable by value.
type readingPool struct {
	indexByKey map[string]int
	readings   []*reading
}

func newReadingPool() *readingPool { return &readingPool{indexByKey: make(map[string]int)} }

func (pool *readingPool) intern(value *reading) int {
	if value == nil {
		return -1
	}
	key := readingKey(value)
	if index, found := pool.indexByKey[key]; found {
		return index
	}
	index := len(pool.readings)
	pool.indexByKey[key] = index
	pool.readings = append(pool.readings, value)
	return index
}

func readingKey(value *reading) string {
	if value == nil {
		return "nil"
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "%d:", value.Count)
	for _, order := range value.Order {
		fmt.Fprintf(&builder, "%d,", order)
	}
	return builder.String()
}

// axisNames pairs the three modifier axes' JSON field groups with the Go field they print into.
type axisSelector struct {
	goName      string
	byType      func(*rootEntry) map[string]*reading
	byNamespace func(*rootEntry) map[string]*reading
	fallback    func(*rootEntry) *reading
	empty       func(*rootEntry) *reading
}

var axes = []axisSelector{
	{
		goName:      "Absent",
		byType:      func(entry *rootEntry) map[string]*reading { return entry.ReadingByType },
		byNamespace: func(entry *rootEntry) map[string]*reading { return entry.ReadingByNamespace },
		fallback:    func(entry *rootEntry) *reading { return entry.Fallback },
		empty:       func(entry *rootEntry) *reading { return entry.Empty },
	},
	{
		goName:      "Alpha",
		byType:      func(entry *rootEntry) map[string]*reading { return entry.ReadingByTypeWithModifier },
		byNamespace: func(entry *rootEntry) map[string]*reading { return entry.ReadingByNamespaceWithModifier },
		fallback:    func(entry *rootEntry) *reading { return entry.FallbackWithModifier },
		empty:       func(entry *rootEntry) *reading { return entry.EmptyWithModifier },
	},
	{
		goName:      "Themed",
		byType:      func(entry *rootEntry) map[string]*reading { return entry.ReadingByTypeWithThemedModifier },
		byNamespace: func(entry *rootEntry) map[string]*reading { return entry.ReadingByNamespaceWithThemedModifier },
		fallback:    func(entry *rootEntry) *reading { return entry.FallbackWithThemedModifier },
		empty:       func(entry *rootEntry) *reading { return entry.EmptyWithThemedModifier },
	},
}

// The two pseudo-namespaces, spelled here as the package spells them. They are not theme namespaces
// and are carried in the same maps because they occupy the same step of the bare-value precedence.
const (
	namespaceColorKeyword = "@colorKeyword"
	namespaceNone         = "@none"
)

func generate(systems []system, packageName string) ([]byte, generationStats) {
	stats := generationStats{}
	pool := newReadingPool()

	// Only roots every system registers. A root one system has and another does not is a repository
	// `@utility` block, which the live evaluator answers rather than this table.
	rootIndex := make([]map[string]*rootEntry, len(systems))
	for index := range systems {
		rootIndex[index] = make(map[string]*rootEntry, len(systems[index].table.Roots))
		for entryIndex := range systems[index].table.Roots {
			entry := &systems[index].table.Roots[entryIndex]
			rootIndex[index][entry.Root] = entry
		}
	}
	sharedRoots := make([]string, 0, len(rootIndex[0]))
	for root := range rootIndex[0] {
		shared := true
		for index := 1; index < len(rootIndex); index++ {
			if _, found := rootIndex[index][root]; !found {
				shared = false
				break
			}
		}
		if shared {
			sharedRoots = append(sharedRoots, root)
			continue
		}
		stats.rootsDropped++
	}
	sort.Strings(sharedRoots)

	// The framework namespaces: those a root consumes identically in every system. Collected before
	// anything is printed, because the same set decides which buckets are kept and is itself emitted
	// so the live constructor can tell a consumed namespace from one it must derive.
	frameworkNamespaces := map[string]bool{}
	for _, root := range sharedRoots {
		for _, axis := range axes {
			maps := make([]map[string]*reading, len(systems))
			for index := range systems {
				maps[index] = axis.byNamespace(rootIndex[index][root])
			}
			for key, value := range maps[0] {
				if key == namespaceColorKeyword || key == namespaceNone {
					continue
				}
				if sameEverywhere(maps, key, value) {
					frameworkNamespaces[key] = true
				}
			}
		}
	}

	// The invariance check, run rather than asserted. Every namespace bucket in every system must be
	// either consumed identically everywhere or exactly equal to its axis's `@none`; a bucket that is
	// neither is a namespace whose reading this file cannot print and the live constructor cannot
	// derive, so it is counted and reported rather than silently dropped.
	for _, root := range sharedRoots {
		for _, axis := range axes {
			for index := range systems {
				entry := rootIndex[index][root]
				namespaceMap := axis.byNamespace(entry)
				noneReading := namespaceMap[namespaceNone]
				maps := make([]map[string]*reading, len(systems))
				for other := range systems {
					maps[other] = axis.byNamespace(rootIndex[other][root])
				}
				for key, value := range namespaceMap {
					if key == namespaceColorKeyword || key == namespaceNone {
						continue
					}
					switch {
					case sameEverywhere(maps, key, value):
						stats.consumedBuckets++
					case readingKey(value) == readingKey(noneReading):
						stats.noneBuckets++
					default:
						stats.violations++
						fmt.Fprintf(os.Stderr,
							"generate_descriptor_base: %s %s namespace %q reads %s, which is neither shared by every system nor equal to its axis @none (%s)\n",
							root, axis.goName, key, readingKey(value), readingKey(noneReading))
					}
				}
			}
		}
	}
	// A violation means the two-way split this file encodes does not hold on these inputs, so the
	// generated table would be wrong in a way no test names. Refuse rather than print it.
	if stats.violations > 0 {
		fmt.Fprintf(os.Stderr, "generate_descriptor_base: %d namespace buckets fit neither half of the split; refusing to generate\n", stats.violations)
		os.Exit(4)
	}

	stats.roots = len(sharedRoots)
	stats.frameworkNamespaces = len(frameworkNamespaces)

	var rootsBuffer bytes.Buffer
	for _, root := range sharedRoots {
		writeDescriptor(&rootsBuffer, rootIndex[0][root], frameworkNamespaces, pool)
	}

	// Statics every system agrees on, by name and by reading. A static one system has and another
	// does not is a repository `@utility` block: ahra declares `markdown-content` and `typing-dots`
	// and www-connected-app declares neither, which is exactly the pair this drops.
	staticSets := make([]map[string]*reading, len(systems))
	for index := range systems {
		staticSets[index] = make(map[string]*reading, len(systems[index].table.Statics))
		for name, value := range systems[index].context.UnlistedStatics {
			staticSets[index][name] = value
		}
		// The registry wins a collision with the probe, since it is the measured source.
		for name, value := range systems[index].table.Statics {
			staticSets[index][name] = value
		}
	}
	staticNames := make([]string, 0, len(staticSets[0]))
	for name, value := range staticSets[0] {
		if value == nil {
			continue
		}
		shared := true
		for index := 1; index < len(staticSets); index++ {
			other, found := staticSets[index][name]
			if !found || other == nil || readingKey(other) != readingKey(value) {
				shared = false
				break
			}
		}
		if shared {
			staticNames = append(staticNames, name)
			continue
		}
		stats.staticsDropped++
	}
	sort.Strings(staticNames)
	stats.statics = len(staticNames)

	var staticsBuffer bytes.Buffer
	for _, name := range staticNames {
		fmt.Fprintf(&staticsBuffer, "\t%q: baseReadings[%d],\n", name, pool.intern(staticSets[0][name]))
	}

	stats.distinctReadings = len(pool.readings)

	var body bytes.Buffer
	sources := make([]string, 0, len(systems))
	for index := range systems {
		sources = append(sources, jsonPathComment(systems[index].table.EntryPoint))
	}
	fmt.Fprintf(&body, `// Code generated by internal/lint/rules/tailwind/tools/generate_descriptor_base. DO NOT EDIT.
//
// Generated against Tailwind %s from %d design systems: %s.
//
// This is the half of the descriptor table that is a fact about Tailwind rather than about any one
// repository, and it is checked in because it was measured to be invariant rather than assumed to
// be. The other half — which theme namespaces exist, which keys are in them, and the repository's
// own `+"`@utility`"+` blocks — is read off the live design system by NewTable in descriptor_live.go.
//
// The seam is measured, and the measurement is the reason the split is exactly here:
//
//   - %d roots registered by every system, agreeing on TypeList, ByType, ByLiteral, the fallbacks
//     and the empties. Against a synthetic third system sharing no submodule with the corpus
//     repositories, 258 of 301 shared roots differ, and every one of them differs *only* in its
//     namespace maps.
//   - %d namespace buckets consumed identically by every system, keyed on the %d namespaces below.
//     Printed here.
//   - %d namespace buckets that read exactly their axis's `+"`@none`"+`. A namespace a root does not
//     consume produces no reading of its own; the value falls through the bare-value precedence to
//     the `+"`@none`"+` step, which is already in this file. Derived live, never printed, because
//     printing them would be printing one repository's token names.
//
// Zero buckets fell outside those two cases. The generator re-runs that check on every invocation
// and exits rather than printing a table if any bucket does, so this comment cannot quietly go stale.
//
// Regenerate with:
//
//	go run ./internal/lint/rules/tailwind/tools/generate_descriptor_base \
//	    -input <a.json> -context <a-context.json> \
//	    -input <b.json> -context <b-context.json> \
//	    -output internal/lint/rules/tailwind/collapse/descriptor_base_table.go

package %s

// baseReadings holds every distinct reading in the invariant table exactly once.
//
// Readings are immutable and shared; nothing in the port writes to one. NewTable copies no reading
// out of this array, it points at them, which is what keeps building a table per run cheap.
var baseReadings = []Reading{
`, systems[0].table.TailwindVersion, len(systems), strings.Join(sources, ", "),
		stats.roots, stats.consumedBuckets, stats.frameworkNamespaces, stats.noneBuckets, packageName)

	for _, value := range pool.readings {
		if len(value.Order) == 0 {
			fmt.Fprintf(&body, "\t{Count: %d},\n", value.Count)
			continue
		}
		orderParts := make([]string, len(value.Order))
		for index, order := range value.Order {
			orderParts[index] = fmt.Sprint(order)
		}
		fmt.Fprintf(&body, "\t{Order: []int{%s}, Count: %d},\n", strings.Join(orderParts, ", "), value.Count)
	}
	fmt.Fprintf(&body, "}\n\n")

	fmt.Fprintf(&body, `// BaseTailwindVersion is the engine version this file was generated against.
//
// Carried so a table built against an upgraded tailwindcss is a loud failure rather than a subtly
// wrong sort. NewTable compares it with the live design system's version.
const BaseTailwindVersion = %q

// FrameworkNamespaces is the set of theme namespaces a utility root actually consumes.
//
// %d of them, and every one is a namespace Tailwind itself defines. Its job is to let NewTable tell
// the two kinds of namespace apart: a namespace in this set has its reading printed per root below,
// and a namespace outside it reads its axis's `+"`@none`"+` on every root, whatever the repository
// chose to call it.
//
// It is emitted rather than inferred from the descriptors because "absent from every descriptor" and
// "consumed but reading like @none" are different facts that happen to look alike in the printed
// table, and only the generator, holding several systems at once, can tell them apart.
var FrameworkNamespaces = map[string]bool{
`, systems[0].table.TailwindVersion, stats.frameworkNamespaces)
	frameworkNames := make([]string, 0, len(frameworkNamespaces))
	for name := range frameworkNamespaces {
		frameworkNames = append(frameworkNames, name)
	}
	sort.Strings(frameworkNames)
	for _, name := range frameworkNames {
		fmt.Fprintf(&body, "\t%q: true,\n", name)
	}
	fmt.Fprintf(&body, "}\n\n")

	fmt.Fprintf(&body, `// baseStatics is every static utility whose reading is a fact about Tailwind.
//
// %d of them. A static one system registers and another does not is a repository `+"`@utility`"+`
// block — ahra declares `+"`markdown-content`"+` and `+"`typing-dots`"+`, www-connected-app declares
// neither — and NewTable adds those from the live evaluator rather than this file carrying one
// repository's utilities as framework facts.
var baseStatics = map[string]Reading{
`, stats.statics)
	body.Write(staticsBuffer.Bytes())
	fmt.Fprintf(&body, "}\n\n")

	fmt.Fprintf(&body, `// baseDescriptors is every functional root the framework registers, keyed by root.
//
// %d roots. The namespace maps hold only the framework namespaces plus the two pseudo-namespaces;
// NewTable adds nothing to them and instead consults FrameworkNamespaces to decide whether a theme
// namespace the repository declared reads here or falls through to `+"`@none`"+`.
var baseDescriptors = map[string]*Descriptor{
`, stats.roots)
	body.Write(rootsBuffer.Bytes())
	fmt.Fprintf(&body, "}\n")

	return body.Bytes(), stats
}

// sameEverywhere reports whether a namespace key holds the same reading in every system's map.
func sameEverywhere(maps []map[string]*reading, key string, value *reading) bool {
	for _, candidate := range maps {
		other, found := candidate[key]
		if !found || readingKey(other) != readingKey(value) {
			return false
		}
	}
	return true
}

// jsonPathComment names a design system without naming the machine it was generated on.
//
// The last three path segments are not enough on their own. Every corpus repository's stylesheet
// ends `_theme/styles/theme.css`, so three segments name two different systems identically, and a
// system generated from a temporary directory leaks a session-specific path into a checked-in file —
// which is what the first version of this table did, baking a scratchpad directory into line 3.
//
// So the repository directory is used where the path has the shape a repository has, and anything
// else is reported by shape rather than by location. A reader wants to know which systems were
// compared, and a path that differs per machine is a spurious diff on every regeneration.
func jsonPathComment(entryPoint string) string {
	if entryPoint == "" {
		return "an unnamed design system"
	}
	parts := strings.Split(entryPoint, "/")
	// A repository stylesheet at `<repository>/app/_theme/styles/theme.css` is named by its
	// repository, which is the segment a reader recognises.
	for index := len(parts) - 1; index > 0; index-- {
		if parts[index] == "app" || parts[index] == "src" {
			return parts[index-1]
		}
	}
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return entryPoint
}

func writeDescriptor(out *bytes.Buffer, entry *rootEntry, frameworkNamespaces map[string]bool, pool *readingPool) {
	fmt.Fprintf(out, "\t%q: {\n\t\tRoot: %q,\n", entry.Root, entry.Root)

	if len(entry.TypeList) > 0 {
		typeParts := make([]string, len(entry.TypeList))
		for index, typeName := range entry.TypeList {
			typeParts[index] = fmt.Sprintf("%q", typeName)
		}
		// The order is the engine's, recovered by topological sort from observation. Written out
		// verbatim and never sorted: InferDataType returns the first match.
		fmt.Fprintf(out, "\t\tTypeList: []DataType{%s},\n", strings.Join(typeParts, ", "))
	}

	writeStringReadingMap(out, "ByLiteral", "\t\t", entry.ReadingByLiteral, nil, nil, pool)

	for _, axis := range axes {
		writeAxis(out, axis, entry, frameworkNamespaces, pool)
	}

	fmt.Fprintf(out, "\t},\n")
}

func writeAxis(out *bytes.Buffer, axis axisSelector, entry *rootEntry, frameworkNamespaces map[string]bool, pool *readingPool) {
	fallback := axis.fallback(entry)
	empty := axis.empty(entry)

	var axisBuffer bytes.Buffer
	writeStringReadingMap(&axisBuffer, "ByType", "\t\t\t", axis.byType(entry), fallback, nil, pool)
	// Namespaces keep their fallback-equal entries: a miss here continues the bare-value precedence
	// to inference rather than ending at the fallback, so eliding one changes the answer. What is
	// filtered instead is which namespaces appear at all — the framework's, plus the two
	// pseudo-namespaces, and never a repository token.
	writeStringReadingMap(&axisBuffer, "ByNamespace", "\t\t\t", axis.byNamespace(entry), nil, frameworkNamespaces, pool)
	if fallback != nil {
		fmt.Fprintf(&axisBuffer, "\t\t\tFallback: baseReadings[%d],\n", pool.intern(fallback))
	}
	if empty != nil {
		fmt.Fprintf(&axisBuffer, "\t\t\tEmpty: &baseReadings[%d],\n", pool.intern(empty))
	}
	if axisBuffer.Len() == 0 {
		return
	}
	fmt.Fprintf(out, "\t\t%s: AxisReadings{\n", axis.goName)
	out.Write(axisBuffer.Bytes())
	fmt.Fprintf(out, "\t\t},\n")
}

// writeStringReadingMap emits one map, dropping entries equal to fallback and, when keep is
// non-nil, entries whose key is not in it.
//
// The fallback elision is safe exactly where a miss ends the lookup at the fallback, which is true
// of ByType and false of everything else, so ByNamespace and ByLiteral pass a nil fallback.
//
// The keep filter is the framework/repository seam, and it applies only to ByNamespace. The two
// pseudo-namespaces are always kept: `@colorKeyword` is Tailwind's own colour keywords and `@none`
// is the last step of the bare path, which is also the reading every unconsumed namespace takes.
func writeStringReadingMap(out *bytes.Buffer, name, indent string, entries map[string]*reading, fallback *reading, keep map[string]bool, pool *readingPool) {
	if len(entries) == 0 {
		return
	}
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var mapBuffer bytes.Buffer
	fallbackKey := readingKey(fallback)
	for _, key := range keys {
		value := entries[key]
		if value == nil {
			continue
		}
		if keep != nil && key != namespaceColorKeyword && key != namespaceNone && !keep[key] {
			continue
		}
		if fallback != nil && readingKey(value) == fallbackKey {
			continue
		}
		fmt.Fprintf(&mapBuffer, "\t%s%q: baseReadings[%d],\n", indent, key, pool.intern(value))
	}
	if mapBuffer.Len() == 0 {
		return
	}

	keyType := "string"
	if name == "ByType" {
		keyType = "DataType"
	}
	fmt.Fprintf(out, "%s%s: map[%s]Reading{\n", indent, name, keyType)
	out.Write(mapBuffer.Bytes())
	fmt.Fprintf(out, "%s},\n", indent)
}

func humanSize(size int) string {
	switch {
	case size >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(size)/(1<<20))
	case size >= 1<<10:
		return fmt.Sprintf("%.0f KB", float64(size)/(1<<10))
	default:
		return fmt.Sprintf("%d B", size)
	}
}
