package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

func TestMapNullableMixedValues(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name    string
		counted bool
		boxes   int
	}{
		{"mixed", false, 0},
		{"scalars", false, 0},
		{"constructor", false, 0},
		{"copy", false, 0},
		{"nullable_array", false, 0},
		{"self_link", true, 0},
		{"array_self_link", true, 0},
		{"find_self_link", true, 2},
		{"find_constructor_link", true, 2},
		{"find_array_link", true, 4},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path := nullableMapFixture(t, probe.name)
			want := nullableMapNode(t, path)
			if want.exitCode != 0 || len(want.stderr) != 0 || len(want.stdout) == 0 {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", want.exitCode, want.stdout, want.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("lowering: %v; Node stdout %q", err, want.stdout)
			}
			if probe.boxes > 0 {
				t.Run("boxed-road", func(t *testing.T) {
					boxes := 0
					for _, fn := range program.Functions {
						if fn.Name == "nullable_object_to_union" {
							boxes++
						}
					}
					if boxes != probe.boxes {
						t.Fatalf("got %d nullable_object_to_union helpers, want %d", boxes, probe.boxes)
					}
				})
			}
			t.Run("agreement", func(t *testing.T) {
				jsPath := filepath.Join(t.TempDir(), "program.mjs")
				if err := os.WriteFile(jsPath, []byte(javascript.JavaScript(program)), 0600); err != nil {
					t.Fatal(err)
				}
				js, release := nullableMapNode(t, jsPath), releasedUncached(t, program)
				t.Logf("Node stdout %q; JavaScript stdout %q; native stdout %q", want.stdout, js.stdout, release.stdout)
				if diff := disagreement(want, js); diff != "" {
					t.Errorf("JavaScript versus Node: %s; exit %d stderr %q", diff, js.exitCode, js.stderr)
				}
				if diff := disagreement(want, release); diff != "" {
					t.Errorf("release native versus Node: %s; exit %d stderr %q", diff, release.exitCode, release.stderr)
				}
			})
			if probe.counted {
				t.Run("counts", func(t *testing.T) {
					checkNullableMapCounts(t, program, want)
				})
			}
		})
	}
}

func checkNullableMapCounts(t *testing.T, program *ir.Program, want run) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "counted")
	if err := native.Build(native.C(program), binary, native_runtime.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	actual := execute(t, binary)
	counts := leakcheck.CountsLine.FindSubmatch(actual.stderr)
	t.Logf("counted native: exit %d stdout %q stderr:\n%s", actual.exitCode, actual.stdout, actual.stderr)
	if len(program.GraphTypes) == 0 {
		t.Error("the cycle must be owned by a graph region")
	}
	if actual.exitCode != want.exitCode || string(actual.stdout) != string(want.stdout) {
		t.Errorf("counted native disagrees with Node: exit %d stdout %q; Node exit %d stdout %q", actual.exitCode, actual.stdout, want.exitCode, want.stdout)
	}
	if counts == nil || string(counts[1]) != string(counts[2]) || string(counts[6]) != "0" {
		t.Error("want allocations equal to frees and zero values in regions")
	}
	if !t.Failed() {
		checkGraphAdmissionCounts(t, program, want)
	}
}

// The shared gate builds a sanitizer archive while constructing its cache key.
// Use the plain Node oracle directly for these fixtures.
func nullableMapNode(t *testing.T, path string) run {
	t.Helper()
	return execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), path)
}

func TestMapNullableMixedUnownedCycle(t *testing.T) {
	t.Parallel()
	path := nullableMapFixture(t, "unowned_cycle")
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "a\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	_, err := lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "a record written into a graph slot") {
		t.Fatalf("want the refusal for a record outside graph ownership, got %v", err)
	}
	t.Log(err)
}

func TestMapNullableMixedBoxedRecordControl(t *testing.T) {
	t.Parallel()
	path := nullableMapFixture(t, "boxed_record_control")
	truth := nullableMapNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "record present\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	_, err := lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "a record written into a graph slot") {
		t.Fatalf("want the refusal for a record outside graph ownership, got %v", err)
	}
	t.Log(err)
}

func nullableMapFixture(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/map_nullable_mixed_values", name+".ts"))
	if err != nil {
		t.Fatal(err)
	}
	return path
}
