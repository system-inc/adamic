package fuzz

import (
	_ "embed"
	"fmt"
	"math/rand/v2"
	goruntime "runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// The breadth scene writes constructs the rest of the generator has no way to write, chosen from the
// blind map (verify/coverage/REPORT.md): the regions of the compiler and the runtime that 2,000 seeds
// never executed. Each construct says which regions it targets, by file and line range as the blind
// map measured them at 7863216ed, so the next measurement can say whether it reached them.
//
// Weighting can't produce what the generator can't write, so this scene is vocabulary first. Which
// constructs a program gets is weighted, by breadth_weights.txt, and cmd/adamic-steer rewrites that
// table from coverage: a construct whose programs lit regions nothing had reached gains weight, one
// whose programs only revisit covered code loses it.
//
// Not written, because Adamic 0.1 refuses them: labeled break and continue (refusals.go, "a label"),
// a case that falls through into the next one's statements (noFallthroughCasesInSwitch; empty cases
// that share a body are written), the in operator and delete. Not written, because they would make
// the oracle inexact: Math.random, time, a program's arguments and a directory listing.
//
// The scene draws from a random stream of its own, and names with a prefix of its own, so adding it
// changes nothing else a seed writes.

// breadthStream is the scene's random stream, beside the generator's own.
const breadthStream = 0x62726561647468

// breadthFeature is the scene's name in Features.
const breadthFeature = "breadth"

// BreadthPerProgram is how many constructs one program gets.
const BreadthPerProgram = 3

//go:embed breadth_weights.txt
var breadthWeightsText string

// BreadthWeights is the checked-in weight table: a construct's name, or construct/variant, and its
// weight. A name missing from the table weighs 1.
func BreadthWeights() map[string]float64 {
	weights, err := ParseWeights(breadthWeightsText)
	if err != nil {
		panic("fuzz: breadth_weights.txt: " + err.Error())
	}
	return weights
}

// ParseWeights reads a weight table: one "name weight" per line, # for comments.
func ParseWeights(text string) (map[string]float64, error) {
	weights := map[string]float64{}
	for number, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("line %d: want a name and a weight, got %q", number+1, line)
		}
		weight, err := strconv.ParseFloat(fields[1], 64)
		if err != nil || weight <= 0 {
			return nil, fmt.Errorf("line %d: weight %q isn't a positive number", number+1, fields[1])
		}
		weights[fields[0]] = weight
	}
	return weights, nil
}

// FormatWeights writes a weight table in the order of BreadthChoices, every construct and variant.
func FormatWeights(weights map[string]float64, header string) string {
	var builder strings.Builder
	for _, line := range strings.Split(strings.TrimRight(header, "\n"), "\n") {
		builder.WriteString("# " + line + "\n")
	}
	for _, name := range BreadthChoices() {
		weight, known := weights[name]
		if !known {
			weight = 1
		}
		fmt.Fprintf(&builder, "%s %s\n", name, strconv.FormatFloat(weight, 'f', 4, 64))
	}
	return builder.String()
}

// BreadthChoices is every name the weight table can weigh: each construct, then its variants.
func BreadthChoices() []string {
	var names []string
	for _, construct := range breadthConstructs() {
		names = append(names, construct.name)
		for _, variant := range construct.variants {
			names = append(names, construct.name+"/"+variant)
		}
	}
	return names
}

// construct is one thing the scene can write.
type construct struct {
	name string
	// targets are the blind map's regions the construct is for, as "file:lines function".
	targets  []string
	variants []string
	write    func(b *breadth, variant string) []string
	// optIn is what the construct needs from OptIn, when stage 0 can't lower it in every program yet;
	// variantOptIn is the same for one variant.
	optIn        string
	variantOptIn map[string]string
}

// The breadth scene's opt-ins: what stage 0 can't lower yet.
const (
	// breadthStatics: a class with static members makes stage 0 decline every method call through a
	// structural signature and every object spread in the program ("in a program with statics" and
	// "with static constructor objects"), and the other scenes make both.
	breadthStatics = "breadth-statics"
	// breadthFrozen: a try around a write to a frozen object, which panics natively but throws a
	// catchable TypeError on Node.
	breadthFrozen = "breadth-frozen"
	// breadthDarwinMath forces the math construct on darwin, where it's off by default. Node's arm64
	// build fuses multiply-adds and Adamic deliberately doesn't (#myatdyv), so on an arm64 Mac a
	// transcendental result can differ from Node's in its last bit for that reason alone:
	// Math.tan(1e22) is -1.628778225606899 natively and on Node for linux/amd64, and
	// -1.6287782256068988 on Node for arm64. Linux x86 is the gate of record, where they agree, so
	// the construct stays on there. Coverage measurements turn it on everywhere (measure.sh).
	breadthDarwinMath = "breadth-darwin-math"
)

// constructAllowed says whether a program may get a construct: its opt-in asked for, and math off on
// darwin unless forced (breadthDarwinMath).
func (g *generator) constructAllowed(construct construct) bool {
	if construct.optIn != "" && !g.with[construct.optIn] {
		return false
	}
	if construct.name == "math" && goruntime.GOOS == "darwin" && !g.with[breadthDarwinMath] {
		return false
	}
	return true
}

// breadth is the scene while it writes one program.
type breadth struct {
	random  *rand.Rand
	weights map[string]float64
	names   int
	// chosen is what the program got, construct and construct/variant, for cmd/adamic-steer.
	chosen []string
}

func (b *breadth) name(prefix string) string {
	b.names++
	return fmt.Sprintf("breadth%s%d", prefix, b.names)
}

func (b *breadth) intn(limit int) int { return b.random.IntN(limit) }

func (b *breadth) between(low, high int) int { return low + b.random.IntN(high-low+1) }

func (b *breadth) pick(options ...string) string { return options[b.random.IntN(len(options))] }

func (b *breadth) weight(name string) float64 {
	if weight, known := b.weights[name]; known {
		return weight
	}
	return 1
}

// weighted picks an index by weight, deterministically from the scene's stream.
func (b *breadth) weighted(weights []float64) int {
	total := 0.0
	for _, weight := range weights {
		total += weight
	}
	target := b.random.Float64() * total
	for index, weight := range weights {
		if target < weight {
			return index
		}
		target -= weight
	}
	return len(weights) - 1
}

