package formatfiles

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Walked paths are absolute, so they cannot exercise a dot at index zero.
// Hold ext itself to filepath.Ext, including bare names and empty inputs.
func TestExtAgreesWithGo(test *testing.T) {
	test.Parallel()
	ctx := formatfilesDeadline(test, "ext agreement")
	cases := []struct {
		name, path string
	}{
		{"bare dot i", ".i"},
		{"single dot", "."},
		{"double dot", ".."},
		{"no dot", "a"},
		{"trailing dot", "a."},
		{"last of multiple dots", "a.b.c"},
		{"absolute nested dotfile", "/a/.i"},
		{"absolute root dotfile", "/.i"},
		{"root", "/"},
		{"empty", ""},
		{"relative nested dotfile", "x/.i"},
		{"dot before separator", "x.y/z"},
		{"nested trailing dot", "x/y."},
		{"bare Unicode dotfile", ".İ"},
		{"Unicode extension", "h.İ"},
		{"dot at zero without slash", ".hidden"},
		{"escaped extension", "a.\"\\\n\t"},
	}
	var names []string
	var expected bytes.Buffer
	for _, each := range cases {
		names = append(names, each.path)
		encoded, encodeError := json.Marshal(filepath.Ext(each.path))
		if encodeError != nil {
			test.Fatal(encodeError)
		}
		expected.Write(encoded)
		expected.WriteByte('\n')
	}
	// Both copies use the package's port layout and its existing Node oracle.
	driver := func(applied *mutant, inputs []string) []byte {
		test.Helper()
		port := portDirectory(test, applied)
		path := filepath.Join(port, "ext_driver.ts")
		source := "import { ext } from './golang.ts';\n" +
			"for (const name of process.argv.slice(2)) { console.log(JSON.stringify(ext(name))); }\n"
		if writeError := os.WriteFile(path, []byte(source), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
		result := formatfilesNode(test, ctx, path, inputs...)
		if result.exitCode != 0 || len(result.stderr) != 0 {
			test.Fatalf("Node: exit %d, stderr %q", result.exitCode, result.stderr)
		}
		return result.stdout
	}
	got := driver(nil, names)
	gotLines := bytes.Split(got, []byte{'\n'})
	wantLines := bytes.Split(expected.Bytes(), []byte{'\n'})
	if len(gotLines) != len(wantLines) {
		test.Fatalf("Node output has %d lines, want %d: %q", len(gotLines), len(wantLines), got)
	}
	for index, each := range cases {
		test.Run(each.name, func(test *testing.T) {
			if !bytes.Equal(gotLines[index], wantLines[index]) {
				test.Errorf("ext(%q): Node %s, filepath.Ext %s", each.path, gotLines[index], wantLines[index])
			}
		})
	}
	if !bytes.Equal(got, expected.Bytes()) {
		test.Errorf("Node and filepath.Ext output differ byte for byte: got %q, want %q", got, expected.Bytes())
	}

	// Do not add this to mutants: that runner only walks absolute tree paths,
	// on which skipping index zero is equivalent to the original loop.
	applied := mutant{
		name: "ext skips index zero",
		file: "golang.ts",
		from: "index >= 0",
		to:   "index > 0",
	}
	mutated := driver(&applied, []string{".i"})
	if bytes.Equal(mutated, append(append([]byte(nil), wantLines[0]...), '\n')) {
		test.Fatal("index-zero mutant agrees with filepath.Ext on .i")
	}
	if string(mutated) != "\"\"\n" {
		test.Fatalf("unexpected mutant output for .i: %q", mutated)
	}
	test.Logf(
		"mutant disagreement: ext(%q): Node %s, filepath.Ext %s",
		".i",
		strings.TrimSuffix(string(mutated), "\n"),
		wantLines[0],
	)
}
