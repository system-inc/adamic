package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os/exec"
)

func init() {
	for _, name := range []string{"discarded", "guarded_snapshot", "compare_missing_scalar", "compare_properties_left", "compare_properties_right", "environment_boundary"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/records_" + name + ".a", true, name == "compare_properties_left" || name == "compare_properties_right" || name == "environment_boundary"})
	}
}

func TestRecordObservationGuardMutants(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"discarded", "guarded_snapshot", "compare_missing_scalar"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			code := native.C(program)
			// Remove the guard and restore the loud lookup, preserving valid C and
			// evaluation order. A missing inherited name now stops where Node prints.
			changed := strings.ReplaceAll(code, "adamic_record_has_own(", "record_mutant_has_own(")
			changed = strings.ReplaceAll(changed, "adamic_record_get_own(", "adamic_record_get(")
			if changed == code {
				t.Fatal("mutant changed nothing")
			}
			changed = insertCollectionMutant(changed, `static bool record_mutant_has_own(const adamic_record *r, const adamic_string *k) { (void)r; (void)k; return true; }`)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(changed, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			truth := onNode(t, path)
			if truth.exitCode != 0 || got.exitCode != 70 || !strings.Contains(string(got.stderr), "records hold own keys only") || disagreement(truth, got) == "" {
				t.Fatalf("guard mutant survived: Node %+v; mutant %+v", truth, got)
			}
			t.Log("guard removal caught: native exit 70 with the missing-member message; Node finishes")
		})
	}
}

// The complete upstream any-valued helper is kept intact. The scalar typed
// slice above compiles; the general recursive helper still has an explicit gap.
func TestRecordCensusComparisonBuckets(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name  string
		equal bool
	}{{"17_compare_missing_object", false}, {"18_compare_missing_scalar", false}, {"19_compare_empty_objects", true}} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_buckets/"+probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			contents, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			// Stock Node executes the unchanged TypeScript source. Adamic's own
			// console declaration accepts strings only, unlike stock tsc's.
			stock := filepath.Join(t.TempDir(), "bucket.ts")
			if err := os.WriteFile(stock, contents, 0644); err != nil {
				t.Fatal(err)
			}
			node, err := exec.LookPath("node")
			if err != nil {
				t.Fatal(err)
			}
			got := execute(t, node, "--disable-warning=ExperimentalWarning", stock)
			result := "false"
			if probe.equal {
				result = "true"
			}
			var want strings.Builder
			for _, key := range []string{"constructor", "toString", "hasOwnProperty", "__proto__"} {
				want.WriteString(key + "\n" + result + "\ntrue\n")
			}
			if got.exitCode != 0 || string(got.stdout) != want.String() || len(got.stderr) != 0 {
				t.Fatalf("Node bucket changed: %+v", got)
			}
			adapted := strings.ReplaceAll(string(contents), "console.log(compareDataObjects(dst, src));", "console.log(String(compareDataObjects(dst, src)));")
			adapted = strings.ReplaceAll(adapted, "console.log(compareDataObjects(dst, own));", "console.log(String(compareDataObjects(dst, own)));")
			checked := filepath.Join(t.TempDir(), "bucket.a")
			if err := os.WriteFile(checked, []byte(adapted), 0644); err != nil {
				t.Fatal(err)
			}
			p, err := load.Load([]string{checked})
			if err != nil {
				t.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), p)
			if err == nil || !strings.Contains(err.Error(), "any") {
				t.Fatalf("want explicit any boundary, got %v", err)
			}
			t.Log(err)
		})
	}
}

func TestRecordHostEnvironmentObservation(t *testing.T) {
	file := filepath.Join(t.TempDir(), "host.mjs")
	if err := os.WriteFile(file, []byte(`console.log(Object.hasOwn(process.env, 'toString')); console.log(typeof process.env['toString']);`), 0644); err != nil {
		t.Fatal(err)
	}
	got := onNode(t, file)
	if got.exitCode != 0 || string(got.stdout) != "false\nfunction\n" || len(got.stderr) != 0 {
		t.Fatalf("host observation changed: %+v", got)
	}
}
