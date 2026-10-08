package oracle

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

const cohereIntervalFixture = "internal/oracle/testdata/regexp_cohere_interval_bounds.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{cohereIntervalFixture, true, false})
}

func TestRegExpCohereIntervalMutants(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, cohereIntervalFixture))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	expected := onNode(t, path)
	parser, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/regexp_compile_parser.c"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"literal minimum zero", "runtime minimum zero"} {
		t.Run(name, func(t *testing.T) {
			changed := source
			if name == "literal minimum zero" {
				if !strings.Contains(source, "UINT64_C(2147483647)") {
					t.Fatal("literal bytecode mutation site missing")
				}
				changed = strings.ReplaceAll(source, "UINT64_C(2147483647)", "UINT64_C(0)")
			} else {
				old := `node->minimum = "2147483647";`
				if bytes.Count(parser, []byte(old)) != 1 {
					t.Fatal("runtime clamp mutation site moved")
				}
				mutated := strings.Replace(string(parser), old, `node->minimum = "0";`, 1)
				changed = strings.Replace(source, `#include "regexp_compile_parser.c"`, mutated, 1)
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
			if bytes.Equal(actual.stdout, expected.stdout) {
				t.Fatal("minimum-zero mutant survived Node comparison")
			}
			t.Logf("caught %s only by Node stdout comparison", name)
		})
	}
}

// Counts are measured only for this family. Unrelated rows retain their Linux
// observations and are written in the full registry's order.
func TestRegExpCohereIntervalCounts(t *testing.T) {
	t.Parallel()
	selected := map[string]string{}
	for _, path := range []string{cohereIntervalFixture, "internal/oracle/testdata/regexp_native_quantifier_bounds.a"} {
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
