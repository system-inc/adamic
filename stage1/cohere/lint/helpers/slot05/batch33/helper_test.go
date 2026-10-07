package batch33

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type mutation struct{ old, replacement string }

func TestComponent(t *testing.T) {
	verify(t, "component", "is_likely_react_component.a",
		mutation{"statement.returnStatement && looksLikeJsx(statement.expression)", "looksLikeJsx(statement.expression)"},
		mutation{"if (index < 0) { return false; }", "if (index < 0) { return true; }"},
		mutation{"if (!node.functionLike) { return false; }", "if (!node.functionLike) { return true; }"},
		mutation{"name.text === 'properties'", "name.text === '__properties__'"},
		mutation{"name.text === 'props'", "name.text === '__props__'"},
		mutation{"node.parameters[0]", "node.parameters[1]"},
		mutation{"if (node.body < 0) { return false; }", "if (node.body < 0) { return true; }"},
		mutation{"statement.returnStatement && looksLikeJsx(statement.expression)", "false && looksLikeJsx(statement.expression)"},
		mutation{"return looksLikeJsx(node.body);", "return false;"},
		mutation{"return false;\n    }\n    return looksLikeJsx(node.body);", "return true;\n    }\n    return looksLikeJsx(node.body);"})
}
func TestType(t *testing.T) {
	verify(t, "type", "type_reference_name.a",
		mutation{"if (index < 0) { return ''; }", "if (index < 0) { return 'wrong'; }"},
		mutation{"if (!node.typeReference) { return ''; }", "if (node.typeReference) { return ''; }"},
		mutation{"if (!name.identifier) { return ''; }", "if (name.identifier) { return ''; }"},
		mutation{"return name.text;", "return 'wrong' + name.text;"})
}
func TestNetwork(t *testing.T) {
	verify(t, "network", "is_network_service_hook_call.a",
		mutation{"while ((nodes[callee] ?? panic('missing callee')).parenthesized)", "while (false)"},
		mutation{"while ((nodes[receiver] ?? panic('missing receiver')).parenthesized)", "while (false)"},
		mutation{"if (!access.property) { return false; }", "if (!access.property) { return true; }"},
		mutation{"object.text !== 'networkService'", "object.text !== 'other'"},
		mutation{"return methods.get(method.text) ?? false;", "return methods.has(method.text);"},
		mutation{"return methods.get(method.text) ?? false;", "return true;"},
		mutation{"return methods.get(method.text) ?? false;", "return false;"})
}

// Not parallel: native compiler builds and corpora share the memory budget.
func verify(t *testing.T, mode, target string, mutations ...mutation) {
	root, _ := filepath.Abs("../../../../../..")
	cohere := filepath.Join(root, "cohere")
	dir, _ := filepath.Abs(".")
	scratch := t.TempDir()
	virtual := filepath.Join(cohere, "adamic_slot05_batch33.go")
	if got := strings.TrimSpace(string(run(t, cohere, "git", "rev-parse", "HEAD"))); got != "7945d102a6c18dd36adf9114a758ce646e8b2359" {
		t.Fatal("Go pin drift")
	}
	replacements := map[string]string{virtual: filepath.Join(dir, "testdata/oracle.go"), filepath.Join(cohere, "internal/lint/rules/structure/adamic_slot05_batch33.go"): filepath.Join(dir, "testdata/structure_export.go")}

	data, _ := json.Marshal(map[string]any{"Replace": replacements})
	overlay := filepath.Join(scratch, "overlay.json")
	if e := os.WriteFile(overlay, data, 0644); e != nil {
		t.Fatal(e)
	}
	oracle := filepath.Join(scratch, "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlay, "-o", oracle, virtual)
	t.Log(strings.TrimSpace(string(run(t, "", oracle, root, scratch, mode))))
	cases := filepath.Join(scratch, "cases.json")
	want, e := os.ReadFile(filepath.Join(scratch, "want.txt"))
	if e != nil {
		t.Fatal(e)
	}
	if evidence := os.Getenv("ADAMIC_SLOT05_BATCH33_EVIDENCE"); evidence != "" {
		for _, name := range []string{"coverage.json", "cases.json", "want.txt"} {
			data, e := os.ReadFile(filepath.Join(scratch, name))
			if e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(filepath.Join(evidence, mode+"-"+name), data, 0644); e != nil {
				t.Fatal(e)
			}
		}
	}
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), filepath.Join(dir, "main.a"), cases), want)
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), slot05JavaScript(t, dir), cases), want)
	compare(t, run(t, "", slot05Build(t, dir), cases), want)
	for _, v := range mutations {
		mutant := t.TempDir()
		entries, e := os.ReadDir(dir)
		if e != nil {
			t.Fatal(e)
		}
		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasSuffix(name, ".a") {
				continue
			}
			data, e := os.ReadFile(filepath.Join(dir, name))
			if e != nil {
				t.Fatal(e)
			}
			if name == target {
				if strings.Count(string(data), v.old) != 1 {
					t.Fatal("mutant anchor drift", v.old)
				}
				data = []byte(strings.Replace(string(data), v.old, v.replacement, 1))
			}
			data = []byte(strings.ReplaceAll(string(data), "'../../options_json.ts'", strconvQuote(filepath.Join(root, "stage1/cohere/lint/helpers/options_json.ts"))))
			for _, dependency := range []string{"return_argument_looks_like_jsx.a", "model.a"} {
				data = []byte(strings.ReplaceAll(string(data), "'../batch32/"+dependency+"'", strconvQuote(filepath.Join(root, "stage1/cohere/lint/helpers/slot05/batch32", dependency))))
			}
			if e = os.WriteFile(filepath.Join(mutant, name), data, 0644); e != nil {
				t.Fatal(e)
			}
		}
		got := run(t, "", slot05Build(t, mutant), cases)
		if bytes.Equal(got, want) {
			t.Fatal("compiled semantic mutant survived", v.old)
		}
		a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
		caught := false
		for i := range a {
			if i < len(b) && a[i] != b[i] {
				t.Logf("%s mutant %q -> %q caught at output line %d: mutant %q, Go %q", target, v.old, v.replacement, i+1, a[i], b[i])
				caught = true
				break
			}
		}
		if !caught {
			t.Fatal("missing mutant witness")
		}
	}
}
func run(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	f, err := os.CreateTemp(t.TempDir(), "output-")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cmd.Stdout = f
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
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

func compare(t *testing.T, got, want []byte) {
	t.Helper()
	if bytes.Equal(got, want) {
		return
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			t.Fatalf("output line %d: got %q, Go %q", i+1, a[i], b[i])
		}
	}
	t.Fatalf("output size: got %d Go %d", len(got), len(want))
}

func slot05Build(t *testing.T, dir string) string {
	t.Helper()
	program, err := load.Load([]string{filepath.Join(dir, "main.a")})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "slot05")
	if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	return binary
}

func strconvQuote(path string) string { data, _ := json.Marshal(path); return string(data) }

func slot05JavaScript(t *testing.T, dir string) string {
	t.Helper()
	program, e := load.Load([]string{filepath.Join(dir, "main.a")})
	if e != nil {
		t.Fatal(e)
	}
	lowered, e := lower.Lower(context.Background(), program)
	if e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "emitted.mjs")
	if e = os.WriteFile(path, []byte(javascript.JavaScript(lowered)), 0644); e != nil {
		t.Fatal(e)
	}
	return path
}