// breadthProgram is the scene: BreadthPerProgram distinct constructs, by weight, each with a variant
// by weight.
func (g *generator) breadthProgram() []*Statement {
	scene := &breadth{random: rand.New(rand.NewPCG(g.seed, breadthStream)), weights: g.weights}
	if scene.weights == nil {
		scene.weights = BreadthWeights()
	}
	var constructs []construct
	for _, construct := range breadthConstructs() {
		if g.constructAllowed(construct) {
			constructs = append(constructs, construct)
		}
	}
	var lines []string
	for range BreadthPerProgram {
		weights := make([]float64, len(constructs))
		for index, construct := range constructs {
			weights[index] = scene.weight(construct.name)
		}
		index := scene.weighted(weights)
		chosen := constructs[index]
		constructs = slices.Delete(slices.Clone(constructs), index, index+1)
		variantWeights := make([]float64, len(chosen.variants))
		for variantIndex, variant := range chosen.variants {
			variantWeights[variantIndex] = scene.weight(chosen.name + "/" + variant)
			if needs := chosen.variantOptIn[variant]; needs != "" && !g.with[needs] {
				variantWeights[variantIndex] = 0
			}
		}
		variant := chosen.variants[scene.weighted(variantWeights)]
		scene.chosen = append(scene.chosen, chosen.name, chosen.name+"/"+variant)
		lines = append(lines, chosen.write(scene, variant)...)
	}
	g.breadthChosen = scene.chosen
	return parseStatements(lines)
}

// parseStatements turns lines of source into statements, each brace block a Block, so the shrinker
// can take a construct apart. A line ending in { opens a block; a line starting with } closes it, and
// opens the next when it ends in { too (} else {, } catch (error) {, } finally {).
func parseStatements(lines []string) []*Statement {
	var trimmed []string
	for _, line := range lines {
		for _, piece := range strings.Split(line, "\n") {
			if piece = strings.TrimSpace(piece); piece != "" {
				trimmed = append(trimmed, piece)
			}
		}
	}
	position := 0
	var block func() []*Statement
	block = func() []*Statement {
		var statements []*Statement
		for position < len(trimmed) {
			line := trimmed[position]
			if strings.HasPrefix(line, "}") {
				return statements
			}
			position++
			if !strings.HasSuffix(line, "{") {
				statements = append(statements, &Statement{Parts: []StatementPart{{Text: line}}})
				continue
			}
			current := &Statement{Parts: []StatementPart{{Text: strings.TrimSuffix(line, "{")}}}
			for {
				inner := block()
				current.Parts = append(current.Parts, StatementPart{Block: &Block{Statements: inner}})
				if position >= len(trimmed) {
					break
				}
				closing := strings.TrimPrefix(trimmed[position], "}")
				position++
				if strings.HasSuffix(closing, "{") {
					current.Parts = append(current.Parts, StatementPart{Text: strings.TrimSuffix(closing, "{")})
					continue
				}
				if closing != "" {
					current.Parts = append(current.Parts, StatementPart{Text: closing})
				}
				break
			}
			statements = append(statements, current)
		}
		return statements
	}
	return block()
}

// caught runs a statement that may throw and prints what was thrown.
func caught(body ...string) []string {
	lines := []string{"try {"}
	lines = append(lines, body...)
	return append(lines, "} catch (error) {", "console.log(error instanceof Error ? `caught ${error.message}` : 'caught something');", "}")
}

// loopShapes are the loops a construct can put its body in. Each declares index, a number, before
// the body runs, and runs it for index 0 up to limit - 1 (do...while at least once).
var loopShapes = []string{"for", "while", "do-while", "for-of", "for-of-map", "for-in"}

// loop wraps a body in a loop shape, with limit the bound's name.
func (b *breadth) loop(shape string, limit string, body []string) []string {
	var lines []string
	switch shape {
	case "for":
		lines = append(lines, "for (let index = 0; index < "+limit+"; index++) {")
		lines = append(lines, body...)
		lines = append(lines, "}")
	case "while":
		lines = append(lines, "let index = -1;", "while (index < "+limit+" - 1) {", "index++;")
		lines = append(lines, body...)
		lines = append(lines, "}")
	case "do-while":
		lines = append(lines, "let index = -1;", "do {", "index++;")
		lines = append(lines, body...)
		lines = append(lines, "} while (index < "+limit+" - 1);")
	case "for-of":
		lines = append(lines, "const items: number[] = [];", "for (let fill = 0; fill < "+limit+"; fill++) {", "items.push(fill);", "}", "for (const index of items) {")
		lines = append(lines, body...)
		lines = append(lines, "}")
	case "for-of-map":
		lines = append(lines, "const entries = new Map<string, number>();", "for (let fill = 0; fill < "+limit+"; fill++) {", "entries.set(`key${fill}`, fill);", "}", "for (const [key, index] of entries) {")
		lines = append(lines, body...)
		lines = append(lines, "}")
	case "for-in":
		lines = append(lines, "const fields = { a: 0, '10': 1, b: 2, '2': 3, c: 4, '0': 5, d: 6, e: 7 };", "let index = -1;", "for (const key in fields) {", "index++;", "if (index >= "+limit+") {", "break;", "}", "trace += key;")
		lines = append(lines, body...)
		lines = append(lines, "}")
	}
	return lines
}

