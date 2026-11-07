package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func jsonEncoderSource(test *testing.T) string {
	test.Helper()
	var code strings.Builder
	for _, path := range []string{
		"internal/native/runtime/json_stringify.c",
		"internal/oracle/testdata/json_encoder_contract/runtime.c",
	} {
		data, err := os.ReadFile(filepath.Join(repository, path))
		if err != nil {
			test.Fatal(err)
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

func jsonEncoderContract(test *testing.T, backend int) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/json_encoder_contract/contract.a"))
	if err != nil {
		test.Fatal(err)
	}
	want := onNodeWith(test, inputRun{}, path)
	code := jsonEncoderSource(test)
	if backend == 0 {
		test.Run("release", func(test *testing.T) {
			binary := filepath.Join(test.TempDir(), "probe")
			if err := native.Build(code, binary, native.Options{}); err != nil {
				test.Fatal(err)
			}
			if difference := disagreement(want, execute(test, binary)); difference != "" {
				test.Fatal(difference)
			}
		})
		return
	}
	if backend == 1 {
		if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
			test.Run("wasi", func(test *testing.T) {
				if difference := disagreement(want, onWASI(test, code)); difference != "" {
					test.Fatal(difference)
				}
			})
		}
		return
	}
	mutations := []struct{ name, old, replacement string }{
		{"control", "", ""},
		{
			"array_descriptor",
			"case adamic_json_boolean: ascii(w, s.value.boolean ? \"true\" : \"false\");",
			"case adamic_json_boolean: ascii(w, s.value.boolean ? \"1\" : \"0\");",
		},
		{
			"tagged_null",
			"case adamic_json_null: ascii(w, \"null\"); return true;",
			"case adamic_json_null: return false;",
		},
		{
			"union_tag",
			"value.boolean = ((adamic_boolean_box *)reference)->boolean",
			"value.boolean = !((adamic_boolean_box *)reference)->boolean",
		},
		{
			"absent_element",
			"if (!present && adamic_thrown == NULL) { ascii(w, \"null\"); }",
			"if (!present && adamic_thrown == NULL) { ascii(w, \"false\"); }",
		},
	}
	for _, mutation := range mutations[backend-2 : backend-1] {
		test.Run(mutation.name, func(test *testing.T) {
			source := code
			if mutation.old != "" {
				source = strings.Replace(source, mutation.old, mutation.replacement, 1)
				if source == code {
					test.Fatal("mutant changed no encoder code")
				}
			}
			binary := filepath.Join(test.TempDir(), "probe")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				test.Fatal(err)
			}
			actual := executeWith(
				test,
				[]string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"},
				binary,
			)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				test.Fatalf("must finish cleanly: exit %d stderr %s", actual.exitCode, actual.stderr)
			}
			difference := disagreement(want, actual)
			if mutation.old == "" {
				if difference != "" {
					test.Fatalf("Node: %s\nNode %s\nencoder %s", difference, want.stdout, actual.stdout)
				}
				for _, probe := range []struct{ arg, reason string }{
					{"empty_array", "JSON.stringify runtime array without complete element descriptors"},
					{"cycle", "NotYet: JSON.stringify cyclic container"},
					{"array", "JSON.stringify runtime array without complete element descriptors"},
					{"element", "JSON.stringify runtime array without complete element descriptors"},
					{"union", "JSON union lacks proven container metadata"},
					{"hook", "JSON toJSON result descriptor is missing"},
					{"missing_hook", "JSON toJSON lacks a typed runtime result provider"},
				} {
					refused := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, probe.arg)
					if refused.exitCode == 0 || !strings.Contains(string(refused.stderr), probe.reason) {
						test.Fatalf(
							"%s must refuse with %q: exit %d stderr %s",
							probe.arg,
							probe.reason,
							refused.exitCode,
							refused.stderr,
						)
					}
				}
			} else if difference != "stdout differs" {
				test.Fatalf("Node alone must catch mutant: %q", difference)
			}
			test.Logf("%s: clean build/run/leak check; Node comparison %q", mutation.name, difference)
		})
	}
}

func TestJSONEncoderContract_000(test *testing.T) { test.Parallel(); jsonEncoderContract(test, 0) }

func TestJSONEncoderContract_001(test *testing.T) { test.Parallel(); jsonEncoderContract(test, 1) }

func TestJSONEncoderContract_002(test *testing.T) { test.Parallel(); jsonEncoderContract(test, 2) }

func TestJSONEncoderContract_003(test *testing.T) { test.Parallel(); jsonEncoderContract(test, 3) }

func TestJSONEncoderContract_004(test *testing.T) { test.Parallel(); jsonEncoderContract(test, 4) }

func TestJSONEncoderContract_005(test *testing.T) { test.Parallel(); jsonEncoderContract(test, 5) }

func TestJSONEncoderContract_006(test *testing.T) { test.Parallel(); jsonEncoderContract(test, 6) }
