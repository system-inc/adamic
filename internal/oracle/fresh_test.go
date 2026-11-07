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
	probes, globError := filepath.Glob(
		filepath.Join(repository, "internal", "oracle", "testdata", "fresh_refused", "*.a"),
	)
	if globError != nil {
		test.Fatal(globError)
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
			path, absolutePathError := filepath.Abs(probe)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			source, readError := os.ReadFile(path)
			if readError != nil {
				test.Fatal(readError)
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
			program, loweredError := lowered(test, path)
			var refused *lower.Refused
			if errors.As(loweredError, &refused) {
				want := fmt.Sprintf("the write at %s:%d:", path, marked)
				if !strings.Contains(refused.Error(), want) ||
					!strings.Contains(refused.Error(), "(adamic/cycle-capable)") {
					test.Errorf("refused, but not naming the marked write (%q):\n%v", want, refused)
				}
				return
			}
			if loweredError != nil {
				test.Fatalf("want the cycle refused, got %v", loweredError)
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
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/fresh_refused/devirt_fresh_method_keeps_argument.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	truth := onNode(test, path)
	if truth.exitCode != 0 || string(truth.stdout) != "1\n" || len(truth.stderr) != 0 {
		test.Fatalf("Node: %+v", truth)
	}
	checkGraphAdmission(test, path, program)
}