func breadthConstructs() []construct {
	return []construct{
		{
			name: "switch",
			targets: []string{
				"internal/native/emit_statements.go:395-435 emitter.switchStatement (never entered)",
				"internal/lower/object.go:970 lowering.switchStatement",
			},
			variants: []string{"grouped", "exits", "strings", "nested", "throws"},
			write:    (*breadth).switchConstruct,
		},
		{
			name: "finally",
			targets: []string{
				"internal/lower/exceptions.go and internal/native/exceptions.go: finally crossing break, continue, return and throw",
				"internal/flow: a finally's edges out of a loop",
			},
			variants: loopShapes,
			write:    (*breadth).finallyConstruct,
		},
		{
			name: "map-foreach",
			targets: []string{
				"internal/native/emit_maps.go:27-74 emitter.mapForEach (never entered)",
				"internal/native/runtime/map_set.c:66-105 collection_next",
			},
			variants: []string{"map", "set", "mutate", "throws"},
			write:    (*breadth).forEachConstruct,
		},
		{
			name:     "group-by",
			targets:  []string{"internal/lower/library_map_set.go:302-437 lowering.libraryMapGroupBy (never entered)", "internal/lower/library_map_set.go:19-106 lowering.librarySetMethod"},
			variants: []string{"numbers", "strings", "set", "letters", "throws"},
			write:    (*breadth).groupByConstruct,
		},
		{
			name: "static",
			targets: []string{
				"internal/lower/class_static.go:48-212 lowering.staticInstance (never entered)",
				"internal/lower/class_static.go:309-436 lowering.staticInitializationReads (never entered)",
				"internal/lower/class_static.go:222-296 lowering.staticDeclaration",
				"internal/native/runtime/class_static.c",
			},
			variants: []string{"order", "inherit", "accessor"},
			write:    (*breadth).staticConstruct,
			optIn:    breadthStatics,
		},
		{
			name: "accessors",
			targets: []string{
				"internal/lower/class_accessors.go:217-346 lowering.finishAccessors",
				"internal/lower/class_accessors.go:141-202 lowering.accessorLiteral",
				"internal/lower/class_accessors.go:58-103 lowering.superAccessor",
				"internal/native/class_accessors.go:14-76 emitter.accessorDeclarations",
			},
			variants: []string{"class", "override", "literal", "throws"},
			write:    (*breadth).accessorConstruct,
		},
		{
			name:     "generics",
			targets:  []string{"internal/lower/generic.go:24-136 lowering.instantiateFunction (never entered)"},
			variants: []string{"values", "callbacks", "recursion"},
			write:    (*breadth).genericConstruct,
		},
		{
			name: "iterators",
			targets: []string{
				"internal/lower/iteration_consume.go:110-203 lowering.destructureIterator (never entered)",
				"internal/lower/iteration_consume.go:11-96 lowering.collectIteration (never entered)",
				"internal/lower/iteration.go:168-297 lowering.planIteration",
				"internal/lower/iteration.go:364-408 lowering.forOfUser",
				"internal/lower/iteration_origin.go:109-171 lowering.iterationOrigin",
				"internal/lower/collections.go:281-355 lowering.destructureFrom",
			},
			variants: []string{"array", "set", "map", "user", "user-break"},
			write:    (*breadth).iteratorConstruct,
		},
		{
			name: "for-in",
			targets: []string{
				"internal/lower/library_for_in.go:13-66 lowering.forIn (never entered)",
				"internal/lower/library_for_in.go:68-153 lowering.plainEnumerableObject (never entered)",
			},
			variants: []string{"exits", "closures", "outer"},
			write:    (*breadth).forInConstruct,
		},
		{
			name: "objects",
			targets: []string{
				"internal/lower/library_object.go:11-147 lowering.objectCall (never entered)",
				"internal/native/runtime/library_object.c:96-108 adamic_object_keys, 34-41 adamic_object_check_write",
				"internal/native/runtime/record.c:119-136 array_index (integer-like keys)",
			},
			variants:     []string{"numbers", "strings", "assign", "frozen"},
			write:        (*breadth).objectConstruct,
			variantOptIn: map[string]string{"frozen": breadthFrozen},
		},
		{
			name: "math",
			targets: []string{
				"internal/native/runtime/ieee754.c: kernel_rem_pio2, ieee754_rem_pio2, kernel_tan, expm1, log1p, atan2, exp, atan, log, asin, acos, cbrt, log2, log10 (never called)",
				"internal/lower/library_math_number.go:67-181 lowering.libraryMathNumberCall",
			},
			variants: []string{"trigonometry", "exponential", "inverse", "pairs"},
			write:    (*breadth).mathConstruct,
		},
		{
			name: "long-sort",
			targets: []string{
				"internal/native/runtime/sort.c:221-311 merge_low, 314-408 merge_high, 119-167 gallop_left, 170-218 gallop_right (never called)",
			},
			variants: []string{"ascending", "descending", "stable"},
			write:    (*breadth).sortConstruct,
		},
		{
			name: "json",
			targets: []string{
				"internal/lower/library_json_stringify.go:12-65 lowering.jsonCall, 72-154 lowering.jsonInput (never entered)",
				"internal/native/library_json_stringify.go emitter.jsonStringify, emitter.jsonSchema (never entered)",
				"internal/native/runtime/json_stringify.c:147-203 write_value (never called)",
			},
			variants: []string{"plain", "indented", "filtered", "values"},
			write:    (*breadth).jsonConstruct,
		},
		{
			name: "edges",
			targets: []string{
				"internal/native/runtime/string_repeat_impl.h:14-20 adamic_string_repeat's range error",
				"internal/native/runtime/number.c:167-175 adamic_number_to_fixed's large branch",
				"internal/native/runtime/radix.c:53-131 double_to_radix",
				"internal/native/runtime/parse.c:69-123 power_of_two",
				"internal/native/runtime/library_math_number.c:57-99 adamic_number_from_string",
			},
			variants: []string{"numbers", "strings"},
			write:    (*breadth).edgesConstruct,
		},
		{
			name: "input",
			targets: []string{
				"internal/lower/input.go:21-91 lowering.input",
				"internal/native/runtime/input.c:168-194 failure, 227-260 read_all, 262-297 adamic_read_text_file",
				"internal/native/runtime/directory.c:37-61 failure, 132-161 adamic_file_status",
			},
			variants: []string{"missing", "present"},
			write:    (*breadth).inputConstruct,
		},
	}
}

// distinct is count different numbers from 0 up to limit - 1, in a random order.
func (b *breadth) distinct(count, limit int) []int {
	order := b.random.Perm(limit)
	return order[:count]
}

