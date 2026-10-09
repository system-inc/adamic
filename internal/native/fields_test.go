package native

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Uniform slots hold across structural literals, extra fields, and generic representations.
// Reordered layouts, including a spread's synthesized undefined layout, must still use lookup.
func TestUniformFieldsMatchNode(t *testing.T) {
	t.Parallel()
	const source = `class Position {
    steady = 1;
    label = 'position'.repeat(2);
    step(): number { this.steady += 2; return this.steady; }
}
class Other {
    steady = 10;
    label = 'other'.repeat(2);
    extra = 99;
}
function advance(point: { steady: number; label: string }): string {
    point.steady += 3;
    point.label += '!';
    return point.label + ':' + point.steady.toString();
}
const position = new Position();
const other = new Other();
const literal = { steady: 20, label: 'literal'.repeat(2), tail: false };
console.log([position.step().toString(), advance(position), advance(other), advance(literal)].join(" "));
class Box<Item> {
    value: Item;
    constructor(value: Item) { this.value = value; }
}
const numberBox = new Box<number>(42);
const stringBox = new Box<string>('boxed'.repeat(2));
console.log(numberBox.value.toString() + " " + stringBox.value);
function pair(point: { left: number; right: number }): number {
    point.left += 10;
    return point.left * 100 + point.right;
}
const first = { left: 1, right: 2 };
const reversed = { right: 3, left: 4 };
const third = { left: 5, right: 6 };
console.log([pair(first), pair(reversed), pair(third)].join(" "));
interface Empty { readonly emptyLeft: number | undefined; readonly emptyRight?: number | undefined; }
function makeEmpty(): Empty | undefined { return undefined; }
const empty = makeEmpty();
const spread = { ...empty, emptyLeft: 7 };
const full: Empty = { emptyLeft: 8, emptyRight: 9 };
function show(value: Empty): void { console.log([value.emptyLeft ?? -1, value.emptyRight ?? -2].join(" ")); }
show(spread);
show(full);
function absent(): Position | undefined { return undefined; }
const missing = absent();
console.log((missing?.steady ?? -3).toString());
`
	directory := t.TempDir()
	path := filepath.Join(directory, "fields.ts")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	generated := C(program)
	// Semantic parity alone would also pass if the optimization disappeared entirely.
	lookup := func(name string) bool {
		return regexp.MustCompile(`adamic_object_(?:(?:data_)?field|maybe_number)\([^\n]+, ` + strconv.Quote(name) + `, &adamic_cache_`).MatchString(generated)
	}
	// Required reads keep the uniform-slot proof. Writes also guard presence and may
	// contain a checked fallback in their slot declaration.
	readLookup := func(name string) bool {
		for _, line := range strings.Split(generated, "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "adamic_value *") {
				continue
			}
			if regexp.MustCompile(`adamic_object_(?:(?:data_)?field|maybe_number)\([^\n]+, ` + strconv.Quote(name) + `, &adamic_cache_`).MatchString(line) {
				return true
			}
		}
		return false
	}
	if readLookup("steady") || readLookup("label") {
		t.Fatal("uniform reads still use shape lookup")
	}
	if !lookup("left") || !lookup("emptyLeft") {
		t.Fatal("conflicting layouts lost their lookup fallback")
	}
	want := runWithInput(t, "", "node", "--disable-warning=ExperimentalWarning", path)
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(directory, "fields")
		if err := Build(generated, binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != want {
			t.Fatalf("sanitize %v: native %q; Node %q", sanitize, got, want)
		}
	}
}

