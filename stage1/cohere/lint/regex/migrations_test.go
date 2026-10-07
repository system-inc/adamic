package regex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"
)

func TestMigrationMatchers(t *testing.T) {
	// Not parallel: builds independent scratch mutant modules for each migration.
	root, _ := filepath.Abs(".")
	data, err := os.ReadFile("testdata/migration_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var samples []struct {
		Rule    string `json:"rule"`
		Matched bool   `json:"matched"`
	}
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatal(err)
	}
	var expected bytes.Buffer
	counts := map[string]int{}
	for i, s := range samples {
		fmt.Fprintf(&expected, "%d:%t\n", i, s.Matched)
		counts[s.Rule]++
	}
	outputs := migrationBackends(t, root, "migration_runner.a")
	for _, output := range outputs {
		if !bytes.Equal(output, expected.Bytes()) {
			t.Fatal("migration matcher traces differ from cohere esregexp fixtures")
		}
	}
	actual := outputs[0]
	t.Logf("%v actual fixture/control matcher calls, %d identical trace bytes on cohere esregexp and source Node", counts, len(actual))
	for _, m := range []struct{ Rule, File, From, To string }{
		{"id-length", "id_length.a", "if(pattern.test(name)) return true;", "if(pattern.test(name)) return false;"},
		{"no-inline-comments", "no_inline_comments.a", "this.pattern !== undefined && this.pattern.test(body)", "this.pattern !== undefined && !this.pattern.test(body)"},
		{"no-warning-comments", "no_warning_comments.a", "new RegExp(source, 'iu')", "new RegExp(source, 'u')"},
	} {
		t.Run(m.Rule, func(t *testing.T) {
			temp := t.TempDir()
			if err := os.MkdirAll(filepath.Join(temp, "testdata"), 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"id_length.a", "no_inline_comments.a", "no_warning_comments.a", "patterns.a", "testdata/migration_runner.a", "testdata/migration_cases.a"} {
				content, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == m.File {
					if bytes.Count(content, []byte(m.From)) != 1 {
						t.Fatal("mutant anchor drift")
					}
					content = bytes.Replace(content, []byte(m.From), []byte(m.To), 1)
				}
				if err := os.WriteFile(filepath.Join(temp, file), content, 0644); err != nil {
					t.Fatal(err)
				}
			}
			for _, changed := range migrationBackends(t, temp, "migration_runner.a") {
				if bytes.Equal(changed, expected.Bytes()) {
					t.Fatal("mutant escaped matcher comparison")
				}
			}

			t.Log("mutant ran cleanly, then cohere comparison caught", m.Rule)
		})
	}

}