func (b *breadth) switchConstruct(variant string) []string {
	name := b.name("Switch")
	modulus := b.between(4, 7)
	cases := b.distinct(4, modulus)
	inputs := make([]string, 0, 8)
	for range 6 + b.intn(4) {
		inputs = append(inputs, strconv.Itoa(b.between(-3, 20)))
	}
	call := "for (const input of [" + strings.Join(inputs, ", ") + "]) {"
	switch variant {
	case "grouped":
		// Empty cases that share a body, and default in the middle, before cases it doesn't catch.
		return []string{
			"function " + name + "(value: number): string {",
			"let out = `v${value}`;",
			fmt.Sprintf("switch (value %% %d) {", modulus),
			fmt.Sprintf("case %d:", cases[0]),
			fmt.Sprintf("case %d:", cases[1]),
			"out += ':low';",
			"break;",
			"default:",
			"out += ':other';",
			"break;",
			fmt.Sprintf("case %d:", cases[2]),
			"out += ':ret';",
			"return out;",
			fmt.Sprintf("case %d: {", cases[3]),
			"const inner = `${out}!`;",
			"out = inner.repeat(2);",
			"break;",
			"}",
			"}",
			"return out;",
			"}",
			"const " + name + "Seen: string[] = [];",
			call,
			name + "Seen.push(" + name + "(input));",
			"}",
			"console.log(" + name + "Seen.join(' '));",
		}
	case "exits":
		// continue, break and return from inside a switch inside a loop, in each loop shape.
		shape := b.pick(loopShapes...)
		body := []string{
			"const word = `w${index}`.repeat(1 + (index % 2));",
			fmt.Sprintf("switch (index %% %d) {", modulus),
			fmt.Sprintf("case %d:", cases[0]),
			"continue;",
			fmt.Sprintf("case %d:", cases[1]),
			"trace += `[${word}]`;",
			"break;",
			fmt.Sprintf("case %d:", cases[2]),
			fmt.Sprintf("if (word.length > %d) {", b.between(2, 4)),
			"return `${trace}|ret${index}`;",
			"}",
			"break;",
			"default:",
			"trace += word;",
			"}",
			"trace += ',';",
		}
		lines := []string{"function " + name + "(limit: number): string {", "let trace = '';"}
		lines = append(lines, b.loop(shape, "limit", body)...)
		lines = append(lines, "return trace;", "}")
		return append(lines, fmt.Sprintf("console.log(`${%s(%d)} ${%s(%d)} ${%s(0)}`);", name, b.between(1, 9), name, b.between(5, 14), name))
	case "strings":
		words := []string{"'Sun'", "'Rain'", "'Snow'", "'Fog'", "'Wind'"}
		return []string{
			"type " + name + "Weather = 'Sun' | 'Rain' | 'Snow' | 'Fog' | 'Wind';",
			"function " + name + "(weather: " + name + "Weather, built: string): string {",
			"switch (weather) {",
			"case " + []string{"'Sun'", "'Fog'", "'Wind'"}[cases[0]%3] + ":",
			"return `hat:${built}`;",
			"case 'Rain':",
			"case 'Snow':",
			"return built.toUpperCase();",
			"default:",
			"return `${weather}?`;",
			"}",
			"}",
			"function " + name + "Text(value: string): number {",
			"switch (value) {",
			"case 'ab':",
			"return 1;",
			"case 'aa':",
			"return 2;",
			"default:",
			"return value.length;",
			"}",
			"}",
			"const " + name + "Forecast: " + name + "Weather[] = [" + strings.Join([]string{words[b.intn(5)], words[b.intn(5)], words[b.intn(5)], words[b.intn(5)]}, ", ") + "];",
			"for (const weather of " + name + "Forecast) {",
			"console.log(`${" + name + "(weather, `x${weather.length}`)} ${" + name + "Text(weather.toLowerCase().slice(0, 2))} ${" + name + "Text('a'.repeat(2))} ${" + name + "Text(`a${'b'}`)}`);",
			"}",
		}
	case "nested":
		// A switch in a switch, each in a loop, and a continue from the inner one.
		return []string{
			"function " + name + "(limit: number): string {",
			"let trace = '';",
			"let round = 0;",
			"while (round < limit) {",
			"round++;",
			fmt.Sprintf("switch (round %% %d) {", modulus),
			fmt.Sprintf("case %d:", cases[0]),
			"for (const letter of `ab${round}`) {",
			"switch (letter) {",
			"case 'a':",
			"continue;",
			"case 'b':",
			"trace += 'B';",
			"break;",
			"default:",
			"trace += letter;",
			"}",
			"}",
			"break;",
			fmt.Sprintf("case %d:", cases[1]),
			"continue;",
			"default: {",
			"const doubled = `${round}`.repeat(2);",
			"trace += doubled;",
			"}",
			"}",
			"trace += ';';",
			"}",
			"return trace;",
			"}",
			fmt.Sprintf("console.log(%s(%d));", name, b.between(3, 12)),
		}
	}
	// throws: a case that throws, out of a loop, past a finally.
	return []string{
		"function " + name + "(value: number): string {",
		"let trace = `s${value}`;",
		"try {",
		fmt.Sprintf("switch (value %% %d) {", modulus),
		fmt.Sprintf("case %d:", cases[0]),
		"throw new Error(`thrown ${trace}`);",
		fmt.Sprintf("case %d:", cases[1]),
		"trace += 'one';",
		"break;",
		"default:",
		"trace += 'default';",
		"}",
		"} finally {",
		"trace += '|finally';",
		"}",
		"return trace;",
		"}",
		call,
		"try {",
		"console.log(" + name + "(input));",
		"} catch (error) {",
		"console.log(error instanceof Error ? error.message : 'other');",
		"}",
		"}",
	}
}

func (b *breadth) finallyConstruct(shape string) []string {
	name := b.name("Finally")
	limit := 8
	exits := b.distinct(4, limit)
	body := []string{
		"try {",
		fmt.Sprintf("if (index === %d) {", exits[0]),
		"break;",
		"}",
		fmt.Sprintf("if (index === %d) {", exits[1]),
		"continue;",
		"}",
		fmt.Sprintf("if (index === %d) {", exits[2]),
		"return `${trace}|return${index}`;",
		"}",
		fmt.Sprintf("if (index === %d) {", exits[3]),
		"throw new Error(`${trace}|throw${index}`);",
		"}",
		"trace += `${index}`;",
	}
	if b.intn(2) == 0 {
		// A try in the try, its finally run on the same way out.
		body = append(body, "try {", fmt.Sprintf("if (index === %d) {", b.intn(limit)), "continue;", "}", "trace += 'i';", "} finally {", "trace += 'F';", "}")
	}
	if b.intn(2) == 0 {
		body = append(body, "} catch (error) {", "trace += error instanceof Error ? 'c' : '?';", "throw error;")
	}
	body = append(body, "} finally {", "trace += 'f';")
	if b.intn(3) == 0 {
		// A finally that leaves on its own, overriding what the try was doing.
		body = append(body, fmt.Sprintf("if (index === %d) {", b.intn(limit)), "break;", "}")
	}
	body = append(body, "}")
	lines := []string{"function " + name + "(limit: number): string {", "let trace = '';"}
	lines = append(lines, b.loop(shape, "limit", body)...)
	lines = append(lines, "return trace;", "}")
	lines = append(lines, fmt.Sprintf("for (const limit of [%d, %d, %d]) {", b.intn(4), b.between(3, 8), limit))
	lines = append(lines, caught("console.log("+name+"(limit));")...)
	return append(lines, "}")
}

func (b *breadth) forEachConstruct(variant string) []string {
	name := b.name("Each")
	count := b.between(3, 9)
	switch variant {
	case "set":
		return []string{
			"const " + name + " = new Set<number>();",
			fmt.Sprintf("for (let fill = 0; fill < %d; fill++) {", count),
			fmt.Sprintf("%s.add((fill * %d) %% %d);", name, b.between(2, 7), b.between(3, 11)),
			"}",
			"let " + name + "Sum = 0;",
			name + ".forEach((value: number, again: number, set: Set<number>) => {",
			name + "Sum += value + again;",
			fmt.Sprintf("if (value %% 2 === 0 && set.size < %d) {", count+4),
			"set.add(value + 101);",
			"}",
			"console.log(`${value}`);",
			"});",
			"console.log(`${" + name + "Sum} ${" + name + ".size}`);",
		}
	case "throws":
		return []string{
			"const " + name + " = new Map<string, number>();",
			fmt.Sprintf("for (let fill = 0; fill < %d; fill++) {", count),
			name + ".set(`k${fill}`, fill);",
			"}",
			"let " + name + "Seen = '';",
			"try {",
			name + ".forEach((value: number, key: string) => {",
			name + "Seen += key;",
			fmt.Sprintf("if (value === %d) {", b.intn(count)),
			"throw new Error(`stopped at ${key}`);",
			"}",
			"});",
			"} catch (error) {",
			"console.log(error instanceof Error ? error.message : 'other');",
			"}",
			"console.log(" + name + "Seen);",
		}
	}
	lines := []string{
		"const " + name + " = new Map<string, number>();",
		fmt.Sprintf("for (let fill = 0; fill < %d; fill++) {", count),
		fmt.Sprintf("%s.set(`k${fill %% %d}`, fill * %d);", name, b.between(2, count), b.between(1, 5)),
		"}",
		"let " + name + "Sum = 0;",
		name + ".forEach((value: number, key: string, map: Map<string, number>) => {",
		name + "Sum += value;",
	}
	if variant == "mutate" {
		// Keys set during the walk are visited; a key deleted before it's reached is not.
		lines = append(lines,
			fmt.Sprintf("if (value %% 3 === 0 && map.size < %d) {", count+5),
			"map.set(`${key}+`, value + 1);",
			"}",
			"if (key === 'k1') {",
			"map.delete('k2');",
			"}",
		)
	}
	return append(lines,
		"console.log(`${key}=${value} ${map.size}`);",
		"});",
		"console.log(`${"+name+"Sum} ${"+name+".size}`);",
	)
}

