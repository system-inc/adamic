package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	for _, name := range []string{"find", "first", "last", "callback", "array"} {
		path := "internal/oracle/testdata/instantiation_" + name + ".a"
		additionalFixtureCounts = append(
			additionalFixtureCounts,
			countCase{path, func(test *testing.T) (string, func() string) {
				return counted(test, path, false, nil, false, false)
			}},
		)
	}
}

func instantiationProofOracle(test *testing.T, name string) {
	test.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/instantiation_"+name+".a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	want := onNode(test, path)
	if want.exitCode != 0 {
		test.Fatalf("Node: %+v", want)
	}
	native, binary := natively(test, program)
	for name, got := range map[string]run{
		"native":     native,
		"JavaScript": onJavaScriptBackend(test, program),
		"release":    released(test, program),
	} {
		if difference := disagreement(want, got); difference != "" {
			test.Errorf("%s: %s; Node %+v; got %+v", name, difference, want, got)
		}
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
}

func TestInstantiationFind(test *testing.T) { test.Parallel(); instantiationProofOracle(test, "find") }
func TestInstantiationFirst(test *testing.T) {
	test.Parallel()
	instantiationProofOracle(test, "first")
}
func TestInstantiationLast(test *testing.T) { test.Parallel(); instantiationProofOracle(test, "last") }
func TestInstantiationCallback(test *testing.T) {
	test.Parallel()
	instantiationProofOracle(test, "callback")
}
func TestInstantiationArray(test *testing.T) {
	test.Parallel()
	instantiationProofOracle(test, "array")
}

// Removing a present nullable return must produce a clean, observable disagreement,
// rather than a compiler or sanitizer failure.
func TestInstantiationPresenceMutant(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/instantiation_first.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	want := onNode(test, path)
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
		test.Fatalf("want one object instantiation return, changed %d", changed)
	}
	native, binary := natively(test, program)
	for name, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(test, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 {
			test.Fatalf("%s mutant did not execute cleanly: %+v", name, got)
		}
		if difference := disagreement(want, got); difference != "stdout differs" {
			test.Fatalf("%s: want Node to catch the mutant, got %q", name, difference)
		}
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
}
