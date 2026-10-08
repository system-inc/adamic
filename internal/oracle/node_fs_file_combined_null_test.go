package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Unknown's null sentinel and front-2's class-aware classifier share one ABI.
func TestNodeFSFileCombinedNullClassifier(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "cloud/reports/host-proof-combined/probes/null_classifier.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	source := native.C(program)
	original := filepath.Join(t.TempDir(), "original")
	if err := native.Build(source, original, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=1"}
	}
	got := executeWith(t, environment, original)
	if difference := disagreement(truth, got); difference != "" {
		t.Fatal(difference)
	}
	for _, mutant := range []struct{ name, function, helper string }{
		{"typeof_null", "adamic_union_typeof", `static adamic_string *mutant_union_typeof(const adamic_heap *value, bool null) { if (value == &adamic_null) { return &adamic_typeof_undefined; } return adamic_union_typeof(value, null); }`},
		{"truthy_null", "adamic_census_to_boolean", `static bool mutant_census_to_boolean(const adamic_heap *value) { if (value == &adamic_null) { return true; } return adamic_census_to_boolean(value); }`},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			mutated := strings.ReplaceAll(source, mutant.function+"(", "mutant_"+strings.TrimPrefix(mutant.function, "adamic_")+"(")
			if mutated == source {
				t.Fatal("mutant changed nothing")
			}
			mutated = strings.Replace(mutated, `#include "adamic.h"`, `#include "adamic.h"`+"\n"+mutant.helper, 1)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(mutated, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			got := executeWith(t, environment, binary)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("mutant failed outside comparison: %+v", got)
			}
			if difference := disagreement(truth, got); difference != "stdout differs" {
				t.Fatalf("mutant caught by %q", difference)
			}
			t.Log("caught only by Node stdout; sanitizers and leaks clean")
		})
	}
}