func (b *breadth) groupByConstruct(variant string) []string {
	name := b.name("Groups")
	modulus := b.between(2, 5)
	var values []string
	for range b.between(3, 12) {
		values = append(values, strconv.Itoa(b.between(-5, 30)))
	}
	show := []string{
		"for (const [key, bucket] of " + name + ") {",
		"console.log(`${key}: ${bucket.join(',')}`);",
		"}",
		"console.log(`${" + name + ".size}`);",
	}
	switch variant {
	case "numbers":
		return append([]string{
			"const " + name + "Values: number[] = [" + strings.Join(values, ", ") + "];",
			"const " + name + " = Map.groupBy(" + name + "Values, (value: number, index: number): number => {",
			fmt.Sprintf("if (index === 0 && %sValues.length < %d) {", name, len(values)+2),
			name + "Values.push(index + 7);",
			"}",
			fmt.Sprintf("return value %% %d;", modulus),
			"});",
		}, show...)
	case "strings":
		return append([]string{
			"const " + name + "Words: string[] = [" + "'a'.repeat(2), 'bb', `c${" + strconv.Itoa(modulus) + "}`, 'dd', 'e'" + "];",
			"const " + name + " = Map.groupBy(" + name + "Words, (word: string, index: number): string => index % " + strconv.Itoa(modulus) + " === 0 ? word.slice(0, 1) : `n${word.length}`);",
		}, show...)
	case "set":
		return append([]string{
			"const " + name + "Live = new Set<number>([" + strings.Join(values, ", ") + "]);",
			"const " + name + " = Map.groupBy(" + name + "Live, (value: number, index: number): number => {",
			"if (index === 0) {",
			name + "Live.add(99);",
			"}",
			fmt.Sprintf("return value %% %d;", modulus),
			"});",
		}, show...)
	case "letters":
		return append([]string{
			"const " + name + " = Map.groupBy(" + b.pick("'abcabc'", "'a🌍aé'", "`x${'yz'.repeat(3)}`") + ", (letter: string, index: number): string => index % 2 === 0 ? letter : 'odd');",
		}, show...)
	}
	return []string{
		"try {",
		"const " + name + " = Map.groupBy([" + strings.Join(values, ", ") + "], (value: number): string => {",
		fmt.Sprintf("if (value === %s) {", values[b.intn(len(values))]),
		"throw new Error(`group failed at ${value}`);",
		"}",
		"return `${value % 3}`;",
		"});",
		"console.log(`${" + name + ".size}`);",
		"} catch (error) {",
		"console.log(error instanceof Error ? error.message : 'other');",
		"}",
	}
}

func (b *breadth) staticConstruct(variant string) []string {
	name := b.name("Static")
	lines := []string{
		"let " + name + "Trace = '';",
		"function " + name + "Mark(label: string): string {",
		name + "Trace += label;",
		"return `${label}${" + name + "Trace.length}`;",
		"}",
		"class " + name + "Base {",
		"static first = " + name + "Mark('a');",
		"static {",
		name + "Trace += 'b';",
		"this.first += 'x';",
		"}",
		"static second = " + name + "Mark('c');",
		fmt.Sprintf("static count = %d;", b.between(0, 9)),
		"static bump(step: number): number {",
		"this.count += step;",
		"return this.count;",
		"}",
	}
	if variant == "accessor" {
		lines = append(lines,
			"static get total(): number {",
			name+"Trace += 'g';",
			"return this.count;",
			"}",
			"static set total(value: number) {",
			name+"Trace += 's';",
			"this.count = value;",
			"}",
		)
	}
	lines = append(lines, "}")
	if variant != "order" {
		lines = append(lines,
			"class "+name+"Child extends "+name+"Base {",
			"static third = "+name+"Mark('d');",
			"static {",
			name+"Trace += 'e';",
			"this.third += this.first;",
			"}",
			"static override bump(step: number): number {",
			"return super.bump(step * 2) + 1;",
			"}",
			"}",
		)
	}
	lines = append(lines, "console.log("+name+"Trace);")
	switch variant {
	case "order":
		lines = append(lines, fmt.Sprintf("console.log(`${%sBase.first} ${%sBase.second} ${%sBase.bump(%d)} ${%sBase.count}`);", name, name, name, b.between(1, 5), name))
	case "inherit":
		lines = append(lines, fmt.Sprintf("console.log(`${%sChild.third} ${%sChild.bump(%d)} ${%sBase.bump(1)} ${%sBase.count} ${%sChild.count}`);", name, name, b.between(1, 5), name, name, name))
		lines = append(lines, name+"Base.first = "+name+"Mark('z');", "console.log(`${"+name+"Child.first} ${"+name+"Trace}`);")
	case "accessor":
		lines = append(lines, fmt.Sprintf("%sChild.total += %d;", name, b.between(1, 9)), "console.log(`${"+name+"Base.total} ${"+name+"Child.total} ${"+name+"Trace}`);")
	}
	return lines
}

func (b *breadth) accessorConstruct(variant string) []string {
	name := b.name("Accessor")
	switch variant {
	case "literal":
		return []string{
			fmt.Sprintf("let %sCaptured = %d;", name, b.between(0, 9)),
			"let " + name + "Trace = '';",
			"const " + name + " = {",
			fmt.Sprintf("stored: %d,", b.between(0, 9)),
			"get gauge(): number {",
			name + "Trace += 'r';",
			"return this.stored + " + name + "Captured;",
			"},",
			"set gauge(next: number) {",
			name + "Trace += 'w';",
			"this.stored = next;",
			"},",
			"label: `L${" + name + "Captured}`,",
			"};",
			fmt.Sprintf("%s.gauge += %d;", name, b.between(1, 9)),
			fmt.Sprintf("%sCaptured = %d;", name, b.between(0, 9)),
			"console.log(`${" + name + ".gauge} ${" + name + ".stored} ${" + name + ".label} ${" + name + "Trace} ${Object.keys(" + name + ").join(',')}`);",
		}
	case "throws":
		return append([]string{
			"class " + name + " {",
			"#count = 0;",
			"get spelled(): string {",
			"this.#count++;",
			"if (this.#count > 1) {",
			"throw new Error(`getter ${this.#count}`);",
			"}",
			"return `read${this.#count}`;",
			"}",
			"set spelled(value: string) {",
			"throw new Error(`setter ${value}`);",
			"}",
			"}",
			"const " + name + "Value = new " + name + "();",
			"console.log(" + name + "Value.spelled);",
		}, append(caught("console.log("+name+"Value.spelled);"), caught(name+"Value.spelled = `input${"+strconv.Itoa(b.intn(9))+"}`;")...)...)
	}
	lines := []string{
		"let " + name + "Trace = '';",
		"class " + name + " {",
		fmt.Sprintf("#value = %d;", b.between(0, 20)),
		"get reading(): number {",
		name + "Trace += 'g';",
		"return this.#value;",
		"}",
		"set reading(next: number) {",
		name + "Trace += 's';",
		"this.#value = next;",
		"}",
		"}",
	}
	if variant == "override" {
		lines = append(lines,
			"class "+name+"Double extends "+name+" {",
			"get reading(): number {",
			name+"Trace += 'G';",
			"return super.reading * 2;",
			"}",
			"set reading(next: number) {",
			name+"Trace += 'S';",
			"super.reading = next / 2;",
			"}",
			"}",
		)
	}
	lines = append(lines,
		"function "+name+"Adjust(meter: "+name+", step: number): number {",
		"meter.reading += step;",
		"meter.reading++;",
		"return meter.reading;",
		"}",
	)
	instance := "new " + name + "()"
	if variant == "override" {
		instance = "new " + name + "Double()"
	}
	return append(lines, fmt.Sprintf("console.log(`${%sAdjust(%s, %d)} ${%sAdjust(new %s(), %d)} ${%sTrace}`);", name, instance, b.between(1, 9), name, name, b.between(1, 9), name))
}

