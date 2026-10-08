package react

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
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func run(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	c := exec.Command(name, args...)
	c.Dir = dir
	b, e := c.CombinedOutput()
	if e != nil {
		t.Fatalf("%s: %v\n%s", name, e, b)
	}
	return b
}
func write(t *testing.T, path string, b []byte) {
	t.Helper()
	if e := os.WriteFile(path, b, 0644); e != nil {
		t.Fatal(e)
	}
}
func read(t *testing.T, path string) []byte {
	t.Helper()
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func TestReactAgreement(t *testing.T) {
	root, _ := filepath.Abs("../../../../..")
	cohere := filepath.Join(root, "cohere")
	dir, _ := filepath.Abs(".")
	scratch := t.TempDir()
	if pin := strings.TrimSpace(string(run(t, cohere, "git", "rev-parse", "HEAD"))); pin != "7945d102a6c18dd36adf9114a758ce646e8b2359" {
		t.Fatal("pin drift", pin)
	}
	// Capture actual asserted cases through Go cohere's own testing hook.
	harness := filepath.Join(cohere, "internal/lint/testing/rule_testing.go")
	original := "return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}"
	s := string(read(t, harness))
	if strings.Count(s, original) != 1 {
		t.Fatal("capture anchor drift")
	}
	replacement := "result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result"
	modified := filepath.Join(scratch, "rule_testing.go")
	write(t, modified, []byte(strings.Replace(s, original, replacement, 1)))
	overlay := filepath.Join(scratch, "capture-overlay.json")
	b, _ := json.Marshal(map[string]any{"Replace": map[string]string{harness: modified}})
	write(t, overlay, b)
	capture := filepath.Join(scratch, "capture")
	var ledger struct {
		Remaining []struct {
			Rule    string
			Helpers []string `json:"remaining_helpers"`
		}
	}
	if e := json.Unmarshal(read(t, filepath.Join(dir, "../readiness.json")), &ledger); e != nil {
		t.Fatal(e)
	}
	consumers := map[string]bool{}
	for _, r := range ledger.Remaining {
		for _, h := range r.Helpers {
			if strings.Contains(h, "/ecmascript/react.") {
				consumers[r.Rule] = true
			}
		}
	}
	var inventory struct {
		Rules []struct {
			Name  string
			Tests struct{ Files []string }
		}
	}
	if e := json.Unmarshal(read(t, filepath.Join(root, "stage1/cohere/lint/inventory/inventory.json")), &inventory); e != nil {
		t.Fatal(e)
	}
	packages := map[string]bool{}
	for _, r := range inventory.Rules {
		if consumers[r.Name] {
			for _, f := range r.Tests.Files {
				packages[filepath.Dir(filepath.Join(root, f))] = true
			}
		}
	}
	for p := range packages {
		c := exec.Command("go", "test", "-overlay="+overlay, ".", "-count=1", "-timeout=10m")
		c.Dir = p
		c.Env = append(os.Environ(), "COHERE_DOCS_CAPTURE="+capture)
		if b, e := c.CombinedOutput(); e != nil {
			t.Fatalf("capture %s: %v\n%s", p, e, b)
		}
	}
	type record struct{ Rule, File, Source string }
	records := []record{}
	observed := map[string]int{}
	files, _ := filepath.Glob(filepath.Join(capture, "*.jsonl"))
	for _, f := range files {
		for _, line := range bytes.Split(read(t, f), []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var r record
			if e := json.Unmarshal(line, &r); e != nil {
				t.Fatal(e)
			}
			if consumers[r.Rule] {
				records = append(records, r)
				observed[r.Rule]++
			}
		}
	}
	for r := range consumers {
		if observed[r] == 0 {
			t.Fatal("missing consumer capture", r)
		}
	}
	sources := filepath.Join(scratch, "sources.json")
	b, _ = json.Marshal(records)
	write(t, sources, b)
	virtual := filepath.Join(cohere, "adamic_react_oracle.go")
	b, _ = json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(dir, "testdata/oracle.go"), filepath.Join(cohere, "internal/lint/ecmascript/react/adamic_exports.go"): filepath.Join(dir, "testdata/exports.go")}})
	overlay = filepath.Join(scratch, "oracle-overlay.json")
	write(t, overlay, b)
	oracle := filepath.Join(scratch, "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlay, "-o", oracle, virtual)
	cases := filepath.Join(scratch, "cases.json")
	wantPath := filepath.Join(scratch, "want.txt")
	t.Log(string(run(t, "", oracle, sources, cases, wantPath)))
	want := read(t, wantPath)
	check := func(entry string, mutant bool, label string) {
		t.Helper()
		node := run(t, "", "node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), entry, cases)
		p, e := load.Load([]string{entry})
		if e != nil {
			t.Fatal(e)
		}
		ir, e := lower.Lower(context.Background(), p)
		if e != nil {
			t.Fatal(e)
		}
		binary := filepath.Join(t.TempDir(), "react")
		if e = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); e != nil {
			t.Fatal(e)
		}
		compiled := run(t, "", binary, cases)
		if mutant {
			if bytes.Equal(node, want) || bytes.Equal(compiled, want) {
				t.Fatal("mutant survived", label)
			}
			t.Logf("%s mutant rejected on Node and sanitized native", label)
		} else {
			js := filepath.Join(t.TempDir(), "emitted.mjs")
			write(t, js, []byte(javascript.JavaScript(ir)))
			emitted := run(t, "", "node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), js, cases)
			for _, got := range [][]byte{node, compiled, emitted} {
				if !bytes.Equal(got, want) {
					a := strings.Split(string(got), "\n")
					b := strings.Split(string(want), "\n")
					for i := range a {
						if i < len(b) && a[i] != b[i] {
							t.Fatalf("line %d got %s want %s", i, a[i], b[i])
						}
					}
					t.Fatal("size mismatch")
				}
			}
			t.Logf("%d Go rows agree on Node, emitted JavaScript and sanitized native", bytes.Count(want, []byte("\n")))
		}
	}
	check(filepath.Join(dir, "main.a"), false, "baseline")
	mutations := [][3]string{{"react_es6_component_class.a", "if (isComponentBase(type.expression))", "if (false)"}, {"react_likely_component_name.a", "return (point - first) % stride === 0;", "return point >= 65 && point <= 90;"}, {"namespaced_member.a", "return name.kind === 'Identifier' && matches(name.text);", "return name.kind === 'Identifier' && !matches(name.text);"}, {"es5_component_call.a", "return isCreateClassName(expression.text);", "return false;"}, {"create_class_name.a", "name === 'createReactClass' || name === 'createClass'", "name === 'createClass'"}, {"component_base_name.a", "name === 'Component' || name === 'PureComponent'", "name === 'Component'"}, {"react_component_base.a", "return matchesBaseName(node.text);", "return false;"}, {"react_identifier_named.a", "node.text === name", "node.text !== name"}, {"react_is_hook_name.a", "return scalar >= range.lo", "return scalar < 128 && scalar >= range.lo"}, {"react_is_hook_call.a", "return isHookName(text(expression));", "return !isHookName(text(expression));"}}
	for _, m := range mutations {
		d := t.TempDir()
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".a") {
				continue
			}
			s := string(read(t, filepath.Join(dir, e.Name())))
			s = strings.ReplaceAll(s, "'../options_json.ts'", fmt.Sprintf("%q", filepath.Join(root, "stage1/cohere/lint/helpers/options_json.ts")))
			if e.Name() == m[0] {
				if strings.Count(s, m[1]) != 1 {
					t.Fatal("mutation anchor", m[0])
				}
				s = strings.Replace(s, m[1], m[2], 1)
			}
			write(t, filepath.Join(d, e.Name()), []byte(s))
		}
		check(filepath.Join(d, "main.a"), true, m[0])
	}
}
