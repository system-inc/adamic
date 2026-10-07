package typeaware

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Not parallel: native components, sanitizer binaries and oracle artifacts share a machine.
// This is component parity, not complete JSX rule parity.
func TestWave30JsxComponents(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE30_JSX_COMPONENTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err = os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	virtual := filepath.Join(repository, "cohere/internal/lint/rules/react/adamic_wave30_jsx_components_test.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(repository, "stage1/cohere/typeaware/testdata/wave_30_jsx_components_oracle_test.go")}})
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "test", "-overlay", h.write("jsx-overlay.json", string(overlay)), "./internal/lint/rules/react", "-run", "^TestWave30JsxComponentOracle$", "-count=1", "-v")
	command.Dir = filepath.Join(repository, "cohere")
	command.Env = append(os.Environ(), "ADAMIC_WAVE30_JSX_COMPONENTS="+directory)
	h.must("jsx-go", command)
	root := filepath.Join(repository, "stage1/cohere/typeaware/wave_30_jsx")
	entry := filepath.Join(root, "components_suite.a")
	binary := filepath.Join(directory, "jsx-components")
	asan := binary + "-asan"
	h.must("jsx-build", exec.Command(stage0, "build", entry, "-o", binary))
	h.must("jsx-asan-build", exec.Command(stage0, "build", entry, "-o", asan, "--sanitize"))
	emitted := h.must("jsx-js-emit", exec.Command(stage0, "js", entry))
	javascript := h.write("jsx-components.js", string(emitted.stdout))
	runtimeDirectory := filepath.Join(directory, "node_modules/adamic")
	if err = os.MkdirAll(runtimeDirectory, 0755); err != nil {
		t.Fatal(err)
	}
	runtime, err := os.ReadFile(filepath.Join(repository, "oracle/adamic.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(runtimeDirectory, "package.json"), []byte(`{"type":"module","exports":"./index.mjs"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(runtimeDirectory, "index.mjs"), runtime, 0600); err != nil {
		t.Fatal(err)
	}
	rows := []struct {
		mode, slug, from, to string
		kinds                []ast.Kind
	}{
		{"fragments", "jsx-fragments", "node.objectName === 'React'", "node.objectName === 'Other'", []ast.Kind{ast.KindJsxFragment, ast.KindJsxElement, ast.KindJsxSelfClosingElement}},
		{"undef", "jsx-no-undef", "name.includes('-')", "name.includes('_')", []ast.Kind{ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement}},
		{"context", "jsx-no-constructed-context-values", "node.construction === 3", "node.construction === 4", []ast.Kind{ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement}},
	}
	for _, row := range rows {
		descriptor, err := os.ReadFile(filepath.Join(root, row.slug, "rule.json"))
		if err != nil {
			t.Fatal(err)
		}
		var metadata struct {
			Name   string
			Kinds  []string
			Status string
		}
		if err = json.Unmarshal(descriptor, &metadata); err != nil {
			t.Fatal(err)
		}
		if metadata.Name != "react/"+row.slug || metadata.Status != "partial components only" || len(metadata.Kinds) != len(row.kinds) {
			t.Fatal("invalid partial component descriptor")
		}
		if !wave30JsxKindsAgree(metadata.Kinds, row.kinds) {
			t.Fatalf("%s named listener differs from pinned Go", row.slug)
		}
		var fields map[string]json.RawMessage
		if err = json.Unmarshal(descriptor, &fields); err != nil {
			t.Fatal(err)
		}
		if _, exists := fields["numericKinds"]; exists {
			t.Fatal("obsolete numeric listener declaration")
		}
		changed := append([]string(nil), metadata.Kinds...)
		changed[0] = "SourceFile"
		if wave30JsxKindsAgree(changed, row.kinds) {
			t.Fatalf("%s named listener mutant escaped Go comparison", row.slug)
		}
		t.Logf("%s: wrong-name descriptor mutant caught by pinned Go kind-name comparison", row.slug)
		truth, err := os.ReadFile(filepath.Join(directory, row.mode+"-go.stdout"))
		if err != nil {
			t.Fatal(err)
		}
		for _, runner := range []struct{ name, path string }{{"native", binary}, {"asan", asan}, {"javascript", "node"}} {
			args := []string{row.mode}
			if runner.name == "javascript" {
				args = []string{javascript, row.mode}
			}
			got := h.must(row.mode+"-"+runner.name, exec.Command(runner.path, args...))
			if len(got.stderr) != 0 || !bytes.Equal(got.stdout, truth) {
				t.Fatalf("%s %s differs at byte %d: %s", row.mode, runner.name, firstDifference(got.stdout, truth), got.stderr)
			}
		}
		scratch := filepath.Join(directory, row.mode+"-mutant-source")
		for _, slug := range []string{"jsx-fragments", "jsx-no-undef", "jsx-no-constructed-context-values"} {
			data, err := os.ReadFile(filepath.Join(root, slug, "components.a"))
			if err != nil {
				t.Fatal(err)
			}
			source := string(data)
			if slug == row.slug {
				if strings.Count(source, row.from) != 1 {
					t.Fatal("nonunique JSX component mutant")
				}
				source = strings.Replace(source, row.from, row.to, 1)
			}
			if err = os.MkdirAll(filepath.Join(scratch, slug), 0755); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(scratch, slug, "components.a"), []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
		}
		printable, err := os.ReadFile(filepath.Join(root, "jsx-no-constructed-context-values", "printable.a"))
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(scratch, "jsx-no-constructed-context-values", "printable.a"), printable, 0600); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(entry)
		if err != nil {
			t.Fatal(err)
		}
		mutantEntry := filepath.Join(scratch, "components_suite.a")
		source := strings.ReplaceAll(string(data), "../../../typescript/", filepath.Join(repository, "stage1/typescript")+"/")
		if err = os.WriteFile(mutantEntry, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		mutant := filepath.Join(directory, row.mode+"-mutant")
		h.must(row.mode+"-mutant-build", exec.Command(stage0, "build", mutantEntry, "-o", mutant))
		got := h.must(row.mode+"-mutant-run", exec.Command(mutant, row.mode))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth) {
			t.Fatalf("%s mutant escaped comparison", row.mode)
		}
		t.Logf("%s: %d Go bytes match native, sanitizer and emitted JavaScript; successful-exit mutant caught only by Go at byte %d", row.mode, len(truth), firstDifference(got.stdout, truth))
	}

	// A successful Unicode escape mutant must be caught by the Go byte comparison.
	scratch := filepath.Join(directory, "context-mutant-source")
	original, err := os.ReadFile(filepath.Join(root, "jsx-no-constructed-context-values", "components.a"))
	if err != nil {
		t.Fatal(err)
	}
	const anchor = "else if(code <= 0xffff)"
	if strings.Count(string(original), anchor) != 1 {
		t.Fatal("nonunique Unicode escape mutant")
	}
	mutated := strings.Replace(string(original), anchor, "else if(code <= 0x10ffff)", 1)
	if err = os.WriteFile(filepath.Join(scratch, "jsx-no-constructed-context-values", "components.a"), []byte(mutated), 0600); err != nil {
		t.Fatal(err)
	}
	mutant := filepath.Join(directory, "unicode-mutant")
	h.must("unicode-mutant-build", exec.Command(stage0, "build", filepath.Join(scratch, "components_suite.a"), "-o", mutant))
	unicodeResult := h.must("unicode-mutant-run", exec.Command(mutant, "context"))
	truth, err := os.ReadFile(filepath.Join(directory, "context-go.stdout"))
	if err != nil {
		t.Fatal(err)
	}
	if len(unicodeResult.stderr) != 0 || bytes.Equal(unicodeResult.stdout, truth) {
		t.Fatal("Unicode escape mutant escaped comparison")
	}
	t.Logf("Unicode astral escape mutant exits successfully and is caught only by Go at byte %d", firstDifference(unicodeResult.stdout, truth))
	got := h.run("unsupported-name", exec.Command(binary, "unsupported"))
	if exit, ok := got.err.(*exec.ExitError); !ok || exit.ExitCode() != 70 || string(got.stderr) != "adamic: panic: NotYet: unpaired surrogate in construction name\n" {
		t.Fatalf("unsupported name silently accepted: %v %s", got.err, got.stderr)
	}
}

// Follow the registry's pinned AST spelling; numeric ordinals are not registry input.
func wave30JsxKindsAgree(names []string, expected []ast.Kind) bool {
	if len(names) != len(expected) {
		return false
	}
	for at, kind := range expected {
		if names[at] != strings.TrimPrefix(kind.String(), "Kind") {
			return false
		}
	}
	return true
}
