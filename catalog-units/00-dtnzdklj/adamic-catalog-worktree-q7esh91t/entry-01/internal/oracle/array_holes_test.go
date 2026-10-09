package oracle

import (
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArrayHolesMilestone(t *testing.T) {
	for _, name := range []string{"library_array_holes_scanner_probe.a", "library_array_holes_length.a", "library_array_holes_range.a"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			native, binary := natively(t, program)
			for _, got := range []run{native, released(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatalf("%s: Node %q %q; got %q %q", difference, truth.stdout, truth.stderr, got.stdout, got.stderr)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Logf("Node stdout: %s", truth.stdout)
		})
	}
}

func init() {
	for _, name := range []string{"library_array_holes_scanner_probe.a", "library_array_holes_length.a", "library_array_holes_range.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}

// Missing indexed slots must stay absent rather than becoming present zero.
func TestArrayHolesAbsentSlotMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_array_holes_length.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	nativeMutant := arrayRuntimeMutant(t, native.C(program), "array_holes.c",
		"return adamic_map_get(array->sparse, (adamic_value){.number = index});",
		"adamic_value *slot = adamic_map_get(array->sparse, (adamic_value){.number = index}); static adamic_value invented = {.number = 0}; return slot == NULL ? &invented : slot;")
	jsSource := strings.ReplaceAll(javascript.JavaScript(program), "new Array(", "adamicMutantDenseArray(")
	jsSource = "const adamicMutantDenseArray = length => { const array = new Array(length); if (length <= 16) array.fill(0); return array; };\n" + jsSource
	mutantPath := filepath.Join(t.TempDir(), "holes-mutant.mjs")
	if err := os.WriteFile(mutantPath, []byte(jsSource), 0644); err != nil {
		t.Fatal(err)
	}
	jsMutant := onNode(t, mutantPath)
	for _, got := range []run{nativeMutant, jsMutant} {
		if got.exitCode != 0 || !strings.Contains(string(got.stdout), "4:false:false") {
			t.Fatalf("absence mutant did not reach the pinned distinguishing observation: exit %d stdout %q stderr %q", got.exitCode, got.stdout, got.stderr)
		}
		if disagreement(truth, got) == "" {
			t.Fatal("absent-slot omission escaped")
		}
	}
	t.Log("absent-slot check removed; Node's 4:true:true catches C and JavaScript mutants")
}

func TestArrayHolesRangeErrorMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_array_holes_range.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	got := arrayRuntimeMutant(t, native.C(program), "array_holes.c",
		"adamic_thrown = adamic_object_new(&range_shape);",
		"return adamic_array_new(0, references);\n        adamic_thrown = adamic_object_new(&range_shape);")
	jsSource := strings.ReplaceAll(javascript.JavaScript(program), "new Array(", "adamicMutantRangeArray(")
	jsSource = "const adamicMutantRangeArray = length => new Array(Number.isInteger(length) && length >= 0 && length <= 4294967295 ? length : 0);\n" + jsSource
	mutantPath := filepath.Join(t.TempDir(), "range-mutant.mjs")
	if err := os.WriteFile(mutantPath, []byte(jsSource), 0644); err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []run{got, onNode(t, mutantPath)} {
		if mutant.exitCode != 0 || strings.Contains(string(mutant.stdout), "range:") || disagreement(truth, mutant) == "" {
			t.Fatalf("RangeError omission escaped: exit %d stdout %q stderr %q", mutant.exitCode, mutant.stdout, mutant.stderr)
		}
	}
	t.Log("invalid-length throw removed; Node's nominal RangeError observations reject both mutants")
}
