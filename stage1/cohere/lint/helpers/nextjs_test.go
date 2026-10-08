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

func nextjsBuild(t *testing.T, entry string) (string, string) {
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
	binary := filepath.Join(dir, "nextjs")
	module := filepath.Join(dir, "nextjs.mjs")
	if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(module, []byte(javascript.JavaScript(ir)), 0644); err != nil {
		t.Fatal(err)
	}
	return binary, module
}
func nextjsCases(t *testing.T) (string, []byte) {
	t.Helper()
	// Requery the current pinned Go bodies for every real consumer invocation.
	output := t.TempDir()
	run(t, "", "python3", "nextjs/testdata/capture.py", output)
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
func TestNextjsPackageAgreementAndMutants(t *testing.T) {
	cases, want := nextjsCases(t)
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("nextjs/main.a")
	binary, module := nextjsBuild(t, entry)
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), want)
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, module, cases), want)
	compare(t, run(t, "", binary, cases), want)
	t.Logf("%d Go calls match Node, emitted JavaScript, ASan/UBSan native", bytes.Count(want, []byte{'\n'}))
	for _, m := range []struct{ file, from, to string }{
		{"last_separator.a", "byte===92", "byte===0"},
		{"split_path.a", "baseName=path.slice(offset+1)", "baseName=path.slice(offset)"},
		{"is_document_file.a", "startsWith('_document.')", "startsWith('_document')"},
		{"is_document_page.a", "page.startsWith('/_document')", "page.startsWith('/_never')"},
		{"is_in_application_directory.a", "path.includes('app/')", "path.includes('never/')"},
		{"split_segments.a", "if(i>start)", "if(i>=start)"},
		{"is_in_pages_directory.a", "segments[i+1]!=='api'", "segments[i+1]==='api'"},
		{"build_route_file_contracts.a", "'generateStaticParams'", "'generateStaticParameters'"},
		{"route_contract_exports.a", "segments[i]==='app'", "segments[i]==='myapp'"},
	} {
		t.Run(m.file, func(t *testing.T) {
			dir := t.TempDir()
			files, err := filepath.Glob("nextjs/*.a")
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
				if err = os.WriteFile(filepath.Join(dir, filepath.Base(file)), []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
			}
			entry := filepath.Join(dir, "main.a")
			binary, module := nextjsBuild(t, entry)
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
