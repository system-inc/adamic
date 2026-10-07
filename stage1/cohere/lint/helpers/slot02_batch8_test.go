package helpers

import (
	"bufio"
	"bytes"
	"compress/gzip"
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

func TestSlot02Batch8(t *testing.T) {
	// Not parallel: baseline and compiling semantic mutants share one large
	// consumer corpus; sequential execution bounds native sanitizer memory.
	base := "slot02/batch8"
	file, err := os.Open(base + "/testdata/sources.jsonl.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	zipped, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer zipped.Close()
	var inputs []map[string]string
	observed := map[string]bool{}
	scan := bufio.NewScanner(zipped)
	scan.Buffer(make([]byte, 65536), 16<<20)
	for scan.Scan() {
		var row map[string]string
		if err = json.Unmarshal(scan.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		inputs = append(inputs, row)
		observed[row["rule"]] = true
	}
	if err = scan.Err(); err != nil {
		t.Fatal(err)
	}
	var ledger struct {
		Remaining []struct {
			Rule    string
			Helpers []string `json:"remaining_helpers"`
		}
	}
	data, err := os.ReadFile("readiness.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	symbols := []string{"regexp.EscapeClassRune", "regexp.identityEscape", "regexp.literalRune"}
	for _, symbol := range symbols {
		count := 0
		for _, row := range ledger.Remaining {
			for _, h := range row.Helpers {
				if strings.HasSuffix(h, "/"+symbol) {
					count++
					if !observed[row.Rule] {
						t.Fatalf("missing captured consumer %s", row.Rule)
					}
				}
			}
		}
		if count != 4 {
			t.Fatalf("consumer count drift: %s: %d", symbol, count)
		}
		t.Logf("%s: %d consumers with actual runtime capture", symbol, count)
	}
	config := smallFixture(t, inputs)
	cohere, _ := filepath.Abs("../../../../cohere")
	oracleSource, _ := filepath.Abs(base + "/testdata/oracle.go")
	export, _ := filepath.Abs(base + "/testdata/regexp_export.go")
	replacements := map[string]string{filepath.Join(cohere, "adamic_slot02_batch8.go"): oracleSource, filepath.Join(cohere, "internal/lint/ecmascript/regexp/adamic_slot02_batch8.go"): export}
	overlay := smallFixture(t, map[string]any{"Replace": replacements})
	oracle := filepath.Join(t.TempDir(), "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlay, "-o", oracle, filepath.Join(cohere, "adamic_slot02_batch8.go"))
	data = run(t, "", oracle, config)
	var corpus struct {
		Want                   string
		Sources, OptionStrings int
		Runes, Contexts        []json.RawMessage
	}
	if err = json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	var input map[string]json.RawMessage
	if err = json.Unmarshal(data, &input); err != nil {
		t.Fatal(err)
	}
	delete(input, "Want")
	path := smallFixture(t, input)
	want := []byte(corpus.Want)
	t.Logf("%d captured source/options inputs; %d rune values; %d contexts; %d runtime option strings", len(inputs), len(corpus.Runes), len(corpus.Contexts), corpus.OptionStrings)
	entry, _ := filepath.Abs(base + "/main.a")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	check := func(t *testing.T, entry string) []byte {
		t.Helper()
		gotNode := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path)
		program, err := load.Load([]string{entry})
		if err != nil {
			t.Fatal(err)
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			t.Fatal(err)
		}
		nativePath := filepath.Join(t.TempDir(), "program")
		if err = native.Build(native.C(ir), nativePath, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		got := run(t, "", nativePath, path)
		jsPath := filepath.Join(t.TempDir(), "program.mjs")
		if err = os.WriteFile(jsPath, []byte(javascript.JavaScript(ir)), 0644); err != nil {
			t.Fatal(err)
		}
		gotJS := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, jsPath, path)
		compare(t, got, gotNode)
		compare(t, gotJS, gotNode)
		return got
	}
	compare(t, check(t, entry), want)
	t.Log("Actual Go, Node source, emitted JavaScript and sanitized native agree")
	mutants := []struct{ name, file, anchor, replacement string }{
		{"class dash escaping", "escape_class_rune.a", "rune === 45", "false"},
		{"class open bracket escaping", "escape_class_rune.a", "rune === 91", "false"},
		{"invalid rune replacement", "escape_class_rune.a", "valid ? rune : 65533", "valid ? rune : 63"},
		{"Unicode identity restriction", "identity_escape.a", "context.unicode && !syntax", "false && !syntax"},
		{"slash identity exemption", "identity_escape.a", "rune !== 47", "true"},
		{"dash outside class", "identity_escape.a", "(!context.inClass || rune !== 45)", "rune !== 45"},
		{"identity width", "identity_escape.a", "width: size,", "width: size + 1,"},
		{"identity error cause", "identity_escape.a", "cause: 'unsupported regexp syntax'", "cause: ''"},
		{"literal lowercase hex", "literal_rune.a", "'0123456789abcdef'", "'0123456789ABCDEF'"},
		{"literal minimum hex width", "literal_rune.a", "negative ? 3 : 4", "negative ? 3 : 3"},
		{"literal supplementary boundary", "literal_rune.a", "rune > 65535", "rune >= 65535"},
		{"literal z passthrough", "literal_rune.a", "rune <= 122", "rune < 122"},
		{"literal negative padding", "literal_rune.a", "negative ? 3 : 4", "negative ? 4 : 4"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			dir := t.TempDir()
			files := []string{"options_json.ts", base + "/main.a", base + "/escape_class_rune.a", base + "/identity_escape.a", base + "/literal_rune.a"}
			for _, name := range files {
				data, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				if name == base+"/"+mutant.file {
					if strings.Count(string(data), mutant.anchor) != 1 {
						t.Fatal("mutant anchor drift")
					}
					data = []byte(strings.Replace(string(data), mutant.anchor, mutant.replacement, 1))
				}
				targetName := name
				if name == "options_json.ts" {
					targetName = "options_json.a"
				}
				if name == base+"/main.a" {
					data = []byte(strings.Replace(string(data), "../../options_json.ts", "../../options_json.a", 1))
				}
				target := filepath.Join(dir, targetName)
				if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(target, data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			got := check(t, filepath.Join(dir, base, "main.a"))
			if bytes.Equal(got, want) {
				t.Fatal("compiled semantic mutant survived")
			}
			a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
			for i, line := range a {
				if i < len(b) && line != b[i] {
					t.Logf("compiled semantic mutant caught at line %d: got %q; Go %q", i+1, line, b[i])
					break
				}
			}
		})
	}
}
