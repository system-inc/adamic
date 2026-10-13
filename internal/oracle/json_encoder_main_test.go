package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// These programs use only the schemas main can prove, without runtime providers.
func TestJSONEncoderMainFixtures(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"scalars", "values", "options", "escapes", "numbers", "undefined", "indent", "keys", "replacer", "census_append_overload"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fixture := "internal/oracle/testdata/json_stringify_" + name + ".a"
			if name == "census_append_overload" {
				fixture = "internal/oracle/testdata/" + name + ".a"
			}
			path, err := filepath.Abs(filepath.Join(repository, fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			recorded, err := os.ReadFile(filepath.Join(repository, "internal/oracle/counts.md"))
			if err != nil {
				t.Fatal(err)
			}
			if row := counted(t, fixture, false, nil, false, false); !strings.Contains(string(recorded), row+"\n") {
				t.Fatalf("counts row must be recorded: %s", row)
			}
			want := onNode(t, path)
			actual, binary := natively(t, program)
			for backend, got := range map[string]run{"native": actual, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s", backend, difference)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Error(report)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				if difference := disagreement(want, onWASI(t, native.C(program))); difference != "" {
					t.Errorf("wasm32-wasi: %s", difference)
				}
			}
		})
	}
}

// Production callers supply no providers; missing metadata must never be approximated.
func TestJSONEncoderMissingProviders(t *testing.T) {
	t.Parallel()
	code := jsonEncoderSource(t)
	for _, target := range []string{"native", "wasm32-wasi"} {
		t.Run(target, func(t *testing.T) {
			t.Parallel()
			options := native.Options{}
			if target == "wasm32-wasi" {
				if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
					t.Skip("set ADAMIC_ORACLE_WASI=1")
				}
				options.Target = target
			}
			binary := filepath.Join(t.TempDir(), "probe")
			if err := native.Build(code, binary, options); err != nil {
				t.Fatal(err)
			}
			for _, probe := range []struct{ arg, reason string }{
				{"empty_array", "JSON.stringify runtime array without complete element descriptors"},
				{"array", "JSON.stringify runtime array without complete element descriptors"},
				{"union", "JSON union lacks proven container metadata"},
				{"missing_hook", "JSON toJSON lacks a typed runtime result provider"},
			} {
				t.Run(probe.arg, func(t *testing.T) {
					var actual run
					if target == "native" {
						actual = execute(t, binary, probe.arg)
					} else {
						actual = execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "wasi.mjs"), binary, probe.arg)
					}
					if actual.exitCode != 70 || !strings.Contains(string(actual.stderr), probe.reason) {
						t.Fatalf("want exit 70 and %q: exit %d stdout %q stderr %q", probe.reason, actual.exitCode, actual.stdout, actual.stderr)
					}
				})
			}
		})
	}
}

func TestJSONEncoderGapNUL(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "gap.a")
	source := `console.log(JSON.stringify([1, [2]], null, 'a\u0000b') ?? 'undefined');
 console.log(JSON.stringify({ k: 1 }, null, '\u0000') ?? 'undefined');`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	actual, binary := natively(t, program)
	for backend, got := range map[string]run{"native": actual, "release": released(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s", backend, difference)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Error(report)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if difference := disagreement(want, onWASI(t, native.C(program))); difference != "" {
			t.Errorf("wasm32-wasi: %s", difference)
		}
	}
}
