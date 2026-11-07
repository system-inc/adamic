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
func TestClosureConventionDropCount(test *testing.T) {
	test.Parallel()
	checked, loadError := load.Load(
		[]string{filepath.Join("..", "oracle", "testdata", "arguments_length_value_count.a")},
	)
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), checked)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	source := C(program)
	if buildError := Build(source, filepath.Join(test.TempDir(), "valid"), Options{}); buildError != nil {
		test.Fatal(buildError)
	}
	lines := strings.Split(source, "\n")
	mutated := false
	for index, line := range lines {
		if !strings.Contains(line, "= adamic_closure_call(") {
			continue
		}
		last := strings.LastIndex(line, ", ")
		close := strings.LastIndex(line, ");")
		if last < 0 || close < last {
			test.Fatal("counted call has unexpected form")
		}
		lines[index] = line[:last] + line[close:]
		mutated = true
		break
	}
	if !mutated {
		test.Fatal("no counted call to mutate")
	}
	buildError := Build(strings.Join(lines, "\n"), filepath.Join(test.TempDir(), "mutant"), Options{})
	if buildError == nil ||
		!strings.Contains(buildError.Error(), "too few arguments to function call") ||
		!strings.Contains(buildError.Error(), "expected 3, have 2") {
		test.Fatalf("drop-count mutant was not rejected for typed arity: %v", buildError)
	}
	test.Log("drop-count mutant rejected by clang under -Werror: expected 3, have 2")
}

