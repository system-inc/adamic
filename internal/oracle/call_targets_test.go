package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/call_targets_override_throw.a", true, false})
}

func TestCallTargetsOverrideIsCaught(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/call_targets_override_throw.a"))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if string(want.stdout) != "before\ncaught override\nafter\n" || want.exitCode != 0 {
		t.Fatalf("Node: %+v", want)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	got, sanitized := natively(t, program)
	if difference := disagreement(want, got); difference != "" {
		t.Fatalf("%s: Node %q; native %q", difference, want.stdout, got.stdout)
	}
	if difference := disagreement(want, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatal(difference)
	}
	if leaked := leaks(t, program, sanitized); leaked != "" {
		t.Fatal(leaked)
	}
	t.Logf("Node and native: %q", got.stdout)
}

// The dynamic method table stays intact. Only MayThrow's target set is reduced
// to the static method, so compilation and dispatch still work and the oracle
// must catch the omitted pending-exception check.
func TestCallTargetsStaticMayThrowMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/call_targets_override_throw.a"))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	mutated := false
	for _, class := range program.Classes {
		if class.Name == "Base" && len(class.Methods) == 1 {
			target := class.Methods[0]
			if len(program.MethodTargets[target]) < 2 || program.Functions[target].MayThrow {
				t.Fatal("fixture must have a nonthrowing base and a throwing override")
			}
			program.MethodTargets[target] = []int{target}
			mutated = true
		}
	}
	if !mutated {
		t.Fatal("fixture has no base target to mutate")
	}
	got, _ := natively(t, program)
	if difference := disagreement(want, got); difference != "stdout differs" {
		t.Fatalf("static MayThrow mutant must fail stdout comparison: %q, result %+v", difference, got)
	}
	if strings.Contains(string(got.stderr), "Sanitizer") || strings.Contains(string(got.stderr), "runtime error:") {
		t.Fatalf("mutant must be caught by Node comparison, not sanitizers: %s", got.stderr)
	}
	t.Logf("caught by Node stdout: want %q; mutant %q (exit %d)", want.stdout, got.stdout, got.exitCode)
}
