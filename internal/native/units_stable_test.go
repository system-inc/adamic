package native

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
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

func TestModuleMainRejectsCrossRangeLocal(t *testing.T) {
	source := `#include "adamic.h"
int main(void) {
// adamic-module "one.a"
int adamic_temporary_shared=1;
// adamic-module "two.a"
(void)adamic_temporary_shared;
 return 0;
}
`
	// The extractor expects the emitter's tabbed return.
	source = strings.Replace(source, " return 0;", "\treturn 0;", 1)
	if _, _, err := splitC(source); err == nil || !strings.Contains(err.Error(), "crosses modules") {
		t.Fatalf("cross-range local accepted: %v", err)
	}
}

// Not parallel: this probe compares source overlays and optionally measures cache rebuilds.
func TestStableLintUnitChanges(t *testing.T) {
	if os.Getenv("ADAMIC_STABLE_LINT_PROBE") != "1" {
		t.Skip("set ADAMIC_STABLE_LINT_PROBE=1 for full lint dependency acceptance")
	}
	entry := "../../stage1/cohere/lint/main.ts"
	module, err := filepath.Abs("../../stage1/typescript/parser/grammar.ts")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(module)
	if err != nil {
		t.Fatal(err)
	}
	original := string(data)
	added := strings.Replace(original, "export function precedence", "function isolatedProbe(): number { return 0; }\nexport function precedence", 1)
	edits := []struct{ name, before, after string }{
		{"add-temporary", original, strings.Replace(original, "export function precedence(kind: string): number {", "export function precedence(kind: string): number {\n const isolatedProbe = [14].length;", 1)},
		{"14-to-array-length", original, strings.Replace(original, "return 14;", "return [14].length;", 1)},
		{"add-function", original, added},
		{"remove-function", added, original},
		{"rename-function", added, strings.Replace(added, "isolatedProbe()", "renamedProbe()", 1)},
	}
	emit := func(source string) string {
		digest := sha256.Sum256([]byte(source))
		archive := os.Getenv("ADAMIC_STABLE_C_INPUTS")
		path := filepath.Join(archive, fmt.Sprintf("%x.c", digest))
		if os.Getenv("ADAMIC_STABLE_READ_INPUTS") == "1" {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			return string(data)
		}
		loaded, err := load.LoadOverlay([]string{entry}, map[string]string{module: source})
		if err != nil {
			t.Fatal(err)
		}
		program, err := lower.Lower(context.Background(), loaded)
		if err != nil {
			t.Fatal(err)
		}
		generated := C(program)
		if archive != "" {
			if err := os.MkdirAll(archive, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(generated), 0644); err != nil {
				t.Fatal(err)
			}
		}
		return generated
	}
	baseline := emit(original)
	withAdded := emit(added)
	directory := t.TempDir()
	// A dedicated object cache makes the actual number of compilations observable.
	t.Setenv("XDG_CACHE_HOME", directory)
	t.Setenv("ADAMIC_GATE_UNCACHED", "0")
	t.Setenv("ADAMIC_NATIVE_SPLIT", "1")
	for _, edit := range edits {
		before := baseline
		if edit.before == added {
			before = withAdded
		}
		after := emit(edit.after)
		header, units, err := splitC(before)
		if err != nil {
			t.Fatal(err)
		}
		nextHeader, nextUnits, err := splitC(after)
		if err != nil {
			t.Fatal(err)
		}
		changed := changedUnitInputs(unitInputs(header, units), unitInputs(nextHeader, nextUnits))
		if os.Getenv("ADAMIC_STABLE_BASELINE") != "1" && (len(changed) != 1 || changed[0] != moduleUnit("../../typescript/parser/grammar.ts")) {
			t.Fatalf("%s changed %v, want only grammar unit", edit.name, changed)
		}
		t.Logf("%s inputs changed=%d units=%d names=%v", edit.name, len(changed), len(nextUnits), changed)
		if os.Getenv("ADAMIC_STABLE_MEASURE") != "1" {
			continue
		}
		// Warm only the unchanged baseline, then remove edited entries between samples.
		if err := os.RemoveAll(filepath.Join(directory, "adamic", "units")); err != nil {
			t.Fatal(err)
		}
		if err := Build(before, filepath.Join(directory, "before"), Options{Split: true, Sanitize: true, Jobs: 4}); err != nil {
			t.Fatal(err)
		}
		cache := filepath.Join(directory, "adamic", "units")
		baselineEntries, err := os.ReadDir(cache)
		if err != nil {
			t.Fatal(err)
		}
		known := map[string]bool{}
		for _, entry := range baselineEntries {
			known[entry.Name()] = true
		}
		rounds := 3
		if os.Getenv("ADAMIC_STABLE_ONE_ROUND") == "1" {
			rounds = 1
		}
		for round := 0; round < rounds; round++ {
			loadBefore, _ := os.ReadFile("/proc/loadavg")
			started := time.Now()
			if err := Build(after, filepath.Join(directory, "after"), Options{Split: true, Sanitize: true, Jobs: 4}); err != nil {
				t.Fatal(err)
			}
			elapsed := time.Since(started)
			loadAfter, _ := os.ReadFile("/proc/loadavg")
			entries, err := os.ReadDir(cache)
			if err != nil {
				t.Fatal(err)
			}
			rebuilt := 0
			for _, entry := range entries {
				if !known[entry.Name()] {
					rebuilt++
					if err := os.RemoveAll(filepath.Join(cache, entry.Name())); err != nil {
						t.Fatal(err)
					}
				}
			}
			if os.Getenv("ADAMIC_STABLE_BASELINE") != "1" && rebuilt != 1 {
				t.Fatalf("%s recompiled %d objects, want 1", edit.name, rebuilt)
			}
			t.Logf("TIMING edit=%s round=%d recompiled=%d seconds=%.6f flags=%v cached=baseline-only load-before=%s load-after=%s", edit.name, round+1, rebuilt, elapsed.Seconds(), Flags(Options{Sanitize: true}), strings.TrimSpace(string(loadBefore)), strings.TrimSpace(string(loadAfter)))
		}
	}
}
