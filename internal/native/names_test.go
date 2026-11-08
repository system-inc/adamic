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

func namedProgram(test *testing.T, entry string, overlay map[string]string) *ir.Program {
	test.Helper()
	loaded, loadOverlayError := load.LoadOverlay([]string{entry}, overlay)
	if loadOverlayError != nil {
		test.Fatal(loadOverlayError)
	}
	program, lowerError := lower.Lower(context.Background(), loaded)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	return program
}

func namedBody(test *testing.T, program *ir.Program, code string, index int) string {
	test.Helper()
	emitter := &emitter{program: program}
	start := strings.Index(code, "static "+emitter.signature(index)+" {\n")
	if start < 0 {
		test.Fatalf("body absent: %s", emitter.functionName(index))
	}
	end := strings.Index(code[start:], "\n}\n")
	if end < 0 {
		test.Fatal("unterminated body")
	}
	return code[start : start+end+3]
}

func TestSourceNamesDoNotMove(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	entry := filepath.Join(directory, "main.a")
	module := filepath.Join(directory, "a-b.a")
	sources := map[string]string{
		entry:  `import { same as first } from './a-b.a'; import { same as second } from './a_b.a'; console.log((first(1) + second(2)).toString());`,
		module: `export const marker = [14].length; export const container = { get value(): number { return 14; } }; export function same(value: number): number { const callback = () => value + 14; return callback(); }`,
		filepath.Join(
			directory,
			"a_b.a",
		): `export const marker = [14].length; export const container = { get value(): number { return 14; } }; export function same(value: number): number { const callback = () => value + 14; return callback(); }`,
	}
	for path, source := range sources {
		if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	before := namedProgram(test, entry, nil)
	globalNames := map[string]bool{}
	globalEmitter := &emitter{program: before}
	for index, local := range before.Locals {
		if !local.Global {
			continue
		}
		name := globalEmitter.namedLocal(index)
		if globalNames[name] {
			test.Fatalf("distinct module globals share stable name %s", name)
		}
		globalNames[name] = true
	}
	original := C(before)
	bodies := map[string]string{}
	e := &emitter{program: before}
	for index, function := range before.Functions {
		if function.Source.Module == "" {
			test.Fatalf("missing source for %s", function.Name)
		}
		name := e.functionName(index)
		if _, found := bodies[name]; found {
			test.Fatalf("colliding declaration %s", name)
		}
		bodies[name] = namedBody(test, before, original, index)
	}
	for _, edit := range []string{
		"function unrelated(): number { return 0; }\n" + sources[module],
		"export const added = [14].length;\n" + sources[module],
		strings.Replace(sources[module], "const callback", "const temporary = [14].length; const callback", 1),
		strings.Replace(sources[module], "value + 14", "value + [14].length", 1),
	} {
		changed := namedProgram(test, entry, map[string]string{module: edit})
		code := C(changed)
		moduleBody := func(code, module string) string {
			declarations, splitDeclarationsError := splitDeclarations(code)
			if splitDeclarationsError != nil {
				test.Fatal(splitDeclarationsError)
			}
			var body strings.Builder
			previous := 0
			for _, declaration := range declarations {
				begin := declaration.tokens[0].start
				marker := "// adamic-module "
				owner, marked := "", false
				var unquoteError error
				if at := strings.LastIndex(code[previous:begin], marker); at >= 0 {
					text := strings.TrimSpace(code[previous:begin][at+len(marker):])
					owner, unquoteError = strconv.Unquote(text)
					marked = true
				}
				if unquoteError != nil {
					test.Fatal(unquoteError)
				}
				end := declaration.tokens[len(declaration.tokens)-1].end
				if declaration.function &&
					marked &&
					owner == module &&
					strings.HasPrefix(declaration.name, "adamic_initialize_") {
					body.WriteString(code[begin:end])
				}
				previous = end
			}
			if body.Len() == 0 {
				test.Fatalf("missing initializer %s", module)
			}
			return body.String()
		}
		secondModule := ""
		for _, function := range before.Functions {
			if function.Position.File == filepath.Join(directory, "a_b.a") {
				secondModule = function.Source.Module
				break
			}
		}
		if moduleBody(original, secondModule) != moduleBody(code, secondModule) {
			test.Fatal("edit moved unrelated module-level code")
		}
		names := &emitter{program: changed}
		checked := 0
		for index, function := range changed.Functions {
			name := names.functionName(index)
			if function.Position.File == module {
				continue
			}
			if body, found := bodies[name]; !found || body != namedBody(test, changed, code, index) {
				test.Fatalf("edit moved unrelated function %s", name)
			}
			checked++
		}
		if checked == 0 {
			test.Fatal("no unaffected declarations checked")
		}
	}
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(directory, "names")
		if buildError := Build(original, binary, Options{Sanitize: sanitize}); buildError != nil {
			test.Fatal(buildError)
		}
		if got := runWithInput(test, "", binary); got != runWithInput(
			test,
			"",
			"node",
			"--disable-warning=ExperimentalWarning",
			"../../oracle/node.mjs",
			entry,
		) {
			test.Fatalf("output %q", got)
		}
	}
}

