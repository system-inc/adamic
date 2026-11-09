package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// recordSourceHandles is records_source_handles, as the test names its fixtures.
const recordSourceHandles = "internal/oracle/testdata/records_source_handles"

// weakFieldReads counts a lowered program's weak_field_read helpers, one per weak read.
func weakFieldReads(program *ir.Program) int {
	helpers := 0
	for _, function := range program.Functions {
		if function.Name == "weak_field_read" {
			helpers++
		}
	}
	return helpers
}

// A record literal is related as the types the program holds it as, never its fresh type, plus `{}`, every
// record's implicit source (Adamic's ruling on #ny0c3gb, lower's entriesRecordHandles and recordEmptyObject).
//
// The first group counts weak reads. The bare, satisfies and generic fixtures read Options, which declares a
// required label, so only a record held as its own widened `{}` can reach it, and Hints, a weak type any record
// reaches: their records have no handle, so both reads are weak_field_read. The control, handle.ts, holds its
// record as Record<string, number>, so only Hints's is, and Options's reads its slot. empty_source.ts holds its
// record as Held, which relates to its all-optional Read in neither direction, so only the `{}` question keeps
// the read checked. Every one lowers and agrees with Node natively and on the JavaScript backend.
//
// The second group is Language's roads on #ny0c3gb, each a record literal seen as `{}` and then as an
// all-optional Read whose number field holds the record's string, which Node prints. Each program's .stop file
// says how it ends: "lowers", then the read is weak_field_read and agrees with Node; "lowers, then stops: " and
// a panic, then the read is weak_field_read and stops there, exit 70, natively and on the JavaScript backend; or
// the stop lowering gives, whole, its path relative to the repository.
//
// The mutants in internal/lower/record-source-handles.mutants.md: `satisfies T` counted as a holding place turns
// satisfies red, and the `{}` question dropped turns empty_source and t2_parameter_return red.
func TestRecordSourceHandles(t *testing.T) {
	t.Parallel()
	for name, checkedReads := range map[string]int{"bare": 2, "satisfies": 2, "generic": 2, "handle": 1, "empty_source": 1} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, recordSourceHandles, name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("%s must lower: %v", name, err)
			}
			helpers := weakFieldReads(program)
			if helpers != checkedReads {
				t.Fatalf("%d weak_field_read helpers, want %d", helpers, checkedReads)
			}
			sites := map[string]int{}
			for _, site := range program.PlaceholderChecks {
				if site.Origin == "record storage" && site.Status == "checked" {
					sites[site.Use]++
				}
			}
			if sites["weak field read storage"] != helpers || sites["weak field read record value"] != helpers {
				t.Fatalf("checked sites %v for %d weak reads, want each one's storage test and value check", sites, helpers)
			}
			node := onNode(t, path)
			if node.exitCode != 0 {
				t.Fatalf("Node: %s", weakReadShown(node))
			}
			native, _ := natively(t, program)
			for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s disagrees with Node (%s): Node %s; %s %s", backend, difference, weakReadShown(node), backend, weakReadShown(got))
				}
			}
		})
	}
	for _, name := range []string{"t1_empty", "t2_parameter_return", "t3_assignment", "spread_copy", "iife_argument", "destructuring_default"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			relative := filepath.Join(recordSourceHandles, name+".ts")
			path, err := filepath.Abs(filepath.Join(repository, relative))
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(path + ".stop")
			if err != nil {
				t.Fatal(err)
			}
			stop := strings.TrimSpace(string(expected))
			node := onNode(t, path)
			if node.exitCode != 0 {
				t.Fatalf("Node: %s", weakReadShown(node))
			}
			program, err := lowered(t, path)
			if stop != "lowers" && !strings.HasPrefix(stop, "lowers, then stops: ") {
				if err == nil {
					t.Fatalf("lowered, want %q", stop)
				}
				if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(path, relative), ""); got != stop {
					t.Fatalf("stop: got %q, want %q", got, stop)
				}
				return
			}
			if err != nil {
				t.Fatalf("want %q, got the stop %v", stop, err)
			}
			if weakFieldReads(program) == 0 {
				t.Fatal("no weak_field_read: the read of a record seen through `{}` went unchecked")
			}
			native, _ := natively(t, program)
			for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if stop == "lowers" {
					if difference := disagreement(node, got); difference != "" {
						t.Fatalf("%s disagrees with Node (%s): Node %s; %s %s", backend, difference, weakReadShown(node), backend, weakReadShown(got))
					}
					continue
				}
				if want := "adamic: panic: " + strings.TrimPrefix(stop, "lowers, then stops: ") + "\n"; got.exitCode != 70 || string(got.stderr) != want {
					t.Fatalf("%s: want exit 70 and %q, got %s", backend, want, weakReadShown(got))
				}
			}
		})
	}
}