// A new runtime-created shape must not silently bypass the whole-program layout proof.
func TestRuntimeFieldLayoutsAreIncluded(t *testing.T) {
	t.Parallel()
	namesPattern := regexp.MustCompile(`static const char \*const (\w+)\[\] = \{([^}]*)\};`)
	shapePattern := regexp.MustCompile(`static const adamic_shape \w+(?:\[\])? = \{([^;]*)\};`)
	fieldPattern := regexp.MustCompile(`"([^"]*)"`)
	offsets := uniformFieldOffsets(&ir.Program{})
	files, err := runtime.ReadDir("runtime")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".c") {
			continue
		}
		contents, err := runtime.ReadFile("runtime/" + file.Name())
		if err != nil {
			t.Fatal(err)
		}
		source := string(contents)
		names := map[string][]string{}
		for _, array := range namesPattern.FindAllStringSubmatch(source, -1) {
			for _, field := range fieldPattern.FindAllStringSubmatch(array[2], -1) {
				names[array[1]] = append(names[array[1]], field[1])
			}
		}
		shapes := shapePattern.FindAllStringSubmatch(source, -1)
		if len(shapes) != strings.Count(source, "static const adamic_shape ") {
			t.Fatalf("%s: unrecognized runtime shape declaration", file.Name())
		}
		for _, declaration := range shapes {
			initializers := []string{declaration[1]}
			if strings.Contains(declaration[1], "{") {
				initializers = nil
				for _, initializer := range regexp.MustCompile(`\{([^{}]*)\}`).FindAllStringSubmatch(declaration[1], -1) {
					initializers = append(initializers, initializer[1])
				}
			}
			for _, initializer := range initializers {
				parts := strings.Split(initializer, ",")
				if len(parts) != 5 {
					t.Fatalf("%s: unrecognized runtime shape %s", file.Name(), initializer)
				}
				count, err := strconv.Atoi(strings.TrimSpace(parts[0]))
				fields := names[strings.TrimSpace(parts[1])]
				if err != nil || count != len(fields) {
					t.Fatalf("%s: unrecognized runtime fields %s", file.Name(), initializer)
				}
				for index, name := range fields {
					if offset, found := offsets[name]; !found || (offset != index && offset != -1) {
						t.Fatalf("%s: runtime field %q at %d is absent or conflicting in the layout proof", file.Name(), name, index)
					}
				}
				checked++
			}
		}
	}
	if checked == 0 {
		t.Fatal("no runtime shapes checked")
	}
	t.Logf("%d runtime layouts included", checked)
}

func TestRegexProgramsKeepCheckedFieldReads(t *testing.T) {
	t.Parallel()
	const source = `const value={uniform:42};console.log(value.uniform.toString());console.log(/(?<uniform>x)/.exec('x')?.groups?.uniform ?? 'missing');`
	directory := t.TempDir()
	path := filepath.Join(directory, "regex-fields.ts")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	generated := C(program)
	if !regexp.MustCompile(`adamic_object_(?:(?:data_)?field|maybe_number)\([^\n]+, "uniform", &adamic_cache_`).MatchString(generated) {
		t.Fatal("regex program specialized a field before named-group layouts entered the proof")
	}
	want := runWithInput(t, "", "node", "--disable-warning=ExperimentalWarning", path)
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(directory, "regex-fields")
		if err := Build(generated, binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != want {
			t.Fatalf("sanitize %v: native %q; Node %q", sanitize, got, want)
		}
	}
}

// Contextual literals reserve optional slots even when passed directly to a function.
func TestOptionalWriteReservedSlotMatchesNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("testdata/field_write_absent.a")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	wantNode := runWithInput(t, string(source), "node", "--disable-warning=ExperimentalWarning", "-e",
		"const fs=require('fs'),m=require('module'); eval(m.stripTypeScriptTypes(fs.readFileSync(0,'utf8')))")
	if wantNode != "2\n2\n" {
		t.Fatalf("Node changed missing-property semantics: %q", wantNode)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	for _, sanitize := range []bool{true, false} {
		binary := filepath.Join(t.TempDir(), "absent-write")
		if err := Build(C(program), binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		if output := runWithInput(t, "", binary); output != wantNode {
			t.Fatalf("sanitize %v: reserved-slot write: got %q, Node %q", sanitize, output, wantNode)
		}
	}
}
