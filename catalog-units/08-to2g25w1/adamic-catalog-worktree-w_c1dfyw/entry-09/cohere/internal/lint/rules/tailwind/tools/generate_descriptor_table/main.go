// Command generate_descriptor_table turns the descriptor extractor's diagnostic JSON into the
// Go table `internal/tailwind` looks readings up in.
//
// Usage:
//
//	go run ./internal/lint/rules/tailwind/tools/generate_descriptor_table -input <table.json> -output <table.go> [-package <name>] [-variable <name>]
//
// The input is what `internal/lint/rules/tailwind/tools/generate_descriptors/extract.mjs --json` writes, which is the
// uncompacted diagnostic shape: a full `{order, count}` per bucket per root across all three
// modifier axes, 465 KB on this repository. Most of it is duplicates, and this tool is where that is
// measured rather than assumed. Two reductions, both of which report what they removed:
//
//   - Interning. There are 1,707 buckets in the input and 379 distinct readings among them, so the
//     readings become a deduplicated array and every bucket becomes an index into it. That also
//     collapses the order slices, which are 641 integers in total across the distinct readings and
//     several thousand as written.
//
//   - Fallback elision. 1,297 of those 1,707 buckets are equal to their own axis's fallback, which
//     the lookup already returns on a miss. Dropping them is not a size trick, it is the fourth
//     refinement of the model made structural: a type earns its place in a root's type list by
//     discriminating on *some* axis, and storing its non-discrimination on the other two is storing
//     the absence of information.
//
// The reason to care is on the record in this slice: an earlier attempt produced a 910 KB table by
// enumerating values instead of readings, and a 6.5 MB fixture by storing trees it did not need.
// Both looked reasonable while being written, so this tool prints its own output size and its
// reduction counts on every run, and the number goes in the commit message.
//
// # The table is generated per repository and is not checked in
//
// `Namespaces` and `KeysByNamespace` are keyed on the repository's own `@theme`, so a committed
// table would bake one repository's tokens into a file claiming to describe Tailwind, which is the
// bug the native port exists to fix. Measured while writing this: generating against
// ~/Projects/ahra and ~/Projects/connected/www-connected-app produces byte-identical `roots` arrays
// and different namespace sets, and a synthetic theme whose only colour token is `--color-weird-1`
// yields a `--color-weird` namespace that changes the reading on 20 roots. So the descriptor
// *structure* is framework-invariant and the namespace axis is not, and only a table generated
// against the repository in front of us is right for it.
//
// # Verifying the generated Go, not just compiling it
//
// A generator that emits a file which compiles has proved nothing about the file's contents: the
// printing is where an index can be written next to the wrong key and still parse. So the suite in
// `internal/tailwind` builds the same table in memory from the same two JSON inputs and asserts
// against the engine's readings, and the generated Go was checked against that in-memory table over
// the whole fixture corpus, 10,977 cases, all agreeing. Rerun that comparison after changing the
// printing here; agreement with the engine is the contract, and compiling is not evidence of it.
//
// Node is required to produce the input and never to use the result.
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

// extractedContext is what `context.mjs` collects: the parts of the table that are not readings.
//
// It is a second input rather than a second pass over the first because `extract.mjs` is the tool
// that proved the model and is left untouched. See context.mjs for what each field is and why it
// cannot be derived from the readings alone.
type extractedContext struct {
	TailwindVersion     string              `json:"tailwindVersion"`
	Namespaces          []string            `json:"namespaces"`
	KeysByNamespace     map[string][]string `json:"keysByNamespace"`
	PropertyOrder       []string            `json:"propertyOrder"`
	PerDeclarationRoots []string            `json:"perDeclarationRoots"`
	// UnlistedStatics are static utilities the engine compiles and getClassList() does not
	// advertise. See context.mjs: one class on this design system, and the table would otherwise
	// decline a real class.
	UnlistedStatics map[string]*reading `json:"unlistedStatics"`
}

