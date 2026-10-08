package unit4

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func unit4MutantFiles(t *testing.T, root, file, old, replacement string) string {
	t.Helper()
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation/passes/unit-4")
	destination := t.TempDir()
	files, err := filepath.Glob(filepath.Join(lane, "*.a"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if filepath.Base(path) == file {
			if strings.Count(text, old) != 1 {
				t.Fatalf("mutant anchor %s occurs %d times", file, strings.Count(text, old))
			}
			text = strings.Replace(text, old, replacement, 1)
		}
		// Scratch files import the actual shared modules by reference, not copied implementations.
		for _, shared := range []string{"replay/index.ts", "core.ts", "dump.ts"} {
			text = strings.ReplaceAll(text, "'../../"+shared+"'", "'"+filepath.Join(root, "stage1/cohere/high_level_intermediate_representation", shared)+"'")
		}
		if err := os.WriteFile(filepath.Join(destination, filepath.Base(path)), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return destination
}
func unit4Mismatch(t *testing.T, output []byte, manifest string) string {
	t.Helper()
	cases := map[string]string{}
	for _, chunk := range strings.Split(string(output), "checkpoint\t")[1:] {
		key, body, ok := strings.Cut(chunk, "\n")
		if !ok {
			t.Fatal("mutant did not finish framing")
		}
		if _, exists := cases[key]; exists {
			t.Fatal("duplicate mutant output")
		}
		cases[key] = body
	}
	first := ""
	for _, line := range strings.Split(manifest, "\n") {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		want, err := os.ReadFile(fields[2])
		if err != nil {
			t.Fatal(err)
		}
		got, exists := cases[fields[0]]
		if !exists {
			t.Fatal("mutant dropped input " + fields[0])
		}
		if got != string(want) && first == "" {
			first = fields[0]
		}
		delete(cases, fields[0])
	}
	if len(cases) != 0 {
		t.Fatal("mutant invented inputs")
	}
	return first
}
func unit4SemanticMutants(t *testing.T, root, pass, manifest, manifestPath string) {
	t.Helper()
	file, old, replacement, entry := "primitive.a", "this.kind(this.root(this.key(index,instruction.lvalue.identifier))) === 1", "this.kind(this.root(this.key(index,instruction.lvalue.identifier))) === 2", "primitive_main.a"
	if pass == "reactive" {
		file = "reactive.a"
		old = "for(const param of this.fn.params) { this.mark(param.identifier); }"
		replacement = "for(const param of this.fn.params) { this.fn.identifier(param.identifier); }"
		entry = "main.a"
	}
	if strings.HasPrefix(pass, "effects") {
		file = "effects.a"
		old = "...v.captures.map((p) => flow('capture',p,l))"
		replacement = "...v.captures.map((p) => flow('immutable-capture',p,l))"
		entry = "effects_main.a"
	}
	directory := unit4MutantFiles(t, root, file, old, replacement)
	output := unit4Run(t, root, "node", "--no-warnings", "oracle/node.mjs", filepath.Join(directory, entry), manifestPath)
	first := unit4Mismatch(t, output, manifest)
	if first == "" {
		t.Fatalf("%s semantic mutant survived", pass)
	}
	t.Logf("%s semantic mutant caught by Go byte comparison on Node: %s", pass, first)
	if !strings.HasPrefix(pass, "effects") {
		unit4MutantBackends(t, root, pass, filepath.Join(directory, entry), manifest, manifestPath)
	}

	if pass == "effects" {
		directory = unit4MutantFiles(t, root, "effects.a", ": signature('read',[],'freeze',true,'frozen')", ": signature('read',[],'freeze',true,'mutable')")
		output = unit4Run(t, root, "node", "--no-warnings", "oracle/node.mjs", filepath.Join(directory, entry), manifestPath)
		if unit4Mismatch(t, output, manifest) == "" {
			t.Fatal("custom-hook signature mutant survived")
		}
		t.Log("custom-hook frozen-result mutant caught by Go byte comparison on Node")
	}
}
func unit4Backends(t *testing.T, root, pass, entry, manifest, manifestPath string) {
	t.Helper()
	if strings.HasPrefix(pass, "effects") {
		t.Logf("%s compiled backends stopped: effects.a recursive initializer gap; no certificate claimed", pass)
		return
	}
	javascript := unit4Run(t, root, "go", "run", "./cmd/adamic", "js", entry)
	js := filepath.Join(t.TempDir(), "primitive.js")
	if err := os.WriteFile(js, javascript, 0600); err != nil {
		t.Fatal(err)
	}
	unit4Compare(t, unit4Run(t, root, "node", "--no-warnings", "oracle/node.mjs", js, manifestPath), manifest)
	t.Logf("%s emitted JavaScript: 1465/1465", pass)
	binary := filepath.Join(t.TempDir(), "primitive")
	unit4Run(t, root, "go", "run", "./cmd/adamic", "build", entry, "-o", binary, "--sanitize")
	unit4Compare(t, unit4Run(t, root, binary, manifestPath), manifest)
	t.Logf("%s sanitized native: 1465/1465", pass)
}

func unit4MutantBackends(t *testing.T, root, pass, entry, manifest, manifestPath string) {
	t.Helper()
	javascript := unit4Run(t, root, "go", "run", "./cmd/adamic", "js", entry)
	js := filepath.Join(t.TempDir(), "mutant.js")
	if err := os.WriteFile(js, javascript, 0600); err != nil {
		t.Fatal(err)
	}
	output := unit4Run(t, root, "node", "--no-warnings", "oracle/node.mjs", js, manifestPath)
	if unit4Mismatch(t, output, manifest) == "" {
		t.Fatalf("%s emitted JavaScript mutant survived", pass)
	}
	t.Logf("%s semantic mutant caught by Go byte comparison on emitted JavaScript", pass)
	binary := filepath.Join(t.TempDir(), "mutant")
	unit4Run(t, root, "go", "run", "./cmd/adamic", "build", entry, "-o", binary, "--sanitize")
	output = unit4Run(t, root, binary, manifestPath)
	if unit4Mismatch(t, output, manifest) == "" {
		t.Fatalf("%s native mutant survived", pass)
	}
	t.Logf("%s semantic mutant caught by Go byte comparison on sanitized native", pass)
}
