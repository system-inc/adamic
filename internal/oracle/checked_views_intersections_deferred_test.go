package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// The root read checks the descendant kind; selecting it owns member selection.
func TestCheckedViewIntersectionDeferredMember(t *testing.T) {
	for _, sample := range []struct{ name, output, expression string }{
		{"unread", "1\n", ""},
		{"read", "42\n", "root.descendant"},
		{"helper", "42\n", "node.descendant"},
		{"callback", "42\n", "root.descendant"},
		{"destructure", "42\n", "descendant: selected (field descendant)"},
	} {
		t.Run(sample.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane7/deferred-member-"+sample.name)
			truth := run{stdout: []byte(sample.output)}
			if difference := disagreement(truth, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			bounded := false
			descendant := ir.ViewContractID(0)
			for _, contract := range program.ViewContracts {
				bounded = bounded || contract.IntersectionBounded
				for _, field := range contract.Fields {
					if field.Name == "descendant" {
						descendant = field.Contract
					}
				}
			}
			if !bounded || descendant == 0 {
				t.Fatal("fixture did not retain bounded root and descendant contract")
			}
			selected := &program.ViewContracts[descendant-1]
			if selected.Kind != ir.ViewUnion || selected.Unsupported != "" || ir.ViewUnionHasDiscriminant(program.ViewContracts, *selected) {
				t.Fatal("fixture must use a supported untagged descendant")
			}
			if os.Getenv("ADAMIC_INTERSECTION_DEFERRED_MUTANT") != "" {
				// Drop only union selection; common field readiness/kind remains checked.
				selected.Kind = ir.ViewObject
			}
			want := truth
			if sample.expression != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: " + sample.expression + " matches no member of Left | Right; expected Left | Right, found object\n")}
			}
			actual, binary := nativelyUncached(t, program)
			if want.exitCode == 0 {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := viewReadDisagreement(want, got, program); difference != "" {
					t.Errorf("%s; got %#v", difference, got)
				}
			}
		})
	}
}

func viewIntersectionDeferredCounts(t *testing.T) []string {
	t.Helper()
	var rows []string
	for _, name := range []string{"unread", "read", "helper", "callback", "destructure"} {
		rows = append(rows, counted(t, checkedViewFixturePath("stage3/interface-downcasts/lane7/deferred-member-"+name+".a"), false, nil, false, false))
	}
	return rows
}

// Not parallel: the update writes the shared measured counts table.
func TestCheckedViewIntersectionDeferredCounts(t *testing.T) {
	rows := viewIntersectionDeferredCounts(t)
	path := filepath.Join(repository, "internal/oracle/counts.md")
	data, err := os.ReadFile(checkedViewFixturePath(path))
	if err != nil {
		t.Fatal(err)
	}
	if !*updateCounts {
		for _, row := range rows {
			if !strings.Contains(string(data), row+"\n") {
				t.Errorf("unrecorded deferred fixture: %s", row)
			}
		}
		return
	}
	parts := strings.SplitN(string(data), "\n## Predicate direction counts", 2)
	lines := strings.Split(strings.TrimSuffix(parts[0], "\n"), "\n")
	for _, row := range rows {
		key := strings.Split(row, " | ")[0] + " | "
		found := false
		for i, line := range lines {
			if strings.HasPrefix(line, key) {
				lines[i] = row
				found = true
				break
			}
		}
		if !found {
			lines = append(lines, row)
		}
	}
	updated := strings.Join(lines, "\n") + "\n"
	if len(parts) == 2 {
		updated += "\n## Predicate direction counts" + parts[1]
	}
	if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
		t.Fatal(err)
	}
}
