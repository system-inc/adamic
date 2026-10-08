package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func jsonEncoderSource(t *testing.T) string {
	t.Helper()
	var code strings.Builder
	for _, path := range []string{"internal/native/runtime/json_stringify.c", "internal/oracle/testdata/json_encoder_contract/runtime.c"} {
		data, err := os.ReadFile(filepath.Join(repository, path))
		if err != nil {
			t.Fatal(err)
		}
		code.Write(data)
		code.WriteByte('\n')
	}
	source := code.String()
	// Give the test copy distinct entry points; other runtime objects may pull in the archive encoder.
	for _, name := range []string{"adamic_json_stringify_runtime", "adamic_json_stringify"} {
		source = strings.ReplaceAll(source, name+"(", "probe_"+name+"(")
	}
	return source
}

func TestJSONEncoderContract(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/json_encoder_contract/contract.a"))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	code := jsonEncoderSource(t)
	t.Run("release", func(t *testing.T) {
		binary := filepath.Join(t.TempDir(), "probe")
		if err := native.Build(code, binary, native.Options{}); err != nil {
			t.Fatal(err)
		}
		if difference := disagreement(want, execute(t, binary)); difference != "" {
			t.Fatal(difference)
		}
	})
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		t.Run("wasi", func(t *testing.T) {
			if difference := disagreement(want, onWASI(t, code)); difference != "" {
				t.Fatal(difference)
			}
		})
	}
	mutations := []struct{ name, old, replacement string }{
		{"control", "", ""},
		{"array_descriptor", "case adamic_json_boolean: ascii(w, s.value.boolean ? \"true\" : \"false\");", "case adamic_json_boolean: ascii(w, s.value.boolean ? \"1\" : \"0\");"},
		{"tagged_null", "case adamic_json_null: ascii(w, \"null\"); return true;", "case adamic_json_null: return false;"},
		{"union_tag", "value.boolean = ((adamic_boolean_box *)reference)->boolean", "value.boolean = !((adamic_boolean_box *)reference)->boolean"},
		{"absent_element", "if (!present && adamic_thrown == NULL) { ascii(w, \"null\"); }", "if (!present && adamic_thrown == NULL) { ascii(w, \"false\"); }"},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			source := code
			if mutation.old != "" {
				source = strings.Replace(source, mutation.old, mutation.replacement, 1)
				if source == code {
					t.Fatal("mutant changed no encoder code")
				}
			}
			binary := filepath.Join(t.TempDir(), "probe")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("must finish cleanly: exit %d stderr %s", actual.exitCode, actual.stderr)
			}
			difference := disagreement(want, actual)
			if mutation.old == "" {
				if difference != "" {
					t.Fatalf("Node: %s\nNode %s\nencoder %s", difference, want.stdout, actual.stdout)
				}
				for _, probe := range []struct{ arg, reason string }{
					{"empty_array", "JSON.stringify runtime array without complete element descriptors"},
					{"cycle", "NotYet: JSON.stringify cyclic container"},
					{"array", "JSON.stringify runtime array without complete element descriptors"},
					{"element", "JSON.stringify runtime array without complete element descriptors"},
					{"union", "JSON union lacks proven container metadata"},
					{"hook", "JSON toJSON result descriptor is missing"},
				} {
					refused := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, probe.arg)
					if refused.exitCode == 0 || !strings.Contains(string(refused.stderr), probe.reason) {
						t.Fatalf("%s must refuse with %q: exit %d stderr %s", probe.arg, probe.reason, refused.exitCode, refused.stderr)
					}
				}
			} else if difference != "stdout differs" {
				t.Fatalf("Node alone must catch mutant: %q", difference)
			}
			t.Logf("%s: clean build/run/leak check; Node comparison %q", mutation.name, difference)
		})
	}
}