func main() {
	inputPath := flag.String("input", "", "path to the JSON written by generate_descriptors/extract.mjs --json")
	contextPath := flag.String("context", "", "path to the JSON written by context.mjs")
	outputPath := flag.String("output", "", "path to write the generated Go file to; - writes to stdout")
	packageName := flag.String("package", "tailwind", "package clause of the generated file")
	variableName := flag.String("variable", "GeneratedTable", "name of the generated table variable")
	flag.Parse()

	if *inputPath == "" || *outputPath == "" || *contextPath == "" {
		fmt.Fprintln(os.Stderr, "generate_descriptor_table: -input, -context and -output are required")
		os.Exit(2)
	}

	inputBytes, err := os.ReadFile(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate_descriptor_table: reading input: %v\n", err)
		os.Exit(1)
	}

	var table extractedTable
	if err := json.Unmarshal(inputBytes, &table); err != nil {
		fmt.Fprintf(os.Stderr, "generate_descriptor_table: parsing input: %v\n", err)
		os.Exit(1)
	}

	// The input is a diagnostic artifact, so it is checked rather than trusted. A table generated
	// from an empty or truncated extraction would compile and produce wrong readings silently,
	// which is the failure mode the volume assertions in the extractor exist to prevent and there
	// is no reason to drop the guard at the boundary between the two tools.
	if len(table.Roots) == 0 || len(table.Statics) == 0 || table.TailwindVersion == "" {
		fmt.Fprintf(os.Stderr, "generate_descriptor_table: input looks empty (%d roots, %d statics, version %q); refusing to generate\n",
			len(table.Roots), len(table.Statics), table.TailwindVersion)
		os.Exit(3)
	}

	contextBytes, err := os.ReadFile(*contextPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate_descriptor_table: reading context: %v\n", err)
		os.Exit(1)
	}

	var tableContext extractedContext
	if err := json.Unmarshal(contextBytes, &tableContext); err != nil {
		fmt.Fprintf(os.Stderr, "generate_descriptor_table: parsing context: %v\n", err)
		os.Exit(1)
	}

	if len(tableContext.Namespaces) == 0 || len(tableContext.PropertyOrder) == 0 {
		fmt.Fprintf(os.Stderr, "generate_descriptor_table: context looks empty (%d namespaces, %d properties); refusing to generate\n",
			len(tableContext.Namespaces), len(tableContext.PropertyOrder))
		os.Exit(3)
	}

	// The two inputs must describe the same engine. They are produced by separate invocations, so
	// nothing but this check stops a stale readings file being paired with a fresh context and
	// generating a table that is internally inconsistent in a way no test would name.
	if tableContext.TailwindVersion != table.TailwindVersion {
		fmt.Fprintf(os.Stderr, "generate_descriptor_table: input is Tailwind %s and context is Tailwind %s; refusing to generate\n",
			table.TailwindVersion, tableContext.TailwindVersion)
		os.Exit(3)
	}

	generated, stats := generate(&table, &tableContext, *packageName, *variableName)

	formatted, err := format.Source(generated)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate_descriptor_table: generated file does not parse: %v\n", err)
		os.Exit(1)
	}

	if *outputPath == "-" {
		os.Stdout.Write(formatted)
	} else if err := os.WriteFile(*outputPath, formatted, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "generate_descriptor_table: writing output: %v\n", err)
		os.Exit(1)
	}

	// Printed on every run, because the size of this table is a thing this slice has already got
	// wrong twice and a number nobody looks at is a number nobody catches.
	fmt.Printf("generate_descriptor_table: %d roots, %d statics (%d recovered from outside the registry), %d namespaces\n", len(table.Roots), stats.statics, len(tableContext.UnlistedStatics), stats.namespaces)
	fmt.Printf("  buckets %d in, %d kept (%d dropped as equal to their axis fallback)\n", stats.bucketsIn, stats.bucketsKept, stats.bucketsIn-stats.bucketsKept)
	fmt.Printf("  readings %d interned, %d order integers\n", stats.distinctReadings, stats.orderIntegers)
	fmt.Printf("  input %s, output %s\n", humanSize(len(inputBytes)), humanSize(len(formatted)))
}

type generationStats struct {
	bucketsIn        int
	bucketsKept      int
	distinctReadings int
	orderIntegers    int
	namespaces       int
	statics          int
}

// readingPool interns readings so that many buckets share one.
//
// The key is the reading's own content rather than a pointer, because the extractor writes an
// independent object per bucket and the duplicates are only discoverable by value.
type readingPool struct {
	indexByKey map[string]int
	readings   []*reading
}

func newReadingPool() *readingPool {
	return &readingPool{indexByKey: make(map[string]int)}
}

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