func (b *breadth) genericConstruct(variant string) []string {
	name := b.name("Generic")
	switch variant {
	case "callbacks":
		return []string{
			"function " + name + "Map<From, To>(items: readonly From[], change: (item: From, index: number) => To): To[] {",
			"const changed: To[] = [];",
			"let index = 0;",
			"for (const item of items) {",
			"changed.push(change(item, index));",
			"index++;",
			"}",
			"return changed;",
			"}",
			"function " + name + "Fold<Item, Total>(items: readonly Item[], start: Total, step: (total: Total, item: Item) => Total): Total {",
			"let total = start;",
			"for (const item of items) {",
			"total = step(total, item);",
			"}",
			"return total;",
			"}",
			fmt.Sprintf("console.log(%sMap([%d, %d, %d], (value, index) => `#${value * index}`).join(' '));", name, b.intn(9), b.intn(9), b.intn(9)),
			"console.log(" + name + "Map(['a', 'bb', 'ccc'], (word) => word.length > 1).join(' '));",
			"console.log(`${" + name + "Fold(['x', 'yy'], 0, (total, word) => total + word.length)} ${" + name + "Fold([1, 2, 3], '', (total, value) => `${total}${value}`)}`);",
		}
	case "recursion":
		return []string{
			"function " + name + "Depth<Item>(item: Item, depth: number): number {",
			"return depth === 0 ? 0 : 1 + " + name + "Depth([item], depth - 1);",
			"}",
			"function " + name + "Repeat<Item>(item: Item, count: number, into: Item[]): Item[] {",
			"if (count === 0) {",
			"return into;",
			"}",
			"into.push(item);",
			"return " + name + "Repeat(item, count - 1, into);",
			"}",
			fmt.Sprintf("console.log(`${%sDepth(1, %d)} ${%sDepth('s', 0)} ${%sRepeat('ab', %d, []).join('-')} ${%sRepeat(%d, 2, [7]).join('-')}`);", name, b.between(1, 5), name, name, b.between(0, 4), name, b.intn(9)),
		}
	}
	return []string{
		"interface " + name + "Point {",
		"readonly x: number;",
		"readonly label: string;",
		"}",
		"function " + name + "Identity<Item>(item: Item): Item {",
		"return item;",
		"}",
		"function " + name + "First<Item>(items: readonly Item[], fallback: Item): Item {",
		"return items[0] ?? fallback;",
		"}",
		"function " + name + "Pair<Left, Right>(left: Left, right: Right): readonly [Left, Right] {",
		"return [left, right];",
		"}",
		fmt.Sprintf("console.log(`${%sIdentity(%d)} ${%sIdentity('three')} ${%sIdentity(true)} ${%sIdentity<number>(4) + 1}`);", name, b.intn(9), name, name, name),
		fmt.Sprintf("console.log(`${%sFirst([%d, 2], 0)} ${%sFirst<string>([], 'none')} ${%sFirst<%sPoint>([], { x: 9, label: 'p' }).label}`);", name, b.intn(9), name, name, name),
		fmt.Sprintf("const [%sLeft, %sRight] = %sPair('left', %d);", name, name, name, b.intn(9)),
		"console.log(`${" + name + "Left} ${" + name + "Right} ${" + name + "Pair(1, 'x')[1]}`);",
	}
}

func (b *breadth) iteratorConstruct(variant string) []string {
	name := b.name("Iterator")
	var values []string
	for range b.between(1, 5) {
		values = append(values, strconv.Itoa(b.between(-3, 20)))
	}
	list := "[" + strings.Join(values, ", ") + "]"
	switch variant {
	case "array":
		// Stage 0 destructures arrays only from tuples.
		return []string{
			fmt.Sprintf("const %sValues: [number, string, boolean] = [%d, `t${%d}`, %d > 3];", name, b.intn(9), b.intn(9), b.intn(9)),
			"const [" + name + "A, , " + name + "C] = " + name + "Values;",
			"const [, " + name + "B] = " + name + "Values;",
			"console.log(`${" + name + "A} ${" + name + "B} ${" + name + "C}`);",
		}
	case "set":
		// Stage 0 destructures only user iterators, and gives Array.from only { length }; a library
		// iterator is spread.
		return []string{
			"const " + name + "Values = new Set<number>(" + list + ");",
			"const " + name + "All = [..." + name + "Values.values()];",
			fmt.Sprintf("const %sDoubled = [...%sValues.keys()].map((value: number, index: number): number => value * %d + index);", name, name, b.between(1, 4)),
			"console.log(`${" + name + "All.join(',')} ${" + name + "Doubled.join(',')}`);",
		}
	case "map":
		return []string{
			"const " + name + "Values = new Map<string, number>();",
			"for (const value of " + list + ") {",
			name + "Values.set(`k${value}`, value);",
			"}",
			"const " + name + "Keys = [..." + name + "Values.keys()];",
			"const " + name + "Lengths = [..." + name + "Values.values()].map((value: number): string => `${value}!`);",
			"console.log(`${" + name + "Keys.join(',')} ${" + name + "Lengths.join(',')}`);",
		}
	}
	// A user iterator: next and return print, so when each is called shows.
	limit := b.between(0, 5)
	lines := []string{
		"interface " + name + "Step {",
		"readonly value: number;",
		"readonly done: boolean;",
		"}",
		"class " + name + " {",
		"index = 0;",
		"readonly limit: number;",
		"constructor(limit: number) {",
		"this.limit = limit;",
		"}",
		"[Symbol.iterator](): " + name + " {",
		"return this;",
		"}",
		"next(): " + name + "Step {",
		"console.log(`next ${this.index}`);",
		"if (this.index >= this.limit) {",
		"return { value: -1, done: true };",
		"}",
		"const value = this.index * 10;",
		"this.index++;",
		"return { value, done: false };",
		"}",
		"return(): " + name + "Step {",
		"console.log(`return ${this.index}`);",
		"return { value: -2, done: true };",
		"}",
		"}",
	}
	if variant == "user-break" {
		return append(lines,
			"let "+name+"Sum = 0;",
			fmt.Sprintf("for (const value of new %s(%d)) {", name, limit),
			name+"Sum += value;",
			fmt.Sprintf("if (value >= %d) {", b.between(0, 4)*10),
			"break;",
			"}",
			"}",
			"console.log(`${"+name+"Sum}`);",
		)
	}
	return append(lines,
		fmt.Sprintf("const [%sA = 7, %sB = 8] = new %s(%d);", name, name, name, limit),
		"console.log(`${"+name+"A} ${"+name+"B}`);",
	)
}

