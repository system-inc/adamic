package reacthelpers

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func run(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	f, err := os.CreateTemp(t.TempDir(), "output-")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cmd.Stdout = f
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = childguard.Run(cmd, runGuard); err != nil {
		t.Fatalf("%s: %s %v: %v\n%s", t.Name(), name, args, err, &stderr)
	}
	if stderr.Len() != 0 {
		t.Fatalf("%s stderr: %s", name, &stderr)
	}
	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func oracle(t *testing.T) string {
	root, _ := filepath.Abs("../../../../../cohere")
	side, _ := filepath.Abs("testdata/oracle.go")
	virtual := filepath.Join(root, "adamic_rules_react_oracle.go")
	data, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	write(t, overlay, data)
	binary := filepath.Join(t.TempDir(), "oracle")
	run(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	return binary
}
func write(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}
func outputsEntry(t *testing.T, directory, path, entry string) [][]byte {
	return entryOutputs(t, directory, path, entry, true)
}

func entryOutputs(t *testing.T, directory, path, entry string, includeNative bool) [][]byte {
	program, err := load.Load([]string{filepath.Join(directory, entry)})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := ""
	if includeNative {
		binary = buildNative(t, ir)
	}
	script := filepath.Join(t.TempDir(), "emitted.mjs")
	write(t, script, []byte(javascript.JavaScript(ir)))
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	results := [][]byte{run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, entry), path)}
	if includeNative {
		results = append(results, run(t, "", binary, path))
	}
	return append(results, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, script, path))
}
func TestGoAgreement(t *testing.T) {
	binary := oracle(t)
	directory, _ := filepath.Abs(".")
	for _, input := range []string{"calls.json", "controls.json"} {
		t.Run(input, func(t *testing.T) {
			path, _ := filepath.Abs("testdata/" + input)
			want := run(t, "", binary, path)
			for i, out := range outputs(t, directory, path) {
				if !bytes.Equal(out, want) {
					t.Fatalf("backend %d disagrees with Go", i)
				}
				t.Logf("backend %d: %d matching calls", i, bytes.Count(out, []byte("\n")))
			}
		})
	}
}
func TestMutants(t *testing.T) {
	binary := oracle(t)
	path, _ := filepath.Abs("testdata/controls.json")
	want := run(t, "", binary, path)
	mutants := []struct{ file, from, to string }{
		{"decode_compiler_rule_options.a", "if(keys.length===0){return {error:''};}", "if(keys.length===0){return {error:'empty object refused'};}"},
		{"decode_no_method_set_state_options.a", "disallowInFunc:true", "disallowInFunc:false"},
		{"is_hook_identifier_name.a", "if(name.length === 3) { return true; }", "if(name.length === 3) { return false; }"},
		{"is_component_identifier_name.a", "first <= 90", "first < 90"},
		{"mentions_ref.a", "name.includes('ref') || name.includes('Ref')", "name.includes('ref')"},
		{"is_java_script_identifier.a", "i > 0 && c >= 48", "i >= 0 && c >= 48"},
		{"comment_value_of.a", "text.slice(0, text.length - 2)", "text"},
		{"jsx_annotation_in.a", "begin > start && begin < text.length", "begin >= start && begin < text.length"},
		{"is_react_component_base_name.a", "name === 'Component' || name === 'PureComponent'", "name === 'Component'"},
	}
	for _, m := range mutants {
		t.Run(m.file, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			dir := filepath.Join(root, "lint/helpers/rules-react")
			if e := os.MkdirAll(dir, 0755); e != nil {
				t.Fatal(e)
			}
			entries, e := os.ReadDir(".")
			if e != nil {
				t.Fatal(e)
			}
			for _, entry := range entries {
				if !strings.HasSuffix(entry.Name(), ".a") {
					continue
				}
				data, e := os.ReadFile(entry.Name())
				if e != nil {
					t.Fatal(e)
				}
				if entry.Name() == m.file {
					if strings.Count(string(data), m.from) != 1 {
						t.Fatal("anchor")
					}
					data = []byte(strings.Replace(string(data), m.from, m.to, 1))
				}
				write(t, filepath.Join(dir, entry.Name()), data)
			}
			data, e := os.ReadFile("../options_json.ts")
			if e != nil {
				t.Fatal(e)
			}
			write(t, filepath.Join(root, "lint/helpers/options_json.ts"), data)
			unicode, e := os.ReadFile("../../unicode.ts")
			if e != nil {
				t.Fatal(e)
			}
			write(t, filepath.Join(root, "lint/unicode.ts"), unicode)
			checkMutant(t, dir, path, "main.a", false, want)
		})
	}
}

