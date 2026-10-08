package regexsyntax

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

func command(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	c := exec.CommandContext(ctx, name, args...)
	c.Dir = dir
	var stderr bytes.Buffer
	file, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	c.Stdout = file
	c.Stderr = &stderr
	if err = c.Run(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	if name != "python3" && stderr.Len() != 0 {
		t.Fatalf("%s stderr: %s", name, &stderr)
	}
	data, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func compile(t *testing.T, dir string, js bool) string {
	t.Helper()
	program, err := load.Load([]string{filepath.Join(dir, "main.a")})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "helpers")
	if js {
		out += ".mjs"
		err = os.WriteFile(out, []byte(javascript.JavaScript(ir)), 0644)
	} else {
		err = native.Build(native.C(ir), out, native.Options{Sanitize: true})
	}
	if err != nil {
		t.Fatal(err)
	}
	return out
}
func assertBytes(t *testing.T, side string, got, want []byte) {
	t.Helper()
	if bytes.Equal(got, want) {
		return
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			t.Fatalf("%s line %d: got %s, Go %s", side, i+1, a[i], b[i])
		}
	}
	t.Fatalf("%s bytes %d vs Go %d", side, len(got), len(want))
}
func TestRegexsyntaxCapture(t *testing.T) {
	root, _ := filepath.Abs("../../../../..")
	dir, _ := filepath.Abs(".")
	corpus := t.TempDir()
	if ready := os.Getenv("ADAMIC_REGEXSYNTAX_CAPTURE"); ready != "" {
		corpus = ready
	} else {
		t.Log(string(command(t, root, "python3", filepath.Join(dir, "testdata/capture.py"), corpus)))
	}
	cases := filepath.Join(corpus, "cases.json")
	want, err := os.ReadFile(filepath.Join(corpus, "want.txt"))
	if err != nil {
		t.Fatal(err)
	}
	counts, err := os.ReadFile(filepath.Join(corpus, "counts.json"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("every actual Go use: %s", counts)
	var hits map[string]int
	if err = json.Unmarshal(counts, &hits); err != nil {
		t.Fatal(err)
	}
	if len(hits) != 13 {
		t.Fatalf("captured %d helpers, need 13", len(hits))
	}
	runner := filepath.Join(root, "oracle/node.mjs")
	node := func(entry string) []byte {
		return command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases)
	}
	assertBytes(t, "Node", node(filepath.Join(dir, "main.a")), want)
	assertBytes(t, "emitted JavaScript", node(compile(t, dir, true)), want)
	assertBytes(t, "sanitized native", command(t, "", compile(t, dir, false), cases), want)
	t.Logf("Node/emitted/native byte agreement: %d rows, %d bytes", bytes.Count(want, []byte("\n")), len(want))
	for _, m := range []struct{ file, from, to string }{
		{"is_hex_digit.a", "byte <= 57", "byte <= 56"},
		{"all_hex_digits.a", "bytes.length === 0", "bytes.length < 0"},
		{"hex_value.a", "return b - 65 + 10", "return b - 65 + 11"},
		{"parse_hex_uint.a", "value << 4", "value << 3"},
		{"parse_regex_flags.a", "bytes.includes(117)", "bytes.includes(103)"},
		{"pattern_and_flags.a", "bytes.slice(lastSlash + 1)", "bytes.slice(lastSlash)"},
		{"regex_flags_uv.a", "flags.unicode || flags.unicodeSets", "flags.unicode && flags.unicodeSets"},
		{"skip_pattern_escape.a", "value: 4, ok: true", "value: 3, ok: true"},
		{"class_end.a", "value: index, ok: true", "value: index - 1, ok: true"},
		{"read_raw_class_char.a", "rune[0] ?? 0", "(rune[0] ?? 0) + 1"},
		{"read_class_escape.a", "value = 8", "value = 9"},
		{"parse_regex_character_class.a", "Kind:1,Value:min.Value", "Kind:0,Value:min.Value"},
		{"parse_regex_character_class_with_end.a", "r.end !== end", "r.end === end"},
	} {
		t.Run(m.file, func(t *testing.T) {
			mutant := t.TempDir()
			files, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range files {
				if !strings.HasSuffix(f.Name(), ".a") {
					continue
				}
				data, err := os.ReadFile(filepath.Join(dir, f.Name()))
				if err != nil {
					t.Fatal(err)
				}
				text := string(data)
				if f.Name() == m.file {
					if strings.Count(text, m.from) != 1 {
						t.Fatal("mutant anchor drift", m.from)
					}
					text = strings.Replace(text, m.from, m.to, 1)
				}
				text = strings.ReplaceAll(text, "'../options_json.ts'", "'"+filepath.Join(root, "stage1/cohere/lint/helpers/options_json.ts")+"'")
				if err = os.WriteFile(filepath.Join(mutant, f.Name()), []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
			}
			for _, side := range []struct {
				name string
				got  []byte
			}{{"Node", node(filepath.Join(mutant, "main.a"))}, {"native", command(t, "", compile(t, mutant, false), cases)}} {
				if bytes.Equal(side.got, want) {
					t.Fatal(side.name, "semantic mutant survived")
				}
				t.Log(side.name, "compiled and ran; output disagrees with Go")
			}
		})
	}
}
