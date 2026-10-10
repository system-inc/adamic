package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	additionalFixtureCounts = append(additionalFixtureCounts, instantiationProofCounts)
}

func instantiationProofCounts(t *testing.T) []string {
	t.Helper()
	var rows []string
	for _, name := range []string{"find", "first", "last", "callback", "array"} {
		rows = append(rows, counted(t, "internal/oracle/testdata/instantiation_"+name+".a", false, nil, false, false))
	}
	return rows
}

func instantiationProofOracle(t *testing.T, name string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/instantiation_"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 0 {
		t.Fatalf("Node: %+v", want)
	}
	native, binary := natively(t, program)
	for name, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program), "release": released(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s; Node %+v; got %+v", name, difference, want, got)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

func TestInstantiationFind(t *testing.T)     { t.Parallel(); instantiationProofOracle(t, "find") }
func TestInstantiationFirst(t *testing.T)    { t.Parallel(); instantiationProofOracle(t, "first") }
func TestInstantiationLast(t *testing.T)     { t.Parallel(); instantiationProofOracle(t, "last") }
func TestInstantiationCallback(t *testing.T) { t.Parallel(); instantiationProofOracle(t, "callback") }
func TestInstantiationArray(t *testing.T)    { t.Parallel(); instantiationProofOracle(t, "array") }

// Removing a present nullable return must produce a clean, observable disagreement,
// rather than a compiler or sanitizer failure.
func TestInstantiationPresenceMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/instantiation_first.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	changed := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		if !strings.HasPrefix(function.Name, "firstOrUndefined_") || function.Returns != ir.Object {
			continue
		}
		for statement, body := range function.Body {
			if _, ok := body.(ir.Return); ok {
				function.Body[statement] = ir.Return{Value: ir.Undefined{Of: function.Returns}}
				changed++
			}
		}
	}
	if changed != 1 {
		t.Fatalf("want one object instantiation return, changed %d", changed)
	}
	native, binary := natively(t, program)
	for name, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 {
			t.Fatalf("%s mutant did not execute cleanly: %+v", name, got)
		}
		if difference := disagreement(want, got); difference != "stdout differs" {
			t.Fatalf("%s: want Node to catch the mutant, got %q", name, difference)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
