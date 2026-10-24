package native

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func namedProgram(t *testing.T, entry string, overlay map[string]string) *ir.Program {
	t.Helper()
	loaded, err := load.LoadOverlay([]string{entry}, overlay)
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	return program
}

func namedBody(t *testing.T, program *ir.Program, code string, index int) string {
	t.Helper()
	e := &emitter{program: program}
	start := strings.Index(code, "static "+e.signature(index)+" {\n")
	if start < 0 {
		t.Fatalf("body absent: %s", e.functionName(index))
	}
	end := strings.Index(code[start:], "\n}\n")
	if end < 0 {
		t.Fatal("unterminated body")
	}
	return code[start : start+end+3]
}

func TestSourceNamesDoNotMove(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	entry := filepath.Join(directory, "main.a")
	module := filepath.Join(directory, "a-b.a")
	sources := map[string]string{
		entry:                             `import { same as first } from './a-b.a'; import { same as second } from './a_b.a'; console.log((first(1) + second(2)).toString());`,
		module:                            `export const marker = [14].length; export const container = { get value(): number { return 14; } }; export function same(value: number): number { const callback = () => value + 14; return callback(); }`,
		filepath.Join(directory, "a_b.a"): `export const marker = [14].length; export const container = { get value(): number { return 14; } }; export function same(value: number): number { const callback = () => value + 14; return callback(); }`,
	}
	for path, source := range sources {
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	before := namedProgram(t, entry, nil)
	original := C(before)
	bodies := map[string]string{}
	e := &emitter{program: before}
	for index, function := range before.Functions {
		if function.Source.Module == "" {
			t.Fatalf("missing source for %s", function.Name)
		}
		name := e.functionName(index)
		if _, found := bodies[name]; found {
			t.Fatalf("colliding declaration %s", name)
		}
		bodies[name] = namedBody(t, before, original, index)
	}
	for _, edit := range []string{
		"function unrelated(): number { return 0; }\n" + sources[module],
		"export const added = [14].length;\n" + sources[module],
		strings.Replace(sources[module], "const callback", "const temporary = [14].length; const callback", 1),
		strings.Replace(sources[module], "value + 14", "value + [14].length", 1),
	} {
		changed := namedProgram(t, entry, map[string]string{module: edit})
		code := C(changed)
		moduleBody := func(code, module string) string {
			code = code[strings.Index(code, "int main("):]
			start := strings.Index(code, "// adamic-module "+strconv.Quote(module))
			if start < 0 {
				t.Fatalf("missing main module %s", module)
			}
			code = code[start:]
			if end := strings.Index(code[1:], "// adamic-module "); end >= 0 {
				code = code[:end+1]
			}
			return code
		}
		if moduleBody(original, "a_b.a") != moduleBody(code, "a_b.a") {
			t.Fatal("edit moved unrelated module-level code")
		}
		names := &emitter{program: changed}
		checked := 0
		for index, function := range changed.Functions {
			name := names.functionName(index)
			if function.Source.Module == "a-b.a" {
				continue
			}
			if body, found := bodies[name]; !found || body != namedBody(t, changed, code, index) {
				t.Fatalf("edit moved unrelated function %s", name)
			}
			checked++
		}
		if checked == 0 {
			t.Fatal("no unaffected declarations checked")
		}
	}
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(directory, "names")
		if err := Build(original, binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != runWithInput(t, "", "node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", entry) {
			t.Fatalf("output %q", got)
		}
	}
}

func TestStableIdentifierBoundsAndEscaping(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Functions: []ir.Function{{Source: ir.SourceIdentity{Module: strings.Repeat("module/", 20) + "main.a", Declaration: []string{"Class", "compileSpecificDeclarationName"}}}}}
	if name := (&emitter{program: program}).functionName(0); !strings.Contains(name, "compileSpecificDeclarationName") {
		t.Fatalf("declaration spelling lost: %s", name)
	}

	seen := map[string]bool{}
	for _, identity := range []string{"a-b/f", "a_b/f", "a/b_f", "a/b/f", "你好", strings.Repeat("long", 1000)} {
		name := stableName("adamic_function", identity, identity)
		if !cName.MatchString(name) || len(name) > 110 {
			t.Fatalf("invalid or unbounded: %q", name)
		}
		if seen[name] {
			t.Fatalf("escaping collision %q", name)
		}
		seen[name] = true
		if name != stableName("adamic_function", identity, identity) {
			t.Fatal("nondeterministic name")
		}
	}
}

