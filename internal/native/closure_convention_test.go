package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// The ruling explicitly requires this mutant to die in the C compiler.
func TestClosureConventionDropCount(t *testing.T) {
	checked, err := load.Load([]string{filepath.Join("..", "oracle", "testdata", "arguments_length_value_count.a")})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	source := C(program)
	if err := Build(source, filepath.Join(t.TempDir(), "valid"), Options{}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(source, "\n")
	mutated := false
	for i, line := range lines {
		if !strings.Contains(line, "= adamic_closure_call(") {
			continue
		}
		last := strings.LastIndex(line, ", ")
		close := strings.LastIndex(line, ");")
		if last < 0 || close < last {
			t.Fatal("counted call has unexpected form")
		}
		lines[i] = line[:last] + line[close:]
		mutated = true
		break
	}
	if !mutated {
		t.Fatal("no counted call to mutate")
	}
	err = Build(strings.Join(lines, "\n"), filepath.Join(t.TempDir(), "mutant"), Options{})
	if err == nil || !strings.Contains(err.Error(), "too few arguments to function call") || !strings.Contains(err.Error(), "expected 3, have 2") {
		t.Fatalf("drop-count mutant was not rejected for typed arity: %v", err)
	}
	t.Log("drop-count mutant rejected by clang under -Werror: expected 3, have 2")
}

// The runtime callback site must obey the same compiler-enforced arity.
func TestClosureConventionRuntimeDropCount(t *testing.T) {
	checked, err := load.Load([]string{filepath.Join("..", "oracle", "testdata", "closure_convention_regexp_count.a")})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := os.ReadFile(filepath.Join("runtime", "regexp_replace.c"))
	if err != nil {
		t.Fatal(err)
	}
	original := "adamic_closure_call(callback, packed, argument_count)"
	if strings.Count(string(runtime), original) != 1 {
		t.Fatal("runtime mutation site moved")
	}
	changed := strings.Replace(string(runtime), original, "adamic_closure_call(callback, packed)", 1)
	changed = strings.ReplaceAll(changed, "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
	source := changed + "\n" + strings.ReplaceAll(C(program), "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
	err = Build(source, filepath.Join(t.TempDir(), "mutant"), Options{})
	if err == nil || !strings.Contains(err.Error(), "expected 3, have 2") {
		t.Fatalf("runtime drop-count mutant escaped the typed convention: %v", err)
	}
	t.Log("runtime callback drop-count mutant rejected by clang under -Werror: expected 3, have 2")
}

func TestClosureConventionRuntimeFeaturesIgnoreLiterals(t *testing.T) {
	program := &ir.Program{Strings: []string{"adamic_regex_replace_callback", "adamic_fs_file_host_cwd"}}
	library, err := RuntimeLibraryForSource("", C(program), Options{})
	if err != nil {
		t.Fatal(err)
	}
	header, err := os.ReadFile(filepath.Join(filepath.Dir(library), "adamic.h"))
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{"#define ADAMIC_REGEXP_REPLACE_CALLBACK 1", "#define ADAMIC_NODE_HOST 1"} {
		if strings.Contains(string(header), marker) {
			t.Fatalf("a source literal selected unused runtime support: %s", marker)
		}
	}
}

// Definition order, call order and explicit signature-erasing casts each fail
// independently. A constructor check alone misses a never-stored definition.
func TestClosureConventionWrongOrder(t *testing.T) {
	checked, err := load.Load([]string{filepath.Join("..", "oracle", "testdata", "arguments_length_value_count.a")})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	source := C(program)
	old := "adamic_value *arguments, size_t argument_count"
	if !strings.Contains(source, old) {
		t.Fatal("no counted definition")
	}
	swapped := strings.Replace(source, old, "size_t argument_count, adamic_value *arguments", 1)
	t.Run("definition", func(t *testing.T) {
		err := Build(swapped, filepath.Join(t.TempDir(), "mutant"), Options{})
		if err == nil || !strings.Contains(err.Error(), "conflicting types") {
			t.Fatalf("swapped definition escaped the header-derived declaration: %v", err)
		}
		t.Log(err)
	})
	t.Run("call", func(t *testing.T) {
		lines := strings.Split(source, "\n")
		changed := false
		for i, line := range lines {
			start := strings.Index(line, "= adamic_closure_call(")
			if start < 0 {
				continue
			}
			first := strings.Index(line[start:], ", ") + start
			last := strings.LastIndex(line, ", ")
			end := strings.LastIndex(line, ");")
			if first >= last || end < last {
				t.Fatal("mutation site changed")
			}
			lines[i] = line[:first+2] + line[last+2:end] + ", " + line[first+2:last] + line[end:]
			changed = true
			break
		}
		if !changed {
			t.Fatal("no counted call")
		}
		err := Build(strings.Join(lines, "\n"), filepath.Join(t.TempDir(), "mutant"), Options{})
		if err == nil || (!strings.Contains(err.Error(), "int-conversion") && !strings.Contains(err.Error(), "incompatible")) {
			t.Fatalf("swapped call escaped typed parameters: %v", err)
		}
		t.Log(err)
	})
	for _, probe := range []struct{ name, code, diagnostic string }{
		{"function-pointer-cast", "static adamic_value wrong(adamic_closure *self, size_t count, adamic_value *arguments) { return (adamic_value){.number = (double)count}; }\nstatic adamic_counted_code erased = (adamic_counted_code)wrong;\n", "cast-function-type-strict"},
		{"void-pointer-cast", "static adamic_value wrong(adamic_closure *self, size_t count, adamic_value *arguments) { return (adamic_value){.number = (double)count}; }\nstatic adamic_closure *erased(void) { return adamic_counted_closure_new((void *)wrong, 0); }\n", "generic association"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			err := Build(strings.Replace(source, "#include \"adamic.h\"", "#include \"adamic.h\"\n"+probe.code, 1), filepath.Join(t.TempDir(), "mutant"), Options{})
			if err == nil || !strings.Contains(err.Error(), probe.diagnostic) {
				t.Fatalf("signature-erasing cast escaped compiler enforcement: %v", err)
			}
			t.Log(err)
		})
	}
}

// Every owner site compiles through the shared typedefs, including when receiver
// and canonical fields change the closure layout. Dropping its count must fail.
func TestClosureConventionOwnerRuntimeSites(t *testing.T) {
	root, err := filepath.Abs("runtime")
	if err != nil {
		t.Fatal(err)
	}
	for _, module := range []string{"node_process.c", "parallel.c"} {
		for _, features := range [][]string{nil, {"-DADAMIC_CLOSURE_CONVENTION=1"}, {"-DADAMIC_CLOSURE_CONVENTION=1", "-DADAMIC_CLOSURE_RECEIVERS=1", "-DADAMIC_CANONICAL_CLOSURES=1"}} {
			arguments := append(Flags(Options{}), features...)
			arguments = append(arguments, "-I", root, "-c", filepath.Join(root, module), "-o", filepath.Join(t.TempDir(), "runtime.o"))
			if output, err := exec.Command(compilerName(Options{}), arguments...).CombinedOutput(); err != nil {
				t.Fatalf("%s with %v: %v\n%s", module, features, err, output)
			}
		}
	}
	for _, probe := range []struct{ module, call string }{
		{"node_process.c", "adamic_closure_call(closure, padded, count)"},
		{"node_process.c", "adamic_closure_call(closure, args, count)"},
		{"parallel.c", "adamic_closure_call(scope->work, arguments, 2)"},
		{"parallel.c", "adamic_closure_call(work, arguments, 2)"},
	} {
		source, err := os.ReadFile(filepath.Join(root, probe.module))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(source), probe.call) != 1 {
			t.Fatalf("mutation site moved: %s", probe.call)
		}
		last := strings.LastIndex(probe.call, ", ")
		mutant := strings.Replace(string(source), probe.call, probe.call[:last]+")", 1)
		path := filepath.Join(t.TempDir(), probe.module)
		if err := os.WriteFile(path, []byte(mutant), 0o644); err != nil {
			t.Fatal(err)
		}
		arguments := append(Flags(Options{}), "-DADAMIC_CLOSURE_CONVENTION=1", "-I", root, "-c", path, "-o", path+".o")
		output, err := exec.Command(compilerName(Options{}), arguments...).CombinedOutput()
		if err == nil || !strings.Contains(string(output), "expected 3, have 2") {
			t.Fatalf("owner drop-count mutant escaped: %s: %v\n%s", probe.call, err, output)
		}
		t.Logf("%s: drop-count mutant rejected under -Werror: expected 3, have 2", probe.call)
	}
}

func TestParserHasNoUnusedOptionalMethodThunks(t *testing.T) {
	checked, err := load.Load([]string{filepath.Join("..", "..", "stage1", "typescript", "parser", "main.ts")})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	code := C(program)
	if !program.ClosureConventionNeeded() {
		t.Fatal("parser optional callbacks require a count for distinct absence representations")
	}
	found := 0
	for index, function := range program.Functions {
		switch function.Name {
		case "Parser_type", "Parser_assignment", "Parser_allowInAssignment":
			found++
			if strings.Contains(code, " adamic_method_"+strconv.Itoa(index)+";") {
				t.Fatalf("unused optional method thunk: %s", function.Name)
			}
		}
	}
	if found != 3 {
		t.Fatalf("expected three direct optional method witnesses, found %d", found)
	}
}

// Required optional methods keep their thunks; same-signature classes which
// cannot inhabit the receiver view do not acquire a table entry from its closure.
func TestOptionalMethodThunksMatchNode(t *testing.T) {
	path, err := filepath.Abs(filepath.Join("testdata", "optional_method_thunks.a"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	want := runWithInput(t, string(source), "node", "--disable-warning=ExperimentalWarning", "-e", "const fs=require('fs'),m=require('module');eval(m.stripTypeScriptTypes(fs.readFileSync(0,'utf8')))")
	if program.ClosureConventionNeeded() {
		t.Fatal("compatible optional absence must use caller padding, without a count")
	}
	code := C(program)
	for index, function := range program.Functions {
		if function.Name == "SameSignature_run" && strings.Contains(code, " adamic_method_"+strconv.Itoa(index)+";") {
			t.Fatal("same-signature class cannot inhabit the facade receiver")
		}
	}
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "valid")
		if err := Build(code, binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != want {
			t.Fatalf("sanitize %v: native %q; Node %q", sanitize, got, want)
		}
	}
	// Real input mutation: omit every required optional method's table thunk.
	// Compilation still succeeds; only comparison with Node catches the failure.
	for function := range program.StructuralMethodThunks {
		program.StructuralMethodThunks[function] = false
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := Build(C(program), binary, Options{}); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(binary).CombinedOutput()
	if err == nil && string(output) == want {
		t.Fatal("Node failed to catch the omitted required method thunk")
	}
	t.Logf("Node catches compiled missing-thunk mutant: %v, output %q, Node %q", err, output, want)
}
