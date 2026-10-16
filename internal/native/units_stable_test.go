package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func unitInputs(header string, units []compilationUnit) map[string]string {
	inputs := map[string]string{}
	for _, unit := range units {
		inputs[unit.name] = header + unit.source
	}
	return inputs
}

func changedUnitInputs(before, after map[string]string) []string {
	var changed []string
	for name, source := range after {
		if before[name] != source {
			changed = append(changed, name)
		}
	}
	for name := range before {
		if _, ok := after[name]; !ok {
			changed = append(changed, name)
		}
	}
	return changed
}

func TestStableUnitInsertion(t *testing.T) {
	t.Parallel()
	source := `#include "adamic.h"
// adamic-module "one.a"
static int first(void) { return 1; }
// adamic-module "two.a"
static int second(void) { return 2; }
int main(void) { return first()+second()-3; }
`
	header, units, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(source, "// adamic-module \"two.a\"", "// adamic-module \"one.a\"\nstatic int added(void) { return 0; }\n// adamic-module \"two.a\"", 1)
	nextHeader, nextUnits, err := splitC(edited)
	if err != nil {
		t.Fatal(err)
	}
	changed := changedUnitInputs(unitInputs(header, units), unitInputs(nextHeader, nextUnits))
	if len(changed) != 1 || changed[0] != moduleUnit("one.a") {
		t.Fatalf("insertion changed %v, want only one.a", changed)
	}
}