func TestStableIdentifierBoundsAndEscaping(test *testing.T) {
	test.Parallel()
	program := &ir.Program{Functions: []ir.Function{{Source: ir.SourceIdentity{
		Module:      strings.Repeat("module/", 20) + "main.a",
		Declaration: []string{"Class", "compileSpecificDeclarationName"},
	}}}}
	if name := (&emitter{program: program}).functionName(0); !strings.Contains(name, "compileSpecificDeclarationName") {
		test.Fatalf("declaration spelling lost: %s", name)
	}

	seen := map[string]bool{}
	for _, identity := range []string{"a-b/f", "a_b/f", "a/b_f", "a/b/f", "你好", strings.Repeat("long", 1000)} {
		name := stableName("adamic_function", identity, identity)
		if !cName.MatchString(name) || len(name) > 110 {
			test.Fatalf("invalid or unbounded: %q", name)
		}
		if seen[name] {
			test.Fatalf("escaping collision %q", name)
		}
		seen[name] = true
		if name != stableName("adamic_function", identity, identity) {
			test.Fatal("nondeterministic name")
		}
	}
}

func TestRegexpSymbolsDoNotMove(test *testing.T) {
	test.Parallel()
	program := &ir.Program{Regexps: []ir.RegExpProgram{{
		Pattern:      "a",
		Flags:        "g",
		Declarations: `static int adamic_regex_0_data = 1; /* adamic_regex_0 */ static char *adamic_regex_0 = "adamic_regex_0";`,
	}}}
	emitter := &emitter{program: program}
	original := emitter.regexDeclarations(0)
	name := emitter.regexName(0)
	if !strings.Contains(original, name+"_data") ||
		!strings.Contains(original, `"adamic_regex_0"`) ||
		!strings.Contains(original, `/* adamic_regex_0 */`) {
		test.Fatalf("token rename damaged declaration: %s", original)
	}
	program.Regexps = append([]ir.RegExpProgram{{Pattern: "other"}}, program.Regexps...)
	program.Regexps[1].Declarations = strings.ReplaceAll(
		program.Regexps[1].Declarations,
		"adamic_regex_0_data",
		"adamic_regex_1_data",
	)
	program.Regexps[1].Declarations = strings.Replace(
		program.Regexps[1].Declarations,
		"*adamic_regex_0 =",
		"*adamic_regex_1 =",
		1,
	)
	if emitter.regexName(1) != name || emitter.regexDeclarations(1) != original {
		test.Fatal("preceding bytecode moved regexp symbols")
	}
}

