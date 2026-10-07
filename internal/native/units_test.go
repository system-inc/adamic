package native

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnitsPreserveSharedState(t *testing.T) {
	source := `#include "adamic.h"
#include <stdio.h>
static int adamic_state = 0;
static int adamic_next(void);
static int adamic_next(void) { return ++adamic_state; }
int main(void) { int first = adamic_next(); int second = adamic_next(); printf("%d %d %d\n", first, second, adamic_state); return first == 1 && second == 2 && adamic_state == 2 ? 0 : 1; }
`
	header, units, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	nextHeader, nextUnits, err := splitC(source)
	if err != nil || header != nextHeader || len(units) != len(nextUnits) {
		t.Fatal("split is not deterministic")
	}
	for index := range units {
		if units[index] != nextUnits[index] {
			t.Fatal("split changed")
		}
	}
	for _, uncached := range []string{"0", "1"} {
		t.Setenv("ADAMIC_GATE_UNCACHED", uncached)
		binary := filepath.Join(t.TempDir(), "program")
		if err := Build(source, binary, Options{Split: true, Sanitize: true, Jobs: 2}); err != nil {
			t.Fatal(err)
		}
		if output, err := exec.Command(binary).CombinedOutput(); err != nil || string(output) != "1 2 2\n" {
			t.Fatalf("shared state (uncached=%s): %v\n%s", uncached, err, output)
		}
	}
}

// The same bytes must be instrumented after a release compile. A flags-key mutant returns the
// release object and loses UBSan's check, while linking and executing normally. No warning kills it.
func TestUnitCacheFlagsHoldSanitizer(t *testing.T) {
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	version, err := exec.Command(compiler, "--version").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	cache := filepath.Join(directory, "cache")
	unit := compilationUnit{"probe.c", `int adamic_probe(int input) { return input + 2147483647; }` + "\n"}
	files := []runtimeFile{{unit.name, []byte(unit.source)}}
	release, err := compileUnit(unit, files, Flags(Options{}), compiler, string(version), cache, directory, false)
	if err != nil {
		t.Fatal(err)
	}
	sanitized, err := compileUnit(unit, files, Flags(Options{Sanitize: true}), compiler, string(version), cache, directory, false)
	if err != nil {
		t.Fatal(err)
	}
	// Do not assert that paths differ: the runtime sanitizer observation is the proof.
	_ = release
	main := filepath.Join(directory, "main.c")
	if err := os.WriteFile(main, []byte(`int adamic_probe(int); int main(void) { volatile int input = 1; volatile int result = adamic_probe(input); (void)result; return 0; }`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "probe")
	args := append(Flags(Options{Sanitize: true}), main, sanitized, "-o", binary)
	if output, err := exec.Command(compiler, args...).CombinedOutput(); err != nil {
		t.Fatalf("link: %v\n%s", err, output)
	}
	output, err := exec.Command(binary).CombinedOutput()
	if err == nil || !bytes.Contains(output, []byte("signed integer overflow")) {
		t.Fatalf("sanitized rebuild reused uninstrumented object: exit=%v output=%s", err, output)
	}
	t.Log("release then sanitized: UBSan caught signed integer overflow")
}

func TestSplitTokensDoNotRewriteLiterals(t *testing.T) {
	source := "#include \"adamic.h\"\nstatic int adamic_x = 1;\nstatic int adamic_f(void) { /* } ; */ const char *s = \"adamic_x } ; \\\"\"; (void)s; return adamic_x; }\nint main(void) { return adamic_f()-1; }\n"
	_, units, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(units[1].source, `"adamic_x } ; \""`) {
		t.Fatal("literal changed")
	}
	if !strings.Contains(units[1].source, "return adamic_unit_adamic_x;") {
		t.Fatal("reference not rewritten")
	}
}