func generate(table *extractedTable, tableContext *extractedContext, packageName, variableName string) ([]byte, generationStats) {
	stats := generationStats{}
	stats.namespaces = len(tableContext.Namespaces)
	pool := newReadingPool()

	perDeclaration := make(map[string]bool, len(tableContext.PerDeclarationRoots))
	for _, root := range tableContext.PerDeclarationRoots {
		perDeclaration[root] = true
	}

	var body bytes.Buffer

	// Roots, in the input's order, which the extractor already sorted.
	var rootsBuffer bytes.Buffer
	for index := range table.Roots {
		entry := &table.Roots[index]
		writeDescriptor(&rootsBuffer, entry, perDeclaration[entry.Root], pool, &stats)
	}

	var staticsBuffer bytes.Buffer
	// The registry's statics, plus the ones it does not advertise. The registry wins on a
	// collision, since it is the measured source and the probe is the supplement.
	allStatics := make(map[string]*reading, len(table.Statics)+len(tableContext.UnlistedStatics))
	for name, value := range tableContext.UnlistedStatics {
		allStatics[name] = value
	}
	for name, value := range table.Statics {
		allStatics[name] = value
	}
	staticNames := make([]string, 0, len(allStatics))
	for name := range allStatics {
		staticNames = append(staticNames, name)
	}
	sort.Strings(staticNames)
	for _, name := range staticNames {
		if allStatics[name] == nil {
			continue
		}
		fmt.Fprintf(&staticsBuffer, "\t\t%q: generatedReadings[%d],\n", name, pool.intern(allStatics[name]))
	}
	stats.statics = len(staticNames)

	stats.distinctReadings = len(pool.readings)
	for _, value := range pool.readings {
		stats.orderIntegers += len(value.Order)
	}

	fmt.Fprintf(&body, `// Code generated by internal/lint/rules/tailwind/tools/generate_descriptor_table. DO NOT EDIT.
//
// Generated from %s against Tailwind %s.
//
// This file is per-repository by construction and is not checked in: Namespaces and KeysByNamespace
// come from this repository's own @theme, and a committed copy would report one repository's tokens
// as facts about Tailwind. See the command's doc comment for the measurement behind that claim.

package %s

// generatedReadings holds every distinct reading in the table exactly once.
//
// %d buckets in the extractor's output hold %d distinct readings between them, so every bucket below
// is an index into this array rather than a copy. Readings are immutable and shared; nothing in the
// port writes to one.
var generatedReadings = []Reading{
`, jsonPathComment(table.EntryPoint), table.TailwindVersion, packageName, stats.bucketsIn, len(pool.readings))

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

	fmt.Fprintf(&body, "}\n\n// %s is this repository's descriptor table.\nvar %s = &Table{\n\tTailwindVersion: %q,\n\tDescriptors: map[string]*Descriptor{\n", variableName, variableName, table.TailwindVersion)
	body.Write(rootsBuffer.Bytes())
	fmt.Fprintf(&body, "\t},\n\tStatics: map[string]Reading{\n")
	body.Write(staticsBuffer.Bytes())
	fmt.Fprintf(&body, "\t},\n")

	// Longest-first, and written in the order context.mjs sorted rather than re-sorted here. The
	// order is the engine's namespace precedence and re-deriving it in a second place is how the two
	// drift.
	fmt.Fprintf(&body, "\tNamespaces: []string{\n")
	for _, namespace := range tableContext.Namespaces {
		fmt.Fprintf(&body, "\t\t%q,\n", namespace)
	}
	fmt.Fprintf(&body, "\t},\n\tKeysByNamespace: map[string]map[string]bool{\n")
	for _, namespace := range tableContext.Namespaces {
		keys := tableContext.KeysByNamespace[namespace]
		if len(keys) == 0 {
			continue
		}
		sorted := append([]string(nil), keys...)
		sort.Strings(sorted)
		fmt.Fprintf(&body, "\t\t%q: {\n", namespace)
		for _, key := range sorted {
			fmt.Fprintf(&body, "\t\t\t%q: true,\n", key)
		}
		fmt.Fprintf(&body, "\t\t},\n")
	}
	fmt.Fprintf(&body, "\t},\n\tPropertyOrder: map[string]int{\n")
	for index, property := range tableContext.PropertyOrder {
		fmt.Fprintf(&body, "\t\t%q: %d,\n", property, index)
	}
	fmt.Fprintf(&body, "\t},\n}\n")

	return body.Bytes(), stats
}

// jsonPathComment keeps an absolute developer path out of a generated file, since the path differs
// per machine and would show up as a spurious diff.
func jsonPathComment(entryPoint string) string {
	if entryPoint == "" {
		return "an unnamed design system"
	}
	parts := strings.Split(entryPoint, "/")
	if len(parts) <= 3 {
		return entryPoint
	}
	return ".../" + strings.Join(parts[len(parts)-3:], "/")
}

