package oracle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

const scout23Directory = "internal/oracle/testdata/scout23_tsc_regex/"

var scout23Fixtures = []string{"literals", "wildcards", "module_specifiers", "pragma", "program_path", "last_index"}

func init() {
	for _, name := range scout23Fixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{scout23Directory + name + ".a", true, false})
	}
	for _, name := range []string{"callback_escape", "callback_wildcard", "module_specifiers_assignment"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{scout23Directory + name + ".a", true, false})
	}
}

// This source corpus is consumed to protect every location and literal, not a
// record of a compiler run. It uses the pinned compiler snapshot's real paths.
func TestScout23SourceCorpus(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join(repository, scout23Directory, "sites.json"))
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Cohere string
		Sites  []struct {
			File, Kind, Expression, Pattern, Flags string
			Line, Column                           int
		}
		Files []struct{ File, SHA256 string }
	}
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	if corpus.Cohere != "7945d102a6c18dd36adf9114a758ce646e8b2359" {
		t.Fatal("unexpected compiler source pin")
	}
	counts := map[string]int{}
	sources := map[string]string{}
	for _, file := range corpus.Files {
		data, err := os.ReadFile(filepath.Join(repository, file.File))
		if err != nil {
			t.Fatal(err)
		}
		if got := sha256.Sum256(data); strings.ToLower(file.SHA256) != strings.ToLower(hex.EncodeToString(got[:])) {
			t.Errorf("source changed: %s", file.File)
		}
		sources[file.File] = strings.ReplaceAll(string(data), "\r\n", "\n")
	}
	literalFixture, err := os.ReadFile(filepath.Join(repository, scout23Directory, "literals.a"))
	if err != nil {
		t.Fatal(err)
	}
	for _, site := range corpus.Sites {
		counts[site.Kind]++
		lines := strings.Split(sources[site.File], "\n")
		if site.Line < 1 || site.Line > len(lines) {
			t.Fatalf("invalid source location: %+v", site)
		}
		// Multiline calls retain their full text; check their first source line.
		first := strings.Split(strings.ReplaceAll(site.Expression, "\r\n", "\n"), "\n")[0]
		if !strings.Contains(lines[site.Line-1], first) {
			t.Errorf("site moved: %s:%d %s", site.File, site.Line, first)
		}
		if site.Kind == "literal" && !bytes.Contains(literalFixture, []byte("literal: "+site.Expression+",")) {
			t.Errorf("literal missing from dual compilation fixture: %+v", site)
		}
	}
	for kind, expected := range map[string]int{"literal": 88, "constructor": 8, "replaceCallback": 12, "replaceOther": 30, "lastIndex": 1, "patternCall": 19, "regexUse": 114} {
		if counts[kind] != expected {
			t.Errorf("%s: got %d, want %d", kind, counts[kind], expected)
		}
	}
}

// Every mutant builds and exits cleanly with sanitizer checks enabled. Only
// comparing its output to the untouched source on Node rejects it.
func TestScout23NodeOnlyMutants(t *testing.T) {
	t.Parallel()
	runtime, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/regexp_compile_parser.c"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range scout23Fixtures {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, scout23Directory, name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			source := native.C(program)
			changed := source
			mutation := "drop runtime ignoreCase flag"
			if name == "last_index" {
				mutation = "report lastIndex one code unit too far"
				lines := strings.Split(source, "\n")
				count := 0
				for i, line := range lines {
					if strings.Contains(line, `"lastIndex"`) && strings.Contains(line, "->number;") {
						lines[i] = strings.Replace(line, "->number;", "->number + 1;", 1)
						count++
					}
				}
				if count != 2 {
					t.Fatalf("expected two lastIndex reads, found %d", count)
				}
				changed = strings.Join(lines, "\n")
			} else {
				old := `case 'i': flag = REGEX_FLAG_I; break;`
				if bytes.Count(runtime, []byte(old)) != 1 {
					t.Fatal("flag canonicalization mutation site moved")
				}
				mutant := strings.Replace(string(runtime), old, `case 'i': flag = 0; break;`, 1)
				changed = strings.Replace(source, `#include "regexp_compile_parser.c"`, mutant, 1)
				if changed == source {
					t.Fatal("runtime compiler include missing")
				}
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(changed, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("mutant failed outside Node comparison: %+v", actual)
			}
			if difference := disagreement(onNode(t, path), actual); difference != "stdout differs" {
				t.Fatalf("mutant survived Node or failed another check: %s", difference)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				wasi := onWASI(t, changed)
				if wasi.exitCode != 0 || len(wasi.stderr) != 0 {
					t.Fatalf("WASI mutant failed outside comparison: %+v", wasi)
				}
				if difference := disagreement(onNode(t, path), wasi); difference != "stdout differs" {
					t.Fatalf("WASI mutant not caught only by Node: %s", difference)
				}
			}
			t.Logf("%s caught only by Node stdout comparison", mutation)
		})
	}
}

// Compiler callback support runs the untouched scout witnesses against Node.
func TestScout23CallbackWitnesses(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"callback_escape", "callback_wildcard"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, scout23Directory, name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			expected := onNode(t, path)
			if expected.exitCode != 0 || len(expected.stderr) != 0 || len(expected.stdout) == 0 {
				t.Fatalf("invalid Node witness: %+v", expected)
			}
			scout23CompileMatchesNode(t, path, expected)
		})
	}
}

func TestScout23AssignmentWitness(t *testing.T) {
	t.Parallel()
	original, err := filepath.Abs(filepath.Join(repository, scout23Directory, "module_specifiers_assignment.a"))
	if err != nil {
		t.Fatal(err)
	}
	adapted, err := filepath.Abs(filepath.Join(repository, scout23Directory, "module_specifiers.a"))
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, original)
	if expected.exitCode != 0 || len(expected.stderr) != 0 {
		t.Fatalf("Node excerpt failed: %+v", expected)
	}
	if difference := disagreement(expected, onNode(t, adapted)); difference != "" {
		t.Fatalf("statement split changed Node behavior: %s", difference)
	}
	scout23CompileMatchesNode(t, original, expected)
}

func scout23CompileMatchesNode(t *testing.T, path string, expected run) {
	t.Helper()
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(expected, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatal("JavaScript: " + difference)
	}
	actual, binary := natively(t, program)
	if difference := disagreement(expected, actual); difference != "" {
		t.Fatal("native: " + difference)
	}
	if failure := leaks(t, program, binary); failure != "" {
		t.Fatal(failure)
	}
}

// Counts are measured only for this family. Unrelated rows retain their Linux
// observations and are written in the full registry's order.
func TestScout23Counts(t *testing.T) {
	t.Parallel()
	selected := map[string]string{}
	for _, name := range scout23Fixtures {
		path := scout23Directory + name + ".a"
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
