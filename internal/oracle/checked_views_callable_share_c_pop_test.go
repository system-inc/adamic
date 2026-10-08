package oracle

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func callableShareCPopRows(t *testing.T) []callableShareCRow {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-c/pop-certified.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []callableShareCRow
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	return rows
}

func TestCheckedViewCallableShareCPop(t *testing.T) {
	for _, row := range callableShareCPopRows(t) {
		t.Run(fmt.Sprintf("rank-%d", row.Rank), func(t *testing.T) {
			for _, variant := range []string{"good", "wrong-element"} {
				t.Run(variant, func(t *testing.T) {
					p, path := interfaceFixture(t, fmt.Sprintf("lane5/share-c/families/rank-%d/%s", row.Rank, variant))
					truth := onNode(t, path)
					if truth.exitCode != 0 || string(truth.stdout) != "completed\n" {
						t.Fatalf("Node: %#v", truth)
					}
					want := truth
					if variant != "good" {
						want = run{exitCode: 70, stderr: []byte(row.Diagnostic)}
					}
					sanitized, binary := nativelyUncached(t, p)
					for _, got := range []run{sanitized, releasedUncached(t, p), onJavaScriptBackend(t, p)} {
						if diff := disagreement(want, got); diff != "" {
							t.Fatalf("%s; stderr %q", diff, got.stderr)
						}
					}
					if variant == "good" {
						if report := leaksUncached(t, p, binary); report != "" {
							t.Fatal(report)
						}
					}
				})
			}
		})
	}
}

func TestCheckedViewCallableShareCPopMutants(t *testing.T) {
	for _, row := range callableShareCPopRows(t) {
		t.Run(fmt.Sprintf("rank-%d", row.Rank), func(t *testing.T) {
			p, _ := interfaceFixture(t, fmt.Sprintf("lane5/share-c/families/rank-%d/wrong-element", row.Rank))
			changed := 0
			omit := func(x ir.Expression) ir.Expression {
				pop, ok := x.(ir.ArrayPop)
				if ok && pop.ViewRead.View != "" {
					pop.ViewRead.View = ""
					changed++
					return pop
				}
				return x
			}
			mutateStringExpressions(reflect.ValueOf(&p.Main).Elem(), omit)
			mutateStringExpressions(reflect.ValueOf(&p.Functions).Elem(), omit)
			if changed != 1 {
				t.Fatalf("changed %d pop reads", changed)
			}
			want := run{exitCode: 70, stderr: []byte(row.Diagnostic)}
			for _, got := range []run{releasedUncached(t, p), onJavaScriptBackend(t, p)} {
				if diff := disagreement(want, got); diff == "" {
					t.Fatal("mutant survived")
				} else {
					t.Logf("caught omitted pop element certificate: %s; exit %d stdout %q stderr %q", diff, got.exitCode, got.stdout, got.stderr)
				}
				if got.exitCode != 0 || string(got.stdout) != "completed\n" {
					t.Fatalf("mutant failed to execute: %#v", got)
				}
			}
		})
	}
}