// The runtime callback site must obey the same compiler-enforced arity.
func TestClosureConventionRuntimeDropCount(test *testing.T) {
	test.Parallel()
	checked, loadError := load.Load(
		[]string{filepath.Join("..", "oracle", "testdata", "closure_convention_regexp_count.a")},
	)
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), checked)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	runtime, readError := os.ReadFile(filepath.Join("runtime", "regexp_replace.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	original := "adamic_closure_call(callback, packed, argument_count)"
	if strings.Count(string(runtime), original) != 1 {
		test.Fatal("runtime mutation site moved")
	}
	changed := strings.Replace(string(runtime), original, "adamic_closure_call(callback, packed)", 1)
	changed = strings.ReplaceAll(changed, "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
	source := changed +
		"\n" +
		strings.ReplaceAll(C(program), "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
	buildError := Build(source, filepath.Join(test.TempDir(), "mutant"), Options{})
	if buildError == nil || !strings.Contains(buildError.Error(), "expected 3, have 2") {
		test.Fatalf("runtime drop-count mutant escaped the typed convention: %v", buildError)
	}
	test.Log("runtime callback drop-count mutant rejected by clang under -Werror: expected 3, have 2")
}

func TestClosureConventionRuntimeFeaturesIgnoreLiterals(test *testing.T) {
	test.Parallel()
	program := &ir.Program{Strings: []string{"adamic_regex_replace_callback", "adamic_fs_file_host_cwd"}}
	library, runtimeLibraryForSourceError := RuntimeLibraryForSource("", C(program), Options{})
	if runtimeLibraryForSourceError != nil {
		test.Fatal(runtimeLibraryForSourceError)
	}
	header, readError := os.ReadFile(filepath.Join(filepath.Dir(library), "adamic.h"))
	if readError != nil {
		test.Fatal(readError)
	}
	for _, marker := range []string{"#define ADAMIC_REGEXP_REPLACE_CALLBACK 1", "#define ADAMIC_NODE_HOST 1"} {
		if strings.Contains(string(header), marker) {
			test.Fatalf("a source literal selected unused runtime support: %s", marker)
		}
	}
}

// Definition order, call order and explicit signature-erasing casts each fail
// independently. A constructor check alone misses a never-stored definition.
func TestClosureConventionWrongOrder(test *testing.T) {
	test.Parallel()
	checked, loadError := load.Load(
		[]string{filepath.Join("..", "oracle", "testdata", "arguments_length_value_count.a")},
	)
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), checked)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	source := C(program)
	old := "adamic_value *arguments, size_t argument_count"
	if !strings.Contains(source, old) {
		test.Fatal("no counted definition")
	}
	swapped := strings.Replace(source, old, "size_t argument_count, adamic_value *arguments", 1)
	test.Run("definition", func(test *testing.T) {
		test.Parallel()
		buildError := Build(swapped, filepath.Join(test.TempDir(), "mutant"), Options{})
		if buildError == nil || !strings.Contains(buildError.Error(), "conflicting types") {
			test.Fatalf("swapped definition escaped the header-derived declaration: %v", buildError)
		}
		test.Log(buildError)
	})
	test.Run("call", func(test *testing.T) {
		test.Parallel()
		lines := strings.Split(source, "\n")
		changed := false
		for index, line := range lines {
			start := strings.Index(line, "= adamic_closure_call(")
			if start < 0 {
				continue
			}
			first := strings.Index(line[start:], ", ") + start
			last := strings.LastIndex(line, ", ")
			end := strings.LastIndex(line, ");")
			if first >= last || end < last {
				test.Fatal("mutation site changed")
			}
			lines[index] = line[:first+2] + line[last+2:end] + ", " + line[first+2:last] + line[end:]
			changed = true
			break
		}
		if !changed {
			test.Fatal("no counted call")
		}
		buildError := Build(strings.Join(lines, "\n"), filepath.Join(test.TempDir(), "mutant"), Options{})
		if buildError == nil ||
			(!strings.Contains(buildError.Error(), "int-conversion") &&
				!strings.Contains(buildError.Error(), "incompatible")) {
			test.Fatalf("swapped call escaped typed parameters: %v", buildError)
		}
		test.Log(buildError)
	})
	for _, probe := range []struct{ name, code, diagnostic string }{
		{
			"function-pointer-cast",
			"static adamic_value wrong(adamic_closure *self, size_t count, adamic_value *arguments) { return (adamic_value){.number = (double)count}; }\nstatic adamic_counted_code erased = (adamic_counted_code)wrong;\n",
			"cast-function-type-strict",
		},
		{
			"void-pointer-cast",
			"static adamic_value wrong(adamic_closure *self, size_t count, adamic_value *arguments) { return (adamic_value){.number = (double)count}; }\nstatic adamic_closure *erased(void) { return adamic_counted_closure_new((void *)wrong, 0); }\n",
			"generic association",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			buildError := Build(
				strings.Replace(source, "#include \"adamic.h\"", "#include \"adamic.h\"\n"+probe.code, 1),
				filepath.Join(test.TempDir(), "mutant"),
				Options{},
			)
			if buildError == nil || !strings.Contains(buildError.Error(), probe.diagnostic) {
				test.Fatalf("signature-erasing cast escaped compiler enforcement: %v", buildError)
			}
			test.Log(buildError)
		})
	}
}

