package fuzz

import (
	"strconv"
	"strings"
)

// parallel-refuse marks a program the compiler must reject. The runner treats a matching
// refusal as success and anything else (acceptance, or a message that misses the path) as a finding.
const parallelRefusePrefix = "// parallel-refuse:"

// parallelRefusal is the path a generated program says the compiler must name, or "" when the
// program is meant to run.
func parallelRefusal(source string) string {
	for _, line := range strings.Split(source, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, parallelRefusePrefix) {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, parallelRefusePrefix))
		}
	}
	return ""
}

// parallelBaseText is long, starts with non-ASCII (including an emoji, a surrogate pair in
// UTF-16), and is only sliced, so every element is a view of the same text.
func parallelBaseText() string {
	return "世界éİß🌍🙂abcdefghijklmnopqrstuvwxyz0123456789-" +
		"hello-world-padding-that-is-long-enough-to-slice-into-many-different-views-" +
		"without-leaving-the-string-and-still-carry-the-non-ascii-prefix-into-short-slices"
}

// parallelSection is the program's parallelMap calls, or a single proof the compiler must refuse.
// One program in five is a refusal. The rest run at least one accepted call, sometimes two, so a
// seed range covers every shape: numbers, long sliced strings, readonly records, nested readonly
// arrays, a Map made and dropped inside the work, and a nested parallelMap.
func (g *generator) parallelSection() ([]*Statement, string) {
	if !g.allowed("parallel") {
		return nil, ""
	}
	if g.random.IntN(5) == 0 {
		switch g.random.IntN(3) {
		case 0:
			return g.refuseCapturedLet()
		case 1:
			return g.refuseMutableArray()
		default:
			return g.refuseGlobalWrite()
		}
	}
	statements := []*Statement{}
	count := 1
	if g.chance(1, 3) {
		count = 2
	}
	for range count {
		statements = append(statements, g.parallelAccepted())
	}
	return statements, ""
}

func (g *generator) parallelAccepted() *Statement {
	switch g.pick("numbers", "strings", "strings", "records", "nested", "fresh-map", "nested-call") {
	case "strings":
		return g.parallelStrings()
	case "records":
		return g.parallelRecords()
	case "nested":
		return g.parallelNested()
	case "fresh-map":
		return g.parallelFreshMap()
	case "nested-call":
		return g.parallelNestedCall()
	default:
		return g.parallelNumbers()
	}
}

// parallelCount is past the runtime's grain, so a second worker can take part of the range.
func (g *generator) parallelCount() string {
	return strconv.Itoa(320 + g.random.IntN(193))
}

func (g *generator) parallelNumbers() *Statement {
	index, prepared, items := g.name("index"), g.name("prepared"), g.name("items")
	lookup, work, result := g.name("lookup"), g.name("work"), g.name("result")
	count := g.parallelCount()
	body := []*Statement{
		statement("// parallel-shape: numbers"),
		statement("const " + lookup + ": ReadonlyMap<string, number> = new Map([['a', 3], ['b', 8], ['c', 1]]);"),
		statement("const " + prepared + ": number[] = [];"),
		statement("for (let "+index+" = 0; "+index+" < "+count+"; "+index+"++) @b", &Block{Statements: []*Statement{
			statement(prepared + ".push((" + index + " * 3 + 1) % 251);"),
		}}),
		statement("const " + items + ": readonly number[] = " + prepared + ";"),
		statement("const "+work+" = (item: number, "+index+": number): number => @b;", &Block{Statements: []*Statement{
			statement("const extra = " + lookup + ".get(" + index + " % 2 === 0 ? 'a' : 'b') ?? 0;"),
			statement("return (item + extra + " + index + ") % 1000;"),
		}}),
		statement("const " + result + " = parallelMap(" + items + ", " + work + ");"),
	}
	body = append(body, g.parallelFinish(result, false)...)
	return statement("@b", &Block{Statements: body})
}

func (g *generator) parallelStrings() *Statement {
	index, prepared, items := g.name("index"), g.name("prepared"), g.name("items")
	base, label, pool, result := g.name("base"), g.name("label"), g.name("pool"), g.name("result")
	fill, at := g.name("fill"), g.name("at")
	count := g.parallelCount()
	// The same runtime strings sit in many slots. Workers then slice and read them together, which
	// is what makes a missing share visible: a literal would be immortal, and a fresh slice per slot
	// would be touched by only one worker.
	body := []*Statement{
		statement("// parallel-shape: strings"),
		statement("const " + base + " = `" + parallelBaseText() + "`;"),
		statement("const " + label + " = " + base + ".slice(0, 6);"),
		statement("const " + pool + ": string[] = [];"),
		statement("for (let "+fill+" = 0; "+fill+" < 8; "+fill+"++) @b", &Block{Statements: []*Statement{
			statement(pool + ".push(" + base + ".slice(" + fill + ", 96 + " + fill + "));"),
		}}),
		statement("const " + prepared + ": string[] = [];"),
		statement("for (let "+index+" = 0; "+index+" < "+count+"; "+index+"++) @b", &Block{Statements: []*Statement{
			statement(prepared + ".push(" + pool + "[" + index + " % " + pool + ".length] ?? " + label + ");"),
		}}),
		statement("const " + items + ": readonly string[] = " + prepared + ";"),
		statement("const "+result+" = parallelMap("+items+", (item, "+at+"): string => @b);", &Block{Statements: []*Statement{
			statement("let mix = 0;"),
			statement("const units = item.length;"),
			statement("for (let step = 0; step < 24; step++) @b", &Block{Statements: []*Statement{
				statement("mix += item.codePointAt((step * 7) % units) ?? 0;"),
			}}),
			statement("return `${" + at + "}:${item.slice(0, 2)}:${item.slice(1, 5)}:${mix}:${" + label + "}`;"),
		}}),
	}
	body = append(body, g.parallelFinish(result, true)...)
	return statement("@b", &Block{Statements: body})
}

