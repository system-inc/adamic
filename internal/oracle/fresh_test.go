package oracle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

// A closing write either remains refused or selects graph ownership. Acceptance
// must agree with source Node and return every value, including the closed cycle.
func TestFreshWriteProbesStayRefused(test *testing.T) {
	test.Parallel()
	probes, err := filepath.Glob(filepath.Join(repository, "internal", "oracle", "testdata", "fresh_refused", "*.a"))
	if err != nil {
		test.Fatal(err)
	}
	if len(probes) < 20 {
		test.Fatalf("found only %d probes", len(probes))
	}
	for _, probe := range probes {
		if _, owned := freshClobberOutputs[filepath.Base(probe)]; owned {
			continue // Each new witness has its own bounded top-level test.
		}
		test.Run(filepath.Base(probe), func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(probe)
			if err != nil {
				test.Fatal(err)
			}
			source, err := os.ReadFile(path)
			if err != nil {
				test.Fatal(err)
			}
			marked := 0
			for index, line := range strings.Split(string(source), "\n") {
				if strings.HasSuffix(line, "// closes the cycle") {
					marked = index + 1
				}
			}
			if marked == 0 {
				test.Fatal("no line is marked // closes the cycle")
			}
			program, err := lowered(test, path)
			var refused *lower.Refused
			if errors.As(err, &refused) {
				want := fmt.Sprintf("the write at %s:%d:", path, marked)
				if !strings.Contains(refused.Error(), want) ||
					!strings.Contains(refused.Error(), "(adamic/cycle-capable)") {
					test.Errorf("refused, but not naming the marked write (%q):\n%v", want, refused)
				}
				return
			}
			if err != nil {
				test.Fatalf("want the cycle refused, got %v", err)
			}
			checkGraphAdmission(test, path, program)
		})
	}
}

// Existing refusal fixtures become admissions only with graph metadata, both
// backends agreeing with Node, and the ordinary leak check still enabled.
func checkGraphAdmission(test *testing.T, path string, program *ir.Program) {
	test.Helper()
	if len(program.GraphTypes) == 0 {
		test.Fatal("cycle admitted without graph ownership")
	}
	truth := onNode(test, path)
	actual, binary := natively(test, program)
	if difference := disagreement(truth, actual); difference != "" {
		test.Fatal(difference)
	}
	if difference := disagreement(truth, onJavaScriptBackend(test, program)); difference != "" {
		test.Fatal(difference)
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
	checkGraphAdmissionCounts(test, program, truth)
}

// Admission cannot weaken the independent escape proof in internal/fresh.
func TestMethodKeepsArgumentAgreesWithNode(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/fresh_refused/devirt_fresh_method_keeps_argument.a"),
	)
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	truth := onNode(test, path)
	if truth.exitCode != 0 || string(truth.stdout) != "1\n" || len(truth.stderr) != 0 {
		test.Fatalf("Node: %+v", truth)
	}
	checkGraphAdmission(test, path, program)
}
