package oracle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

const runtimeConstructorsDirectory = "internal/oracle/testdata/regexp_runtime_constructors/"

var runtimeConstructorFixtures = []string{"enums", "option_errors", "schema", "error_messages", "mobile_detect", "emoji"}

func init() {
	for _, name := range runtimeConstructorFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{runtimeConstructorsDirectory + name + ".a", true, false})
	}
}

func TestRuntimeConstructorFixtures(t *testing.T) {
	if got := execute(t, "node", "--version"); string(got.stdout) != "v24.19.0\n" {
		t.Fatal("requires Node 24.19.0")
	}
	for _, name := range runtimeConstructorFixtures {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, _ := filepath.Abs(filepath.Join(repository, runtimeConstructorsDirectory, name+".a"))
			expected := onNode(t, path)
			if expected.exitCode != 0 || len(expected.stderr) != 0 || len(expected.stdout) == 0 {
				t.Fatalf("invalid Node witness: %+v", expected)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			source := native.C(program)
			if !strings.Contains(source, "adamic_regex_compile_new(") || !strings.Contains(source, `#include "regexp_compile_runtime.c"`) {
				t.Fatal("fixture bypassed runtime compiler")
			}
			for backend, actual := range map[string]run{"javascript": onJavaScriptBackend(t, program), "native": func() run {
				r, b := natively(t, program)
				if failure := leaks(t, program, b); failure != "" {
					t.Fatal(failure)
				}
				return r
			}()} {
				if difference := disagreement(expected, actual); difference != "" {
					t.Fatalf("%s: %s\nNode %q\nactual %q stderr %q", backend, difference, expected.stdout, actual.stdout, actual.stderr)
				}
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				if difference := disagreement(expected, onWASI(t, source)); difference != "" {
					t.Fatal("WASI: " + difference)
				}
			}
		})
	}
}

func TestRuntimeConstructorCounts(t *testing.T) {
	t.Parallel()
	selected := map[string]string{}
	for _, name := range runtimeConstructorFixtures {
		path := runtimeConstructorsDirectory + name + ".a"
		selected[path] = counted(t, path, false, nil, false, false)
	}
	data, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	rows := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Split(line, " | ")
		if len(fields) == 7 && fields[0] != "| Fixture" {
			rows[strings.TrimPrefix(fields[0], "| ")] = line
		}
	}
	if !*updateCounts {
		for path, row := range selected {
			if rows[path] != row {
				t.Errorf("recorded %s, measured %s", rows[path], row)
			}
		}
		return
	}
	for path, row := range selected {
		rows[path] = row
	}
	var out strings.Builder
	out.WriteString(countsHeader)
	write := func(path string) {
		if row, ok := rows[path]; ok {
			out.WriteString(row + "\n")
			delete(rows, path)
		}
	}
	for _, fixture := range fixtures {
		if fixture.lowers && !uncounted[fixture.path] {
			write(fixture.path)
		}
	}
	for _, fixture := range slowRegExpFixtures {
		if fixture.lowers && !uncounted[fixture.path] {
			write(fixture.path)
		}
	}
	for _, fixture := range inputFixtures {
		write(fixture.path)
	}
	for _, name := range fsFileFixtures {
		write("internal/oracle/testdata/node_fs_file_" + name + ".a")
	}
	if len(rows) != 0 {
		t.Fatalf("unregistered existing rows: %v", rows)
	}
	if err := os.WriteFile(countsPath, []byte(out.String()), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestRuntimeConstructorSites(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repository, runtimeConstructorsDirectory, "sites.json"))
	if err != nil {
		t.Fatal(err)
	}
	var sites struct {
		Cohere  string
		Files   []struct{ File, SHA256 string }
		Dynamic []struct {
			File, Expression, Fixture string
			Line                      int
		}
	}
	if err := json.Unmarshal(data, &sites); err != nil {
		t.Fatal(err)
	}
	if sites.Cohere != "7945d102a6c18dd36adf9114a758ce646e8b2359" || len(sites.Dynamic) != 19 {
		t.Fatal("source pin or site count changed")
	}
	inputs, err := os.ReadFile(filepath.Join(repository, runtimeConstructorsDirectory, "input_sources.json"))
	if err != nil {
		t.Fatal(err)
	}
	var inputFiles []struct{ File, SHA256 string }
	if err := json.Unmarshal(inputs, &inputFiles); err != nil {
		t.Fatal(err)
	}
	sites.Files = append(sites.Files, inputFiles...)
	sources := map[string]string{}
	for _, file := range sites.Files {
		data, err := os.ReadFile(filepath.Join(repository, file.File))
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(data)
		if hex.EncodeToString(hash[:]) != file.SHA256 {
			t.Fatal("source changed: " + file.File)
		}
		sources[file.File] = strings.ReplaceAll(string(data), "\r\n", "\n")
	}
	for _, site := range sites.Dynamic {
		lines := strings.Split(sources[site.File], "\n")
		if site.Line < 1 || site.Line > len(lines) || !strings.Contains(lines[site.Line-1], site.Expression) {
			t.Fatalf("constructor site moved: %+v", site)
		}
		path := runtimeConstructorsDirectory + site.Fixture + ".a"
		if strings.HasPrefix(site.Fixture, "scout:") {
			path = scout23Directory + strings.TrimPrefix(site.Fixture, "scout:") + ".a"
		}
		if _, err := os.Stat(filepath.Join(repository, path)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRuntimeConstructorInputsGenerated(t *testing.T) {
	if os.Getenv("ADAMIC_TYPESCRIPT") == "" {
		t.Skip("set ADAMIC_TYPESCRIPT to the pinned TypeScript 6.0.3 module")
	}
	command := exec.Command("node", filepath.Join(runtimeConstructorsDirectory, "generate.mjs"), "--check")
	command.Dir = repository
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("generated input drift: %v %s", err, output)
	}
}

func runtimeConstructorArgs(t *testing.T, path, source string, args ...string) (run, run) {
	t.Helper()
	node := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path, args[0], args[1])
	js := filepath.Join(t.TempDir(), "program.mjs")
	if err := os.WriteFile(js, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	return node, execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), js, args[0], args[1])
}