func (g *generator) parallelRecords() *Statement {
	index, prepared, items := g.name("index"), g.name("prepared"), g.name("items")
	row, base, shared, result := g.name("Row"), g.name("base"), g.name("shared"), g.name("result")
	count := g.parallelCount()
	body := []*Statement{
		statement("// parallel-shape: records"),
		statement("interface " + row + " { readonly name: string; readonly score: number; readonly tags: readonly (readonly string[])[]; }"),
		statement("const " + base + " = `" + parallelBaseText() + "`;"),
		statement("const " + shared + " = " + base + ".slice(0, 48);"),
		statement("const " + prepared + ": " + row + "[] = [];"),
		statement("for (let "+index+" = 0; "+index+" < "+count+"; "+index+"++) @b", &Block{Statements: []*Statement{
			statement(prepared + ".push({ name: " + shared + ", score: " + index + " % 200, tags: [[" + base + ".slice(0, 2), `${" + index + "}`]] });"),
		}}),
		statement("const " + items + ": readonly " + row + "[] = " + prepared + ";"),
		statement("const "+result+" = parallelMap("+items+", (item, "+index+"): string => @b);", &Block{Statements: []*Statement{
			statement("let mix = 0;"),
			statement("const units = item.name.length;"),
			statement("for (let step = 0; step < 16; step++) @b", &Block{Statements: []*Statement{
				statement("mix += item.name.codePointAt(step % units) ?? 0;"),
			}}),
			statement("const group = item.tags[" + index + " % 2];"),
			statement("const text = group === undefined ? '' : group.join('|');"),
			statement("return `${" + index + "}:${item.name.slice(0, 3)}:${item.score}:${text}:${mix}`;"),
		}}),
	}
	body = append(body, g.parallelFinish(result, true)...)
	return statement("@b", &Block{Statements: body})
}

func (g *generator) parallelNested() *Statement {
	index, prepared, items := g.name("index"), g.name("prepared"), g.name("items")
	result := g.name("result")
	count := g.parallelCount()
	body := []*Statement{
		statement("// parallel-shape: nested"),
		statement("const " + prepared + ": (readonly number[])[] = [];"),
		statement("for (let "+index+" = 0; "+index+" < "+count+"; "+index+"++) @b", &Block{Statements: []*Statement{
			statement(prepared + ".push([" + index + " % 97, (" + index + " * 2) % 97, (" + index + " + 5) % 13]);"),
		}}),
		statement("const " + items + ": readonly (readonly number[])[] = " + prepared + ";"),
		statement("const "+result+" = parallelMap("+items+", (item, "+index+"): number => @b);", &Block{Statements: []*Statement{
			statement("const first = item[0] ?? 0;"),
			statement("const second = item[1] ?? 0;"),
			statement("const third = item[3] ?? 0;"),
			statement("return first + second + third + item.length + " + index + ";"),
		}}),
	}
	body = append(body, g.parallelFinish(result, false)...)
	return statement("@b", &Block{Statements: body})
}

func (g *generator) parallelFreshMap() *Statement {
	index, prepared, items := g.name("index"), g.name("prepared"), g.name("items")
	bias, result := g.name("bias"), g.name("result")
	count := g.parallelCount()
	body := []*Statement{
		statement("// parallel-shape: fresh-map"),
		statement("const " + bias + " = " + g.pick("2", "4", "7") + ";"),
		statement("const " + prepared + ": number[] = [];"),
		statement("for (let "+index+" = 0; "+index+" < "+count+"; "+index+"++) @b", &Block{Statements: []*Statement{
			statement(prepared + ".push(" + index + " % 180);"),
		}}),
		statement("const " + items + ": readonly number[] = " + prepared + ";"),
		statement("const "+result+" = parallelMap("+items+", (item, "+index+"): number => @b);", &Block{Statements: []*Statement{
			statement("const seen = new Map<string, number>();"),
			statement("const key = `k${" + index + " % 4}`;"),
			statement("seen.set(key, item);"),
			statement("seen.set('n', item + " + index + " + " + bias + ");"),
			statement("const left = seen.get(key) ?? 0;"),
			statement("const right = seen.get('missing') ?? 0;"),
			statement("seen.delete(key);"),
			statement("return left + right + seen.size;"),
		}}),
	}
	body = append(body, g.parallelFinish(result, false)...)
	return statement("@b", &Block{Statements: body})
}

