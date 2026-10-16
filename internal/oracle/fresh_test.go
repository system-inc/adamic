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
func TestFreshWriteProbesStayRefused(t *testing.T) {
	t.Parallel()
	probes, err := filepath.Glob(filepath.Join(repository, "internal", "oracle", "testdata", "fresh_refused", "*.a"))
	if err != nil {
		t.Fatal(err)
	}
	if len(probes) < 20 {
		t.Fatalf("found only %d probes", len(probes))
	}
	for _, probe := range probes {
		if _, owned := freshClobberOutputs[filepath.Base(probe)]; owned {
			continue // Each new witness has its own bounded top-level test.
		}
		t.Run(filepath.Base(probe), func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(probe)
			if err != nil {
				t.Fatal(err)
			}
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			marked := 0
			for index, line := range strings.Split(string(source), "\n") {
				if strings.HasSuffix(line, "// closes the cycle") {
					marked = index + 1
				}
			}
			if marked == 0 {
				t.Fatal("no line is marked // closes the cycle")
			}
			program, err := lowered(t, path)
			var refused *lower.Refused
			if errors.As(err, &refused) {
				want := fmt.Sprintf("the write at %s:%d:", path, marked)
				if !strings.Contains(refused.Error(), want) || !strings.Contains(refused.Error(), "(adamic/cycle-capable)") {
					t.Errorf("refused, but not naming the marked write (%q):\n%v", want, refused)
				}
				return
			}
			if err != nil {
				t.Fatalf("want the cycle refused, got %v", err)
			}
			checkGraphAdmission(t, path, program)
		})
	}
}

// Existing refusal fixtures become admissions only with graph metadata, both
// backends agreeing with Node, and the ordinary leak check still enabled.
func checkGraphAdmission(t *testing.T, path string, program *ir.Program) {
	t.Helper()
	if len(program.GraphTypes) == 0 {
		t.Fatal("cycle admitted without graph ownership")
	}
	truth := onNode(t, path)
	actual, binary := natively(t, program)
	if difference := disagreement(truth, actual); difference != "" {
		t.Fatal(difference)
	}
	if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatal(difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
