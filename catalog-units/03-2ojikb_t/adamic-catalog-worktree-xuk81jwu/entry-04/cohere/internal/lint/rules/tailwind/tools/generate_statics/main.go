// Command generate_statics turns the engine's own static-utility registrations into Go.
//
// Usage:
//
//	go run ./internal/lint/rules/tailwind/tools/generate_statics \
//	    -input <a.json> -input <b.json> -input <c.json> \
//	    -output internal/lint/rules/tailwind/collapse/framework_statics_table.go
//
// Each input is written by enumerate.mjs against one design system. Several are required and the
// reason is the same one generate_descriptor_base states: one extraction cannot tell a fact
// about Tailwind from a fact about a repository, so a table generated from a single system would
// ship one repository's `@utility` blocks as framework registrations. This refuses fewer than two
// and re-checks the split on every invocation rather than trusting the split it was built with.
//
// The emitted table holds only the intersection: names every input system registers, agreeing on
// every declaration. A name present in some systems and not others is a repository utility and is
// dropped with its origin reported, which is how `markdown-content` stays out.
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
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// staticDeclaration mirrors one declaration of enumerate.mjs's output.
type staticDeclaration struct {
	Property     string  `json:"property"`
	Value        *string `json:"value"`
	ValuePresent bool    `json:"valuePresent"`
	Important    bool    `json:"important"`
}

type staticReading struct {
	Order []int `json:"order"`
	Count int   `json:"count"`
}

type staticEntry struct {
	Name         string              `json:"name"`
	Declarations []staticDeclaration `json:"declarations"`
	Reading      *staticReading      `json:"reading"`
}

type staticsFile struct {
	TailwindVersion   string        `json:"tailwindVersion"`
	EntryPoint        string        `json:"entryPoint"`
	RegisteredStatics int           `json:"registeredStatics"`
	Emitted           int           `json:"emitted"`
	Statics           []staticEntry `json:"statics"`
}

type inputPaths []string

func (paths *inputPaths) String() string { return strings.Join(*paths, ", ") }

func (paths *inputPaths) Set(value string) error {
	*paths = append(*paths, value)
	return nil
}

func main() {
	var inputs inputPaths
	flag.Var(&inputs, "input", "path to a JSON file written by enumerate.mjs; repeatable, at least two required")
	output := flag.String("output", "", "path to write the generated Go file")
	flag.Parse()

	if len(inputs) < 2 {
		fail("at least two design systems are required: one extraction cannot tell a framework registration from a repository one, and a table built from a single system would ship that repository's utilities as framework facts")
	}

	systems := make([]staticsFile, 0, len(inputs))
	for _, path := range inputs {
		raw, err := os.ReadFile(path)
		if err != nil {
			fail("read %s: %v", path, err)
		}
		var file staticsFile
		if err := json.Unmarshal(raw, &file); err != nil {
			fail("parse %s: %v", path, err)
		}
		if len(file.Statics) == 0 {
			fail("%s holds no statics, so an intersection against it is empty for a reason that is not a measurement", path)
		}
		systems = append(systems, file)
	}

	// One Tailwind, or the comparison is between two different engines and proves nothing.
	for index := 1; index < len(systems); index++ {
		if systems[index].TailwindVersion != systems[0].TailwindVersion {
			fail("inputs were generated against different Tailwind versions (%s and %s); a check run against a different engine than the table came from cannot detect a difference in the table",
				systems[0].TailwindVersion, systems[index].TailwindVersion)
		}
	}

	shared, dropped := intersect(systems)
	if len(shared) == 0 {
		fail("no static utility is registered by every input system, which means the inputs describe unrelated design systems rather than one framework")
	}

	// The split has to be visible, not merely performed. A run where every system registers exactly
	// the same names has not demonstrated that the intersection is doing anything, and that is the
	// state two repositories vendoring one submodule are in.
	if len(dropped) == 0 {
		fail("every input system registers an identical name set, so the intersection removed nothing and the framework/repository split is unverified; include a design system that shares no stylesheet with the others")
	}

	source, err := render(systems, shared, dropped)
	if err != nil {
		fail("render: %v", err)
	}

	if *output == "" {
		os.Stdout.Write(source)
	} else if err := os.WriteFile(*output, source, 0o644); err != nil {
		fail("write %s: %v", *output, err)
	}

	fmt.Fprintf(os.Stderr, "generate_statics: %d design systems, %d statics registered by all of them (%d dropped as repository-specific)\n",
		len(systems), len(shared), len(dropped))
	fmt.Fprintf(os.Stderr, "  dropped: %s\n", strings.Join(firstFew(dropped, 8), ", "))
}