// Every owner site compiles through the shared typedefs, including when receiver
// and canonical fields change the closure layout. Dropping its count must fail.
// node_process.c's two sites join this list when library's area brings the file.
func TestClosureConventionOwnerRuntimeSites(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs("runtime")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for _, module := range []string{"parallel.c"} {
		for _, features := range [][]string{
			nil,
			{"-DADAMIC_CLOSURE_CONVENTION=1"},
			{"-DADAMIC_CLOSURE_CONVENTION=1", "-DADAMIC_CLOSURE_RECEIVERS=1", "-DADAMIC_CANONICAL_CLOSURES=1"},
		} {
			arguments := append(Flags(Options{}), features...)
			arguments = append(
				arguments,
				"-I",
				root,
				"-c",
				filepath.Join(root, module),
				"-o",
				filepath.Join(test.TempDir(), "runtime.o"),
			)
			if output, combinedOutputError := exec.Command(
				compilerName(Options{}),
				arguments...,
			).CombinedOutput(); combinedOutputError != nil {
				test.Fatalf("%s with %v: %v\n%s", module, features, combinedOutputError, output)
			}
		}
	}
	for _, probe := range []struct{ module, call string }{
		{"parallel.c", "adamic_closure_call(scope->work, arguments, 2)"},
		{"parallel.c", "adamic_closure_call(work, arguments, 2)"},
	} {
		source, readError := os.ReadFile(filepath.Join(root, probe.module))
		if readError != nil {
			test.Fatal(readError)
		}
		if strings.Count(string(source), probe.call) != 1 {
			test.Fatalf("mutation site moved: %s", probe.call)
		}
		last := strings.LastIndex(probe.call, ", ")
		mutant := strings.Replace(string(source), probe.call, probe.call[:last]+")", 1)
		path := filepath.Join(test.TempDir(), probe.module)
		if writeError := os.WriteFile(path, []byte(mutant), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
		arguments := append(Flags(Options{}), "-DADAMIC_CLOSURE_CONVENTION=1", "-I", root, "-c", path, "-o", path+".o")
		output, combinedOutputError := exec.Command(compilerName(Options{}), arguments...).CombinedOutput()
		if combinedOutputError == nil || !strings.Contains(string(output), "expected 3, have 2") {
			test.Fatalf("owner drop-count mutant escaped: %s: %v\n%s", probe.call, combinedOutputError, output)
		}
		test.Logf("%s: drop-count mutant rejected under -Werror: expected 3, have 2", probe.call)
	}
}

func TestParserHasNoUnusedOptionalMethodThunks(test *testing.T) {
	test.Parallel()
	checked, loadError := load.Load([]string{filepath.Join("..", "..", "stage1", "typescript", "parser", "main.ts")})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), checked)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	code := C(program)
	if !program.ClosureConventionNeeded() {
		test.Fatal("parser optional callbacks require a count for distinct absence representations")
	}
	found := 0
	for index, function := range program.Functions {
		switch function.Name {
		case "Parser_type", "Parser_assignment", "Parser_allowInAssignment":
			found++
			if strings.Contains(code, " adamic_method_"+strconv.Itoa(index)+";") {
				test.Fatalf("unused optional method thunk: %s", function.Name)
			}
		}
	}
	if found != 3 {
		test.Fatalf("expected three direct optional method witnesses, found %d", found)
	}
}

// Required optional methods keep their thunks; same-signature classes which
// cannot inhabit the receiver view do not acquire a table entry from its closure.
func TestOptionalMethodThunksMatchNode(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join("testdata", "optional_method_thunks.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	source, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	checked, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), checked)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	want := runWithInput(
		test,
		string(source),
		"node",
		"--disable-warning=ExperimentalWarning",
		"-e",
		"const fs=require('fs'),m=require('module');eval(m.stripTypeScriptTypes(fs.readFileSync(0,'utf8')))",
	)
	if program.ClosureConventionNeeded() {
		test.Fatal("compatible optional absence must use caller padding, without a count")
	}
	code := C(program)
	for index, function := range program.Functions {
		if function.Name == "SameSignature_run" && strings.Contains(code, " adamic_method_"+strconv.Itoa(index)+";") {
			test.Fatal("same-signature class cannot inhabit the facade receiver")
		}
	}
	for _, sanitize := range []bool{false, true} {
		test.Run("sanitize="+strconv.FormatBool(sanitize), func(test *testing.T) {
			test.Parallel()
			binary := filepath.Join(test.TempDir(), "valid")
			if buildError := Build(code, binary, Options{Sanitize: sanitize}); buildError != nil {
				test.Fatal(buildError)
			}
			if got := runWithInput(test, "", binary); got != want {
				test.Fatalf("sanitize %v: native %q; Node %q", sanitize, got, want)
			}
		})
	}
	// Real input mutation: omit every required optional method's table thunk.
	// Compilation still succeeds; only comparison with Node catches the failure.
	for function := range program.StructuralMethodThunks {
		program.StructuralMethodThunks[function] = false
	}
	binary := filepath.Join(test.TempDir(), "mutant")
	if buildError := Build(C(program), binary, Options{}); buildError != nil {
		test.Fatal(buildError)
	}
	output, combinedOutputError := exec.Command(binary).CombinedOutput()
	if combinedOutputError == nil && string(output) == want {
		test.Fatal("Node failed to catch the omitted required method thunk")
	}
	test.Logf("Node catches compiled missing-thunk mutant: %v, output %q, Node %q", combinedOutputError, output, want)
}
