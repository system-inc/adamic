package helpers

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func slot01Wave10Fixture(t *testing.T, dependency string, consumers int, mode string) (string, []byte) {
	t.Helper()
	var readiness struct {
		Remaining []struct {
			Rule             string
			RemainingHelpers []string `json:"remaining_helpers"`
		}
	}
	data, err := os.ReadFile("readiness.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &readiness); err != nil {
		t.Fatal(err)
	}
	root, _ := filepath.Abs("../../../../cohere")
	var paths []string
	for _, rule := range readiness.Remaining {
		if !slicesContain(rule.RemainingHelpers, dependency) {
			continue
		}
		namespace, name, has := strings.Cut(rule.Rule, "/")
		if !has {
			namespace = "core"
			name = rule.Rule
		}
		if namespace == "@next" {
			name = strings.TrimPrefix(name, "next/")
			namespace = "next"
		}
		if namespace == "better-tailwindcss" {
			namespace = "tailwind"
		}
		if namespace == "@typescript-eslint" {
			namespace = "typescript"
		}
		path := filepath.Join(root, "internal/lint/rules", namespace, strings.ReplaceAll(name, "-", "_")+"_test.go")
		if name == "react-element-no-anchor" || name == "react-element-no-horizontal-rule" {
			path = filepath.Join(root, "internal/lint/rules/structure/react_no_intrinsic_element_test.go")
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("consumer %s: %v", rule.Rule, err)
		}
		paths = append(paths, path)
	}
	if len(paths) != consumers {
		t.Fatalf("consumer ledger drift: %d", len(paths))
	}
	directory := t.TempDir()
	manifest, _ := json.Marshal(paths)
	manifestPath := filepath.Join(directory, "manifest.json")
	if err := os.WriteFile(manifestPath, manifest, 0644); err != nil {
		t.Fatal(err)
	}
	source, _ := filepath.Abs("testdata/slot01_wave10_oracle.go")
	virtual := filepath.Join(root, "adamic_slot01_wave10.go")
	bridge, _ := filepath.Abs("testdata/slot01_wave10_regexp.go")
	rewritePath := filepath.Join(root, "internal/lint/ecmascript/regexp/rewrite.go")
	rewrite, err := os.ReadFile(rewritePath)
	if err != nil {
		t.Fatal(err)
	}
	anchor := "word := wordCharacters(options)"
	if strings.Count(string(rewrite), anchor) != 1 {
		t.Fatal("boundary observation anchor drift")
	}
	observedPath := filepath.Join(directory, "rewrite.go")
	if err := os.WriteFile(observedPath, []byte(strings.Replace(string(rewrite), anchor, "word := adamicWave10WordCharacters(options)", 1)), 0644); err != nil {
		t.Fatal(err)
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source, filepath.Join(root, "internal/lint/ecmascript/regexp/adamic_slot01_wave10.go"): bridge, rewritePath: observedPath}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "go-oracle")
	run(t, root, "go", "build", "-overlay="+overlayPath, "-o", binary, virtual)
	cases := filepath.Join(directory, "cases.jsonl")
	expected := filepath.Join(directory, "expected.log")
	cmd := exec.Command(binary, manifestPath, cases, expected, mode)
	log, err := os.Create(filepath.Join(directory, "generation.log"))
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stdout = log
	cmd.Stderr = log
	err = cmd.Run()
	log.Close()
	generated, _ := os.ReadFile(log.Name())
	t.Logf("Go fixture observations:\n%s", generated)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(expected)
	if err != nil {
		t.Fatal(err)
	}
	if len(want) == 0 {
		t.Fatal("empty Go observations")
	}
	return cases, want
}
func slot01Wave10Check(t *testing.T, dependency, mode, file, old, replacement string) {
	t.Helper()
	cases, want := slot01Wave10Fixture(t, dependency, 4, mode)
	entry, _ := filepath.Abs("slot01_wave10_main.a")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases, mode), want)
	compare(t, run(t, "", slot01Build(t, entry), cases, mode), want)
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, slot01Wave3JavaScript(t, entry), cases, mode), want)
	t.Logf("%d Go output lines matched source Node, native and emitted JavaScript", bytes.Count(want, []byte("\n")))
	dir := t.TempDir()
	for _, name := range []string{"slot01_wave10_main.a", "regexp_word_class_atoms.a", "regexp_nonword_class_atoms.a", "regexp_word_boundary.a", "options_json.ts"} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if name == file {
			if strings.Count(string(data), old) != 1 {
				t.Fatal("mutant anchor drift")
			}
			data = []byte(strings.Replace(string(data), old, replacement, 1))
		}
		if name == "slot01_wave10_main.a" {
			data = []byte(strings.ReplaceAll(string(data), "./options_json.ts", "./options_json.a"))
		}
		if name == "options_json.ts" {
			name = "options_json.a"
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	mutant := run(t, "", slot01Build(t, filepath.Join(dir, "slot01_wave10_main.a")), cases, mode)
	if bytes.Equal(mutant, want) {
		t.Fatal("compiled mutant survived")
	}
	slot01MutantWitness(t, mutant, want, file)
}

