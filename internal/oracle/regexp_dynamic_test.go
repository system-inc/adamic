package oracle

import (
	"bytes"
	"errors"
	"github.com/system-inc/adamic/internal/native"
	reference "github.com/system-inc/adamic/internal/regexp"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	for _, name := range []string{"dynamic_gap", "id_length", "inline_comments", "warning_comments", "flags_errors", "ownership", "order"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/regexp_dynamic/" + name + ".a", true, false})
	}
}

func TestDynamicRegExpRuntimeRefusals(t *testing.T) {
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/regexp_dynamic/runtime_refusals.a"))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	node, js := onNode(t, path), onJavaScriptBackend(t, program)
	if diff := disagreement(node, js); diff != "" {
		t.Fatal(diff)
	}
	if string(node.stdout) != strings.Repeat("accepted\n", 5) {
		t.Fatalf("Node changed its divergences: %+v", node)
	}
	var expected strings.Builder
	for _, row := range [][2]string{{`[\q{a}]`, "iv"}, {`(?i:a)[b]`, "v"}, {`[\q{ab|a|}]`, "v"}, {`a{9223372036854775808,9223372036854775807}`, ""}} {
		p, err := reference.Compile(row[0], row[1])
		if err == nil {
			err = p.NativeCompatibility()
		}
		var divergence *reference.V8DivergenceError
		if !errors.As(err, &divergence) {
			t.Fatal(err)
		}
		expected.WriteString("SyntaxError: " + err.Error() + "\n")
	}
	wide, err := reference.Compile("a{18446744073709551616}", "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = wide.NativeDeclarations("wide")
	if err == nil {
		t.Fatal("expected counter-width refusal")
	}
	expected.WriteString("Error: " + err.Error() + "\n")
	actual, _ := natively(t, program)
	firstReason := strings.TrimPrefix(strings.Split(expected.String(), "\n")[0], "SyntaxError: ")
	if actual.exitCode != 70 || len(actual.stdout) != 0 || !bytes.Contains(actual.stderr, []byte("for /[\\q{a}]/iv: ")) || !bytes.Contains(actual.stderr, []byte(firstReason)) {
		t.Fatalf("Node-valid refusal became catchable or lost its reason: %+v", actual)
	}
	// The first refusal stops execution. TestRuntimeConstructorFailureSplit
	// runs all five independently, including on WASI.
	if os.Getenv("ADAMIC_ORACLE_WASI") != "" {
		wasm := onWASI(t, native.C(program))
		if diff := disagreement(actual, wasm); diff != "" {
			t.Fatal(diff)
		}
	}

}

func TestDynamicRegExpConstructionMutants(t *testing.T) {
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/regexp_dynamic/dynamic_gap.a"))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	runtime, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/regexp_compile_runtime.c"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct{ name, old, new, diagnostic string }{
		{"drop compiled storage owner", "adamic_regex_new_owned(program,source,flag_string,storage)", "adamic_regex_new(program,source,flag_string)", "heap-use-after-free"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			if bytes.Count(runtime, []byte(mutant.old)) != 1 {
				t.Fatal("mutation site moved")
			}
			changed := strings.Replace(string(runtime), mutant.old, mutant.new, 1)
			changedSource := strings.Replace(source, `#include "regexp_compile_runtime.c"`, changed, 1)
			if changedSource == source {
				t.Fatal("construction include missing")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(changedSource, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1"}, binary)
			if actual.exitCode == 0 || !bytes.Contains(actual.stderr, []byte(mutant.diagnostic)) {
				t.Fatalf("mutant did not reach ownership assertion: %+v", actual)
			}
			t.Logf("caught %s: %s", mutant.name, mutant.diagnostic)
		})
	}
}
