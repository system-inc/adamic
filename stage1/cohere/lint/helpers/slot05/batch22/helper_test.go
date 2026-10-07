package batch22

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

func TestProperty(t *testing.T) {
	verify(t, "property", "decode_property_escape.a", mutation{"if(!context.unicode)", "if(context.unicode)"}, mutation{"source[index+size] !== 123", "source[index+size] === 123"}, mutation{"if(end<0)", "if(false)"}, mutation{"size+end+1", "size+end"}, mutation{"indexBrace(index+size)", "indexBrace(index+size+1)"})
}
func TestNumeric(t *testing.T) {
	verify(t, "numeric", "decode_numeric_escape.a", mutation{"!isDecimalDigit(index+1)", "isDecimalDigit(index+1)"}, mutation{"!context.inClass", "context.inClass"}, mutation{"value.rune<=context.groups", "value.rune<context.groups"}, mutation{"if(context.unicode)", "if(false)"}, mutation{"if(octal.width>0)", "if(false)"}, mutation{"rune:byte,width:1", "rune:0,width:1"}, mutation{"decimalEscape(index)", "decimalEscape(index+1)"})
}
func TestReadClass(t *testing.T) {
	verify(t, "class", "read_class.a", mutation{"source[index] === 94", "source[index] === 95"}, mutation{"index+=1+decodeSize(index+1)", "index+=1"}, mutation{"width:index+1", "width:index"}, mutation{"body:slice(start,index)", "body:slice(1,index)"}, mutation{"negated:false,width:0", "negated,width:0"})
}

// Not parallel: native compiler builds and corpora share the memory budget.
func verify(t *testing.T, mode, target string, mutations ...mutation) {
	root, _ := filepath.Abs("../../../../../..")
	cohere := filepath.Join(root, "cohere")
	dir, _ := filepath.Abs(".")
	scratch := t.TempDir()
	virtual := filepath.Join(cohere, "adamic_slot05_batch22.go")
	if got := strings.TrimSpace(string(run(t, cohere, "git", "rev-parse", "HEAD"))); got != "715ba94f3608a6500086b1076ce5cb7e51b836db" {
		t.Fatal("Go pin drift")
	}
	replacements := map[string]string{virtual: filepath.Join(dir, "testdata/oracle.go"), filepath.Join(cohere, "internal/lint/ecmascript/regexp/adamic_slot05_batch22.go"): filepath.Join(dir, "testdata/regexp_export.go")}
	for _, item := range []struct {
		file, signature string
		calls           [][2]string
	}{
		{"escape.go", "func decodePropertyEscape(", [][2]string{{"identityEscape(", "adamicIdentity("}, {"strings.IndexByte(", "adamicIndexByte("}}},
		{"escape.go", "func decodeNumericEscape(", [][2]string{{"isDecimalDigit(", "adamicDigit("}, {"decimalEscape(", "adamicDecimal("}, {"decodeLegacyOctal(", "adamicOctal("}}},
		{"class.go", "func readClass(", [][2]string{{"utf8.DecodeRuneInString(", "adamicDecodeRune("}}},
	} {
		original := filepath.Join(cohere, "internal/lint/ecmascript/regexp", item.file)
		modified := filepath.Join(scratch, item.file)
		data, e := os.ReadFile(original)
		if existing, err := os.ReadFile(modified); err == nil {
			data = existing
		}
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
		if e = os.WriteFile(modified, []byte(text[:start]+body+text[end:]), 0644); e != nil {
			t.Fatal(e)
		}
		replacements[original] = modified
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
	if evidence := os.Getenv("ADAMIC_SLOT05_BATCH22_EVIDENCE"); evidence != "" {
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
