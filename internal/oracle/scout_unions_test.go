package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"testing"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/scout_hir_optional_boolean.a",
		"internal/oracle/testdata/scout_tsc_source_content.a",
		"internal/oracle/testdata/scout_tsc_nullish_content.a",
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
	t.Parallel()
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
