package native

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
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
	globalNames := map[string]bool{}
	globalEmitter := &emitter{program: before}
	for index, local := range before.Locals {
		if !local.Global {
			continue
		}
		name := globalEmitter.namedLocal(index)
		if globalNames[name] {
			t.Fatalf("distinct module globals share stable name %s", name)
		}
		globalNames[name] = true
	}
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
		secondModule := ""
		for _, function := range before.Functions {
			if function.Position.File == filepath.Join(directory, "a_b.a") {
				secondModule = function.Source.Module
				break
			}
		}
		if moduleBody(original, secondModule) != moduleBody(code, secondModule) {
			t.Fatal("edit moved unrelated module-level code")
		}
		names := &emitter{program: changed}
		checked := 0
		for index, function := range changed.Functions {
			name := names.functionName(index)
			if function.Position.File == module {
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

// Definition identity is independent of the split layout. Keep the historical
// test name so the named identity mutants continue to exercise this control.
func stableDefinitions(t *testing.T, code string) map[string]string {
	t.Helper()
	declarations, err := splitDeclarations(code)
	if err != nil {
		t.Fatal(err)
	}
	definitions := map[string]string{}
	for _, declaration := range declarations {
		if !declaration.function || declaration.name == "main" {
			continue
		}
		tokens := declaration.tokens
		if _, exists := definitions[declaration.name]; exists {
			t.Fatalf("duplicate definition %s", declaration.name)
		}
		definitions[declaration.name] = code[tokens[0].start:tokens[len(tokens)-1].end]
	}
	return definitions
}

func stableDefinitionsSurvive(t *testing.T, before, after string, minimum int) {
	t.Helper()
	original, next := stableDefinitions(t, before), stableDefinitions(t, after)
	if len(original) < minimum {
		t.Fatalf("only %d definitions for %d functions", len(original), minimum)
	}
	for name, definition := range original {
		if got, found := next[name]; !found || got != definition {
			t.Fatalf("insertion changed emitted definition %s", name)
		}
	}
	t.Logf("insertion preserved all %d existing function and scheduler definitions byte for byte", len(original))
}

func stableProgramMatchesNode(t *testing.T, entry, code string) {
	t.Helper()
	loader, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	run := func(name string, arguments ...string) []byte {
		t.Helper()
		command, release := boundedrun.Command(15*time.Second, name, arguments...)
		defer release()
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		if err := command.Run(); err != nil || stderr.Len() != 0 {
			t.Fatalf("%s: %v; stderr %s", name, err, &stderr)
		}
		return stdout.Bytes()
	}
	want := run("node", "--disable-warning=ExperimentalWarning", loader, entry)
	binary := filepath.Join(t.TempDir(), "program")
	if err := Build(code, binary, Options{Split: true, Sanitize: true, Jobs: 2}); err != nil {
		t.Fatal(err)
	}
	if got := run(binary); !bytes.Equal(got, want) {
		t.Fatalf("native %q differs from Node %q", got, want)
	}
}

func TestStableUnitsSurviveMiddleDeclaration(t *testing.T) {
	t.Parallel()
	entry := filepath.Join(t.TempDir(), "main.a")
	source := `function first(n: number): number { return n + 1; }
async function middle(n: number): Promise<number> { return await Promise.resolve(n + 2); }
function last(n: number): number { return first(n) + 3; }
console.log(last(4).toString()); console.log((await middle(5)).toString());`
	if err := os.WriteFile(entry, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	before := namedProgram(t, entry, nil)
	changed := strings.Replace(source, "async function middle", "function inserted(n: number): number { return n * 7; }\nasync function middle", 1)
	after := namedProgram(t, entry, map[string]string{entry: changed})
	original, next := C(before), C(after)
	stableDefinitionsSurvive(t, original, next, len(before.Functions))
	stableProgramMatchesNode(t, entry, original)
	if err := os.WriteFile(entry, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	stableProgramMatchesNode(t, entry, next)
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
	source, err := os.ReadFile("../lower/testdata/predicates/overload_checked.a")
	if err != nil {
		t.Fatal(err)
	}
	// The original fixture deliberately lies for a string, which a checked native
	// predicate rejects. Use its number arm so both programs can agree with Node.
	if strings.Count(string(source), `use("text");`) != 1 {
		t.Fatal("predicate witness lost its call anchor")
	}
	safe := strings.Replace(string(source), `use("text");`, "use(7);", 1)
	entry := filepath.Join(t.TempDir(), "main.ts")
	safe = "\n" + safe // Reserve a line so insertion leaves checked diagnostics unchanged.
	if err := os.WriteFile(entry, []byte(safe), 0600); err != nil {
		t.Fatal(err)
	}
	before := namedProgram(t, entry, nil)
	wrappers := 0
	for _, function := range before.Functions {
		if strings.Contains(function.Name, "_predicate_check") {
			wrappers++
		}
	}
	if wrappers == 0 {
		t.Fatal("missing checked predicate helper")
	}
	changed := "function unrelated(): number { return 42; }" + safe
	if changed == safe {
		t.Fatal("missing middle-insertion anchor")
	}
	after := namedProgram(t, entry, map[string]string{entry: changed})
	original, next := C(before), C(after)
	stableDefinitionsSurvive(t, original, next, len(before.Functions))
	stableProgramMatchesNode(t, entry, original)
	if err := os.WriteFile(entry, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	stableProgramMatchesNode(t, entry, next)
	t.Logf("%d checked predicate helpers retain declaration/content identity after middle insertion", wrappers)
}

// Both modules deliberately leave their predicates unproven in checked .ts mode.
// Equal helper IR names must resolve to distinct semantic C definitions.
func TestStableSameNamedCheckedPredicates(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	for _, name := range []string{"main.ts", "number.ts", "string.ts"} {
		source, err := os.ReadFile(filepath.Join("testdata/stable_checked_predicates", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, name), source, 0600); err != nil {
			t.Fatal(err)
		}
	}
	entry := filepath.Join(directory, "main.ts")
	before := namedProgram(t, entry, nil)
	original := C(before)
	helpers := map[string]bool{}
	names := &emitter{program: before}
	for index, function := range before.Functions {
		if function.Name == "claim_predicate_check" {
			helpers[names.functionName(index)] = true
		}
	}
	if len(helpers) != 2 || before.PredicateChecks.Checked == 0 {
		t.Fatalf("want two same-named checked helpers with distinct stable C names, got %d; checks %+v", len(helpers), before.PredicateChecks)
	}
	source, err := os.ReadFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	// The first line is reserved; leave every checked call's diagnostic location fixed.
	changed := "function unrelated(): number { return 42; }" + string(source)
	after := namedProgram(t, entry, map[string]string{entry: changed})
	next := C(after)
	stableDefinitionsSurvive(t, original, next, len(before.Functions))
	stableProgramMatchesNode(t, entry, original)
	if err := os.WriteFile(entry, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	stableProgramMatchesNode(t, entry, next)
	t.Log("two claim_predicate_check IR names have distinct stable C definitions; sanitized native matches Node before and after insertion")
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

// Cross checkouts and working directories with copies of the entry tree or just
// its entry. Imports outside the copied directory keep referring to the checkout.
// Not parallel: it changes the process working directory
func TestStableCFromDifferentCheckoutsAndWorkingDirectories(t *testing.T) {
	fixture, err := filepath.Abs("testdata/stable_checkout")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	checkouts := []string{filepath.Join(root, "checkout-one"), filepath.Join(root, "deeper", "checkout-two")}
	working := []string{filepath.Join(root, "cwd-one"), filepath.Join(root, "other", "cwd-two")}
	entries := []string{}
	for _, checkout := range checkouts {
		for _, name := range []string{"src", "shared/left", "shared/right", "private/tree", "private/entry"} {
			if err := os.MkdirAll(filepath.Join(checkout, name), 0700); err != nil {
				t.Fatal(err)
			}
		}
		main, err := os.ReadFile(filepath.Join(fixture, "main.a"))
		if err != nil {
			t.Fatal(err)
		}
		// Two different definitions have the same basename and declaration name.
		source := string(main) + "\nimport { value as left } from '../shared/left/library.a';\nimport { value as right } from '../shared/right/library.a';\nconsole.log((left() + right()).toString());\n"
		library, err := os.ReadFile(filepath.Join(fixture, "library.a"))
		if err != nil {
			t.Fatal(err)
		}
		files := map[string]string{"src/main.a": source, "src/library.a": string(library), "shared/left/library.a": "export function value(): number { return 11; }\n", "shared/right/library.a": "export function value(): number { return 22; }\n"}
		for name, text := range files {
			if err := os.WriteFile(filepath.Join(checkout, name), []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
		}
		entries = append(entries, filepath.Join(checkout, "src/main.a"))
		// This is mutantPort's shape: copied local files, imports into a shared tree.
		copied := source
		for _, name := range []string{"left", "right"} {
			copied = strings.ReplaceAll(copied, "../shared/"+name+"/library.a", filepath.ToSlash(filepath.Join(checkout, "shared", name, "library.a")))
		}
		if err := os.WriteFile(filepath.Join(checkout, "private/tree/main.a"), []byte(copied), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(checkout, "private/tree/library.a"), library, 0600); err != nil {
			t.Fatal(err)
		}
		entries = append(entries, filepath.Join(checkout, "private/tree/main.a"))
		// Also move only the entry: a formerly adjacent import now stays outside it.
		copied = strings.ReplaceAll(copied, "./library.a", filepath.ToSlash(filepath.Join(checkout, "src/library.a")))
		if err := os.WriteFile(filepath.Join(checkout, "private/entry/main.a"), []byte(copied), 0600); err != nil {
			t.Fatal(err)
		}
		entries = append(entries, filepath.Join(checkout, "private/entry/main.a"))
	}
	for _, cwd := range working {
		if err := os.MkdirAll(cwd, 0700); err != nil {
			t.Fatal(err)
		}
	}
	var wantC, wantJS string
	for _, absolute := range entries {
		for _, cwd := range working {
			t.Run(filepath.ToSlash(strings.TrimPrefix(absolute, root))+"/"+filepath.Base(cwd), func(t *testing.T) {
				t.Chdir(cwd)
				entry, err := filepath.Rel(cwd, absolute)
				if err != nil {
					t.Fatal(err)
				}
				loaded, err := load.Load([]string{entry})
				if err != nil {
					t.Fatal(err)
				}
				if len(loaded.Files()) != 1 || loaded.FileName(loaded.Files()[0]) != absolute {
					t.Fatal("Files did not retain the explicit entry")
				}
				lowered, err := lower.Lower(context.Background(), loaded)
				if err != nil {
					t.Fatal(err)
				}
				gotC, gotJS := C(lowered), javascript.JavaScript(lowered)
				if strings.Contains(gotC, root) || strings.Contains(gotJS, root) {
					t.Fatal("emitted code contains checkout or working directory")
				}
				if wantC == "" {
					wantC, wantJS = gotC, gotJS
				} else if gotC != wantC || gotJS != wantJS {
					t.Fatal("source path or working directory changed C or JS bytes")
				}
				t.Logf("%d C bytes and %d JS bytes identical", len(gotC), len(gotJS))
			})
		}
	}
	stableProgramMatchesNode(t, entries[0], wantC)
}

// A module generic called during class lowering must not inherit its caller's
// class specialization in the identity of its nested callback.
func TestStableGenericCallbackIgnoresCallingClass(t *testing.T) {
	t.Parallel()
	entry := filepath.Join(t.TempDir(), "main.a")
	source := `function collect<T>(item: T): T[] {
 const values: T[] = [];
 const add = () => { values.push(item); };
 add();
 return values;
}
class Box<T> {
 readonly value: T;
 constructor(value: T) { this.value = value; }
 run(): T[] { return collect(this.value); }
}
function run<T>(value: T): T[] { return new Box<T>(value).run(); }
console.log((collect(11)[0] ?? 0).toString());
console.log((run(22)[0] ?? 0).toString());`
	if err := os.WriteFile(entry, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program := namedProgram(t, entry, nil)
	code := C(program)
	instances := 0
	keys := map[string]bool{}
	for _, function := range program.Functions {
		if len(function.Source.Declaration) == 1 && function.Source.Declaration[0] == "collect" {
			instances++
			keys[sourceKey(function.Source)] = true
		}
		if len(function.Source.Declaration) > 1 && function.Source.Declaration[0] == "collect" && strings.Contains(function.Source.Specialization, "Box") {
			t.Fatalf("module callback inherited calling class: %#v", function.Source)
		}
	}
	// Separate generic cache buckets lower the same concrete module generic twice:
	// once directly and once while its same-file generic class caller is active.
	if instances != 2 || len(keys) != 1 {
		t.Fatalf("collect: %d instances with %d source keys, want 2 with one shared key", instances, len(keys))
	}
	truth := runWithInput(t, "", "node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", entry)
	if truth != "11\n22\n" {
		t.Fatalf("Node fixture output: %q", truth)
	}
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(filepath.Dir(entry), "calling-class")
		if err := Build(code, binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != truth {
			t.Fatalf("native sanitize=%t: %q, Node: %q", sanitize, got, truth)
		}
		t.Logf("two same-file collect instances; native sanitize=%t matches Node: %q", sanitize, truth)
	}
}
