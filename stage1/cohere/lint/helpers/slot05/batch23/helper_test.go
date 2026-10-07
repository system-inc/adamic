package batch23

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

func TestDecodeEscape(t *testing.T) {
	verify(t, "decode", "decode_escape.a", mutation{"case 98:if(context.inClass)", "case 98:if(!context.inClass)"}, mutation{"case 66:if(context.inClass)", "case 66:if(!context.inClass)"}, mutation{"if(context.named)", "if(!context.named)"}, mutation{"rune>=48 && rune<=57", "rune>=49 && rune<=57"}, mutation{"case 110:return escapeValue(10", "case 110:return escapeValue(13"}, mutation{"ops.hexEscape(index,size,2,context)", "ops.hexEscape(index,size,4,context)"})
}
func TestClassAtoms(t *testing.T) {
	verify(t, "atoms", "class_atoms.a", mutation{"inClass:true", "inClass:false"}, mutation{"rune === 45 ? 3 : 0", "rune === 45 ? 0 : 3"}, mutation{"exact=false", "exact=true"}, mutation{"ops.slice(index,end)", "ops.slice(index+size,end)"}, mutation{"ops.word(options)", "ops.nonword(options)"})
}
func TestRegExpTest(t *testing.T) {
	verify(t, "test", "regexp_test.a", mutation{"return false", "return true"}, mutation{"return false", "return regexp !== undefined"}, mutation{"!result.error && result.matched", "result.error || result.matched"}, mutation{"result.matched;", "!result.matched;"})
}

// Not parallel: native compiler builds and corpora share the memory budget.
func verify(t *testing.T, mode, target string, mutations ...mutation) {
	root, _ := filepath.Abs("../../../../../..")
	cohere := filepath.Join(root, "cohere")
	dir, _ := filepath.Abs(".")
	scratch := t.TempDir()
	virtual := filepath.Join(cohere, "adamic_slot05_batch23.go")
	if got := strings.TrimSpace(string(run(t, cohere, "git", "rev-parse", "HEAD"))); got != "7945d102a6c18dd36adf9114a758ce646e8b2359" {
		t.Fatal("Go pin drift")
	}
	replacements := map[string]string{virtual: filepath.Join(dir, "testdata/oracle.go"), filepath.Join(cohere, "internal/lint/ecmascript/regexp/adamic_slot05_batch23.go"): filepath.Join(dir, "testdata/regexp_export.go")}
	for _, item := range []struct {
		file, signature string
		calls           [][2]string
	}{
		{"escape.go", "func decodeEscape(", [][2]string{{"utf8.DecodeRuneInString(", "adamicRune("}, {"decodeUnicodeEscape(", "adamicUnicode("}, {"decodeFixedHex(", "adamicHex("}, {"decodeControlEscape(", "adamicControl("}, {"decodePropertyEscape(", "adamicProperty("}, {"identityEscape(", "adamicIdentity("}, {"decodeNumericEscape(", "adamicNumeric("}}},
		{"class.go", "func classAtoms(", [][2]string{{"utf8.DecodeRuneInString(", "adamicRune("}, {"decodeEscape(", "adamicEscape("}, {"wordClassAtoms(", "adamicWord("}, {"nonWordClassAtoms(", "adamicNonword("}, {"joinRanges(", "adamicJoin("}}},
		{"regexp.go", "func (r *RegExp) Test(", [][2]string{{"r.re.MatchString(s)", "adamicMatch(r.re,s)"}}},
	} {
		original := filepath.Join(cohere, "internal/lint/ecmascript/regexp", item.file)
		data, e := os.ReadFile(original)
		if e != nil {
			t.Fatal(e)
		}
		text := string(data)
		start := strings.Index(text, item.signature)
		if start < 0 {
			t.Fatal("overlay signature drift")
		}
		end := strings.Index(text[start:], "\n}\n") + start + 3
		body := text[start:end]
		for _, pair := range item.calls {
			if !strings.Contains(body, pair[0]) {
				t.Fatal("overlay call drift")
			}
			body = strings.ReplaceAll(body, pair[0], pair[1])
		}
		path := filepath.Join(scratch, item.file)
		if e = os.WriteFile(path, []byte(text[:start]+body+text[end:]), 0644); e != nil {
			t.Fatal(e)
		}
		replacements[original] = path
	}

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
	if evidence := os.Getenv("ADAMIC_SLOT05_BATCH23_EVIDENCE"); evidence != "" {
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
