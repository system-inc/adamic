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
		if got := runWithInput(t, "", binary); got != "31\n" {
			t.Fatalf("output %q", got)
		}
	}
}

func TestStableIdentifierBoundsAndEscaping(t *testing.T) {
	program := &ir.Program{Functions: []ir.Function{{Source: ir.SourceIdentity{Module: strings.Repeat("module/", 20) + "main.a", Declaration: []string{"Class", "compileSpecificDeclarationName"}}}}}
	if name := (&emitter{program: program}).functionName(0); !strings.Contains(name, "compileSpecificDeclarationName") {
		t.Fatalf("declaration spelling lost: %s", name)
	}

	t.Parallel()
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
	if count != 2 {
		t.Fatalf("got %d Holder specializations", count)
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
	if got := runWithInput(t, "", binary); got != "10\n11\n" {
		t.Fatalf("output %q", got)
	}
}

func TestMethodAdaptersShareStableDefinition(t *testing.T) {
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
	if got := runWithInput(t, "", binary); got != "20\n" {
		t.Fatalf("output %q", got)
	}
}