func TestRuntimeConstructorFailureSplit(t *testing.T) {
	path, _ := filepath.Abs(filepath.Join(repository, runtimeConstructorsDirectory, "construction_failure.a"))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	binary := filepath.Join(t.TempDir(), "failures")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	cases := []struct{ pattern, flags, reason string }{
		{"[", "u", ""}, {"a", "ii", ""}, {"\\p{NotAProperty}", "u", ""}, {"a{2,1}", "u", ""},
		{"[\\q{a}]", "iv", "does not case-fold singleton"},
		{"(?i:a)[b]", "v", "scoped i modifier"},
		{"[\\q{ab|a|}]", "v", "mixed-length"},
		{"a{9223372036854775808,9223372036854775807}", "", "clamps quantifier bounds"},
	}
	for _, row := range cases {
		t.Run(row.pattern+row.flags, func(t *testing.T) {
			expected, js := runtimeConstructorArgs(t, path, javascript.JavaScript(program), row.pattern, row.flags)
			if d := disagreement(expected, js); d != "" {
				t.Fatal("JavaScript: " + d)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, row.pattern, row.flags)
			if row.reason == "" {
				if !bytes.Contains(expected.stdout, []byte("SyntaxError: ")) {
					t.Fatal("Node no longer rejects this pattern")
				}
				if d := disagreement(expected, actual); d != "" {
					t.Fatalf("syntax: %s %+v", d, actual)
				}
			} else {
				if !bytes.Contains(expected.stdout, []byte("accepted")) {
					t.Fatal("Node stopped accepting refusal witness")
				}
				if actual.exitCode != 70 || len(actual.stdout) != 0 || !bytes.Contains(actual.stderr, []byte(row.pattern)) || !bytes.Contains(actual.stderr, []byte(row.reason)) || bytes.Contains(actual.stderr, []byte("Sanitizer")) {
					t.Fatalf("valid pattern entered catch or lost its diagnostic: %+v", actual)
				}
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				wasm := filepath.Join(t.TempDir(), "failure.wasm")
				buildWASI(t, source, wasm)
				wasi := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), wasm, row.pattern, row.flags)
				if d := disagreement(actual, wasi); d != "" {
					t.Fatalf("WASI failure split: %s native %+v WASI %+v", d, actual, wasi)
				}
			}
		})
	}
}

// Run the scout's compiler shapes and the remaining constructors through the VM
// with a finite budget too. Setting a limit disables optimized matcher paths.
func TestRuntimeConstructorScoutBudget(t *testing.T) {
	names := append([]string{}, scout23Fixtures...)
	for _, name := range runtimeConstructorFixtures {
		names = append(names, "new:"+name)
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			directory := scout23Directory
			fixture := name
			if strings.HasPrefix(name, "new:") {
				directory = runtimeConstructorsDirectory
				fixture = strings.TrimPrefix(name, "new:")
			}
			path, _ := filepath.Abs(filepath.Join(repository, directory, fixture+".a"))
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected := onNode(t, path)
			if d := disagreement(expected, onJavaScriptBackend(t, program)); d != "" {
				t.Fatal("JavaScript: " + d)
			}
			source := native.C(program)
			old := "adamic_start(argc, argv);"
			if strings.Count(source, old) != 1 {
				t.Fatal("main initialization moved")
			}
			source = strings.Replace(source, old, old+"adamic_regex_set_step_limit(UINT64_C(100000000));", 1)
			binary := filepath.Join(t.TempDir(), "budget")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if d := disagreement(expected, actual); d != "" {
				t.Fatalf("budget native: %s %+v", d, actual)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				if d := disagreement(expected, onWASI(t, source)); d != "" {
					t.Fatal("budget WASI: " + d)
				}
			}
			t.Log("native and WASI VM step-limit=100000000 budget-stops=0 disagreements=0")
		})
	}
}
