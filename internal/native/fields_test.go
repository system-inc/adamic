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
const scannerLike = { text: 'local text', pos: 0, start: 0, fullStart: 0, kind: 'local kind' };
console.log(scannerLike.text + ':' + scannerLike.kind);
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
		return regexp.MustCompile(`adamic_object_field\([^\n]+, ` + strconv.Quote(name) + `, &adamic_cache_`).MatchString(generated)
	}
	if lookup("steady") || lookup("label") || lookup("text") || lookup("kind") {
		t.Fatal("uniform fields still use shape lookup")
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
	offsets := uniformFieldOffsets(&ir.Program{Main: []ir.Statement{ir.Evaluate{Value: ir.ReadTextFile{Path: ir.StringConstant{Index: 0}}}}})
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
				if len(parts) != 4 {
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
	if !regexp.MustCompile(`adamic_object_field\([^\n]+, "uniform", &adamic_cache_`).MatchString(generated) {
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

func TestInputLayoutReachability(t *testing.T) {
	t.Parallel()
	literal := ir.ObjectLiteral{Fields: []ir.Field{{Name: "text"}, {Name: "pos"}, {Name: "start"}, {Name: "fullStart"}, {Name: "kind"}}}
	program := &ir.Program{Main: []ir.Statement{ir.Evaluate{Value: literal}}}
	offsets := uniformFieldOffsets(program)
	if offsets["text"] != 0 || offsets["kind"] != 4 {
		t.Fatalf("unused input API conflicts with layout: %v", offsets)
	}
	for _, call := range []ir.Expression{
		ir.ReadTextFile{Path: ir.StringConstant{}}, ir.WriteTextFile{Path: ir.StringConstant{}, Text: ir.StringConstant{}},
		ir.ReadDirectory{Path: ir.StringConstant{}}, ir.FileStatus{Path: ir.StringConstant{}},
	} {
		// All bodies count, including a call inside an otherwise unused callback.
		program.Functions = []ir.Function{{Body: []ir.Statement{ir.If{Condition: ir.BooleanConstant{Value: true}, Then: []ir.Statement{ir.Evaluate{Value: call}}}}}}
		offsets = uniformFieldOffsets(program)
		if offsets["text"] != -1 || offsets["kind"] != -1 {
			t.Fatalf("%T omitted input layout: %v", call, offsets)
		}
	}
}

// The file API's kind/text slots conflict with this scanner-like class. A missing
// runtime layout must fail by actual execution, not just by an emission assertion.
func TestInputAPIFieldLayouts(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	payload := filepath.Join(directory, "payload.txt")
	if err := os.WriteFile(payload, []byte("disk text"), 0o644); err != nil {
		t.Fatal(err)
	}
	source := `import { readTextFile, panic } from 'adamic';
class ScannerLike {
    text = 'local text';
    pos = 0;
    start = 0;
    fullStart = 0;
    kind = 'local kind';
}
const local = new ScannerLike();
const result = readTextFile(` + strconv.Quote(payload) + `);
if (result.kind === 'Error') { panic(result.message); }
console.log(local.text + ':' + local.kind + '|' + result.kind + ':' + result.text);
`
	path := filepath.Join(directory, "input-layout.a")
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
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(directory, "input-layout")
		if err := Build(C(program), binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != "local text:local kind|Ok:disk text\n" {
			t.Fatalf("sanitize %v: runtime input fields differ: %q", sanitize, got)
		}
	}
}