func TestImportedSpecializationsHaveSourceIdentity(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	entry := filepath.Join(directory, "main.a")
	sources := map[string]string{
		entry: `import { Box as A } from './a.a'; import { Box as B } from './b.a'; class Holder<T> { readonly value: T; constructor(value: T) { this.value = value; } get(): T { return this.value; } } const first = new Holder<A>(new A(10)); const second = new Holder<B>(new B(10)); console.log(first.get().read().toString()); console.log(second.get().read().toString());`,
		filepath.Join(
			directory,
			"a.a",
		): `export class Box { readonly value: number; constructor(value: number) { this.value = value; } read(): number { return this.value; } }`,
		filepath.Join(
			directory,
			"b.a",
		): `export class Box { readonly value: number; constructor(value: number) { this.value = value; } read(): number { return this.value + 1; } }`,
	}
	for path, source := range sources {
		if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	program := namedProgram(test, entry, nil)
	names := &emitter{program: program}
	seen := map[string]bool{}
	count := 0
	for index, function := range program.Functions {
		if len(function.Source.Declaration) != 2 ||
			function.Source.Declaration[0] != "Holder" ||
			function.Source.Declaration[1] != "allocate" {
			continue
		}
		name := names.functionName(index)
		if seen[name] {
			test.Fatal("imported same-spelling types collided")
		}
		seen[name] = true
		count++
	}
	// Main shares representation-equivalent object instantiations. Preserve
	// that policy; imported Box methods still require distinct source identities.
	if count != 1 {
		test.Fatalf("got %d shared Holder specializations", count)
	}
	code := C(program)
	relocated := test.TempDir()
	for path, source := range sources {
		if writeError := os.WriteFile(
			filepath.Join(relocated, filepath.Base(path)),
			[]byte(source),
			0644,
		); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if other := C(namedProgram(test, filepath.Join(relocated, "main.a"), nil)); other != code {
		test.Fatal("program root relocation changed emitted text")
	}
	binary := filepath.Join(directory, "imported")
	if buildError := Build(code, binary, Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	if got := runWithInput(
		test,
		"",
		binary,
	); got != runWithInput(test, "", "node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", entry) {
		test.Fatalf("output %q", got)
	}
}

func TestMethodAdaptersShareStableDefinition(test *testing.T) {
	test.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{
			{Name: "this", Type: ir.Object, Function: 0},
			{Name: "this", Type: ir.Object, Function: 1},
		},
		Functions: []ir.Function{
			{Name: "helper", Parameters: []int{0}, Returns: ir.Number},
			{Name: "helper", Parameters: []int{1}, Returns: ir.Number},
		},
	}
	emitter := &emitter{program: program, reuse: &reusePlan{}}
	if emitter.methodThunk(0) != emitter.methodThunk(1) {
		test.Fatal("identical helper adapters have different names")
	}
	if len(emitter.declarations) != 1 {
		test.Fatalf("duplicate helper adapters: %d", len(emitter.declarations))
	}
}

func TestGeneratedRolesDoNotCollide(test *testing.T) {
	test.Parallel()
	entry := filepath.Join(test.TempDir(), "roles.a")
	source := `class Box { readonly value: number; constructor(value: number) { this.value = value; } allocate(value: number): Box { return new Box(value); } initialize(): number { return this.value + 1; } } function outer(): number { return 17; } const fn = outer; console.log((new Box(1).allocate(2).initialize() + fn()).toString());`
	if writeError := os.WriteFile(entry, []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	program := namedProgram(test, entry, nil)
	emitter := &emitter{program: program}
	seen := map[string]bool{}
	for index := range program.Functions {
		name := emitter.functionName(index)
		if seen[name] {
			test.Fatalf("generated role collided: %s", name)
		}
		seen[name] = true
	}
	binary := filepath.Join(filepath.Dir(entry), "roles")
	if buildError := Build(C(program), binary, Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	if got := runWithInput(
		test,
		"",
		binary,
	); got != runWithInput(test, "", "node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", entry) {
		test.Fatalf("output %q", got)
	}
}

func TestGenericSpecializationsHaveSourceIdentity(test *testing.T) {
	test.Parallel()
	entry := filepath.Join(test.TempDir(), "generic.a")
	source := `function identity<T>(value:T):T {return value;} console.log(identity<number>(7).toString()); console.log(identity<string>('seven'));`
	if writeError := os.WriteFile(entry, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	program := namedProgram(test, entry, nil)
	emitter := &emitter{program: program}
	seen := map[string]bool{}
	count := 0
	for index, function := range program.Functions {
		if len(function.Source.Declaration) == 1 && function.Source.Declaration[0] == "identity" {
			name := emitter.functionName(index)
			if seen[name] {
				test.Fatal("concrete specializations share an identity")
			}
			seen[name] = true
			count++
		}
	}
	if count != 2 {
		test.Fatalf("got %d concrete specializations", count)
	}
	binary := filepath.Join(test.TempDir(), "generic")
	if buildError := Build(C(program), binary, Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	if got, want := runWithInput(test, "", binary), runWithInput(
		test,
		"",
		"node",
		"--disable-warning=ExperimentalWarning",
		"../../oracle/node.mjs",
		entry,
	); got != want {
		test.Fatalf("native %q, Node %q", got, want)
	}
}

func TestStableShapeIdentity(test *testing.T) {
	test.Parallel()
	first := &emitter{program: &ir.Program{}}
	second := &emitter{program: &ir.Program{}}
	want := first.shapeOf([]string{"label"}, []ir.Type{ir.String})
	second.shapeOf([]string{"other"}, []ir.Type{ir.Number})
	if got := second.shapeOf([]string{"label"}, []ir.Type{ir.String}); got != want {
		test.Fatal("preceding layout moved shape identity")
	}
	if got := second.shapeOf([]string{"label"}, []ir.Type{ir.Boolean}); got == want {
		test.Fatal("different storage kinds collided")
	}
}

func TestStableLiteralIdentity(test *testing.T) {
	test.Parallel()
	program := &ir.Program{
		Source:  "literal.a",
		Strings: []string{"kept"},
		Main:    []ir.Statement{ir.WriteLine{Value: ir.StringConstant{Index: 0}}},
	}
	before := C(program)
	program.Strings = []string{"preceding", "kept"}
	program.Main = []ir.Statement{ir.WriteLine{Value: ir.StringConstant{Index: 1}}}
	after := C(program)
	name := stringName("kept")
	if stringName("preceding") == name || !strings.Contains(before, "&"+name) || !strings.Contains(after, "&"+name) {
		test.Fatal("literal contents lost their identity")
	}
	a, b := strings.Index(before, "int main("), strings.Index(after, "int main(")
	if before[a:] != after[b:] {
		test.Fatal("preceding literal moved emitted reference")
	}
}

func TestStableCacheOwnersAndCounters(test *testing.T) {
	test.Parallel()
	program := &ir.Program{Functions: []ir.Function{{Name: "first"}, {Name: "second"}}}
	emitter := &emitter{program: program}
	emitter.function = &program.Functions[1]
	emitter.functionIndex = 1
	emitter.regionValues = map[string]bool{"old": true}
	emitter.resetCounters()
	if len(emitter.regionValues) != 0 {
		test.Fatal("region facts outlived their temporary names")
	}
	want := emitter.cache()
	emitter.function = &program.Functions[0]
	emitter.functionIndex = 0
	emitter.resetCounters()
	other := emitter.cache()
	emitter.cache()
	emitter.temporary()
	emitter.function = &program.Functions[1]
	emitter.functionIndex = 1
	emitter.resetCounters()
	if emitter.cache() != want || other == want || !strings.HasSuffix(emitter.temporary(), "_1") {
		test.Fatal("cache owner or body-local counter moved")
	}
}

func TestStableEmissionIsDeterministic(test *testing.T) {
	test.Parallel()
	entry, absolutePathError := filepath.Abs("../oracle/testdata/async_methods.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	first, second := namedProgram(test, entry, nil), namedProgram(test, entry, nil)
	want := C(first)
	if C(first) != want || C(second) != want {
		test.Fatal("same source changed emitted C")
	}
	h, a, splitCError := splitC(want)
	if splitCError != nil {
		test.Fatal(splitCError)
	}
	h2, b, splitCError := splitC(C(second))
	if splitCError != nil || h != h2 || len(a) != len(b) {
		test.Fatal("same source changed split output")
	}
	for index := range a {
		if a[index] != b[index] {
			test.Fatal("same source changed split unit")
		}
	}
}

func TestStableClassIdentity(test *testing.T) {
	test.Parallel()
	identity := ir.SourceIdentity{Module: "kept.a", Declaration: []string{"Holder"}}
	program := &ir.Program{Classes: []ir.Class{{Name: "Holder", Source: identity, Constructor: -1}}}
	first := (&emitter{program: program}).className(1)
	program.Classes = append(
		[]ir.Class{{Name: "Other", Source: ir.SourceIdentity{Module: "other.a"}, Constructor: -1}},
		program.Classes...,
	)
	if (&emitter{program: program}).className(2) != first {
		test.Fatal("preceding class moved identity")
	}
}

func TestStableSchemaIdentity(test *testing.T) {
	test.Parallel()
	first, second := &emitter{program: &ir.Program{}}, &emitter{program: &ir.Program{}}
	schema := &ir.JSONSchema{Kind: "string"}
	want := first.jsonSchema(schema)
	second.jsonSchema(&ir.JSONSchema{Kind: "number"})
	if second.jsonSchema(schema) != want {
		test.Fatal("preceding schema moved identity")
	}
	if first.jsonSchema(schema) != want || len(first.declarations) != 1 {
		test.Fatal("identical schemas did not share identity")
	}
}

func TestStableRequestHandlerIdentity(test *testing.T) {
	test.Parallel()
	entry, absolutePathError := filepath.Abs("wasm/request.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program := namedProgram(test, entry, nil)
	match := wasiRequestHandler(C(program))
	if len(match) != 2 {
		test.Fatal("WASI reader lost stable handler identity")
	}
	emitter := &emitter{program: program}
	found := false
	for index, function := range program.Functions {
		if function.Name == "handleRequest" {
			found = true
			if match[1] != emitter.functionName(index) {
				test.Fatal("WASI reader selected another function")
			}
		}
	}
	if !found {
		test.Fatal("missing request handler")
	}
}

func TestStableClosureComparatorIdentity(test *testing.T) {
	test.Parallel()
	program := &ir.Program{
		Locals:    []ir.Local{{Type: ir.Number}, {Type: ir.Number}, {Type: ir.Number}},
		Functions: []ir.Function{{Closure: true, Parameters: []int{0, 1, 2}, Returns: ir.Number}},
	}
	emitter := &emitter{program: program, reuse: &reusePlan{}}
	sort := ir.ArraySort{Callback: ir.MakeClosure{Function: 0}, Element: ir.Number}
	want := emitter.closureComparator(sort)
	emitter.temporaries = 100
	if emitter.closureComparator(sort) != want || len(emitter.declarations) != 1 {
		test.Fatal("closure comparator lost its content identity")
	}
}

func TestStableDeclarationAncestry(test *testing.T) {
	test.Parallel()
	program := &ir.Program{Functions: []ir.Function{
		{Source: ir.SourceIdentity{Module: "same.a", Declaration: []string{"a_b", "c"}}},
		{Source: ir.SourceIdentity{Module: "same.a", Declaration: []string{"a", "b_c"}}},
	}}
	emitter := &emitter{program: program}
	if emitter.sourceReadable(0) != emitter.sourceReadable(1) {
		test.Fatal("control must collide after readable escaping")
	}
	if emitter.functionName(0) == emitter.functionName(1) {
		test.Fatal("declaration ancestry boundaries lost identity")
	}
}

// Definition identity is independent of the split layout. Keep the historical
// test name so the named identity mutants continue to exercise this control.
func stableDefinitions(test *testing.T, code string) map[string]string {
	test.Helper()
	declarations, splitDeclarationsError := splitDeclarations(code)
	if splitDeclarationsError != nil {
		test.Fatal(splitDeclarationsError)
	}
	definitions := map[string]string{}
	for _, declaration := range declarations {
		if !declaration.function || declaration.name == "main" {
			continue
		}
		tokens := declaration.tokens
		if _, exists := definitions[declaration.name]; exists {
			test.Fatalf("duplicate definition %s", declaration.name)
		}
		definitions[declaration.name] = code[tokens[0].start:tokens[len(tokens)-1].end]
	}
	return definitions
}

func stableDefinitionsSurvive(test *testing.T, before, after string, minimum int) {
	test.Helper()
	original, next := stableDefinitions(test, before), stableDefinitions(test, after)
	if len(original) < minimum {
		test.Fatalf("only %d definitions for %d functions", len(original), minimum)
	}
	for name, definition := range original {
		if got, found := next[name]; !found || got != definition {
			test.Fatalf("insertion changed emitted definition %s", name)
		}
	}
	test.Logf("insertion preserved all %d existing function and scheduler definitions byte for byte", len(original))
}

func stableProgramMatchesNode(test *testing.T, entry, code string) {
	test.Helper()
	loader, absolutePathError := filepath.Abs("../../oracle/node.mjs")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	run := func(name string, arguments ...string) []byte {
		test.Helper()
		command, release := boundedrun.Command(15*time.Second, name, arguments...)
		defer release()
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		if runError := command.Run(); runError != nil || stderr.Len() != 0 {
			test.Fatalf("%s: %v; stderr %s", name, runError, &stderr)
		}
		return stdout.Bytes()
	}
	want := run("node", "--disable-warning=ExperimentalWarning", loader, entry)
	binary := filepath.Join(test.TempDir(), "program")
	if buildError := Build(code, binary, Options{Split: true, Sanitize: true, Jobs: 2}); buildError != nil {
		test.Fatal(buildError)
	}
	if got := run(binary); !bytes.Equal(got, want) {
		test.Fatalf("native %q differs from Node %q", got, want)
	}
}

func TestStableUnitsSurviveMiddleDeclaration(test *testing.T) {
	test.Parallel()
	entry := filepath.Join(test.TempDir(), "main.a")
	source := `function first(n: number): number { return n + 1; }
async function middle(n: number): Promise<number> { return await Promise.resolve(n + 2); }
function last(n: number): number { return first(n) + 3; }
console.log(last(4).toString()); console.log((await middle(5)).toString());`
	if writeError := os.WriteFile(entry, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	before := namedProgram(test, entry, nil)
	changed := strings.Replace(
		source,
		"async function middle",
		"function inserted(n: number): number { return n * 7; }\nasync function middle",
		1,
	)
	after := namedProgram(test, entry, map[string]string{entry: changed})
	original, next := C(before), C(after)
	stableDefinitionsSurvive(test, original, next, len(before.Functions))
	stableProgramMatchesNode(test, entry, original)
	if writeError := os.WriteFile(entry, []byte(changed), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	stableProgramMatchesNode(test, entry, next)
}

func TestStableDecoderAndCallableDescriptors(test *testing.T) {
	test.Parallel()
	first, second := &emitter{program: &ir.Program{}}, &emitter{program: &ir.Program{}}
	schema := ir.JSONDecodeSchema{Nodes: []ir.JSONDecodeNode{{Kind: "string", Of: ir.String, Expected: "string"}}}
	want := first.jsonDecodeSchema(schema)
	second.temporaries = 19
	second.jsonDecodeSchema(ir.JSONDecodeSchema{Nodes: []ir.JSONDecodeNode{{Kind: "number", Of: ir.Number}}})
	if second.jsonDecodeSchema(schema) != want {
		test.Fatal("preceding decoder changed descriptor identity")
	}
	count := len(first.declarations)
	if first.jsonDecodeSchema(schema) != want || len(first.declarations) != count {
		test.Fatal("decoder descriptor not shared")
	}
	if first.jsonDataSchema(schema, true) == want {
		test.Fatal("encoder and decoder descriptor kinds collide")
	}
	signature := first.unionCallableSignature([]ir.Type{ir.String}, ir.Number, "read")
	second.temporaries = 23
	second.unionCallableSignature(nil, ir.Boolean, "other")
	if second.unionCallableSignature([]ir.Type{ir.String}, ir.Number, "read") != signature {
		test.Fatal("preceding callable changed signature identity")
	}
	count = len(first.declarations)
	if first.unionCallableSignature([]ir.Type{ir.String}, ir.Number, "read") != signature ||
		len(first.declarations) != count {
		test.Fatal("callable signature not shared")
	}
}

func TestStablePredicateHelperUnits(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("../lower/testdata/predicates/overload_checked.a")
	if readError != nil {
		test.Fatal(readError)
	}
	// The original fixture deliberately lies for a string, which a checked native
	// predicate rejects. Use its number arm so both programs can agree with Node.
	if strings.Count(string(source), `use("text");`) != 1 {
		test.Fatal("predicate witness lost its call anchor")
	}
	safe := strings.Replace(string(source), `use("text");`, "use(7);", 1)
	entry := filepath.Join(test.TempDir(), "main.ts")
	safe = "\n" + safe // Reserve a line so insertion leaves checked diagnostics unchanged.
	if writeError := os.WriteFile(entry, []byte(safe), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	before := namedProgram(test, entry, nil)
	wrappers := 0
	for _, function := range before.Functions {
		if strings.Contains(function.Name, "_predicate_check") {
			wrappers++
		}
	}
	if wrappers == 0 {
		test.Fatal("missing checked predicate helper")
	}
	changed := "function unrelated(): number { return 42; }" + safe
	if changed == safe {
		test.Fatal("missing middle-insertion anchor")
	}
	after := namedProgram(test, entry, map[string]string{entry: changed})
	original, next := C(before), C(after)
	stableDefinitionsSurvive(test, original, next, len(before.Functions))
	stableProgramMatchesNode(test, entry, original)
	if writeError := os.WriteFile(entry, []byte(changed), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	stableProgramMatchesNode(test, entry, next)
	test.Logf("%d checked predicate helpers retain declaration/content identity after middle insertion", wrappers)
}

// Both modules deliberately leave their predicates unproven in checked .ts mode.
// Equal helper IR names must resolve to distinct semantic C definitions.
func TestStableSameNamedCheckedPredicates(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	for _, name := range []string{"main.ts", "number.ts", "string.ts"} {
		source, readError := os.ReadFile(filepath.Join("testdata/stable_checked_predicates", name))
		if readError != nil {
			test.Fatal(readError)
		}
		if writeError := os.WriteFile(filepath.Join(directory, name), source, 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	entry := filepath.Join(directory, "main.ts")
	before := namedProgram(test, entry, nil)
	original := C(before)
	helpers := map[string]bool{}
	names := &emitter{program: before}
	for index, function := range before.Functions {
		if function.Name == "claim_predicate_check" {
			helpers[names.functionName(index)] = true
		}
	}
	if len(helpers) != 2 || before.PredicateChecks.Checked == 0 {
		test.Fatalf(
			"want two same-named checked helpers with distinct stable C names, got %d; checks %+v",
			len(helpers),
			before.PredicateChecks,
		)
	}
	source, readError := os.ReadFile(entry)
	if readError != nil {
		test.Fatal(readError)
	}
	// The first line is reserved; leave every checked call's diagnostic location fixed.
	changed := "function unrelated(): number { return 42; }" + string(source)
	after := namedProgram(test, entry, map[string]string{entry: changed})
	next := C(after)
	stableDefinitionsSurvive(test, original, next, len(before.Functions))
	stableProgramMatchesNode(test, entry, original)
	if writeError := os.WriteFile(entry, []byte(changed), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	stableProgramMatchesNode(test, entry, next)
	test.Log(
		"two claim_predicate_check IR names have distinct stable C definitions; sanitized native matches Node before and after insertion",
	)
}

func TestStableGenericArrowObjectShapes(test *testing.T) {
	test.Parallel()
	entry, absolutePathError := filepath.Abs("../oracle/testdata/generic_arrow_object_shapes.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program := namedProgram(test, entry, nil)
	_ = C(program)
}

// Cross checkouts and working directories with copies of the entry tree or just
// its entry. Imports outside the copied directory keep referring to the checkout.
// Not parallel: it changes the process working directory
func TestStableCFromDifferentCheckoutsAndWorkingDirectories(test *testing.T) {
	fixture, absolutePathError := filepath.Abs("testdata/stable_checkout")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	root := test.TempDir()
	checkouts := []string{filepath.Join(root, "checkout-one"), filepath.Join(root, "deeper", "checkout-two")}
	working := []string{filepath.Join(root, "cwd-one"), filepath.Join(root, "other", "cwd-two")}
	entries := []string{}
	for _, checkout := range checkouts {
		for _, name := range []string{"src", "shared/left", "shared/right", "private/tree", "private/entry"} {
			if createDirectoryError := os.MkdirAll(filepath.Join(checkout, name), 0700); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
		}
		main, readError := os.ReadFile(filepath.Join(fixture, "main.a"))
		if readError != nil {
			test.Fatal(readError)
		}
		// Two different definitions have the same basename and declaration name.
		source := string(main) +
			"\nimport { value as left } from '../shared/left/library.a';\nimport { value as right } from '../shared/right/library.a';\nconsole.log((left() + right()).toString());\n"
		library, readError := os.ReadFile(filepath.Join(fixture, "library.a"))
		if readError != nil {
			test.Fatal(readError)
		}
		files := map[string]string{
			"src/main.a":             source,
			"src/library.a":          string(library),
			"shared/left/library.a":  "export function value(): number { return 11; }\n",
			"shared/right/library.a": "export function value(): number { return 22; }\n",
		}
		for name, text := range files {
			if writeError := os.WriteFile(filepath.Join(checkout, name), []byte(text), 0600); writeError != nil {
				test.Fatal(writeError)
			}
		}
		entries = append(entries, filepath.Join(checkout, "src/main.a"))
		// This is mutantPort's shape: copied local files, imports into a shared tree.
		copied := source
		for _, name := range []string{"left", "right"} {
			copied = strings.ReplaceAll(
				copied,
				"../shared/"+name+"/library.a",
				filepath.ToSlash(filepath.Join(checkout, "shared", name, "library.a")),
			)
		}
		if writeError := os.WriteFile(
			filepath.Join(checkout, "private/tree/main.a"),
			[]byte(copied),
			0600,
		); writeError != nil {
			test.Fatal(writeError)
		}
		if writeError := os.WriteFile(
			filepath.Join(checkout, "private/tree/library.a"),
			library,
			0600,
		); writeError != nil {
			test.Fatal(writeError)
		}
		entries = append(entries, filepath.Join(checkout, "private/tree/main.a"))
		// Also move only the entry: a formerly adjacent import now stays outside it.
		copied = strings.ReplaceAll(copied, "./library.a", filepath.ToSlash(filepath.Join(checkout, "src/library.a")))
		if writeError := os.WriteFile(
			filepath.Join(checkout, "private/entry/main.a"),
			[]byte(copied),
			0600,
		); writeError != nil {
			test.Fatal(writeError)
		}
		entries = append(entries, filepath.Join(checkout, "private/entry/main.a"))
	}
	for _, cwd := range working {
		if createDirectoryError := os.MkdirAll(cwd, 0700); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
	}
	var wantC, wantJs string
	for _, absolute := range entries {
		for _, cwd := range working {
			test.Run(filepath.ToSlash(strings.TrimPrefix(absolute, root))+
				"/"+
				filepath.Base(cwd), func(test *testing.T) {
				test.Chdir(cwd)
				entry, relativePathError := filepath.Rel(cwd, absolute)
				if relativePathError != nil {
					test.Fatal(relativePathError)
				}
				loaded, loadError := load.Load([]string{entry})
				if loadError != nil {
					test.Fatal(loadError)
				}
				if len(loaded.Files()) != 1 || loaded.FileName(loaded.Files()[0]) != absolute {
					test.Fatal("Files did not retain the explicit entry")
				}
				lowered, lowerError := lower.Lower(context.Background(), loaded)
				if lowerError != nil {
					test.Fatal(lowerError)
				}
				gotC, gotJs := C(lowered), javascript.JavaScript(lowered)
				if strings.Contains(gotC, root) || strings.Contains(gotJs, root) {
					test.Fatal("emitted code contains checkout or working directory")
				}
				if wantC == "" {
					wantC, wantJs = gotC, gotJs
				} else if gotC != wantC || gotJs != wantJs {
					test.Fatal("source path or working directory changed C or JS bytes")
				}
				test.Logf("%d C bytes and %d JS bytes identical", len(gotC), len(gotJs))
			})
		}
	}
	stableProgramMatchesNode(test, entries[0], wantC)
}

// A module generic called during class lowering must not inherit its caller's
// class specialization in the identity of its nested callback.
func TestStableGenericCallbackIgnoresCallingClass(test *testing.T) {
	test.Parallel()
	entry := filepath.Join(test.TempDir(), "main.a")
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
	if writeError := os.WriteFile(entry, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	program := namedProgram(test, entry, nil)
	code := C(program)
	instances := 0
	keys := map[string]bool{}
	for _, function := range program.Functions {
		if len(function.Source.Declaration) == 1 && function.Source.Declaration[0] == "collect" {
			instances++
			keys[sourceKey(function.Source)] = true
		}
		if len(function.Source.Declaration) > 1 &&
			function.Source.Declaration[0] == "collect" &&
			strings.Contains(function.Source.Specialization, "Box") {
			test.Fatalf("module callback inherited calling class: %#v", function.Source)
		}
	}
	// Separate generic cache buckets lower the same concrete module generic twice:
	// once directly and once while its same-file generic class caller is active.
	if instances != 2 || len(keys) != 1 {
		test.Fatalf("collect: %d instances with %d source keys, want 2 with one shared key", instances, len(keys))
	}
	truth := runWithInput(test, "", "node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", entry)
	if truth != "11\n22\n" {
		test.Fatalf("Node fixture output: %q", truth)
	}
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(filepath.Dir(entry), "calling-class")
		if buildError := Build(code, binary, Options{Sanitize: sanitize}); buildError != nil {
			test.Fatal(buildError)
		}
		if got := runWithInput(test, "", binary); got != truth {
			test.Fatalf("native sanitize=%t: %q, Node: %q", sanitize, got, truth)
		}
		test.Logf("two same-file collect instances; native sanitize=%t matches Node: %q", sanitize, truth)
	}
}
