package oracle

import (
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func init() {
	for _, name := range []string{"library_date_string", "library_date_string_utc"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name + ".a", true, false})
	}
}

// Compile once and run the same artifact in two zones. No parent environment
// changes: this can run alongside the ordinary parallel oracle safely.
func TestDateStringRuntimeZones(t *testing.T) {
	t.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/library_date_string.a")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	binary := filepath.Join(directory, "program")
	script := filepath.Join(directory, "program.mjs")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0644); err != nil {
		t.Fatal(err)
	}
	for _, zone := range []string{"UTC", "America/Denver"} {
		environment := []string{"TZ=" + zone}
		if runtime.GOOS == "linux" {
			environment = append(environment, "ASAN_OPTIONS=detect_leaks=1")
		}
		expected := executeWith(t, environment, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
		if expected.exitCode != 0 {
			t.Fatalf("source Node: %s", expected.stderr)
		}
		for _, actual := range []run{executeWith(t, environment, binary), executeWith(t, environment, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), script)} {
			if difference := disagreement(expected, actual); difference != "" {
				t.Fatalf("%s: %s; backend stderr %s", zone, difference, actual.stderr)
			}
		}
		t.Log(zone + ": native and JavaScript agree with Node")
	}
}

func TestDateStringNamedRefusal(t *testing.T) {
	t.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/library_date_string_refusal.a")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	binary := filepath.Join(directory, "program")
	script := filepath.Join(directory, "program.mjs")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0644); err != nil {
		t.Fatal(err)
	}
	environment := []string{"TZ=Europe/Paris"}
	if runtime.GOOS == "linux" {
		environment = append(environment, "ASAN_OPTIONS=detect_leaks=1")
	}
	expected := run{stdout: []byte("right operand\nDateStringNotYet\nDate.toString: exact ICU long zone name unavailable for runtime TZ\nInvalid Date\n")}
	for _, actual := range []run{executeWith(t, environment, binary), executeWith(t, environment, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), script)} {
		if difference := disagreement(expected, actual); difference != "" {
			t.Fatalf("named catchable refusal: %s; stdout %q, stderr %s", difference, actual.stdout, actual.stderr)
		}
	}
	source := executeWith(t, environment, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
	if source.exitCode != 0 || strings.Contains(string(source.stdout), "DateStringNotYet") {
		t.Fatal("stock Node should format Paris normally")
	}
	t.Log("Both backends loudly refuse where stock Node can format the zone")
}

// Each mutant changes only a private copy of the runtime and must finish cleanly.
// No implementation assertion decides the answer: the source on Node does.
func TestDateStringRuntimeMutants(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, before, after string }{
		{"offset sign", "offset<0?'-':'+'", "offset<0?'+':'-'"},
		{"weekday", "days[parts[6]]", "days[(parts[6]+1)%7]"},
		{"invalid spelling", "date_string_text(\"Invalid Date\")", "date_string_text(\"Invalid date\")"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(repository, "internal/oracle/testdata/library_date_string.a")
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			directory := t.TempDir()
			entries, err := os.ReadDir(filepath.Join(repository, "internal/native/runtime"))
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			for _, entry := range entries {
				if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".c") && !strings.HasSuffix(entry.Name(), ".h")) {
					continue
				}
				contents, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime", entry.Name()))
				if err != nil {
					t.Fatal(err)
				}
				if entry.Name() == "date_string.c" {
					if strings.Count(string(contents), probe.before) != 1 {
						t.Fatal("mutant must replace one runtime operation")
					}
					contents = []byte(strings.Replace(string(contents), probe.before, probe.after, 1))
					changed = true
				}
				if err := os.WriteFile(filepath.Join(directory, entry.Name()), contents, 0644); err != nil {
					t.Fatal(err)
				}
			}
			if !changed {
				t.Fatal("no runtime changed")
			}
			options := native.Options{Sanitize: true}
			library, err := native.RuntimeLibrary(directory, options)
			if err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(directory, "main.c")
			binary := filepath.Join(directory, "program")
			if err := os.WriteFile(source, []byte(native.C(program)), 0644); err != nil {
				t.Fatal(err)
			}
			arguments := append(native.Flags(options), "-I", filepath.Dir(library), "-o", binary, source)
			arguments = append(arguments, native.RuntimeLinkFlags(library)...)
			arguments = append(arguments, "-lm")
			if compiled := execute(t, "clang", arguments...); compiled.exitCode != 0 {
				t.Fatalf("mutant failed compilation: %s", compiled.stderr)
			}
			environment := []string{"TZ=UTC"}
			if runtime.GOOS == "linux" {
				environment = append(environment, "ASAN_OPTIONS=detect_leaks=1")
			}
			expected := executeWith(t, environment, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
			actual := executeWith(t, environment, binary)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("mutant failed execution: %d %s", actual.exitCode, actual.stderr)
			}
			if difference := disagreement(expected, actual); difference != "stdout differs" {
				t.Fatalf("Node must catch mutant only by stdout: %s", difference)
			}
			t.Log("compiled and finished without sanitizer/leak findings; caught only by Node stdout")
		})
	}
}

func TestDateStringWASIRefusal(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path := filepath.Join(repository, "internal/oracle/testdata/library_date_string.a")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "program.wasm")
	if err := native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"TZ=America/Denver"}, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
	expected := run{exitCode: 70, stderr: []byte("adamic: panic: DateStringNotYet: Date.toString: America/Denver requires a system zone database unavailable on wasm32-wasi\n")}
	if difference := disagreement(expected, actual); difference != "" {
		t.Fatalf("WASI target refusal: %s, stdout %q, stderr %s", difference, actual.stdout, actual.stderr)
	}
	t.Log("whole archive builds; Denver conversion loudly refuses missing WASI zone database")
}
