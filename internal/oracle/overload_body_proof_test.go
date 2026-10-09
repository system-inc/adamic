package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, name := range []string{"block", "evaluate"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/overload_body_proof/" + name + ".a", true, false})
	}
}

func checkOverloadBodyProof(t *testing.T, name, output string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/overload_body_proof", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != output {
		t.Fatalf("source Node: %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if program.PredicateChecks.Checked != 0 {
		t.Fatalf("inserted result checks: %+v", program.PredicateChecks)
	}
	for _, constant := range program.Strings {
		if strings.Contains(constant, "overload 1 of") {
			t.Fatalf("inserted overload boundary: %s", constant)
		}
	}
	native, sanitized := natively(t, program)
	for backend, actual := range map[string]run{"javascript": onJavaScriptBackend(t, program), "native": native, "release": released(t, program)} {
		if difference := disagreement(truth, actual); difference != "" {
			t.Fatalf("%s: %s; truth=%+v actual=%+v", backend, difference, truth, actual)
		}
	}
	if report := leaks(t, program, sanitized); report != "" {
		t.Fatal(report)
	}
}

func TestOverloadBodyProofBlock(t *testing.T) {
	t.Parallel()
	checkOverloadBodyProof(t, "block", "Block\ntrue\n3\nBlock\nBlock\nBlock\nExpression\n")
}

func TestOverloadBodyProofEvaluate(t *testing.T) {
	t.Parallel()
	checkOverloadBodyProof(t, "evaluate", "word\ntrue:true:false\nfalse\nmissing\n7\n")
}

func TestOverloadBodyProofRefusesNumericReturn(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/overload_body_proof/evaluate-negative.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "1\nfalse:false:false\nfalse\nmissing\n7\n" {
		t.Fatalf("source Node: %+v", truth)
	}
	program, err := lowered(t, path)
	var refused *lower.Refused
	if program != nil || !errors.As(err, &refused) {
		t.Fatalf("numeric return must be refused, got program=%v error=%v", program != nil, err)
	}
	for _, part := range []string{"overload 1 of evaluate", "result.value", "reachable return { value: 1", "evaluate-negative.a:"} {
		if !strings.Contains(err.Error(), part) {
			t.Fatalf("missing %q: %v", part, err)
		}
	}
}