// Not parallel: bounded external Go capture and sanitizer builds.
func TestSlot01Wave10Word(t *testing.T) {
	slot01Wave10Check(t, "github.com/system-inc/cohere/internal/lint/ecmascript/regexp.wordClassAtoms", "word", "regexp_word_class_atoms.a", "options.ignoreCase && options.unicode", "options.ignoreCase || options.unicode")
}

// Not parallel: bounded external Go capture and sanitizer builds.
func TestSlot01Wave10NonWord(t *testing.T) {
	slot01Wave10Check(t, "github.com/system-inc/cohere/internal/lint/ecmascript/regexp.nonWordClassAtoms", "nonword", "regexp_nonword_class_atoms.a", "high: 1114111", "high: 65535")
}

// Not parallel: bounded external Go capture and sanitizer builds.
func TestSlot01Wave10Boundary(t *testing.T) {
	slot01Wave10Check(t, "github.com/system-inc/cohere/internal/lint/ecmascript/regexp.wordBoundary", "boundary", "regexp_word_boundary.a", "if (negated)", "if (!negated)")
}

// Not parallel: bounded external Go capture and sanitizer builds.
func TestSlot01Wave10WordFresh(t *testing.T) {
	slot01Wave10Check(t, "github.com/system-inc/cohere/internal/lint/ecmascript/regexp.wordClassAtoms", "word", "regexp_word_class_atoms.a", "return atoms;\n}", "if (mutantCache.length === 0) { for (const atom of atoms) { mutantCache.push(atom); } }\n    return mutantCache;\n}\nconst mutantCache: RegexpClassAtom[] = [];")
}

// Not parallel: bounded external Go capture and sanitizer builds.
func TestSlot01Wave10NonWordFresh(t *testing.T) {
	slot01Wave10Check(t, "github.com/system-inc/cohere/internal/lint/ecmascript/regexp.nonWordClassAtoms", "nonword", "regexp_nonword_class_atoms.a", "return atoms;\n}", "if (mutantCache.length === 0) { for (const atom of atoms) { mutantCache.push(atom); } }\n    return mutantCache;\n}\nconst mutantCache: RegexpClassAtom[] = [];")
}

// Not parallel: bounded external Go capture and sanitizer builds.
func TestSlot01Wave10BoundaryForward(t *testing.T) {
	slot01Wave10Check(t, "github.com/system-inc/cohere/internal/lint/ecmascript/regexp.wordBoundary", "boundary", "regexp_word_boundary.a", "wordCharacters(options)", "wordCharacters({ ignoreCase: false, unicode: options.unicode, multiline: options.multiline, dotAll: options.dotAll })")
}

// Not parallel: bounded external Go capture and sanitizer builds.
func TestSlot01Wave10BoundaryCount(t *testing.T) {
	slot01Wave10Check(t, "github.com/system-inc/cohere/internal/lint/ecmascript/regexp.wordBoundary", "boundary", "regexp_word_boundary.a", "const word = wordCharacters(options);", "wordCharacters(options);\n    const word = wordCharacters(options);")
}