func (g *generator) parallelNestedCall() *Statement {
	index, prepared, items := g.name("index"), g.name("prepared"), g.name("items")
	result, sum, value := g.name("result"), g.name("sum"), g.name("value")
	count := g.parallelCount()
	body := []*Statement{
		statement("// parallel-shape: nested-call"),
		statement("const " + prepared + ": number[] = [];"),
		statement("for (let "+index+" = 0; "+index+" < "+count+"; "+index+"++) @b", &Block{Statements: []*Statement{
			statement(prepared + ".push((" + index + " * 5) % 90);"),
		}}),
		statement("const " + items + ": readonly number[] = " + prepared + ";"),
		statement("const "+result+" = parallelMap("+items+", (item, "+index+"): number => @b);", &Block{Statements: []*Statement{
			statement("const children: readonly number[] = [item, " + index + " % 50, 1, 2, 3];"),
			statement("const inner = parallelMap(children, (child, childIndex) => child + childIndex + " + index + ");"),
			statement("let " + sum + " = 0;"),
			statement("for (const "+value+" of inner) @b", &Block{Statements: []*Statement{
				statement(sum + " += " + value + ";"),
			}}),
			statement("return " + sum + ";"),
		}}),
	}
	body = append(body, g.parallelFinish(result, false)...)
	return statement("@b", &Block{Statements: body})
}

// parallelFinish reads the result in index order and prints a digest plus a prefix, so a reorder
// changes the output without printing the whole array.
func (g *generator) parallelFinish(result string, text bool) []*Statement {
	place, folded, value := g.name("place"), g.name("folded"), g.name("value")
	fallback := "0"
	step := folded + " = (" + folded + " + (" + place + " + 1) * ((Math.abs(Math.trunc(" + value + ")) % 997) + 1)) % 1000003;"
	if text {
		fallback = "''"
		step = folded + " = (" + folded + " + (" + place + " + 1) * (((" + value + ".length + (" + value + ".codePointAt(0) ?? 0)) % 997) + 1)) % 1000003;"
	}
	return []*Statement{
		statement("let " + folded + " = 0;"),
		statement("for (let "+place+" = 0; "+place+" < "+result+".length; "+place+"++) @b", &Block{Statements: []*Statement{
			statement("const " + value + " = " + result + "[" + place + "] ?? " + fallback + ";"),
			statement(step),
		}}),
		statement("console.log(`${" + result + ".length}:${" + folded + "}:${" + result + ".at(0) ?? " + fallback + "}:${" + result + ".at(-1) ?? " + fallback + "}`);"),
		statement("console.log(" + result + ".slice(0, 5).join('|'));"),
	}
}

func (g *generator) refuseCapturedLet() ([]*Statement, string) {
	step, items, result := g.name("step"), g.name("items"), g.name("result")
	expected := "task capture '" + step + "' is not shareable: " + step + " is not an immutable binding"
	body := &Block{Statements: []*Statement{
		statement("let " + step + " = " + g.pick("1", "2", "5", "10") + ";"),
		statement("const " + items + ": readonly number[] = [1, 2, 3, 4, 5];"),
		statement("const " + result + " = parallelMap(" + items + ", (item) => item + " + step + ");"),
		statement("console.log(" + result + ".join(','));"),
	}}
	return []*Statement{statement("@b", body)}, expected
}

func (g *generator) refuseMutableArray() ([]*Statement, string) {
	if g.chance(1, 2) {
		items, result := g.name("items"), g.name("result")
		expected := "parallelMap items are not shareable: items is a mutable number[]"
		body := &Block{Statements: []*Statement{
			statement("const " + items + ": number[] = [1, 2, 3, 4];"),
			statement("const " + result + " = parallelMap(" + items + ", (item) => item + 1);"),
			statement("console.log(" + result + ".join(','));"),
		}}
		return []*Statement{statement("@b", body)}, expected
	}
	bag, items, result := g.name("Bag"), g.name("items"), g.name("result")
	expected := bag + ".tags is a mutable string[]"
	body := &Block{Statements: []*Statement{
		statement("interface " + bag + " { readonly name: string; readonly tags: string[]; }"),
		statement("const " + items + ": readonly " + bag + "[] = [{ name: 'a世界', tags: ['x', 'y'] }];"),
		statement("const " + result + " = parallelMap(" + items + ", (item) => item.tags.length + item.name.length);"),
		statement("console.log(`${" + result + "[0] ?? 0}`);"),
	}}
	return []*Statement{statement("@b", body)}, expected
}

func (g *generator) refuseGlobalWrite() ([]*Statement, string) {
	total, record, items := g.name("total"), g.name("record"), g.name("items")
	result := g.name("result")
	expected := "writes the global '" + total + "'"
	return []*Statement{
		statement("let " + total + " = 0;"),
		statement("function " + record + "(item: number): number { " + total + " += item; return item; }"),
		statement("const " + items + ": readonly number[] = [1, 2, 3, 4];"),
		statement("const " + result + " = parallelMap(" + items + ", " + record + ");"),
		statement("console.log(" + result + ".join(','));"),
	}, expected
}
