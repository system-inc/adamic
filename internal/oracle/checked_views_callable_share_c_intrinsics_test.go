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

func callableShareCArrayRows(t *testing.T) []callableShareCRow {
	t.Helper()
	data, e := os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-c/array-certified.json"))
	if e != nil {
		t.Fatal(e)
	}
	var rows []callableShareCRow
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	return rows
}
func TestCheckedViewCallableShareCArrayIntrinsics(t *testing.T) {
	for _, row := range callableShareCArrayRows(t) {
		t.Run(fmt.Sprintf("rank-%d", row.Rank), func(t *testing.T) {
			for _, variant := range []string{"good", "wrong-element"} {
				t.Run(variant, func(t *testing.T) {
					p, path := interfaceFixture(t, fmt.Sprintf("lane5/share-c/families/rank-%d/%s", row.Rank, variant))
					truth := onNode(t, path)
					if truth.exitCode != 0 || string(truth.stdout) != "2\n" {
						t.Fatalf("Node: %#v", truth)
					}
					want := truth
					if variant != "good" {
						want = run{exitCode: 70, stderr: []byte(row.Diagnostic)}
					}
					got, binary := nativelyUncached(t, p)
					for _, r := range []run{got, releasedUncached(t, p), onJavaScriptBackend(t, p)} {
						if diff := disagreement(want, r); diff != "" {
							t.Fatalf("%s; stderr %q", diff, r.stderr)
						}
					}
					if variant == "good" {
						if r := leaksUncached(t, p, binary); r != "" {
							t.Fatal(r)
						}
					}
				})
			}
		})
	}
}
func TestCheckedViewCallableShareCArrayMutants(t *testing.T) {
	for _, row := range callableShareCArrayRows(t) {
		t.Run(fmt.Sprintf("rank-%d", row.Rank), func(t *testing.T) {
			p, _ := interfaceFixture(t, fmt.Sprintf("lane5/share-c/families/rank-%d/wrong-element", row.Rank))
			changed := 0
			omit := func(x ir.Expression) ir.Expression {
				v, ok := x.(ir.ArrayPush)
				if ok {
					v.DictionaryProduction = true
					changed++
					return v
				}
				return x
			}
			mutateStringExpressions(reflect.ValueOf(&p.Main).Elem(), omit)
			mutateStringExpressions(reflect.ValueOf(&p.Functions).Elem(), omit)
			if changed != 1 {
				t.Fatalf("changed %d pushes", changed)
			}
			want := run{exitCode: 70, stderr: []byte(row.Diagnostic)}
			for _, m := range []run{releasedUncached(t, p), onJavaScriptBackend(t, p)} {
				if diff := disagreement(want, m); diff == "" {
					t.Fatal("mutant survived")
				} else {
					t.Logf("caught omitted Array.push write certificate: %s; exit %d stdout %q stderr %q", diff, m.exitCode, m.stdout, m.stderr)
				}
				if m.exitCode != 0 || string(m.stdout) != "2\n" {
					t.Fatalf("mutant did not execute: %#v", m)
				}
			}
		})
	}
}
