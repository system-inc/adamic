package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"strings"
	"testing"
)

// Historical port programs are remeasured without claiming to undo their corpus workarounds.
func TestScoutUnionSourceOutcomes(t *testing.T) {
	for _, probe := range []struct{ name, output, stop string }{
		{"selector-optional-boolean.a", "true\n", "a field of type boolean | undefined"},
		{"values-optional-boolean.a", "true\n", "a field of type boolean | undefined"},
		{"hir-optional-boolean.a", "true\n", ""},
		{"tsc-source-content.a", "source\nmissing\n", ""},
		{"tsc-nullish-content.a", "string\nsource\nobject\nnull\nundefined\nundefined\n", ""},
		{"presence.a", "false\n[]\nfalse\ntrue\n[first]\ntrue\n", "refuses delete"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "docs/step-17-unions/probes", probe.name))
			if err != nil {
				t.Fatal(err)
			}
			source := onNode(t, path)
			if difference := disagreement(run{stdout: []byte(probe.output)}, source); difference != "" {
				t.Fatal(difference)
			}
			program, err := lowered(t, path)
			if probe.stop != "" {
				if err == nil || !strings.Contains(err.Error(), probe.stop) {
					t.Fatalf("expected stop %q, got %v", probe.stop, err)
				}
				t.Logf("CURRENT STOP: %v", err)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			backend := onJavaScriptBackend(t, program)
			native, sanitized := natively(t, program)
			for name, result := range map[string]run{"JavaScript": backend, "native": native, "release": released(t, program)} {
				if difference := disagreement(source, result); difference != "" {
					t.Errorf("%s: %s", name, difference)
				}
			}
			if leak := leaks(t, program, sanitized); leak != "" {
				t.Fatal(leak)
			}
			t.Log("CURRENT OUTCOME: both backends match source Node; sanitizers and leak check passed")
		})
	}
}

func init() {
	for _, path := range []string{
		"docs/step-17-unions/probes/hir-optional-boolean.a",
		"docs/step-17-unions/probes/tsc-source-content.a",
		"docs/step-17-unions/probes/tsc-nullish-content.a",
		"internal/oracle/testdata/scout_nullable_strings.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/scout_nullable_string_stale_field.a", true, true})
}

func TestNullableStringFieldCheckMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_nullable_string_stale_field.a"))
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("null\n")}, source); difference != "" {
		t.Fatal(difference)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back\n")}
	native, _ := natively(t, program)
	for name, result := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, result); difference != "" {
			t.Fatalf("%s: %s", name, difference)
		}
	}
	removed := 0
	for i := range program.Functions {
		f := &program.Functions[i]
		if f.Name == "narrowed_union_member" && len(f.Body) > 0 {
			if _, check := f.Body[0].(ir.If); check {
				f.Body = f.Body[1:]
				removed++
			}
		}
	}
	if removed != 1 {
		t.Fatalf("mutant removed %d checks, want 1", removed)
	}
	mutant := onJavaScriptBackend(t, program)
	if disagreement(want, mutant) == "" {
		t.Fatal("field tag check mutant survived")
	}
	if mutant.exitCode != 0 || string(mutant.stdout) != "undefined\n" || len(mutant.stderr) != 0 {
		t.Fatalf("mutant must complete with its unchecked field result: %+v", mutant)
	}
	t.Log("Dropped field tag check mutant caught: JavaScript exits 0 and prints undefined; checked baseline panics with exit 70")
}