// objectKeys is a pool of keys, integer-like and not, in Object's own order rules.
var objectKeys = []string{"z", "alpha", "b", "'10'", "'2'", "'0'", "'01'", "'-0'", "'4294967294'", "'4294967295'", "'1e0'", "beta"}

func (b *breadth) objectLiteral(count int, value func(index int) string) string {
	var fields []string
	for index, key := range b.distinct(count, len(objectKeys)) {
		fields = append(fields, objectKeys[key]+": "+value(index))
	}
	return "{ " + strings.Join(fields, ", ") + " }"
}

func (b *breadth) forInConstruct(variant string) []string {
	name := b.name("ForIn")
	count := b.between(3, 8)
	literal := b.objectLiteral(count, func(index int) string { return strconv.Itoa(index * 3) })
	switch variant {
	case "closures":
		return []string{
			"const " + name + " = " + literal + ";",
			"const " + name + "Calls: (() => string)[] = [];",
			"for (let key in " + name + ") {",
			name + "Calls.push(() => `${key}!`);",
			"key = `${key}?`;",
			"}",
			"console.log(" + name + "Calls.map((call) => call()).join('|'));",
		}
	case "outer":
		return []string{
			"const " + name + " = " + literal + ";",
			"let " + name + "Last = 'untouched';",
			"for (" + name + "Last in " + name + ") {",
			"if (" + name + "Last.length > 3) {",
			"break;",
			"}",
			"}",
			"console.log(" + name + "Last);",
		}
	}
	return []string{
		"const " + name + " = " + literal + ";",
		"const " + name + "Keys: string[] = [];",
		"for (const key in " + name + ") {",
		"if (key === '2' || key === 'b') {",
		"continue;",
		"}",
		name + "Keys.push(key);",
		"if (key === '10') {",
		"break;",
		"}",
		"}",
		"console.log(" + name + "Keys.join('|'));",
	}
}

func (b *breadth) objectConstruct(variant string) []string {
	name := b.name("Object")
	count := b.between(2, 8)
	switch variant {
	case "strings":
		literal := b.objectLiteral(count, func(index int) string { return fmt.Sprintf("'v'.repeat(%d)", index+1) })
		return []string{
			"const " + name + " = " + literal + ";",
			"console.log(Object.keys(" + name + ").join('|'));",
			"console.log(Object.values(" + name + ").join('|'));",
			"for (const [key, value] of Object.entries(" + name + ")) {",
			"console.log(`${key}:${value}`);",
			"}",
		}
	case "assign":
		return []string{
			fmt.Sprintf("const %s = { a: %d, b: 'old'.repeat(2), c: false };", name, b.intn(9)),
			fmt.Sprintf("const %sFirst = { b: `new${%d}`, a: %d };", name, b.intn(9), b.intn(9)),
			fmt.Sprintf("const %sLast = { a: %d, c: true };", name, b.intn(9)),
			"const " + name + "Result = Object.assign(" + name + ", " + name + "First, " + name + "Last);",
			"console.log(`${Object.is(" + name + "Result, " + name + ")} ${" + name + ".a} ${" + name + ".b} ${" + name + ".c} ${Object.keys(" + name + ").join('|')}`);",
			"const " + name + "Itself = Object.assign(" + name + ", " + name + ");",
			"console.log(`${Object.is(" + name + "Itself, " + name + ")} ${" + name + ".b}`);",
		}
	case "frozen":
		return append([]string{
			fmt.Sprintf("const %s = { b: %d, '2': %d, a: %d };", name, b.intn(9), b.intn(9), b.intn(9)),
			"console.log(`${Object.isFrozen(" + name + ")}`);",
			"Object.freeze(" + name + ");",
			"console.log(`${Object.isFrozen(" + name + ")}`);",
		}, caught(fmt.Sprintf("Object.assign(%s, { a: %d, '2': %d });", name, b.intn(9), b.intn(9)), "console.log(`${"+name+".a}`);")...)
	}
	literal := b.objectLiteral(count, func(index int) string { return strconv.Itoa(b.between(-5, 50)) })
	return []string{
		"const " + name + " = " + literal + ";",
		"console.log(Object.keys(" + name + ").join('|'));",
		"console.log(Object.values(" + name + ").join('|'));",
		"for (const [key, value] of Object.entries(" + name + ")) {",
		"console.log(`${key}:${value}`);",
		"}",
		"const " + name + "Copy = { ..." + name + " };",
		"console.log(Object.keys(" + name + "Copy).join('|'));",
	}
}

// mathInputs are where the transcendental functions change behavior: zeros, tiny and huge values,
// the infinities and NaN, multiples of pi that need a careful reduction, and the overflow edges.
var mathInputs = []string{"0", "-0", "1", "-1", "0.5", "-0.5", "1e-300", "-1e-300", "1e300", "NaN", "Infinity", "-Infinity", "Math.PI / 2", "Math.PI", "1e10", "3", "710", "-745", "1e22", "0.7", "2.5", "-2", "1e-8"}

func (b *breadth) mathConstruct(variant string) []string {
	name := b.name("Math")
	families := map[string][]string{
		"trigonometry": {"sin", "cos", "tan"},
		"exponential":  {"exp", "expm1", "log", "log1p", "log2", "log10", "cbrt", "sinh", "cosh", "tanh"},
		"inverse":      {"asin", "acos", "atan", "asinh", "acosh"},
	}
	inputs := b.distinct(8, len(mathInputs))
	var chosen []string
	for _, index := range inputs {
		chosen = append(chosen, mathInputs[index])
	}
	lines := []string{"const " + name + ": number[] = [" + strings.Join(chosen, ", ") + "];"}
	if variant == "pairs" {
		return append(lines,
			"for (const left of "+name+") {",
			"const row: string[] = [];",
			"for (const right of "+name+") {",
			"row.push(`${Math.atan2(left, right)}/${Math.pow(left, right)}/${Math.hypot(left, right)}`);",
			"}",
			"console.log(row.join(' '));",
			"}",
		)
	}
	functions := families[variant]
	var shown []string
	for _, function := range functions {
		shown = append(shown, "${Math."+function+"(value)}")
	}
	return append(lines,
		"for (const value of "+name+") {",
		"console.log(`"+strings.Join(shown, " ")+"`);",
		"}",
	)
}

