package wave10

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func unescapeOracle(t *testing.T) string {
	t.Helper()
	root := absolute(t, "../../../../../cohere")
	scratch := t.TempDir()
	main := filepath.Join(root, "adamic_css_unescape.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{main: absolute(t, "testdata/unescape_oracle.go.txt"), filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_css_unescape.go"): absolute(t, "testdata/unescape_exports.go.txt")}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(scratch, "overlay.json")
	write(t, path, overlay)
	binary := filepath.Join(scratch, "go-oracle")
	run(t, root, "go", "build", "-overlay="+path, "-o", binary, main)
	return binary
}
func unescapeCases(t *testing.T) string {
	t.Helper()
	file, err := os.Open("testdata/unescape_cases.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	out, err := os.Create(filepath.Join(t.TempDir(), "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	if _, err = io.Copy(out, reader); err != nil {
		t.Fatal(err)
	}
	return out.Name()
}

// Not parallel: exhaustive scalar inputs allocate large JSON arenas on sanitized native.
func TestCSSIdentifierUnescape(t *testing.T) {
	checkCoverage(t)
	path := unescapeCases(t)
	want := run(t, "", unescapeOracle(t), path)
	entry := absolute(t, "unescape_main.a")
	for i, cmd := range commands(t, entry, path) {
		got := run(t, "", cmd[0], cmd[1:]...)
		if !bytes.Equal(got, want) {
			t.Fatalf("runtime %d differs: got %d Go %d", i, len(got), len(want))
		}
		t.Logf("runtime %d matched %d bytes", i, len(got))
	}
	t.Log("1116695 cases, every scalar through 0x110000 plus six consumers' literals and CSS boundary controls")
}
func TestCSSIdentifierUnescapeMutant(t *testing.T) {
	// Keep exhaustive scalar proof in the baseline; this independent bounded corpus targets the mutant.
	values := []string{`\0 `, `\d800 `, `\dfff `, `\110000 `, `\41 `, `\1f600 `, `\41`}
	data, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "mutant-cases.json")
	write(t, path, data)
	want := run(t, "", unescapeOracle(t), path)
	directory := t.TempDir()
	main, err := os.ReadFile("unescape_main.a")
	if err != nil {
		t.Fatal(err)
	}
	main = []byte(strings.ReplaceAll(string(main), "../options_json.ts", filepath.ToSlash(absolute(t, "../options_json.ts"))))
	write(t, filepath.Join(directory, "main.a"), main)
	source, err := os.ReadFile("css_identifier_unescape.a")
	if err != nil {
		t.Fatal(err)
	}
	old := "if(codePoint === 0 || codePoint > 0x10ffff || (codePoint >= 0xd800 && codePoint <= 0xdfff))"
	replacement := "if(codePoint <= 1 || codePoint > 0x10ffff || (codePoint >= 0xd800 && codePoint <= 0xdfff))"
	// U+0001 is valid in Go; changing it to replacement remains a successful execution.
	values = append(values, `\1 `)
	data, err = json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	write(t, path, data)
	want = run(t, "", unescapeOracle(t), path)
	if strings.Count(string(source), old) != 1 {
		t.Fatal("mutant anchor changed")
	}
	write(t, filepath.Join(directory, "css_identifier_unescape.a"), []byte(strings.Replace(string(source), old, replacement, 1)))
	for i, cmd := range commands(t, filepath.Join(directory, "main.a"), path) {
		got := run(t, "", cmd[0], cmd[1:]...)
		if bytes.Equal(got, want) {
			t.Fatalf("runtime %d mutant survived", i)
		}
		t.Logf("runtime %d compiling scalar_one_replaced mutant caught only by output comparison", i)
	}
}
