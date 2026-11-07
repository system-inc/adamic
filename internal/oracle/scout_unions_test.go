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

func TestNullableStringFieldCheckMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/scout_nullable_string_stale_field.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	source := onNode(test, path)
	if difference := disagreement(run{stdout: []byte("null\n")}, source); difference != "" {
		test.Fatal(difference)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := run{exitCode: 70, stderr: []byte(
		"adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back\n",
	)}
	native, _ := natively(test, program)
	for name, result := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(test, program)} {
		if difference := disagreement(want, result); difference != "" {
			test.Fatalf("%s: %s", name, difference)
		}
	}
	removed := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name == "narrowed_union_member" && len(function.Body) > 0 {
			if _, check := function.Body[0].(ir.If); check {
				function.Body = function.Body[1:]
				removed++
			}
		}
	}
	if removed != 1 {
		test.Fatalf("mutant removed %d checks, want 1", removed)
	}
	mutant := onJavaScriptBackend(test, program)
	if disagreement(want, mutant) == "" {
		test.Fatal("field tag check mutant survived")
	}
	if mutant.exitCode != 0 || string(mutant.stdout) != "undefined\n" || len(mutant.stderr) != 0 {
		test.Fatalf("mutant must complete with its unchecked field result: %+v", mutant)
	}
	test.Log(
		"Dropped field tag check mutant caught: JavaScript exits 0 and prints undefined; checked baseline panics with exit 70",
	)
}