func outputs(t *testing.T, directory, path string) [][]byte {
	return outputsEntry(t, directory, path, "main.a")
}
func TestAstGoAgreement(t *testing.T) {
	data, err := os.ReadFile("testdata/ast.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	expanded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	reader.Close()
	var captured struct{ Calls []struct{ Name, Want string } }
	if err = json.Unmarshal(expanded, &captured); err != nil {
		t.Fatal(err)
	}
	want := []byte{}
	counts := map[string]int{}
	for _, call := range captured.Calls {
		want = append(want, []byte(call.Want+"\n")...)
		counts[call.Name]++
	}
	input := filepath.Join(t.TempDir(), "ast.json")
	write(t, input, expanded)
	directory, _ := filepath.Abs(".")
	for i, out := range outputsEntry(t, directory, input, "ast_main.a") {
		if !bytes.Equal(out, want) {
			a, b := strings.Split(string(out), "\n"), strings.Split(string(want), "\n")
			for j := 0; j < len(a) && j < len(b); j++ {
				if a[j] != b[j] {
					t.Fatalf("backend %d call %d: %s, Go %s", i, j, a[j], b[j])
				}
			}
			t.Fatalf("backend %d lengths %d/%d", i, len(out), len(want))
		}
		t.Logf("backend %d: %d Go calls, %v", i, len(captured.Calls), counts)
	}
}

func TestAstMutants(t *testing.T) {
	data, err := os.ReadFile("testdata/ast.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	expanded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	reader.Close()
	var captured struct{ Calls []struct{ Name, Want string } }
	if err = json.Unmarshal(expanded, &captured); err != nil {
		t.Fatal(err)
	}
	want := []byte{}
	for _, call := range captured.Calls {
		want = append(want, []byte(call.Want+"\n")...)
	}
	input := filepath.Join(t.TempDir(), "ast.json")
	write(t, input, expanded)
	var mutations []struct{ File, From, To string }
	data, err = os.ReadFile("testdata/ast-mutants.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &mutations); err != nil {
		t.Fatal(err)
	}
	hasCanary := false
	for _, m := range mutations {
		hasCanary = hasCanary || m.File == nativeCanaryMutant
	}
	if !hasCanary {
		t.Fatalf("native canary mutant %s is missing", nativeCanaryMutant)
	}
	for _, m := range mutations {
		t.Run(m.File, func(t *testing.T) {
			t.Parallel()
			directory := copyAstMutantTree(t)
			path := filepath.Join(directory, m.File)
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(source), m.From) != 1 {
				t.Fatal("anchor")
			}
			write(t, path, []byte(strings.Replace(string(source), m.From, m.To, 1)))
			checkMutant(t, directory, input, "ast_main.a", m.File == nativeCanaryMutant, want)
		})
	}
}

func TestCaptureCoverage(t *testing.T) {
	var metadata struct {
		Cohere      string
		ActualCalls map[string]int `json:"actual_calls"`
		Controls    int
		GoSources   map[string]string `json:"go_sources"`
	}
	data, err := os.ReadFile("testdata/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &metadata); err != nil {
		t.Fatal(err)
	}
	root, _ := filepath.Abs("../../../../../cohere")
	if strings.TrimSpace(string(run(t, root, "git", "rev-parse", "HEAD"))) != metadata.Cohere {
		t.Fatal("cohere pin drift")
	}
	for file, hash := range metadata.GoSources {
		data, err = os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != hash {
			t.Fatalf("Go source drift: %s", file)
		}
	}
	counts := map[string]int{}
	var leaf []struct{ Name string }
	data, err = os.ReadFile("testdata/calls.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &leaf); err != nil {
		t.Fatal(err)
	}
	for _, row := range leaf {
		counts[row.Name]++
	}
	data, err = os.ReadFile("testdata/ast.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	expanded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	reader.Close()
	var ast struct{ Calls []struct{ Name string } }
	if err = json.Unmarshal(expanded, &ast); err != nil {
		t.Fatal(err)
	}
	for _, row := range ast.Calls {
		counts[row.Name]++
	}
	if len(counts) != 35 || len(metadata.ActualCalls) != 35 {
		t.Fatal("helper omitted")
	}
	for name, count := range metadata.ActualCalls {
		if counts[name] != count || count == 0 {
			t.Fatalf("capture omitted: %s %d/%d", name, counts[name], count)
		}
	}
	var controls []struct{ Name string }
	data, err = os.ReadFile("testdata/controls.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &controls); err != nil {
		t.Fatal(err)
	}
	if len(controls) != metadata.Controls {
		t.Fatal("controls omitted")
	}
}
