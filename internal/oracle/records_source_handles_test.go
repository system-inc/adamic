package oracle

import (
	"path/filepath"
	"testing"
)

// A record literal is related as the types the program holds it as, never its fresh type (Adamic's ruling on
// #ny0c3gb, lower's entriesRecordHandles). Each fixture of records_source_handles reads Options, which declares a
// required label, so only a record held as its own widened `{}` can reach it, and Hints, a weak type any record
// reaches. The bare, satisfies and generic forms have no handle, so both reads are weak_field_read; the control,
// handle.ts, holds its record as Record<string, number>, so only Hints's is, and Options's reads its slot. Every
// fixture lowers and agrees with Node natively and on the JavaScript backend. The mutant in
// internal/lower/record-source-handles.mutants.md, `satisfies T` counted as a holding place, turns satisfies red.
func TestRecordSourceHandles(t *testing.T) {
	t.Parallel()
	for name, checkedReads := range map[string]int{"bare": 2, "satisfies": 2, "generic": 2, "handle": 1} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_source_handles", name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("%s must lower: %v", name, err)
			}
			helpers := 0
			for _, function := range program.Functions {
				if function.Name == "weak_field_read" {
					helpers++
				}
			}
			if helpers != checkedReads {
				t.Fatalf("%d weak_field_read helpers, want %d: Options's read is checked only where the record has no handle, Hints's always", helpers, checkedReads)
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
}
