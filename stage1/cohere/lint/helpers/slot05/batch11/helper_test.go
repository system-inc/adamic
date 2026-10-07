package batch11

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

func TestSlot05ClassAtomWrite(t *testing.T) {
	verify(t, "atom", "class_atom_write.a", mutation{"atom.kind === 2", "atom.kind === 3"}, mutation{"literalRune(atom.hi)", "literalRune(atom.lo)"}, mutation{" + '-' + ", " + ':' + "})
}
func TestSlot05WordCharacters(t *testing.T) {
	verify(t, "word", "word_characters.a", mutation{"wordClassAtoms(options)", "wordClassAtoms({unicode: false, ignoreCase: false, multiline: false, dotAll: false})"}, mutation{"), false,", "), true,"}, mutation{"unicode: false, ignoreCase: false", "unicode: true, ignoreCase: false"})
}
func TestSlot05ExpandsOnUppercase(t *testing.T) {
	verify(t, "expand", "expands_on_uppercase.a", mutation{"rune <= 0x1f87", "rune < 0x1f87"}, mutation{"rune === 0x1ff3", "rune === 0x1ff4"})
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
	virtual := filepath.Join(cohere, "adamic_slot05_batch11.go")
	replacements := map[string]string{virtual: filepath.Join(dir, "testdata", "oracle.go")}

	replacements[filepath.Join(cohere, "internal/lint/ecmascript/regexp/adamic_slot05_batch11.go")] = filepath.Join(dir, "testdata/regexp_export.go")

	// Instrument calls only; the actual Go algorithm and all returned bytes stay unchanged.
	for _, file := range []string{"class.go", "rewrite.go"} {
		path := filepath.Join(cohere, "internal/lint/ecmascript/regexp", file)
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(source)
		if file == "class.go" {
			start := strings.Index(text, "func (a classAtom) write() string {")
			end := strings.Index(text[start:], "\n}\n") + start + 3
			text = text[:start] + strings.ReplaceAll(text[start:end], "literalRune(", "adamicTraceRune(") + text[end:]
		} else {
			old := "return writeClass(wordClassAtoms(options), false, rewriteOptions{})"
			if strings.Count(text, old) != 1 {
				t.Fatal("word instrumentation drift")
			}
			text = strings.Replace(text, old, "return adamicTraceWriteClass(adamicTraceWordAtoms(options), false, rewriteOptions{})", 1)
		}
		instrumented := filepath.Join(scratch, file)
		if err := os.WriteFile(instrumented, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		replacements[path] = instrumented
	}

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
	if evidence := os.Getenv("ADAMIC_SLOT05_BATCH11_EVIDENCE"); evidence != "" {
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
				t.Logf("%s mutant %q -> %q caught at verdict %d: mutant %q, Go %q", target, old, replacement, i+1, a[i], b[i])
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