func writeDescriptor(out *bytes.Buffer, entry *rootEntry, perDeclaration bool, pool *readingPool, stats *generationStats) {
	fmt.Fprintf(out, "\t\t%q: {\n\t\t\tRoot: %q,\n", entry.Root, entry.Root)
	if perDeclaration {
		// A root the table cannot answer. See Descriptor.PerDeclaration.
		fmt.Fprintf(out, "\t\t\tPerDeclaration: true,\n")
	}

	if len(entry.TypeList) > 0 {
		typeParts := make([]string, len(entry.TypeList))
		for index, typeName := range entry.TypeList {
			typeParts[index] = fmt.Sprintf("%q", typeName)
		}
		// The order is the engine's, recovered by topological sort from observation. It is written
		// out verbatim and must not be sorted: InferDataType returns the first match.
		fmt.Fprintf(out, "\t\t\tTypeList: []DataType{%s},\n", strings.Join(typeParts, ", "))
	}

	writeStringReadingMap(out, "ByLiteral", entry.ReadingByLiteral, nil, pool, stats)

	writeAxis(out, "Absent", entry.ReadingByType, entry.ReadingByNamespace, entry.Fallback, entry.Empty, pool, stats)
	writeAxis(out, "Alpha", entry.ReadingByTypeWithModifier, entry.ReadingByNamespaceWithModifier, entry.FallbackWithModifier, entry.EmptyWithModifier, pool, stats)
	writeAxis(out, "Themed", entry.ReadingByTypeWithThemedModifier, entry.ReadingByNamespaceWithThemedModifier, entry.FallbackWithThemedModifier, entry.EmptyWithThemedModifier, pool, stats)

	fmt.Fprintf(out, "\t\t},\n")
}

func writeAxis(out *bytes.Buffer, name string, byType, byNamespace map[string]*reading, fallback, empty *reading, pool *readingPool, stats *generationStats) {
	var axisBuffer bytes.Buffer
	writeStringReadingMap(&axisBuffer, "ByType", byType, fallback, pool, stats)
	// Namespaces keep their fallback-equal entries. A miss here continues the bare-value
	// precedence to inference rather than ending at the fallback, so eliding one changes the answer;
	// see AxisReadings.ByNamespace for the measurement.
	writeStringReadingMap(&axisBuffer, "ByNamespace", byNamespace, nil, pool, stats)
	if fallback != nil {
		fmt.Fprintf(&axisBuffer, "\t\t\t\tFallback: generatedReadings[%d],\n", pool.intern(fallback))
	}
	if empty != nil {
		fmt.Fprintf(&axisBuffer, "\t\t\t\tEmpty: &generatedReadings[%d],\n", pool.intern(empty))
	}
	if axisBuffer.Len() == 0 {
		return
	}
	fmt.Fprintf(out, "\t\t\t%s: AxisReadings{\n", name)
	out.Write(axisBuffer.Bytes())
	fmt.Fprintf(out, "\t\t\t},\n")
}

// writeStringReadingMap emits one map, dropping every entry equal to fallback.
//
// The elision is safe exactly where a miss ends the lookup at the fallback, which is true of ByType
// and false of everything else. Pass a nil fallback to keep every entry, which is what ByNamespace
// and ByLiteral need: a miss in either continues the bare-value precedence to a later step that
// answers differently, so an elided entry is not the same answer as an absent one.
func writeStringReadingMap(out *bytes.Buffer, name string, entries map[string]*reading, fallback *reading, pool *readingPool, stats *generationStats) {
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
		stats.bucketsIn++
		if value == nil {
			continue
		}
		if fallback != nil && readingKey(value) == fallbackKey {
			continue
		}
		stats.bucketsKept++
		fmt.Fprintf(&mapBuffer, "\t\t\t\t\t%q: generatedReadings[%d],\n", key, pool.intern(value))
	}
	if mapBuffer.Len() == 0 {
		return
	}

	valueType := "Reading"
	keyType := "string"
	if name == "ByType" {
		keyType = "DataType"
	}
	indent := "\t\t\t\t"
	if name == "ByLiteral" {
		indent = "\t\t\t"
	}
	fmt.Fprintf(out, "%s%s: map[%s]%s{\n", indent, name, keyType, valueType)
	out.Write(mapBuffer.Bytes())
	fmt.Fprintf(out, "%s},\n", indent)
}

func humanSize(bytes int) string {
	switch {
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.0f KB", float64(bytes)/(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