func (b *breadth) sortConstruct(variant string) []string {
	name := b.name("Sort")
	// Two long runs with overlapping values, one shorter than the other either way, so the merge
	// takes both directions and gallops when one run wins many times in a row.
	first, second := b.between(40, 220), b.between(40, 220)
	start, step := b.between(0, 50), b.between(1, 4)
	otherStart, otherStep := b.between(0, 120), b.between(1, 3)
	fill := []string{
		"const " + name + ": number[] = [];",
		fmt.Sprintf("for (let index = 0; index < %d; index++) {", first),
		fmt.Sprintf("%s.push(%d + index * %d);", name, start, step),
		"}",
		fmt.Sprintf("for (let index = 0; index < %d; index++) {", second),
		fmt.Sprintf("%s.push(%d + index * %d);", name, otherStart, otherStep),
		"}",
	}
	if b.intn(2) == 0 {
		fill = append(fill, fmt.Sprintf("for (let index = 0; index < %d; index++) {", b.between(10, 80)), fmt.Sprintf("%s.push(%d - index);", name, b.between(100, 300)), "}")
	}
	check := []string{
		"let " + name + "Check = 0;",
		"for (const value of " + name + ") {",
		name + "Check = (" + name + "Check * 31 + value) % 1000003;",
		"}",
		"console.log(`${" + name + ".length} ${" + name + "Check} ${" + name + ".slice(0, 5).join(',')} ${" + name + ".slice(-5).join(',')}`);",
	}
	switch variant {
	case "descending":
		fill = append(fill, name+".sort((left, right) => right - left);")
	case "stable":
		// Equal keys keep their order: ids sorted by a key with many ties.
		return append(fill,
			"const "+name+"Records = "+name+".map((value, index) => ({ key: value % "+strconv.Itoa(b.between(3, 17))+", id: index }));",
			name+"Records.sort((left, right) => left.key - right.key);",
			"let "+name+"Check = 0;",
			"for (const record of "+name+"Records) {",
			name+"Check = ("+name+"Check * 31 + record.id) % 1000003;",
			"}",
			"console.log(`${"+name+"Records.length} ${"+name+"Check}`);",
		)
	default:
		fill = append(fill, name+".sort((left, right) => left - right);")
	}
	return append(fill, check...)
}

func (b *breadth) jsonConstruct(variant string) []string {
	name := b.name("Json")
	value := fmt.Sprintf("{ b: %d, a: [true, false], inner: { z: %d, a: 'q\"\\\\n'.repeat(%d) }, none: null, '2': 'two', s: `t${%d}\\u0001`, n: -0, big: 1e21, small: 1e-7, nan: NaN, nested: { list: [1, 2.5, -3] }, ignored: undefined }",
		b.intn(9), b.intn(9), b.between(1, 3), b.intn(9))
	// Stage 0 stringifies an object written at the call, not one held in a variable.
	switch variant {
	case "indented":
		return []string{fmt.Sprintf("console.log(JSON.stringify(%s, undefined, %s) ?? '<undefined>');", value, b.pick("2", "0", "'->'", "10", "'\\t'"))}
	case "filtered":
		return []string{"console.log(JSON.stringify(" + value + ", ['a', 'b', '2', 'inner', 'z', 'missing']) ?? '<undefined>');"}
	case "values":
		return []string{
			fmt.Sprintf("console.log(`${JSON.stringify(%d)} ${JSON.stringify('x\\u2028y')} ${JSON.stringify([NaN, -0, Infinity, 1e-7, 123456789012345680000])} ${JSON.stringify([[1], [], [2, 3]])} ${JSON.stringify(true)}`);", b.intn(99)),
		}
	}
	return []string{"console.log(JSON.stringify(" + value + ") ?? '<undefined>');", "console.log(`" + name + " ${JSON.stringify([1, 2]) ?? ''}`);"}
}

func (b *breadth) edgesConstruct(variant string) []string {
	if variant == "strings" {
		// Not repeat(-1): its RangeError is a panic natively, which stage 0 won't put in a try yet.
		return []string{
			fmt.Sprintf("console.log(`${'x'.repeat(%d).padStart(%d, 'ab')}|${'yz'.padEnd(%d)}|`);", b.intn(4), b.between(0, 9), b.between(0, 6)),
			fmt.Sprintf("console.log(`${String.fromCharCode(%d, %d)} ${'%s'.codePointAt(%d) ?? -1}`);", b.between(0xd800, 0xdbff), b.between(0xdc00, 0xdfff), b.pick("a🌍", "é", "z"), b.intn(3)),
		}
	}
	value := b.pick("1e21", "-1e21", "123456789012345680000", "1.5e300", "0.000001", "-0")
	radix := b.between(2, 36)
	return []string{
		fmt.Sprintf("console.log(`${(%s).toFixed(%d)} ${(%d.25).toString(%d)} ${(-%d.5e-3).toString(%d)} ${(2 ** %d).toString(%s)}`);", value, b.intn(5), b.intn(99), radix, b.intn(99), radix, b.between(50, 70), b.pick("2", "8", "16", "32")),
		fmt.Sprintf("console.log(`${Number('%s')} ${Number.parseFloat('%s')} ${Number.parseInt('%s', %s)}`);", b.pick("0x1f", "0b101", "0o17", " 12.5e3 ", "-Infinity", "1_0", ""), b.pick("1e-400", "4.9e-324", "1.7976931348623157e308", ".5", "-.5e2x"), b.pick("ff", "777", "zz", "1e3", "-12"), b.pick("2", "8", "16", "36")),
	}
}

func (b *breadth) inputConstruct(variant string) []string {
	name := b.name("Input")
	path := "'program.a'"
	if variant == "missing" {
		path = "'" + name + "-missing.txt'"
	}
	return []string{
		"import { fileStatus as " + name + "Status, readTextFile as " + name + "Read } from 'adamic';",
		"const " + name + " = " + name + "Read(" + path + ");",
		"if (" + name + ".kind === 'Ok') {",
		"console.log(`read ${" + name + ".text.length}`);",
		"} else {",
		"console.log(`error: ${" + name + ".message}`);",
		"}",
		"const " + name + "Info = " + name + "Status(" + path + ");",
		"if (" + name + "Info.kind === 'Ok') {",
		"console.log(`${" + name + "Info.type} ${" + name + "Info.size} ${" + name + "Info.symbolicLink}`);",
		"} else {",
		"console.log(`status: ${" + name + "Info.message}`);",
		"}",
	}
}

// GenerateWeighted makes the program a seed names with the breadth scene weighted by weights instead
// of the checked-in table, and says what the scene chose: each construct, and each construct/variant,
// sorted. cmd/adamic-steer generates with it.
func GenerateWeighted(seed uint64, without []string, with []string, weights map[string]float64) (*Program, []string) {
	generator := newGenerator(seed, without, with)
	generator.weights = weights
	program := generator.program()
	chosen := slices.Clone(generator.breadthChosen)
	sort.Strings(chosen)
	return program, chosen
}
