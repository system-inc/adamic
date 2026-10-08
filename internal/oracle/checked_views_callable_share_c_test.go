package oracle

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type callableShareCRow struct {
	Rank, Reads, FoundArity               int
	Status, Read, Declaration, Diagnostic string
}

func callableShareCRows(t *testing.T) []callableShareCRow {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-c/certified.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []callableShareCRow
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	return rows
}
func TestCheckedViewCallableShareC(t *testing.T) {
	for _, row := range callableShareCRows(t) {
		t.Run(fmt.Sprintf("rank-%d", row.Rank), func(t *testing.T) {
			for _, variant := range []string{"good", "wrong-arity"} {
				t.Run(variant, func(t *testing.T) {
					program, path := interfaceFixture(t, fmt.Sprintf("lane5/share-c/families/rank-%d/%s", row.Rank, variant))
					truth := onNode(t, path)
					if truth.exitCode != 0 || string(truth.stdout) != "completed\n" {
						t.Fatalf("Node: %#v", truth)
					}
					want := truth
					if variant != "good" {
						want = run{exitCode: 70, stderr: []byte(row.Diagnostic)}
					}
					sanitized, binary := nativelyUncached(t, program)
					for _, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
						if difference := disagreement(want, got); difference != "" {
							t.Fatalf("%s; stderr %q", difference, got.stderr)
						}
					}
					if variant == "good" {
						if report := leaksUncached(t, program, binary); report != "" {
							t.Fatal(report)
						}
					}
				})
			}
		})
	}
}
func TestCheckedViewCallableShareCMutants(t *testing.T) {
	for _, row := range callableShareCRows(t) {
		t.Run(fmt.Sprintf("rank-%d", row.Rank), func(t *testing.T) {
			program, _ := interfaceFixture(t, fmt.Sprintf("lane5/share-c/families/rank-%d/wrong-arity", row.Rank))
			changed := 0
			omit := func(value ir.Expression) ir.Expression {
				p, ok := value.(ir.Property)
				if ok && p.View == row.Read && p.ViewContract != 0 {
					p.ViewContract = 0
					changed++
					return p
				}
				return value
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), omit)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), omit)
			if changed != 1 {
				t.Fatalf("mutant changed %d reads", changed)
			}
			want := run{exitCode: 70, stderr: []byte(row.Diagnostic)}
			for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference == "" {
					t.Fatal("mutant survived")
				} else {
					t.Logf("caught missing callable certificate: %s; exit %d stdout %q stderr %q", difference, got.exitCode, got.stdout, got.stderr)
				}
				if got.exitCode != 0 || string(got.stdout) != "completed\n" {
					t.Fatalf("mutant failed to execute successfully: %#v", got)
				}
			}
		})
	}
}
func TestCheckedViewCallableShareCCounts(t *testing.T) {
	path := filepath.Join(repository, "internal/oracle/counts.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	for _, row := range callableShareCRows(t) {
		for _, variant := range []string{"good", "wrong-arity"} {
			p := fmt.Sprintf("stage3/interface-downcasts/lane5/share-c/families/rank-%d/%s.a", row.Rank, variant)
			measured := counted(t, checkedViewFixturePath(p), false, nil, false, false)
			if !*updateCounts {
				if !strings.Contains(string(data), measured+"\n") {
					t.Errorf("unrecorded counts: %s", measured)
				}
				continue
			}
			key := strings.Split(measured, " | ")[0] + " | "
			found := false
			for i, line := range lines {
				if strings.HasPrefix(line, key) {
					lines[i] = measured
					found = true
					break
				}
			}
			if !found {
				lines = append(lines, measured)
			}
		}
	}
	if *updateCounts {
		if err = os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
}
