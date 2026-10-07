package batch17

import (
	"bytes"
	"context"
	"crypto/sha256"
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
	"time"
)

type mutation struct{ old, replacement string }

func TestSlot05VisitUnknown(t *testing.T) {
	verify(t, "visit", "visit_unknown.a", mutation{"if(isNil(node))", "if(false)"}, mutation{"if(isStatement(node))", "if(false)"}, mutation{"expr(node);", "statement(node);"}, mutation{"statement(node);return;", "statement(node);expr(node);return;"})
}
func TestSlot05ForkOptionalChain(t *testing.T) {
	verify(t, "chain", "fork_optional_chain.a", mutation{"state.joins.length === 0 || !isOptionalChainRoot(node)", "!isOptionalChainRoot(node) || state.joins.length === 0"}, mutation{"state.joins.length === 0", "state.joins.length <= 1"}, mutation{"!isOptionalChainRoot(node)", "false"}, mutation{"state.joins.length - 1", "0"}, mutation{"link(state.cur,next);", "if(false) {link(state.cur,next);}"}, mutation{"enter(next);", "if(false) {enter(next);}"})
}
func TestSlot05IsHookCall(t *testing.T) {
	verify(t, "hook", "react_is_hook_call.a", mutation{"if(call === undefined || call.expression === undefined) {return false;}", "if(call === undefined || call.expression === undefined) {return true;}"}, mutation{"\n return false;\n}", "\n return true;\n}"}, mutation{"if(isIdentifier(expression))", "if(false)"}, mutation{"if(isPropertyAccess(expression))", "if(false)"}, mutation{"return isHookName(text(expression));", "return !isHookName(text(expression));"}, mutation{"return isNamespacedMember(expression,isHookName);", "return !isNamespacedMember(expression,isHookName);"})
}

