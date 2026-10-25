package oracle

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

var recordUnionNullMutant = flag.Bool("record-union-null-mutant", false, "run the throwaway null-pointer record store")

func init() {
	for _, name := range []string{"null", "presence"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/records_union_" + name + ".a", true, false})
	}
}

// Only this declared product builds the fixture binaries and the changed store.
// Runtime archives retain their existing oracle owners in product-order.json.
func recordUnionSlotsProduct(t *testing.T) string {
	t.Helper()
	inputs := buildcache.Inputs{Name: "record-union-slots", Files: compilerinputs.Lowering()}
	variants := []struct {
		name    string
		options native.Options
	}{{"sanitized", native.Options{Sanitize: true}}, {"release", native.Options{}}, {"counted", native.Options{Count: true}}}
	for _, variant := range variants {
		request, err := native.RuntimeRequestFor(variant.options)
		if err != nil {
			t.Fatal(err)
		}
		inputs.Flags = append(inputs.Flags, request.Portable...)
		inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool(request.Compiler, "--version"), buildcache.Tool(request.Archiver, "--version"))
	}
	return buildcache.Product(t, inputs, func(directory string) error {
		for _, name := range []string{"null", "presence"} {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_union_"+name+".a"))
			if err != nil {
				return err
			}
			program, err := lowered(t, path)
			if err != nil {
				return err
			}
			code := native.C(program)
			for _, variant := range variants {
				if err := native.Build(code, filepath.Join(directory, name+"-"+variant.name), variant.options); err != nil {
					return err
				}
			}
			if name == "null" {
				// Change only the value being written. Keep the null comparison's singleton.
				const store = " = &adamic_null;"
				if strings.Count(code, store) != 1 {
					return fmt.Errorf("expected one null store, found %d", strings.Count(code, store))
				}
				mutant := strings.Replace(code, store, " = NULL;", 1)
				if err := native.Build(mutant, filepath.Join(directory, "null-mutant"), native.Options{Sanitize: true}); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func TestProduct_RecordUnionSlots(t *testing.T) {
	t.Parallel()
	recordUnionSlotsProduct(t)
}

func checkRecordUnionSlots(t *testing.T, name, expected string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_union_"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != expected {
		t.Fatalf("Node result: %+v", truth)
	}
	directory := recordUnionSlotsProduct(t)
	for _, variant := range []string{"sanitized", "release", "counted"} {
		binary := name + "-" + variant
		if *recordUnionNullMutant && name == "null" && variant == "sanitized" {
			binary = "null-mutant"
		}
		got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, filepath.Join(directory, binary))
		if variant == "counted" {
			if report := leakcheck.Unbalanced(leakRun(got)); report != "" {
				t.Fatal(report)
			}
			got.stderr = nil // Counts are checked independently of program output.
		}
		if diff := disagreement(truth, got); diff != "" {
			t.Fatalf("%s versus Node: %s; got stdout %q stderr %q exit %d; want stdout %q", variant, diff, got.stdout, got.stderr, got.exitCode, truth.stdout)
		}
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if got := onJavaScriptBackend(t, program); disagreement(truth, got) != "" {
		t.Fatalf("JavaScript versus Node: %+v; want %+v", got, truth)
	}
	t.Logf("Node, release, ASan, counted and JavaScript agree: %q", truth.stdout)
}

func TestRecordUnionSlotsNull(t *testing.T) {
	t.Parallel()
	checkRecordUnionSlots(t, "null", "true\nfalse\nobject\n")
}
func TestRecordUnionSlotsPresence(t *testing.T) {
	t.Parallel()
	checkRecordUnionSlots(t, "presence", "a\n\ntrue\nfalse\n{}\n{}\ntrue\nfalse\n")
}
func TestRecordUnionSlotsNullPointerMutant(t *testing.T) {
	t.Parallel()
	recordUnionSlotsProduct(t)
	got := execute(t, os.Args[0], "-test.run=^TestRecordUnionSlotsNull$", "-test.v", "-test.timeout=30s", "-record-union-null-mutant")
	if got.exitCode == 0 || !bytes.Contains(got.stdout, []byte("--- FAIL: TestRecordUnionSlotsNull")) || !bytes.Contains(got.stdout, []byte("sanitized versus Node: stdout differs")) {
		t.Fatalf("mutant did not fail its own oracle: %+v", got)
	}
	t.Logf("%s", got.stdout)
}
