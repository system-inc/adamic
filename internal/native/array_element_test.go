package native_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// The element byte (adamic.h's adamic_array) is read natively from the arrays a program itself
// makes. The emitted calls to adamic_array_reverse and adamic_array_join are routed through a
// test-only probe compiled into the witness, which writes the byte of the array it was handed to
// standard error and then does the real call. So the byte read is the real emitted array's, every
// probed array is one the program reverses or joins, in the program's order, and a mutant of the
// emitted C is seen the way the array it changes is.
const elementProbes = `#include "adamic.h"
static adamic_array *element_probe_reverse(adamic_array *array) {
	fprintf(stderr, "element=%u\n", (unsigned)array->element);
	return adamic_array_reverse(array);
}
static adamic_string *element_probe_join(const adamic_array *array, const adamic_string *separator, enum adamic_join kind) {
	fprintf(stderr, "element=%u\n", (unsigned)array->element);
	return adamic_array_join(array, separator, kind);
}
`

// elementName is the enum adamic_ir_type name of an element byte, and 0 for unknown.
func elementName(element string) string {
	for name, tag := range ir.TypeTags() {
		if fmt.Sprint(int(tag)) == element {
			return name
		}
	}
	return element
}

// arrayElementWitness holds source to Node through both backends, then builds it sanitized with the
// probes and gives back the element kinds the probes read, in order. mutant, when given, changes the
// emitted C after the probes are in, and must change it.
func arrayElementWitness(
	test *testing.T,
	source string,
	instrument func(string) string,
	mutant func(string) string,
) []string {
	test.Helper()
	directory := test.TempDir()
	path := filepath.Join(directory, "main.a")
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	run := func(env []string, name string, args ...string) ([]byte, []byte, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, name, args...)
		command.Env = append(os.Environ(), env...)
		var out, diagnostic bytes.Buffer
		command.Stdout = &out
		command.Stderr = &diagnostic
		runError := command.Run()
		return out.Bytes(), diagnostic.Bytes(), runError
	}
	// The source decodes JSON, so Node takes its type descriptors from the prebuilt json-types product.
	truth, diagnostic, runError := run(
		[]string{"ADAMIC_JSON_TYPES=" + jsonTypes(test)},
		"node",
		"--disable-warning=ExperimentalWarning",
		"../../oracle/node.mjs",
		path,
	)
	if runError != nil {
		test.Fatalf("source Node: %v\n%s", runError, diagnostic)
	}
	checked, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerWithOptionsError := lower.LowerWithOptions(context.Background(), checked, lower.Options{})
	if lowerWithOptionsError != nil {
		test.Fatal(lowerWithOptionsError)
	}
	js := filepath.Join(directory, "main.mjs")
	if writeError := os.WriteFile(js, []byte(javascript.JavaScript(program)), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	got, diagnostic, runError := run(nil, "node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", js)
	if runError != nil || !bytes.Equal(got, truth) {
		test.Fatalf("JavaScript: %v\n%s\n%s, Node %s", runError, diagnostic, got, truth)
	}
	code := native.C(program)
	code = strings.ReplaceAll(code, "adamic_array_reverse(", "element_probe_reverse(")
	code = strings.ReplaceAll(code, "adamic_array_join(", "element_probe_join(")
	code = strings.Replace(code, "#include \"adamic.h\"\n", elementProbes, 1)
	if instrument != nil {
		code = instrument(code)
	}
	if mutant != nil {
		changed := mutant(code)
		if changed == code {
			test.Fatal("mutant did not change C")
		}
		code = changed
	}
	binary := filepath.Join(directory, "native")
	if buildError := native.Build(code, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	got, diagnostic, runError = run([]string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if runError != nil || !bytes.Equal(got, truth) {
		test.Fatalf("native: %v\n%s\n%s, Node %s", runError, diagnostic, got, truth)
	}
	elements := []string{}
	for _, line := range strings.Split(string(diagnostic), "\n") {
		if element, found := strings.CutPrefix(line, "element="); found {
			elements = append(elements, elementName(element))
		} else if strings.HasPrefix(line, "reused=") {
			elements = append(elements, line)
		}
	}
	return elements
}

// elementKindSource makes an array by each path the emitter and runtime have, and reverses (or, for
// a readonly one, joins) it once, so each probe line is one path's array.
const elementKindSource = `import { decodeJson, parallelMap } from 'adamic';
function rest(...values: number[]): number {
	values.reverse();
	return values.length;
}
function tripled(shared: readonly number[]): number[] {
	return parallelMap(shared, (value) => value * 3);
}
const numbers = [1, 2, 3];
numbers.reverse();
const flags = [true, false];
flags.reverse();
const maybes: (number | undefined)[] = [1, undefined];
maybes.reverse();
const words = [` + "`a${1}`, `b${2}`, `c${3}`" + `];
words.reverse();
const doubled = numbers.map((value) => value * 2);
doubled.reverse();
const kept = words.filter((word) => word.length > 0);
kept.reverse();
const made = Array.from({ length: 3 }, (_, index) => index > 0);
made.reverse();
const filled = new Array<number>(2).fill(7);
filled.reverse();
const sliced = maybes.slice(1);
sliced.reverse();
const removed = words.splice(0, 1);
removed.reverse();
const both = maybes.concat(maybes);
both.reverse();
const table = new Map<number, boolean>();
table.set(1, true);
table.set(2, false);
const keys = [...table.keys()];
keys.reverse();
const values = [...table.values()];
values.reverse();
const set = new Set<number>([3, 4]);
const members = [...set];
members.reverse();
const spreadCall = rest(...numbers);
const times = tripled([5, 6]);
times.reverse();
const decoded = decodeJson<readonly number[]>('[8,9]');
const decodedText = decoded.kind === 'Error' ? decoded.message : decoded.value.join(',');
const point = { x: 1, y: 2 };
const coordinates = Object.values(point);
coordinates.reverse();
console.log(` + "`${numbers[0]} ${flags[0]} ${maybes[0]} ${words[0]} ${doubled[0]} ${kept[0]} ${made[0]} ${filled[0]}`" + `);
console.log(` + "`${sliced.length} ${removed[0]} ${both.length} ${keys[0]} ${values[0]} ${members[0]} ${spreadCall} ${times[0]} ${decodedText} ${coordinates[0]}`" + `);
`

// elementKinds is what elementKindSource's probes read, in order, each named by its path.
var elementKinds = []struct{ path, element string }{
	{"literal of numbers", "Number"},
	{"literal of booleans", "Boolean"},
	{"literal of number | undefined", "MaybeNumber"},
	{"literal of strings", "Union"},
	{"map", "Number"},
	{"filter", "Union"},
	{"Array.from", "Boolean"},
	{"new Array(n).fill", "Number"},
	{"slice", "MaybeNumber"},
	{"splice", "Union"},
	{"concat of number | undefined", "MaybeNumber"},
	{"Map keys", "Number"},
	{"Map values", "Boolean"},
	{"Set values", "Number"},
	{"rest arguments from a spread", "Number"},
	{"parallel map", "Number"},
	{"decodeJson into readonly number[]", "Number"},
	{"Object.values", "Number"},
}

func expectedKinds() []string {
	kinds := []string{}
	for _, each := range elementKinds {
		kinds = append(kinds, each.element)
	}
	return kinds
}

func describeKinds(read []string) string {
	lines := []string{}
	for index, each := range read {
		path := "(extra)"
		if index < len(elementKinds) {
			path = elementKinds[index].path
		}
		lines = append(lines, fmt.Sprintf("%s: %s", path, each))
	}
	return strings.Join(lines, "\n")
}

// TestArrayElementKind reads every creation path's byte. Mutants, each red here: the Array.from
// site made with 0 (an emitter site passing unknown); the boolean literal made with Number.
func TestArrayElementKind(test *testing.T) {
	test.Parallel()
	read := arrayElementWitness(test, elementKindSource, nil, nil)
	if !slices.Equal(read, expectedKinds()) {
		test.Fatalf("element kinds:\n%s\nwant:\n%s", describeKinds(read), describeKinds(expectedKinds()))
	}
}

// arrayElementMutant requires a mutant of elementKindSource's C to be read as a wrong kind.
func arrayElementMutant(test *testing.T, pattern, replacement string) {
	test.Helper()
	read := arrayElementWitness(test, elementKindSource, nil, func(code string) string {
		matches := regexp.MustCompile(pattern).FindAllStringIndex(code, -1)
		if len(matches) != 1 {
			test.Fatalf("mutant %s matches %d sites, want 1", pattern, len(matches))
		}
		return regexp.MustCompile(pattern).ReplaceAllString(code, replacement)
	})
	if slices.Equal(read, expectedKinds()) {
		test.Fatalf("mutant %s survived: every kind read as expected", pattern)
	}
	test.Logf("mutant caught:\n%s", describeKinds(read))
}

func TestArrayElementKindEmitterUnknownMutant(test *testing.T) {
	test.Parallel()
	arrayElementMutant(test, `adamic_array_make\(([A-Za-z_]\w*), adamic_ir_Boolean\)`, `adamic_array_make($1, 0)`)
}

func TestArrayElementKindBooleanLiteralMutant(test *testing.T) {
	test.Parallel()
	arrayElementMutant(test, `adamic_array_make\(2, adamic_ir_Boolean\)`, `adamic_array_make(2, adamic_ir_Number)`)
}

// reuseAllocation is reuse.go's in-place allocation: the source array itself when it's unique.
var reuseAllocation = regexp.MustCompile(
	`(adamic_array \*\s*(\w+) = \((\w+) \? adamic_retain\((\w+)\) : adamic_array_make\([^\n]+;)`,
)

// reusedProbe zeroes the reused source's byte just before the allocation that may take it over, and
// says after it whether the result is that same array. So a result that reads its kind read the byte
// the reuse wrote, never one it inherited from the source.
func reusedProbe(test *testing.T) func(string) string {
	return func(code string) string {
		matches := reuseAllocation.FindAllStringSubmatch(code, -1)
		if len(matches) != 1 {
			test.Fatalf("want one reused array allocation, got %d", len(matches))
		}
		result, source := matches[0][2], matches[0][4]
		return strings.Replace(code, matches[0][1], fmt.Sprintf(
			"%s->element = 0; %s fprintf(stderr, \"reused=%%d\\n\", %s == %s);",
			source,
			matches[0][1],
			result,
			source,
		), 1)
	}
}

const reuseObjectsSource = `interface Item {
	name: string;
}
function names(count: number): string {
	const items: Item[] = [];
	for (let index = 0; index < count; index++) {
		items.push({ name: ` + "`n${index}`" + ` });
	}
	const labels = items.map((item) => item.name);
	labels.reverse();
	return labels[0] ?? '';
}
console.log(names(3));
`

const reuseNumbersSource = `function grown(count: number): number {
	const numbers: number[] = [];
	for (let index = 0; index < count; index++) {
		numbers.push(index);
	}
	const doubled = numbers.map((value) => value * 2);
	doubled.reverse();
	return doubled[0] ?? -1;
}
console.log(` + "`${grown(3)}`" + `);
`

const reuseSpreadSource = `function grown(count: number): string {
	const list: string[] = [];
	for (let index = 0; index < count; index++) {
		list.push(` + "`item${index}`" + `);
	}
	const longer = [...list, ` + "`last${count}`" + `];
	longer.reverse();
	return longer[0] ?? '';
}
console.log(grown(3));
`

// TestArrayElementKindReuse pins the arrays reuse.go takes over in place: an object[] mapped to
// string[], a number[] mapped to number[], and a string[] spread first into a literal. Each was
// reused (the result is the source) and reads the kind the reuse wrote over a zeroed byte. Mutant:
// the unique branch not writing the byte reads 0, red in TestArrayElementKindReuseWriteMutant.
func TestArrayElementKindReuse(test *testing.T) {
	test.Parallel()
	for _, row := range []struct{ name, source, element string }{
		{"objects mapped to strings", reuseObjectsSource, "Union"},
		{"numbers mapped to numbers", reuseNumbersSource, "Number"},
		{"strings spread first", reuseSpreadSource, "Union"},
	} {
		test.Run(row.name, func(test *testing.T) {
			test.Parallel()
			read := arrayElementWitness(test, row.source, reusedProbe(test), nil)
			if want := []string{"reused=1", row.element}; !slices.Equal(read, want) {
				test.Fatalf("reuse read %q, want %q", read, want)
			}
		})
	}
}

func TestArrayElementKindReuseWriteMutant(test *testing.T) {
	test.Parallel()
	read := arrayElementWitness(test, reuseObjectsSource, reusedProbe(test), func(code string) string {
		return regexp.MustCompile(`\n\s*\w+->element = adamic_ir_Union;`).ReplaceAllString(code, "")
	})
	if want := []string{"reused=1", "Union"}; slices.Equal(read, want) {
		test.Fatal("a reuse that writes no byte still read Union")
	}
	test.Logf("mutant caught: %q", read)
}

// A string[] mapped to number[] is a scalar array over a reference one, which sameSlots (reuse.go)
// never lets reuse take over: the map makes a new array of numbers, and no reused allocation is
// emitted.
func TestArrayElementKindScalarOverReferenceIsNotReused(test *testing.T) {
	test.Parallel()
	source := `function lengths(count: number): number {
	const words: string[] = [];
	for (let index = 0; index < count; index++) {
		words.push(` + "`w${index}`" + `);
	}
	const sizes = words.map((word) => word.length);
	sizes.reverse();
	return sizes[0] ?? -1;
}
console.log(` + "`${lengths(3)}`" + `);
`
	read := arrayElementWitness(test, source, func(code string) string {
		if reuseAllocation.MatchString(code) {
			test.Fatal("a string[] mapped to number[] was planned for reuse")
		}
		return code
	}, nil)
	if want := []string{"Number"}; !slices.Equal(read, want) {
		test.Fatalf("read %q, want %q", read, want)
	}
}

// The static type at a read and the array's storage can't differ for these: lowering refuses each
// before an array could be held one way and read another.
func TestArrayElementKindStorageRefusals(test *testing.T) {
	test.Parallel()
	for _, row := range []struct{ name, source, refusal string }{
		{"number[] seen as (number | undefined)[]", `const numbers: number[] = [1, 2];
const maybes: (number | undefined)[] = numbers;
maybes.push(undefined);
console.log(` + "`${numbers.length}`" + `);
`, "seen as (number | undefined)[]"},
		{"number[] passed as (number | undefined)[]", `function widen(values: (number | undefined)[]): void {
	values.push(undefined);
}
const numbers: number[] = [1, 2];
widen(numbers);
console.log(` + "`${numbers.length}`" + `);
`, "seen as (number | undefined)[]"},
		{"number | undefined concat number", `const numbers: number[] = [1, 2];
const maybes: (number | undefined)[] = [3, undefined];
const joined = maybes.concat(numbers);
console.log(` + "`${joined.length}`" + `);
`, "concat of arrays of different elements"},
		{"number[] spread into (number | string)[]", `const numbers: number[] = [1, 2];
const mixed: (number | string)[] = [...numbers, 'x'];
console.log(` + "`${mixed.length}`" + `);
`, "spreading an array of other elements"},
	} {
		test.Run(row.name, func(test *testing.T) {
			test.Parallel()
			path := filepath.Join(test.TempDir(), "main.a")
			if writeError := os.WriteFile(path, []byte(row.source), 0600); writeError != nil {
				test.Fatal(writeError)
			}
			checked, loadError := load.Load([]string{path})
			if loadError != nil {
				test.Fatal(loadError)
			}
			_, lowerWithOptionsError := lower.LowerWithOptions(context.Background(), checked, lower.Options{})
			if lowerWithOptionsError == nil || !strings.Contains(lowerWithOptionsError.Error(), row.refusal) {
				test.Fatalf("want a refusal naming %q, got %v", row.refusal, lowerWithOptionsError)
			}
		})
	}
}

// scalarElementFunction is array.c's adamic_array_scalar_element, its text as written.
var scalarElementFunction = regexp.MustCompile(
	`(?s)enum adamic_ir_type adamic_array_scalar_element\(const adamic_array \*array\) \{\n.*?\n\}\n`,
)

// scalarElementHarness asks for the scalar kind of an array made with the kind its first argument
// names, through the runtime's own function, or with its second argument "copy", through array.c's
// text of it compiled here (so a mutant of that text is what runs).
func scalarElementHarness(test *testing.T, mutant func(string) string) string {
	test.Helper()
	contents, readError := os.ReadFile("runtime/array.c")
	if readError != nil {
		test.Fatal(readError)
	}
	function := scalarElementFunction.FindString(string(contents))
	if function == "" {
		test.Fatal("array.c has no adamic_array_scalar_element")
	}
	function = strings.Replace(function, "adamic_array_scalar_element(", "copied_scalar_element(", 1)
	if mutant != nil {
		changed := mutant(function)
		if changed == function {
			test.Fatal("mutant did not change adamic_array_scalar_element")
		}
		function = changed
	}
	source := `#include "adamic.h"
#include <stdio.h>
#include <string.h>
static ` + function + `
int main(int argc, char **argv) {
	if (argc < 3) return 2;
	enum adamic_ir_type kind = strcmp(argv[1], "number") == 0 ? adamic_ir_Number : strcmp(argv[1], "boolean") == 0 ? adamic_ir_Boolean : strcmp(argv[1], "maybe") == 0 ? adamic_ir_MaybeNumber : strcmp(argv[1], "string") == 0 ? adamic_ir_String : 0;
	adamic_array *array = adamic_array_make(2, kind);
	enum adamic_ir_type read = strcmp(argv[2], "copy") == 0 ? copied_scalar_element(array) : adamic_array_scalar_element(array);
	printf("%d\n", (int)read);
	adamic_release(array);
	return 0;
}
`
	binary := filepath.Join(test.TempDir(), "scalar")
	if buildError := native.Build(source, binary, native.Options{}); buildError != nil {
		test.Fatal(buildError)
	}
	return binary
}

// scalarElementRows are what adamic_array_scalar_element answers: a scalar array's kind, and a stop
// naming why for an array of references or of a kind not known.
var scalarElementRows = []struct{ kind, answer string }{
	{"number", "1"},
	{"boolean", "2"},
	{"maybe", "7"},
	{"string", "panic: compiler bug: a scalar element kind was asked of a reference array"},
	{"unknown", "panic: compiler bug: a scalar element kind was asked of an array of unknown element kind"},
}

func scalarElementAnswer(test *testing.T, binary, kind, which string) string {
	test.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var out, diagnostic bytes.Buffer
	command := exec.CommandContext(ctx, binary, kind, which)
	command.Stdout = &out
	command.Stderr = &diagnostic
	if runError := command.Run(); runError != nil {
		if line, found := strings.CutPrefix(strings.TrimSpace(diagnostic.String()), "adamic: "); found {
			return line
		}
		test.Fatalf("%s %s: %v\n%s", kind, which, runError, diagnostic.String())
	}
	return strings.TrimSpace(out.String())
}

// TestArrayScalarElement asks adamic_array_scalar_element of arrays made with each kind, the linked
// runtime's and array.c's text compiled beside it, which must agree. Mutant: the unknown branch
// returning Number, red in TestArrayScalarElementUnknownMutant.
func TestArrayScalarElement(test *testing.T) {
	test.Parallel()
	binary := scalarElementHarness(test, nil)
	for _, row := range scalarElementRows {
		for _, which := range []string{"runtime", "copy"} {
			if answer := scalarElementAnswer(test, binary, row.kind, which); answer != row.answer {
				test.Errorf("%s through the %s function: %q, want %q", row.kind, which, answer, row.answer)
			}
		}
	}
}

func TestArrayScalarElementUnknownMutant(test *testing.T) {
	test.Parallel()
	binary := scalarElementHarness(test, func(function string) string {
		return strings.Replace(
			function,
			"if (array->element == 0) {",
			"if (array->element == 0) {\n\t\treturn adamic_ir_Number;",
			1,
		)
	})
	if answer := scalarElementAnswer(test, binary, "unknown", "copy"); answer == scalarElementRows[4].answer {
		test.Fatal("a scalar element read of an unknown array still stopped")
	}
}