func TestRegexpSymbolsDoNotMove(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Regexps: []ir.RegExpProgram{{Pattern: "a", Flags: "g", Declarations: `static int adamic_regex_0_data = 1; /* adamic_regex_0 */ static char *adamic_regex_0 = "adamic_regex_0";`}}}
	e := &emitter{program: program}
	original := e.regexDeclarations(0)
	name := e.regexName(0)
	if !strings.Contains(original, name+"_data") || !strings.Contains(original, `"adamic_regex_0"`) || !strings.Contains(original, `/* adamic_regex_0 */`) {
		t.Fatalf("token rename damaged declaration: %s", original)
	}
	program.Regexps = append([]ir.RegExpProgram{{Pattern: "other"}}, program.Regexps...)
	program.Regexps[1].Declarations = strings.ReplaceAll(program.Regexps[1].Declarations, "adamic_regex_0_data", "adamic_regex_1_data")
	program.Regexps[1].Declarations = strings.Replace(program.Regexps[1].Declarations, "*adamic_regex_0 =", "*adamic_regex_1 =", 1)
	if e.regexName(1) != name || e.regexDeclarations(1) != original {
		t.Fatal("preceding bytecode moved regexp symbols")
	}
}

func TestImportedSpecializationsHaveSourceIdentity(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	entry := filepath.Join(directory, "main.a")
	sources := map[string]string{
		entry:                           `import { Box as A } from './a.a'; import { Box as B } from './b.a'; class Holder<T> { readonly value: T; constructor(value: T) { this.value = value; } get(): T { return this.value; } } const first = new Holder<A>(new A(10)); const second = new Holder<B>(new B(10)); console.log(first.get().read().toString()); console.log(second.get().read().toString());`,
		filepath.Join(directory, "a.a"): `export class Box { readonly value: number; constructor(value: number) { this.value = value; } read(): number { return this.value; } }`,
		filepath.Join(directory, "b.a"): `export class Box { readonly value: number; constructor(value: number) { this.value = value; } read(): number { return this.value + 1; } }`,
	}
	for path, source := range sources {
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	program := namedProgram(t, entry, nil)
	names := &emitter{program: program}
	seen := map[string]bool{}
	count := 0
	for i, function := range program.Functions {
		if len(function.Source.Declaration) != 2 || function.Source.Declaration[0] != "Holder" || function.Source.Declaration[1] != "allocate" {
			continue
		}
		name := names.functionName(i)
		if seen[name] {
			t.Fatal("imported same-spelling types collided")
		}
		seen[name] = true
		count++
	}
	// Main shares representation-equivalent object instantiations. Preserve
	// that policy; imported Box methods still require distinct source identities.
	if count != 1 {
		t.Fatalf("got %d shared Holder specializations", count)
	}
	code := C(program)
	relocated := t.TempDir()
	for path, source := range sources {
		if err := os.WriteFile(filepath.Join(relocated, filepath.Base(path)), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if other := C(namedProgram(t, filepath.Join(relocated, "main.a"), nil)); other != code {
		t.Fatal("program root relocation changed emitted text")
	}
	binary := filepath.Join(directory, "imported")
	if err := Build(code, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if got := runWithInput(t, "", binary); got != runWithInput(t, "", "node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", entry) {
		t.Fatalf("output %q", got)
	}
}

func TestMethodAdaptersShareStableDefinition(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals:    []ir.Local{{Name: "this", Type: ir.Object, Function: 0}, {Name: "this", Type: ir.Object, Function: 1}},
		Functions: []ir.Function{{Name: "helper", Parameters: []int{0}, Returns: ir.Number}, {Name: "helper", Parameters: []int{1}, Returns: ir.Number}},
	}
	e := &emitter{program: program, reuse: &reusePlan{}}
	if e.methodThunk(0) != e.methodThunk(1) {
		t.Fatal("identical helper adapters have different names")
	}
	if len(e.declarations) != 1 {
		t.Fatalf("duplicate helper adapters: %d", len(e.declarations))
	}
}

func TestGeneratedRolesDoNotCollide(t *testing.T) {
	t.Parallel()
	entry := filepath.Join(t.TempDir(), "roles.a")
	source := `class Box { readonly value: number; constructor(value: number) { this.value = value; } allocate(value: number): Box { return new Box(value); } initialize(): number { return this.value + 1; } } function outer(): number { return 17; } const fn = outer; console.log((new Box(1).allocate(2).initialize() + fn()).toString());`
	if err := os.WriteFile(entry, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program := namedProgram(t, entry, nil)
	e := &emitter{program: program}
	seen := map[string]bool{}
	for i := range program.Functions {
		name := e.functionName(i)
		if seen[name] {
			t.Fatalf("generated role collided: %s", name)
		}
		seen[name] = true
	}
	binary := filepath.Join(filepath.Dir(entry), "roles")
	if err := Build(C(program), binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if got := runWithInput(t, "", binary); got != runWithInput(t, "", "node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", entry) {
		t.Fatalf("output %q", got)
	}
}

func TestGenericSpecializationsHaveSourceIdentity(t *testing.T) {
	t.Parallel()
	entry := filepath.Join(t.TempDir(), "generic.a")
	source := `function identity<T>(value:T):T {return value;} console.log(identity<number>(7).toString()); console.log(identity<string>('seven'));`
	if err := os.WriteFile(entry, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program := namedProgram(t, entry, nil)
	e := &emitter{program: program}
	seen := map[string]bool{}
	count := 0
	for index, function := range program.Functions {
		if len(function.Source.Declaration) == 1 && function.Source.Declaration[0] == "identity" {
			name := e.functionName(index)
			if seen[name] {
				t.Fatal("concrete specializations share an identity")
			}
			seen[name] = true
			count++
		}
	}
	if count != 2 {
		t.Fatalf("got %d concrete specializations", count)
	}
	binary := filepath.Join(t.TempDir(), "generic")
	if err := Build(C(program), binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if got, want := runWithInput(t, "", binary), runWithInput(t, "", "node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", entry); got != want {
		t.Fatalf("native %q, Node %q", got, want)
	}
}

func TestStableShapeIdentity(t *testing.T) {
	t.Parallel()
	first := &emitter{program: &ir.Program{}}
	second := &emitter{program: &ir.Program{}}
	want := first.shapeOf([]string{"label"}, []ir.Type{ir.String})
	second.shapeOf([]string{"other"}, []ir.Type{ir.Number})
	if got := second.shapeOf([]string{"label"}, []ir.Type{ir.String}); got != want {
		t.Fatal("preceding layout moved shape identity")
	}
	if got := second.shapeOf([]string{"label"}, []ir.Type{ir.Boolean}); got == want {
		t.Fatal("different storage kinds collided")
	}
}

func TestStableLiteralIdentity(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Source: "literal.a", Strings: []string{"kept"}, Main: []ir.Statement{ir.WriteLine{Value: ir.StringConstant{Index: 0}}}}
	before := C(program)
	program.Strings = []string{"preceding", "kept"}
	program.Main = []ir.Statement{ir.WriteLine{Value: ir.StringConstant{Index: 1}}}
	after := C(program)
	name := stringName("kept")
	if stringName("preceding") == name || !strings.Contains(before, "&"+name) || !strings.Contains(after, "&"+name) {
		t.Fatal("literal contents lost their identity")
	}
	a, b := strings.Index(before, "int main("), strings.Index(after, "int main(")
	if before[a:] != after[b:] {
		t.Fatal("preceding literal moved emitted reference")
	}
}

func TestStableCacheOwnersAndCounters(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Functions: []ir.Function{{Name: "first"}, {Name: "second"}}}
	e := &emitter{program: program}
	e.function = &program.Functions[1]
	e.functionIndex = 1
	e.regionValues = map[string]bool{"old": true}
	e.resetCounters()
	if len(e.regionValues) != 0 {
		t.Fatal("region facts outlived their temporary names")
	}
	want := e.cache()
	e.function = &program.Functions[0]
	e.functionIndex = 0
	e.resetCounters()
	other := e.cache()
	e.cache()
	e.temporary()
	e.function = &program.Functions[1]
	e.functionIndex = 1
	e.resetCounters()
	if e.cache() != want || other == want || !strings.HasSuffix(e.temporary(), "_1") {
		t.Fatal("cache owner or body-local counter moved")
	}
}

func TestStableEmissionIsDeterministic(t *testing.T) {
	t.Parallel()
	entry, err := filepath.Abs("../oracle/testdata/async_methods.a")
	if err != nil {
		t.Fatal(err)
	}
	first, second := namedProgram(t, entry, nil), namedProgram(t, entry, nil)
	want := C(first)
	if C(first) != want || C(second) != want {
		t.Fatal("same source changed emitted C")
	}
	h, a, err := splitC(want)
	if err != nil {
		t.Fatal(err)
	}
	h2, b, err := splitC(C(second))
	if err != nil || h != h2 || len(a) != len(b) {
		t.Fatal("same source changed split output")
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("same source changed split unit")
		}
	}
}

func TestStableClassIdentity(t *testing.T) {
	t.Parallel()
	identity := ir.SourceIdentity{Module: "kept.a", Declaration: []string{"Holder"}}
	program := &ir.Program{Classes: []ir.Class{{Name: "Holder", Source: identity, Constructor: -1}}}
	first := (&emitter{program: program}).className(1)
	program.Classes = append([]ir.Class{{Name: "Other", Source: ir.SourceIdentity{Module: "other.a"}, Constructor: -1}}, program.Classes...)
	if (&emitter{program: program}).className(2) != first {
		t.Fatal("preceding class moved identity")
	}
}

func TestStableSchemaIdentity(t *testing.T) {
	t.Parallel()
	first, second := &emitter{program: &ir.Program{}}, &emitter{program: &ir.Program{}}
	schema := &ir.JSONSchema{Kind: "string"}
	want := first.jsonSchema(schema)
	second.jsonSchema(&ir.JSONSchema{Kind: "number"})
	if second.jsonSchema(schema) != want {
		t.Fatal("preceding schema moved identity")
	}
	if first.jsonSchema(schema) != want || len(first.declarations) != 1 {
		t.Fatal("identical schemas did not share identity")
	}
}

func TestStableRequestHandlerIdentity(t *testing.T) {
	t.Parallel()
	entry, err := filepath.Abs("wasm/request.a")
	if err != nil {
		t.Fatal(err)
	}
	program := namedProgram(t, entry, nil)
	match := wasiRequestHandler(C(program))
	if len(match) != 2 {
		t.Fatal("WASI reader lost stable handler identity")
	}
	e := &emitter{program: program}
	found := false
	for i, function := range program.Functions {
		if function.Name == "handleRequest" {
			found = true
			if match[1] != e.functionName(i) {
				t.Fatal("WASI reader selected another function")
			}
		}
	}
	if !found {
		t.Fatal("missing request handler")
	}
}

func TestStableClosureComparatorIdentity(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Locals: []ir.Local{{Type: ir.Number}, {Type: ir.Number}, {Type: ir.Number}}, Functions: []ir.Function{{Closure: true, Parameters: []int{0, 1, 2}, Returns: ir.Number}}}
	e := &emitter{program: program, reuse: &reusePlan{}}
	sort := ir.ArraySort{Callback: ir.MakeClosure{Function: 0}, Element: ir.Number}
	want := e.closureComparator(sort)
	e.temporaries = 100
	if e.closureComparator(sort) != want || len(e.declarations) != 1 {
		t.Fatal("closure comparator lost its content identity")
	}
}

func TestStableDeclarationAncestry(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Functions: []ir.Function{
		{Source: ir.SourceIdentity{Module: "same.a", Declaration: []string{"a_b", "c"}}},
		{Source: ir.SourceIdentity{Module: "same.a", Declaration: []string{"a", "b_c"}}},
	}}
	e := &emitter{program: program}
	if e.sourceReadable(0) != e.sourceReadable(1) {
		t.Fatal("control must collide after readable escaping")
	}
	if e.functionName(0) == e.functionName(1) {
		t.Fatal("declaration ancestry boundaries lost identity")
	}
}

// Insertion shifts the IR indexes of the following declarations, including async
// scheduler helpers. Compare the actual split translation units, not just names.
func TestStableUnitsSurviveMiddleDeclaration(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	entry := filepath.Join(directory, "main.a")
	source := `function first(n: number): number { return n + 1; }
async function middle(n: number): Promise<number> { return await Promise.resolve(n + 2); }
function last(n: number): number { return first(n) + 3; }
console.log(last(4).toString()); console.log((await middle(5)).toString());`
	if err := os.WriteFile(entry, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	before := namedProgram(t, entry, nil)
	_, units, err := splitC(C(before))
	if err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(source, "async function middle", "function inserted(n: number): number { return n * 7; }\nasync function middle", 1)
	after := namedProgram(t, entry, map[string]string{entry: changed})
	_, next, err := splitC(C(after))
	if err != nil {
		t.Fatal(err)
	}
	// compilationUnit has two strings, its filename and its exact emitted bytes.
	found := map[string]compilationUnit{}
	for _, unit := range next {
		found[unit.name] = unit
	}
	checked := 0
	for _, unit := range units {
		if unit.name == "main.c" || unit.name == "state.c" {
			continue
		}
		if got, ok := found[unit.name]; !ok || got != unit {
			t.Fatalf("middle insertion changed emitted unit %s", unit.name)
		}
		checked++
	}
	if checked < len(before.Functions) {
		t.Fatalf("checked %d units for %d functions", checked, len(before.Functions))
	}
	t.Logf("middle declaration insertion preserved all %d existing function and scheduler units byte for byte", checked)
}

func TestStableDecoderAndCallableDescriptors(t *testing.T) {
	t.Parallel()
	first, second := &emitter{program: &ir.Program{}}, &emitter{program: &ir.Program{}}
	schema := ir.JSONDecodeSchema{Nodes: []ir.JSONDecodeNode{{Kind: "string", Of: ir.String, Expected: "string"}}}
	want := first.jsonDecodeSchema(schema)
	second.temporaries = 19
	second.jsonDecodeSchema(ir.JSONDecodeSchema{Nodes: []ir.JSONDecodeNode{{Kind: "number", Of: ir.Number}}})
	if second.jsonDecodeSchema(schema) != want {
		t.Fatal("preceding decoder changed descriptor identity")
	}
	count := len(first.declarations)
	if first.jsonDecodeSchema(schema) != want || len(first.declarations) != count {
		t.Fatal("decoder descriptor not shared")
	}
	if first.jsonDataSchema(schema, true) == want {
		t.Fatal("encoder and decoder descriptor kinds collide")
	}
	signature := first.unionCallableSignature([]ir.Type{ir.String}, ir.Number, "read")
	second.temporaries = 23
	second.unionCallableSignature(nil, ir.Boolean, "other")
	if second.unionCallableSignature([]ir.Type{ir.String}, ir.Number, "read") != signature {
		t.Fatal("preceding callable changed signature identity")
	}
	count = len(first.declarations)
	if first.unionCallableSignature([]ir.Type{ir.String}, ir.Number, "read") != signature || len(first.declarations) != count {
		t.Fatal("callable signature not shared")
	}
}

func TestStablePredicateHelperUnits(t *testing.T) {
	t.Parallel()
	entry, err := filepath.Abs("../lower/testdata/predicates/overload_checked.a")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	before := namedProgram(t, entry, nil)
	wrappers := 0
	for _, function := range before.Functions {
		if strings.Contains(function.Name, "_overload_") && strings.Contains(function.Name, "_check") {
			wrappers++
		}
	}
	if wrappers == 0 {
		t.Fatal("missing checked predicate helper")
	}
	_, units, err := splitC(C(before))
	if err != nil {
		t.Fatal(err)
	}
	after := namedProgram(t, entry, map[string]string{entry: "function unrelated(): number { return 42; }\n" + string(source)})
	_, next, err := splitC(C(after))
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]compilationUnit{}
	for _, unit := range next {
		found[unit.name] = unit
	}
	for _, unit := range units {
		if unit.name == "state.c" || unit.name == "main.c" {
			continue
		}
		if got, ok := found[unit.name]; !ok || got != unit {
			t.Fatalf("insertion moved predicate declaration unit %s", unit.name)
		}
	}
	t.Logf("%d checked predicate helpers retain declaration/content identity after insertion", wrappers)
}

func TestStableGenericArrowObjectShapes(t *testing.T) {
	t.Parallel()
	entry, err := filepath.Abs("../oracle/testdata/generic_arrow_object_shapes.a")
	if err != nil {
		t.Fatal(err)
	}
	program := namedProgram(t, entry, nil)
	_ = C(program)
}

// Copy one multi-module source checkout twice, then cross both copies with two
// unrelated working directories. Relative CLI entries must emit identical C.
func TestStableCFromDifferentCheckoutsAndWorkingDirectories(t *testing.T) {
	fixture, err := filepath.Abs("testdata/stable_checkout")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	checkouts := []string{filepath.Join(root, "checkout-one"), filepath.Join(root, "checkout-two")}
	working := []string{filepath.Join(root, "cwd-one"), filepath.Join(root, "other", "cwd-two")}
	for _, checkout := range checkouts {
		if err := os.MkdirAll(filepath.Join(checkout, "src"), 0700); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"main.a", "library.a"} {
			source, err := os.ReadFile(filepath.Join(fixture, name))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(checkout, "src", name), source, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, cwd := range working {
		if err := os.MkdirAll(cwd, 0700); err != nil {
			t.Fatal(err)
		}
	}
	var want string
	for _, checkout := range checkouts {
		for _, cwd := range working {
			t.Run(filepath.Base(checkout)+"/"+filepath.Base(cwd), func(t *testing.T) {
				t.Chdir(cwd)
				entry, err := filepath.Rel(cwd, filepath.Join(checkout, "src", "main.a"))
				if err != nil {
					t.Fatal(err)
				}
				loaded, err := load.Load([]string{entry})
				if err != nil {
					t.Fatal(err)
				}
				if len(loaded.Files()) != 1 || loaded.FileName(loaded.Files()[0]) != filepath.Join(checkout, "src", "main.a") {
					t.Fatal("Files did not retain the explicit entry ahead of imported and bundled libraries")
				}
				lowered, err := lower.Lower(context.Background(), loaded)
				if err != nil {
					t.Fatal(err)
				}
				got := C(lowered)
				if strings.Contains(got, root) {
					t.Fatal("emitted C contains checkout or working directory")
				}
				if want == "" {
					want = got
				} else if got != want {
					t.Fatal("checkout or working directory changed emitted C bytes")
				}
				t.Logf("%d identical C bytes", len(got))
			})
		}
	}
}
