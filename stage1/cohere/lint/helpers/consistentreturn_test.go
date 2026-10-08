package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func consistentreturnBuild(t *testing.T, entry string) (string, string) {
	t.Helper()
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "consistentreturn")
	module := filepath.Join(dir, "consistentreturn.mjs")
	if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(module, []byte(javascript.JavaScript(ir)), 0644); err != nil {
		t.Fatal(err)
	}
	return binary, module
}
func consistentreturnCases(t *testing.T) (string, []byte) {
	t.Helper()
	// Requery the current pinned Go bodies for every real consumer invocation.
	output := t.TempDir()
	run(t, "", "python3", "consistentreturn/testdata/capture.py", output)
	path := filepath.Join(output, "cases.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct{ Kind, Path, Want string }
	if err = json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	var want strings.Builder
	counts := map[string]int{}
	for _, row := range rows {
		want.WriteString(row.Want + "\n")
		counts[row.Kind]++
	}
	if len(counts) != 9 {
		t.Fatal("incomplete helpers", counts)
	}
	t.Logf("Go agreement inputs: %d calls, %v", len(rows), counts)
	return path, []byte(want.String())
}
func TestConsistentReturnPackageAgreementAndMutants(t *testing.T) {
	cases, want := consistentreturnCases(t)
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("consistentreturn/main.a")
	binary, module := consistentreturnBuild(t, entry)
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), want)
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, module, cases), want)
	compare(t, run(t, "", binary, cases), want)
	t.Logf("%d Go calls match Node, emitted JavaScript, ASan/UBSan native", bytes.Count(want, []byte{'\n'}))
	for _, m := range []struct{ file, from, to string }{
		{"is_scope.a", "case 'ArrowFunction':", "case 'NeverArrow':"},
		{"is_generator.a", "return node.generator", "return false"},
		{"has_value.a", "node.text==='undefined'", "node.text==='neverUndefined'"},
		{"is_exempt_from_end_judgment.a", "first>=128", "first>=999999"},
		{"capitalise_first.a", ".toUpperCase()", ".toLowerCase()"},
		{"verb.a", "messageId==='missingReturnValue'", "messageId==='missingReturn'"},
		{"static_name.a", "parent.name,Static", "node.name,Static"},
		{"name.a", "tokens.push('static')", "tokens.push('never-static')"},
		{"report_range.a", "{pos:0,end:0}", "{pos:1,end:1}"},
	} {
		t.Run(m.file, func(t *testing.T) {
			dir := t.TempDir()
			files, err := filepath.Glob("consistentreturn/*.a")
			if err != nil {
				t.Fatal(err)
			}
			reader, _ := filepath.Abs("options_json.ts")
			for _, file := range files {
				raw, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				text := string(raw)
				if filepath.Base(file) == m.file {
					if strings.Count(text, m.from) != 1 {
						t.Fatal("anchor", m.file)
					}
					text = strings.Replace(text, m.from, m.to, 1)
				}
				text = strings.ReplaceAll(text, "../options_json.ts", filepath.ToSlash(reader))
				for _, dep := range []string{"property_name.a", "slot02_ast.a"} {
					absolute, _ := filepath.Abs(dep)
					text = strings.ReplaceAll(text, "../"+dep, filepath.ToSlash(absolute))
				}
				if err = os.WriteFile(filepath.Join(dir, filepath.Base(file)), []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
			}
			entry := filepath.Join(dir, "main.a")
			binary, module := consistentreturnBuild(t, entry)
			for name, got := range map[string][]byte{
				"Node":               run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases),
				"emitted JavaScript": run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, module, cases),
				"ASan/UBSan native":  run(t, "", binary, cases),
			} {
				if bytes.Equal(got, want) {
					t.Fatal("survived", name)
				}
				t.Logf("semantic mutant caught on %s", name)
			}
		})
	}
}