func migrationWritten(text string) string {
	var result strings.Builder
	for _, unit := range utf16.Encode([]rune(text)) {
		if unit >= 32 && unit <= 126 && unit != '\\' {
			result.WriteRune(rune(unit))
		} else {
			fmt.Fprintf(&result, `\u%04x`, unit)
		}
	}
	return result.String()
}
func TestWarningMigrationFindings(t *testing.T) {
	root, _ := filepath.Abs(".")
	data, err := os.ReadFile("testdata/warning_fixtures.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct {
		Source   string `json:"source"`
		Findings []struct {
			Start, End int
			ID         string `json:"id"`
			Message    string `json:"message"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	var expected bytes.Buffer
	findings := 0
	for index, fixture := range fixtures {
		fmt.Fprintf(&expected, "case %d\n", index)
		for _, f := range fixture.Findings {
			fmt.Fprintf(&expected, "%d:%d:%s:%s\n", f.Start, f.End, f.ID, migrationWritten(f.Message))
			findings++
		}
		fmt.Fprintf(&expected, "fixed %s\n", migrationWritten(fixture.Source))
	}
	outputs := migrationBackends(t, root, "warning_runner.a")
	for _, output := range outputs {
		if !bytes.Equal(output, expected.Bytes()) {
			t.Fatal("complete warning findings differ on backend")
		}
	}
	actual := outputs[0]
	if !bytes.Equal(actual, expected.Bytes()) {
		left, right := strings.Split(expected.String(), "\n"), strings.Split(string(actual), "\n")
		for i, line := range left {
			if i >= len(right) || line != right[i] {
				t.Fatalf("production warning finding comparison line %d: Go %q Node %q", i, line, right[min(i, len(right)-1)])
			}
		}
		t.Fatal("finding output length differs")
	}
	t.Logf("%d full warning fixtures, %d findings, %d identical bytes: IDs, messages, UTF-16 spans and whole fixed sources", len(fixtures), findings, len(actual))
}

func TestIdMigrationFindings(t *testing.T) {
	root, _ := filepath.Abs(".")
	data, err := os.ReadFile("testdata/id_fixtures.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct {
		Source   string `json:"source"`
		Findings []struct {
			Start, End int
			ID         string `json:"id"`
			Message    string `json:"message"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	var expected bytes.Buffer
	findings := 0
	for index, fixture := range fixtures {
		fmt.Fprintf(&expected, "case %d\n", index)
		for _, f := range fixture.Findings {
			fmt.Fprintf(&expected, "%d:%d:%s:%s\n", f.Start, f.End, f.ID, migrationWritten(f.Message))
			findings++
		}
		fmt.Fprintf(&expected, "fixed %s\n", migrationWritten(fixture.Source))
	}
	outputs := migrationBackends(t, root, "id_runner.a")
	for _, output := range outputs {
		if !bytes.Equal(output, expected.Bytes()) {
			t.Fatal("complete id-length findings differ on backend")
		}
	}
	actual := outputs[0]
	if !bytes.Equal(actual, expected.Bytes()) {
		left, right := strings.Split(expected.String(), "\n"), strings.Split(string(actual), "\n")
		for i, line := range left {
			if i >= len(right) || line != right[i] {
				t.Fatalf("id-length finding comparison line %d: Go %q Node %q", i, line, right[min(i, len(right)-1)])
			}
		}
		t.Fatal("finding output length differs")
	}
	t.Logf("%d full id-length fixtures, %d findings, %d identical bytes: IDs, messages, UTF-16 spans and whole fixed sources", len(fixtures), findings, len(actual))
}

func TestInlineMigrationFindings(t *testing.T) {
	root, _ := filepath.Abs(".")
	data, err := os.ReadFile("testdata/inline_fixtures.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct {
		Source   string `json:"source"`
		Findings []struct {
			Start, End int
			ID         string `json:"id"`
			Message    string `json:"message"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	var expected bytes.Buffer
	findings := 0
	for index, fixture := range fixtures {
		fmt.Fprintf(&expected, "case %d\n", index)
		for _, f := range fixture.Findings {
			fmt.Fprintf(&expected, "%d:%d:%s:%s\n", f.Start, f.End, f.ID, migrationWritten(f.Message))
			findings++
		}
		fmt.Fprintf(&expected, "fixed %s\n", migrationWritten(fixture.Source))
	}
	outputs := migrationBackends(t, root, "inline_runner.a")
	for _, output := range outputs {
		if !bytes.Equal(output, expected.Bytes()) {
			t.Fatal("complete projected no-inline-comments findings differ on backend")
		}
	}
	actual := outputs[0]
	if !bytes.Equal(actual, expected.Bytes()) {
		left, right := strings.Split(expected.String(), "\n"), strings.Split(string(actual), "\n")
		for i, line := range left {
			if i >= len(right) || line != right[i] {
				t.Fatalf("projected no-inline-comments finding comparison line %d: Go %q Node %q", i, line, right[min(i, len(right)-1)])
			}
		}
		t.Fatal("finding output length differs")
	}
	t.Logf("%d full projected no-inline-comments fixtures, %d findings, %d identical bytes: IDs, messages, UTF-16 spans and whole fixed sources", len(fixtures), findings, len(actual))
}

func migrationBackends(t *testing.T, root, runner string) [][]byte {
	t.Helper()
	repo, _ := filepath.Abs("../../../..")
	entry := filepath.Join(root, "testdata", runner)
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	outputs := [][]byte{run(t, root, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repo, "oracle/node.mjs"), entry)}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		if !strings.Contains(err.Error(), "RegExp with a nonconstant pattern") {
			t.Fatalf("unexpected migration compiler blocker: %v", err)
		}
		t.Logf("PENDING emitted JavaScript and sanitized native on codex/regex-runtime-compiler: %v", err)
		return outputs
	}
	// Automatically enforce both legs once the shared compiler dependency lands.
	temp := t.TempDir()
	emitted := filepath.Join(temp, "runner.js")
	binary := filepath.Join(temp, "runner")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	outputs = append(outputs, run(t, root, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repo, "oracle/node.mjs"), emitted), run(t, root, binary))
	t.Log("emitted JavaScript and sanitized native migration legs executed")
	return outputs
}