// Not parallel: large generated corpora and multiple sanitized compiler builds share the memory budget.
func verify(t *testing.T, mode, target string, mutations ...mutation) {
	root, _ := filepath.Abs("../../../../../..")
	cohere := filepath.Join(root, "cohere")
	dir, _ := filepath.Abs(".")
	if got := strings.TrimSpace(string(run(t, cohere, "git", "rev-parse", "HEAD"))); got != "715ba94f3608a6500086b1076ce5cb7e51b836db" {
		t.Fatal("Go pin drift", got)
	}

	scratch := t.TempDir()
	virtual := filepath.Join(cohere, "adamic_slot05_batch17.go")
	replacements := map[string]string{virtual: filepath.Join(dir, "testdata", "oracle.go")}

	replacements[filepath.Join(cohere, "internal/lint/ecmascript/control_flow_graph/adamic_slot05_batch17.go")] = filepath.Join(dir, "testdata", "cfg_export.go")
	for file, signatures := range map[string][]string{"expressions.go": {"visitUnknown", "forkOptionalChain"}} {
		original := filepath.Join(cohere, "internal/lint/ecmascript/control_flow_graph", file)
		source, err := os.ReadFile(original)
		if err != nil {
			t.Fatal(err)
		}
		text := string(source)
		for _, signature := range signatures {
			start := strings.Index(text, "func (b *Builder[E]) "+signature+"(")
			if start < 0 {
				t.Fatal("signature drift")
			}
			end := strings.Index(text[start:], "\n}\n") + start + 3
			body := text[start:end]
			for _, pair := range [][2]string{{"ast.IsStatement(node)", "adamicIsStatement(node)"}, {"b.statement(node)", "b.adamicStatement(node)"}, {"b.expr(node)", "b.adamicExpr(node)"}, {"ast.IsOptionalChainRoot(node)", "adamicIsChainRoot(node)"}, {"b.link(b.cur,", "b.adamicLink(b.cur,"}, {"b.newBlock()", "b.adamicNewBlock()"}, {"b.enter(next)", "b.adamicEnter(next)"}} {
				body = strings.ReplaceAll(body, pair[0], pair[1])
			}
			text = text[:start] + body + text[end:]
		}
		modified := filepath.Join(scratch, file)
		if e := os.WriteFile(modified, []byte(text), 0644); e != nil {
			t.Fatal(e)
		}
		replacements[original] = modified
	}
	original := filepath.Join(cohere, "internal/lint/ecmascript/react/names.go")
	source, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	start := strings.Index(text, "func IsHookCall(")
	end := strings.Index(text[start:], "\n}\n") + start + 3
	body := text[start:end]
	body = strings.ReplaceAll(body, "call.Expression.Text()", "adamicText(call.Expression)")
	body = strings.ReplaceAll(body, "IsHookName(adamicText(call.Expression))", "adamicHookName(adamicText(call.Expression))")
	body = strings.ReplaceAll(body, "IsNamespacedMember(call.Expression, IsHookName)", "adamicNamespaced(call.Expression, IsHookName)")
	text = text[:start] + body + text[end:]
	modified := filepath.Join(scratch, "names.go")
	if e := os.WriteFile(modified, []byte(text), 0644); e != nil {
		t.Fatal(e)
	}
	replacements[original] = modified
	replacements[filepath.Join(cohere, "internal/lint/ecmascript/react/adamic_slot05_batch17.go")] = filepath.Join(dir, "testdata", "react_export.go")

	overlay, _ := json.Marshal(map[string]any{"Replace": replacements})
	overlayPath := filepath.Join(scratch, "overlay.json")
	if e := os.WriteFile(overlayPath, overlay, 0644); e != nil {
		t.Fatal(e)
	}
	oracle := filepath.Join(scratch, "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlayPath, "-o", oracle, virtual)
	t.Log(strings.TrimSpace(string(run(t, "", oracle, root, scratch, mode))))
	cases := filepath.Join(scratch, "cases.json")
	want, e := os.ReadFile(filepath.Join(scratch, "want.txt"))
	if e != nil {
		t.Fatal(e)
	}
	coverage, e := os.ReadFile(filepath.Join(scratch, "coverage.json"))
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("consumer coverage: %s", coverage)
	if evidence := os.Getenv("ADAMIC_SLOT05_BATCH17_EVIDENCE"); evidence != "" {
		if e = os.WriteFile(filepath.Join(evidence, mode+"-coverage.json"), coverage, 0644); e != nil {
			t.Fatal(e)
		}
		data, e := os.ReadFile(cases)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(evidence, mode+"-corpus.sha256"), []byte(fmt.Sprintf("%x  cases.json\n", sha256.Sum256(data))), 0644); e != nil {
			t.Fatal(e)
		}
	}
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), filepath.Join(dir, "main.a"), cases), want)
	emitted := slot05JavaScript(t, dir)
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), emitted, cases), want)
	binary := slot05Build(t, dir)
	compare(t, run(t, "", binary, cases), want)
	for _, variant := range mutations {
		old, replacement := variant.old, variant.replacement
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
				if strings.Count(string(data), old) != 1 {
					t.Fatal("mutant anchor drift")
				}
				data = []byte(strings.Replace(string(data), old, replacement, 1))
			}
			data = []byte(strings.ReplaceAll(string(data), "'../../options_json.ts'", strconvQuote(filepath.Join(root, "stage1/cohere/lint/helpers/options_json.ts"))))

			data = []byte(strings.ReplaceAll(string(data), "'../batch14/types.a'", strconvQuote(filepath.Join(root, "stage1/cohere/lint/helpers/slot05/batch14/types.a"))))

			if e = os.WriteFile(filepath.Join(mutant, name), data, 0644); e != nil {
				t.Fatal(e)
			}
		}
		got := run(t, "", slot05Build(t, mutant), cases)
		if bytes.Equal(got, want) {
			t.Fatal("compiled semantic mutant survived")
		}
		a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
		for i := range a {
			if i < len(b) && a[i] != b[i] {
				offset := 0
				for offset < len(a[i]) && offset < len(b[i]) && a[i][offset] == b[i][offset] {
					offset++
				}
				start := max(0, offset-20)
				t.Logf("%s mutant %q -> %q caught at output line %d byte %d: mutant %q, Go %q", target, old, replacement, i+1, offset, a[i][min(start, len(a[i])):min(start+120, len(a[i]))], b[i][min(start, len(b[i])):min(start+120, len(b[i]))])
				break
			}
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
