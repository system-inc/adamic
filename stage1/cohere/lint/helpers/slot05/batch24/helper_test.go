package batch24

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

func TestRewrite(t *testing.T) {
	verify(t, "rewrite", "rewrite.a",
		mutation{"if(name.yes) {if(flag(current,2))", "if(name.yes) {if(!flag(current,2))"},
		mutation{"case 3:if(flag(current,2))", "case 3:if(!flag(current,2))"},
		mutation{"escape.set === 2 ? 1 : 0", "escape.set === 2 ? 0 : 1"},
		mutation{"const suffix=source.slice(end*2);", "const suffix=source.slice((end+1)*2);"},
		mutation{"read.negated ? 1 : 0,current,0", "read.negated ? 0 : 1,current,0"},
		mutation{"current=modifier.options;", "current=options;"},
		mutation{"dependency('case','',rune,flag(current,1) ? 1 : 0,0)", "dependency('case','',rune,flag(current,1) ? 0 : 1,0)"},
		mutation{"return failed(escape.error);", "return {text:out,exact,error:escape.error};"},
		mutation{"case 46:out+=encoded(flag(current,8)", "case 46:out+=encoded(!flag(current,8)"},
		mutation{"current=group.options;", "current=options;"},
		mutation{"escape.set === 3 && flag(current,2)", "escape.set === 3 && !flag(current,2)"},
		mutation{"group.kind,options,0", "group.kind,current,0"},
		mutation{"if(!atoms.yes){exact=false;}", "if(!atoms.yes){exact=true;}"},
		mutation{"flag(current,4) ? '(?:\\\\A|", "!flag(current,4) ? '(?:\\\\A|"})
}
func TestIsHexDigit(t *testing.T) {
	verify(t, "hex", "is_hex_digit.a", mutation{"byte <= 57", "byte < 57"}, mutation{"byte >= 65", "byte > 65"}, mutation{"byte >= 97", "byte > 97"}, mutation{"byte <= 102", "byte < 102"})
}
func TestAllHexDigits(t *testing.T) {
	verify(t, "all", "all_hex_digits.a", mutation{"bytes.length === 0) { return false", "bytes.length === 0) { return true"}, mutation{"if (!isHexDigit(byte))", "if (isHexDigit(byte))"}, mutation{"for (const byte of bytes)", "for (const byte of bytes.slice(0,-1))"})
}

// Not parallel: native compiler builds and corpora share the memory budget.
func verify(t *testing.T, mode, target string, mutations ...mutation) {
	root, _ := filepath.Abs("../../../../../..")
	cohere := filepath.Join(root, "cohere")
	dir, _ := filepath.Abs(".")
	scratch := t.TempDir()
	virtual := filepath.Join(cohere, "adamic_slot05_batch24.go")
	if got := strings.TrimSpace(string(run(t, cohere, "git", "rev-parse", "HEAD"))); got != "7945d102a6c18dd36adf9114a758ce646e8b2359" {
		t.Fatal("Go pin drift")
	}
	replacements := map[string]string{virtual: filepath.Join(dir, "testdata/oracle.go"), filepath.Join(cohere, "internal/lint/ecmascript/regexp/adamic_slot05_batch24.go"): filepath.Join(dir, "testdata/regexp_export.go")}
	original := filepath.Join(cohere, "internal/lint/ecmascript/regexp/rewrite.go")
	raw, e := os.ReadFile(original)
	if e != nil {
		t.Fatal(e)
	}
	text := string(raw)
	start := strings.Index(text, "func rewrite(")
	end := strings.Index(text[start:], "\n}\n") + start + 3
	body := text[start:end]
	for _, pair := range [][2]string{
		{"countGroups(", "adamicCount("}, {"utf8.DecodeRuneInString(", "adamicRune("}, {"namedBackreference(", "adamicBack("}, {"namedGroupOpener(", "adamicOpener("}, {"decodeEscape(", "adamicEscape("}, {"CaseClass(", "adamicCase("}, {"literalRune(", "adamicLiteral("}, {"quantifierWidth(", "adamicQuantifier("}, {"errNothingToRepeat(", "adamicRepeat("}, {"wordBoundary(", "adamicBoundary("}, {"wordClassAtoms(", "adamicWord("}, {"writeClass(", "adamicWrite("}, {"readClass(", "adamicClass("}, {"classAtoms(", "adamicAtoms("}, {"modifierGroup(", "adamicModifier("}, {"checkGroupConstruct(", "adamicConstruct("}, {"groupKindOf(", "adamicGroup("}, {"checkGroupQuantifier(", "adamicGroupQuantifier("},
	} {
		if !strings.Contains(body, pair[0]) {
			t.Fatal("overlay call drift", pair[0])
		}
		body = strings.ReplaceAll(body, pair[0], pair[1])
	}
	path := filepath.Join(scratch, "rewrite.go")
	if e = os.WriteFile(path, []byte(strings.Replace(text[:start]+body+text[end:], "\"unicode/utf8\"", "", 1)), 0644); e != nil {
		t.Fatal(e)
	}
	replacements[original] = path

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
	if evidence := os.Getenv("ADAMIC_SLOT05_BATCH24_EVIDENCE"); evidence != "" {
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
