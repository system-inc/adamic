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

// Prototype comparisons and the environment boundary use TestRecordRefusal* pins.
func init() {
	for _, name := range []string{"discarded", "guarded_snapshot", "compare_missing_scalar"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/records_" + name + ".a", true, false})
	}
}

func TestRecordObservationGuardMutants(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"discarded", "guarded_snapshot", "compare_missing_scalar"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/records_"+name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			code := native.C(program)
			// Remove the guard and restore the loud lookup, preserving valid C and
			// evaluation order. A missing inherited name now stops where Node prints.
			changed := strings.ReplaceAll(code, "adamic_record_has_own(", "record_mutant_has_own(")
			changed = strings.ReplaceAll(changed, "adamic_record_get_own(", "adamic_record_get(")
			if changed == code {
				test.Fatal("mutant changed nothing")
			}
			changed = insertCollectionMutant(
				changed,
				`static bool record_mutant_has_own(const adamic_record *r, const adamic_string *k) { (void)r; (void)k; return true; }`,
			)
			binary := filepath.Join(test.TempDir(), "mutant")
			if buildError := native.Build(changed, binary, native.Options{Sanitize: true}); buildError != nil {
				test.Fatal(buildError)
			}
			got := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			truth := onNode(test, path)
			if truth.exitCode != 0 ||
				got.exitCode != 70 ||
				!strings.Contains(string(got.stderr), "records hold own keys only") ||
				disagreement(truth, got) == "" {
				test.Fatalf("guard mutant survived: Node %+v; mutant %+v", truth, got)
			}
			test.Log("guard removal caught: native exit 70 with the missing-member message; Node finishes")
		})
	}
}

// The complete upstream any-valued helper is kept intact. The scalar typed
// slice above compiles; the general recursive helper still has an explicit gap.
func TestRecordCensusComparisonBuckets(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		name  string
		equal bool
	}{{"17_compare_missing_object", false}, {"18_compare_missing_scalar", false}, {"19_compare_empty_objects", true}} {
		test.Run(probe.name, func(test *testing.T) {
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/records_buckets/"+probe.name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			contents, readError := os.ReadFile(path)
			if readError != nil {
				test.Fatal(readError)
			}
			// Stock Node executes the unchanged TypeScript source. Adamic's own
			// console declaration accepts strings only, unlike stock tsc's.
			stock := filepath.Join(test.TempDir(), "bucket.ts")
			if writeError := os.WriteFile(stock, contents, 0644); writeError != nil {
				test.Fatal(writeError)
			}
			node, lookPathError := exec.LookPath("node")
			if lookPathError != nil {
				test.Fatal(lookPathError)
			}
			got := execute(test, node, "--disable-warning=ExperimentalWarning", stock)
			result := "false"
			if probe.equal {
				result = "true"
			}
			var want strings.Builder
			for _, key := range []string{"constructor", "toString", "hasOwnProperty", "__proto__"} {
				want.WriteString(key + "\n" + result + "\ntrue\n")
			}
			if got.exitCode != 0 || string(got.stdout) != want.String() || len(got.stderr) != 0 {
				test.Fatalf("Node bucket changed: %+v", got)
			}
			adapted := strings.ReplaceAll(
				string(contents),
				"console.log(compareDataObjects(dst, src));",
				"console.log(String(compareDataObjects(dst, src)));",
			)
			adapted = strings.ReplaceAll(
				adapted,
				"console.log(compareDataObjects(dst, own));",
				"console.log(String(compareDataObjects(dst, own)));",
			)
			checked := filepath.Join(test.TempDir(), "bucket.a")
			if writeError := os.WriteFile(checked, []byte(adapted), 0644); writeError != nil {
				test.Fatal(writeError)
			}
			program, loadError := load.Load([]string{checked})
			if loadError != nil {
				test.Fatal(loadError)
			}
			_, lowerError := lower.Lower(context.Background(), program)
			if lowerError == nil || !strings.Contains(lowerError.Error(), "any") {
				test.Fatalf("want explicit any boundary, got %v", lowerError)
			}
			test.Log(lowerError)
		})
	}
}

func TestRecordHostEnvironmentObservation(test *testing.T) {
	test.Parallel()
	file := filepath.Join(test.TempDir(), "host.mjs")
	if writeError := os.WriteFile(
		file,
		[]byte(`console.log(Object.hasOwn(process.env, 'toString')); console.log(typeof process.env['toString']);`),
		0644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	got := onNode(test, file)
	if got.exitCode != 0 || string(got.stdout) != "false\nfunction\n" || len(got.stderr) != 0 {
		test.Fatalf("host observation changed: %+v", got)
	}
}
