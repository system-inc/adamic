package typeaware

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: component builds and sanitizer processes share the machine.
// This proves components only. It is not an end-to-end React rule gate.
func TestWave30ReactComponents(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE30_COMPONENT_ORACLE")
	if directory == "" {
		directory = t.TempDir()
	}
	if err = os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	for _, row := range []struct{ pkg, source, filter string }{
		{"internal/lint/rules/react", "wave_30_react_components_oracle_test.go", "^TestWave30ReactComponentOracle$"},
		{"internal/lint/ecmascript/high_level_intermediate_representation", "wave_30_memo_scope_oracle_test.go", "^TestWave30MemoScopeOracle$"},
	} {
		virtual := filepath.Join(repository, "cohere", row.pkg, "adamic_wave30_components_test.go")
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(repository, "stage1/cohere/typeaware/testdata", row.source)}})
		if err != nil {
			t.Fatal(err)
		}
		command := exec.Command("go", "test", "-overlay", h.write(filepath.Base(row.source)+".json", string(overlay)), "./"+row.pkg, "-run", row.filter, "-count=1", "-v")
		command.Dir = filepath.Join(repository, "cohere")
		command.Env = append(os.Environ(), "ADAMIC_WAVE30_COMPONENT_ORACLE="+directory)
		h.must(row.filter, command)
	}
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_30_react_components_suite.a")
	binary := filepath.Join(directory, "react-components")
	asan := binary + "-asan"
	h.must("components-build", exec.Command(stage0, "build", entry, "-o", binary))
	h.must("components-asan-build", exec.Command(stage0, "build", entry, "-o", asan, "--sanitize"))
	emitted := h.must("components-js-emit", exec.Command(stage0, "js", entry))
	javascript := h.write("react-components.js", string(emitted.stdout))
	runtimeDirectory := filepath.Join(directory, "node_modules/adamic")
	if err = os.MkdirAll(runtimeDirectory, 0755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(runtimeDirectory, "package.json"), []byte(`{"type":"module","exports":"./index.mjs"}`), 0600); err != nil {
		t.Fatal(err)
	}
	runtime, err := os.ReadFile(filepath.Join(repository, "oracle/adamic.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(runtimeDirectory, "index.mjs"), runtime, 0600); err != nil {
		t.Fatal(err)
	}

	for _, row := range []struct{ mode, file, from, to string }{
		{"refs", "react_refs_lattice.a", "if(a.kind === 3) { return true; }", "if(a.kind === 3) { return a.refId === b.refId; }"},
		{"purity", "react_purity_properties.a", "object.kind !== 1", "object.kind === 1"},
		{"memo", "react_manual_memo_scope.a", "kind === 0 && scope.pruned", "scope.pruned"},
	} {
		truth, err := os.ReadFile(filepath.Join(directory, row.mode+"-go.stdout"))
		if err != nil {
			t.Fatal(err)
		}
		node := h.must(row.mode+"-javascript", exec.Command("node", javascript, row.mode))
		if len(node.stderr) != 0 || !bytes.Equal(node.stdout, truth) {
			t.Fatalf("%s emitted JavaScript differs at byte %d: %s", row.mode, firstDifference(node.stdout, truth), node.stderr)
		}
		for _, runner := range []struct{ name, path string }{{row.mode, binary}, {row.mode + "-asan", asan}} {
			got := h.must(runner.name+"-native", exec.Command(runner.path, row.mode))
			if len(got.stderr) != 0 || !bytes.Equal(got.stdout, truth) {
				t.Fatalf("%s differs at byte %d: %s", runner.name, firstDifference(got.stdout, truth), got.stderr)
			}
		}
		scratch := filepath.Join(directory, row.mode+"-mutant")
		if err = os.MkdirAll(scratch, 0755); err != nil {
			t.Fatal(err)
		}
		for _, file := range []string{"wave_30_react_components_suite.a", "react_refs_lattice.a", "react_purity_properties.a", "react_manual_memo_scope.a"} {
			data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware", file))
			if err != nil {
				t.Fatal(err)
			}
			source := string(data)
			if file == row.file {
				if strings.Count(source, row.from) != 1 {
					t.Fatal("nonunique component mutant")
				}
				source = strings.Replace(source, row.from, row.to, 1)
			}
			if err = os.WriteFile(filepath.Join(scratch, file), []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
		}
		mutant := filepath.Join(scratch, "component-mutant")
		h.must(row.mode+"-mutant-build", exec.Command(stage0, "build", filepath.Join(scratch, "wave_30_react_components_suite.a"), "-o", mutant))
		got := h.must(row.mode+"-mutant-run", exec.Command(mutant, row.mode))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth) {
			t.Fatalf("%s component mutant survived", row.mode)
		}
		t.Logf("%s: %d exact bytes agree, sanitizer and emitted JavaScript agree; mutant exits 0 with empty stderr, Go catches byte %d", row.mode, len(truth), firstDifference(got.stdout, truth))
	}
}
