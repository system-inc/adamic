package oracle

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"multimap", "multimap_values"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"internal/oracle/testdata/scout19_slice2_" + name + ".a", true, false,
		})
	}
}

// These are source witnesses, not stubs. Acceptance trips the expected refusal
// so the owning compiler proof can replace it with the three-backend oracle.
func TestScout19Slice2ExpectedRefusals(t *testing.T) {
	for _, name := range []string{"create_set", "presence", "range", "brands", "path", "generator"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout19_slice2_refused/scout19_slice2_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("source Node: %+v", truth)
			}
			loaded, err := load.Load([]string{path})
			if err == nil {
				_, err = lower.Lower(context.Background(), loaded)
			}
			if err == nil {
				t.Fatal("compiler gap closed: move this witness into the three-backend fixture registry")
			}
			expected := map[string]string{
				"create_set": "error TS2345", "range": "error TS2345",
				"presence": "refuses the non-null assertion", "brands": "refuses a cast the runtime can't check",
				"path":      "refuses a cast the runtime can't check",
				"generator": "refuses a generator function",
			}[name]
			if !strings.Contains(err.Error(), expected) {
				t.Fatalf("wrong gap, want %q: %v", expected, err)
			}
			t.Logf("source Node agrees with its protocol; expected compiler refusal: %v", err)
		})
	}
}

func TestScout19Slice2ExpandoStaysRefused(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "stage3/scout/19-slice2/multimap-composition/original.a"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	path = filepath.Join(t.TempDir(), "expando.a")
	if err := os.WriteFile(path, append(source, []byte("\ndeclare function unorderedRemoveItem<T>(array: T[], item: T): boolean;\n")...), 0644); err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) {
		t.Fatalf("expando must be refused: %v", err)
	}
	t.Logf("expando: %v", refused)
}

func TestScout19Slice2MultimapMutants(t *testing.T) {
	for _, name := range []string{"keep_empty_bucket", "skip_last_element_copy", "wrong_forEach_receiver"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout19_slice2_multimap.a"))
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			original := native.C(program)
			code, helper := "", ""
			switch name {
			case "wrong_forEach_receiver":
				changed := false
				var mutate func([]ir.Statement)
				mutate = func(body []ir.Statement) {
					for index, statement := range body {
						switch node := statement.(type) {
						case ir.Block:
							mutate(node.Body)
							body[index] = node
						case ir.ForOf:
							mutate(node.Body)
							body[index] = node
						case ir.Evaluate:
							call, ok := node.Value.(ir.CallClosure)
							if !ok || len(call.Arguments) != 3 {
								continue
							}
							call.Arguments[2] = ir.Undefined{Of: ir.Object}
							node.Value = call
							body[index] = node
							changed = true
						}
					}
				}
				for index, function := range program.Functions {
					if !strings.HasPrefix(function.Name, "multiMapForEach") {
						continue
					}
					mutate(function.Body)
					program.Functions[index] = function
				}
				if !changed {
					t.Fatal("forEach receiver mutant changed no callback")
				}
				code = native.C(program)
				if difference := disagreement(onNode(t, path), onJavaScriptBackend(t, program)); difference != "stdout differs" {
					t.Fatalf("Node must catch JavaScript receiver mutant: %q", difference)
				}
			case "keep_empty_bucket":
				code = strings.ReplaceAll(original, "adamic_map_delete(", "scout19_slice2_delete(")
				helper = `static bool scout19_slice2_delete(adamic_map *map, adamic_value key) { (void)map; (void)key; return false; }`
			case "skip_last_element_copy":
				code = strings.ReplaceAll(original, "adamic_array_set(", "scout19_slice2_copy(")
				helper = `static void scout19_slice2_copy(adamic_array *array, double target, adamic_value value) { (void)array; (void)target; (void)value; }`
			}
			if code == original {
				t.Fatal("mutant changed no code")
			}
			if helper != "" {
				code = insertCollectionMutant(code, helper)
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("mutant must be sanitizer-clean: %+v", actual)
			}
			truth := onNode(t, path)
			if difference := disagreement(truth, actual); difference != "stdout differs" {
				t.Fatalf("Node must catch mutant: %q", difference)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				actual = onWASI(t, code)
				if actual.exitCode != 0 || len(actual.stderr) != 0 {
					t.Fatalf("WASI mutant must finish cleanly: %+v", actual)
				}
				if difference := disagreement(truth, actual); difference != "stdout differs" {
					t.Fatalf("Node must catch WASI mutant: %q", difference)
				}
			}
			t.Log("clean exit 0; only Node stdout comparison caught mutant")
		})
	}
}

// This only proves the Node witness for a currently refused custom collection.
// It is not a claim that the native custom-Set protocol is implemented.
func TestScout19Slice2CustomSetSnapshotWitness(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout19_slice2_create_set.a"))
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mutated := strings.Replace(string(source), "arrayFrom(multiMap.values())", "multiMap.values()", 1)
	if mutated == string(source) {
		t.Fatal("snapshot mutant changed no source")
	}
	mutant := filepath.Join(t.TempDir(), "live-buckets.a")
	if err := os.WriteFile(mutant, []byte(mutated), 0644); err != nil {
		t.Fatal(err)
	}
	observe := func(path string) run {
		return execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
	}
	truth, actual := observe(path), observe(mutant)
	if truth.exitCode != 0 || actual.exitCode != 0 || len(truth.stderr) != 0 || len(actual.stderr) != 0 {
		t.Fatalf("Node witnesses must finish cleanly: %+v %+v", truth, actual)
	}
	if !strings.Contains(string(truth.stdout), "snapshot first:true:true,third:true:true,same-bucket:true:true,second:true:true") {
		t.Fatalf("source snapshot contract changed: %s", truth.stdout)
	}
	if difference := disagreement(truth, actual); difference != "stdout differs" {
		t.Fatalf("Node must catch replacing the shallow snapshot by live buckets: %q", difference)
	}
	t.Log("Node caught live bucket traversal; original outer snapshot retains live inner bucket aliases")
}