func TestUnitDeclarationDisagreement(t *testing.T) {
	t.Parallel()
	source := `#include "adamic.h"
// adamic-module "one.a"
static int first(void) { return 1; }
// adamic-module "two.a"
static int second(void) { return first(); }
int main(void) { return second()-1; }
`
	header, units, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	// This mutant is valid C in both units. Only the declaration consistency relocation kills it.
	declaration := "int adamic_unit_first(void);\n"
	mutant := "long adamic_unit_first(void);\n"
	changed := false
	for index := range units {
		if units[index].name == moduleUnit("two.a") {
			units[index].source = strings.ReplaceAll(units[index].source, declaration, mutant)
			units[index].source = strings.ReplaceAll(units[index].source, declarationGuard("first", declaration), declarationGuard("first", mutant))
			changed = true
		}
	}
	if !changed {
		t.Fatal("mutant did not change caller")
	}
	directory := t.TempDir()
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	version, err := exec.Command(compiler, "--version").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	var objects []string
	runtimeFiles, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	for _, unit := range units {
		files := []runtimeFile{{unit.name, []byte(unit.source)}, {"units.h", []byte(header)}}
		for _, file := range runtimeFiles {
			if strings.HasSuffix(file.name, ".h") {
				files = append(files, file)
			}
		}
		object, err := compileUnit(unit, files, Flags(Options{}), compiler, string(version), filepath.Join(directory, "cache"), directory, true)
		if err != nil {
			t.Fatalf("mutant must compile cleanly: %v", err)
		}
		objects = append(objects, object)
	}
	library, err := RuntimeLibrary("", Options{})
	if err != nil {
		t.Fatal(err)
	}
	arguments := append(Flags(Options{}), objects...)
	arguments = append(arguments, RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-lm", "-o", filepath.Join(directory, "mutant"))
	output, err := exec.Command(compiler, arguments...).CombinedOutput()
	if err == nil || !strings.Contains(string(output), declarationGuard("first", mutant)) {
		t.Fatalf("declaration mutant silently linked: %v\n%s", err, output)
	}
	t.Logf("declaration mutant: units compiled cleanly; link rejected missing ABI guard")
}

func TestSharedIdentitiesDoNotMigrate(t *testing.T) {
	t.Parallel()
	source := `#include "adamic.h"
static int adamic_string_probe=7;
// adamic-module ""
static int helper(void) { return adamic_string_probe; }
// adamic-module "one.a"
static int first(void) { return helper(); }
// adamic-module "two.a"
static int second(void) { return 0; }
int main(void) { return first()-7; }
`
	header, units, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(source, "int main(void)", "// adamic-module \"two.a\"\nstatic int added(void) { return helper()+adamic_string_probe; }\nint main(void)", 1)
	nextHeader, nextUnits, err := splitC(edited)
	if err != nil {
		t.Fatal(err)
	}
	changed := changedUnitInputs(unitInputs(header, units), unitInputs(nextHeader, nextUnits))
	if len(changed) != 1 || changed[0] != moduleUnit("two.a") {
		t.Fatalf("shared identity migrated: changed %v", changed)
	}
}

func TestConstantDescriptorDoesNotMigrate(t *testing.T) {
	t.Parallel()
	source := "#include \"adamic.h\"\nstatic const int adamic_class_probe=7;\n// adamic-module \"one.a\"\nstatic int first(void){return adamic_class_probe;}\n// adamic-module \"two.a\"\nstatic int second(void){return 0;}\nint main(void){return first()-7;}\n"
	header, units, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(source, "int main(void)", "// adamic-module \"two.a\"\nstatic int added(void){return adamic_class_probe;}\nint main(void)", 1)
	nextHeader, nextUnits, err := splitC(edited)
	if err != nil {
		t.Fatal(err)
	}
	changed := changedUnitInputs(unitInputs(header, units), unitInputs(nextHeader, nextUnits))
	if len(changed) != 1 || changed[0] != moduleUnit("two.a") {
		t.Fatalf("constant descriptor migrated: %v", changed)
	}
}
func TestDeclarationAdapterFollowsCallee(t *testing.T) {
	t.Parallel()
	source := "#include \"adamic.h\"\n// adamic-module \"one.a\"\nstatic int first(void){return 7;}\nstatic int adapter(void){return first();}\nstatic int outer_adapter(void){return adapter();}\n// adamic-module \"two.a\"\nstatic int second(void){return outer_adapter();}\nint main(void){return second()-7;}\n"
	_, units, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, unit := range units {
		if unit.name == moduleUnit("one.a") && strings.Contains(unit.source, "int adamic_unit_adapter(void){") && strings.Contains(unit.source, "int adamic_unit_outer_adapter(void){") {
			found = true
		}
	}
	if !found {
		t.Fatal("adapter did not follow its source declaration")
	}
}

func TestForwarderGlobalsDoNotMigrate(t *testing.T) {
	t.Parallel()
	source := "#include \"adamic.h\"\nstatic int adamic_global_probe=7;\n// adamic-module \"one.a\"\nstatic int first(void){return 0;}\nint main(void){return adamic_global_probe-7;}\n"
	header, units, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(source, "int main(void)", "// adamic-module \"two.a\"\nstatic int added(void){return adamic_global_probe;}\nint main(void)", 1)
	nextHeader, nextUnits, err := splitC(edited)
	if err != nil {
		t.Fatal(err)
	}
	changed := changedUnitInputs(unitInputs(header, units), unitInputs(nextHeader, nextUnits))
	if len(changed) != 1 || changed[0] != moduleUnit("two.a") {
		t.Fatalf("forwarder global migrated: %v", changed)
	}
}

// Function-type prototypes and explicit definitions share an ABI, including
// optional argument counts. Only the shared header may redeclare these functions.
func checkSplitFunctionPrototypes(t *testing.T, method bool) {
	t.Helper()
	for _, counted := range []bool{false, true} {
		alias, count := "adamic_code_function", ""
		if counted {
			alias, count = "adamic_counted_code_function", ", size_t argument_count"
		}
		receiver := "adamic_closure"
		if method {
			receiver = "adamic_object"
			alias = strings.Replace(alias, "code", "method", 1)
		}
		source := fmt.Sprintf(`#include "adamic.h"
static %s adamic_function_reader;
// adamic-module "reader.a"
static adamic_value adamic_function_reader(adamic_closure *self, adamic_value *arguments%s) { return arguments[0]; }
// adamic-module "caller.a"
static adamic_value call_reader(adamic_closure *self, adamic_value *arguments%s) { return adamic_function_reader(self, arguments%s); }
int main(void) { return 0; }
`, alias, count, count, map[bool]string{false: "", true: ", argument_count"}[counted])
		source = strings.ReplaceAll(source, "adamic_closure", receiver)
		header, units, err := splitC(source)
		if err != nil {
			t.Fatal(err)
		}
		declaration := fmt.Sprintf("adamic_value adamic_unit_adamic_function_reader(%s *self, adamic_value *arguments%s);\n", receiver, count)
		if strings.Count(header, declaration) != 1 {
			t.Fatalf("want one shared prototype: %s", header)
		}
		definitions := 0
		for _, unit := range units {
			if strings.Contains(unit.source, declaration) || !strings.HasPrefix(unit.source, "#include \"units.h\"\n") {
				t.Fatalf("unit %s has a private closure prototype", unit.name)
			}
			definitions += strings.Count(unit.source, strings.TrimSuffix(declaration, ";\n")+" {")
		}
		if definitions != 1 {
			t.Fatalf("want one definition, got %d", definitions)
		}
		wrongCount := ""
		if !counted {
			wrongCount = ", size_t argument_count"
		}
		wrong := strings.Replace(source, "adamic_value adamic_function_reader("+receiver+" *self, adamic_value *arguments"+count+")", "adamic_value adamic_function_reader("+receiver+" *self, adamic_value *arguments"+wrongCount+")", 1)
		if _, _, err := splitC(wrong); err == nil || !strings.Contains(err.Error(), "conflicting declaration") {
			t.Fatalf("ABI mismatch survived: %v", err)
		}
	}
}

func TestSplitClosurePrototypesUseSharedHeader(t *testing.T) {
	t.Parallel()
	checkSplitFunctionPrototypes(t, false)
}
func TestSplitMethodPrototypesUseSharedHeader(t *testing.T) {
	t.Parallel()
	checkSplitFunctionPrototypes(t, true)
}

func TestStableUnitFunctionEdit(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	entry := filepath.Join(directory, "main.a")
	module := filepath.Join(directory, "one.a")
	sources := map[string]string{
		entry:                             `import { first } from './one.a'; import { second } from './two.a'; console.log((first(7)+second(8)).toString());`,
		module:                            `export function first(value: number): number { return value + 14; }`,
		filepath.Join(directory, "two.a"): `export function second(value: number): number { return value + 15; }`,
	}
	for path, source := range sources {
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	before := C(namedProgram(t, entry, nil))
	after := C(namedProgram(t, entry, map[string]string{module: strings.Replace(sources[module], "value + 14", "value + [14].length", 1)}))
	header, units, err := splitC(before)
	if err != nil {
		t.Fatal(err)
	}
	nextHeader, nextUnits, err := splitC(after)
	if err != nil {
		t.Fatal(err)
	}
	changed := changedUnitInputs(unitInputs(header, units), unitInputs(nextHeader, nextUnits))
	if len(units) < 3 || len(changed) != 1 || changed[0] != moduleUnit("one.a") {
		t.Fatalf("function edit changed %v of %d units", changed, len(units))
	}
	// Comparing the full header+source bytes also holds the object cache's inputs.
	for label, source := range map[string]string{"before": before, "after": after} {
		binary := filepath.Join(directory, label)
		if err := Build(source, binary, Options{Split: true, Sanitize: true, Count: true, Jobs: 2}); err != nil {
			t.Fatal(err)
		}
		text := sources[module]
		if label == "after" {
			text = strings.Replace(text, "value + 14", "value + [14].length", 1)
		}
		if err := os.WriteFile(module, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		truth := runWithInput(t, "", "node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", entry)
		if got := runWithInput(t, "", binary); got != truth {
			t.Fatalf("%s: native %q Node %q", label, got, truth)
		}
	}
}

func TestStableUnitOrderDoesNotNameUnits(t *testing.T) {
	t.Parallel()
	source := "#include \"adamic.h\"\n// adamic-module \"z.a\"\nstatic int z(void){return 7;}\nint main(void){return z()-7;}\n"
	header, units, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	added := strings.Replace(source, "// adamic-module", "// adamic-module \"a.a\"\nstatic int a(void){return 1;}\n// adamic-module", 1)
	nextHeader, nextUnits, err := splitC(added)
	if err != nil {
		t.Fatal(err)
	}
	changed := changedUnitInputs(unitInputs(header, units), unitInputs(nextHeader, nextUnits))
	sort.Strings(changed)
	if len(changed) != 1 || changed[0] != moduleUnit("a.a") {
		t.Fatalf("positional identities changed %v", changed)
	}
}

func TestCanonicalAsyncPrivateIdentity(t *testing.T) {
	t.Parallel()
	source := "typedef struct {int x;} adamic_generated_frame_0;\nstatic void adamic_async_resume_0(void){ const char *s=\"adamic_async_resume_0\"; (void)s;}\n"
	identity := ir.SourceIdentity{Module: "async.a", Declaration: []string{"work"}}
	first := &emitter{program: &ir.Program{Functions: []ir.Function{{Source: identity}}}}
	second := &emitter{program: &ir.Program{Functions: []ir.Function{{Source: ir.SourceIdentity{Module: "other.a", Declaration: []string{"added"}}}, {Source: identity}}}}
	before := first.canonicalAsyncDeclarations(source, 0)
	after := second.canonicalAsyncDeclarations(strings.ReplaceAll(strings.ReplaceAll(source, "frame_0", "frame_1"), "resume_0(void)", "resume_1(void)"), 1)
	if before != after || !strings.Contains(before, `"adamic_async_resume_0"`) {
		t.Fatalf("async identity moved or literal changed:\n%s\n%s", before, after)
	}
}