// intersect keeps the names every system registers with identical declarations, and reports the rest.
//
// Declaration equality rather than reading equality is deliberate. Two utilities can agree on
// `{order, count}` while declaring different properties, since the reading is a sorted set of
// positions and a total; comparing the declarations is what makes a disagreement visible at the
// place it happens rather than at the place it survives.
func intersect(systems []staticsFile) (shared []staticEntry, dropped []string) {
	byName := make([]map[string]staticEntry, len(systems))
	for index, system := range systems {
		byName[index] = make(map[string]staticEntry, len(system.Statics))
		for _, entry := range system.Statics {
			byName[index][entry.Name] = entry
		}
	}

	droppedSet := map[string]bool{}
	for name, entry := range byName[0] {
		agrees := true
		for index := 1; index < len(byName); index++ {
			other, found := byName[index][name]
			if !found || !declarationsEqual(entry.Declarations, other.Declarations) {
				agrees = false
				break
			}
		}
		if agrees {
			shared = append(shared, entry)
			continue
		}
		droppedSet[name] = true
	}
	// Names absent from the first system entirely are repository utilities of the others.
	for index := 1; index < len(byName); index++ {
		for name := range byName[index] {
			if _, found := byName[0][name]; !found {
				droppedSet[name] = true
			}
		}
	}

	sort.Slice(shared, func(left, right int) bool { return shared[left].Name < shared[right].Name })
	for name := range droppedSet {
		dropped = append(dropped, name)
	}
	sort.Strings(dropped)
	return shared, dropped
}

func declarationsEqual(left []staticDeclaration, right []staticDeclaration) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Property != right[index].Property ||
			left[index].ValuePresent != right[index].ValuePresent ||
			left[index].Important != right[index].Important {
			return false
		}
		if (left[index].Value == nil) != (right[index].Value == nil) {
			return false
		}
		if left[index].Value != nil && *left[index].Value != *right[index].Value {
			return false
		}
	}
	return true
}

func render(systems []staticsFile, shared []staticEntry, dropped []string) ([]byte, error) {
	sources := make([]string, 0, len(systems))
	for _, system := range systems {
		sources = append(sources, filepath.Base(system.EntryPoint))
	}

	var body bytes.Buffer
	fmt.Fprintf(&body, `// Code generated by internal/lint/rules/tailwind/tools/generate_statics. DO NOT EDIT.
//
// Generated against Tailwind %s from %d design systems: %s.
//
// Every static utility the framework registers, with the declarations it compiles to. A static
// utility takes no value, so its body is a constant and its reading is PropertySort over that body.
// Carrying the declarations rather than the reading is what lets the Go side compute the reading
// with the same walk it uses for a repository's own `+"`@utility`"+` blocks, so the two paths cannot
// disagree about what a declaration list means.
//
// # Why this is checked in when the port exists to remove checked-in tables
//
// These are registrations, not readings of a repository's tokens. A static utility declares literal
// property/value pairs from Tailwind's own source and consults no theme, which is what makes it a
// framework fact; the generator proves that rather than asserting it, by keeping only the names every
// input system registers with identical declarations and refusing a run where the intersection
// removed nothing.
//
//   - %d statics registered identically by every input system, printed here.
//   - %d dropped as repository-specific, including %s.
//
// # Read from the registry, not from the class list
//
// `+"`getClassList()`"+` enumerates classes, so a utility advertising none is invisible to it. That
// hid fourteen functional roots until 33e12df and 6347b07, and it hides 21 statics here: the
// deprecated-but-registered ones such as `+"`bg-gradient-to-r`"+`, `+"`break-words`"+` and
// `+"`max-w-screen`"+`, which the base descriptor table never carried. enumerate.mjs reads
// `+"`utilities.keys('static')`"+` instead.
//
// Regenerate with:
//
//	node internal/lint/rules/tailwind/tools/generate_statics/enumerate.mjs <theme.css> > <system>.json   (once per system)
//	go run ./internal/lint/rules/tailwind/tools/generate_statics -input <a.json> -input <b.json> -input <c.json> \
//	    -output internal/lint/rules/tailwind/collapse/framework_statics_table.go

package tailwind

// FrameworkStaticDeclarations is the compiled body of every static utility the framework registers.
//
// Keyed by class name. The value is the declaration sequence in the order PropertySort visits it,
// which is the order the engine emitted, so a consumer walks it rather than re-deriving one.
var FrameworkStaticDeclarations = map[string][]StaticDeclaration{
`, systems[0].TailwindVersion, len(systems), strings.Join(sources, ", "),
		len(shared), len(dropped), strings.Join(firstFew(dropped, 3), ", "))

	for _, entry := range shared {
		fmt.Fprintf(&body, "\t%s: {", strconv.Quote(entry.Name))
		for index, declaration := range entry.Declarations {
			if index > 0 {
				body.WriteString(", ")
			}
			value := ""
			if declaration.Value != nil {
				value = *declaration.Value
			}
			fmt.Fprintf(&body, "{Property: %s, Value: %s", strconv.Quote(declaration.Property), strconv.Quote(value))
			if declaration.ValuePresent {
				body.WriteString(", ValuePresent: true")
			}
			if declaration.Important {
				body.WriteString(", Important: true")
			}
			body.WriteString("}")
		}
		body.WriteString("},\n")
	}
	body.WriteString("}\n")

	return format.Source(body.Bytes())
}

func firstFew(values []string, count int) []string {
	if len(values) < count {
		return values
	}
	return values[:count]
}

func fail(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "generate_statics: "+format+"\n", arguments...)
	os.Exit(1)
}
