package typeaware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: native and sanitizer builds share the machine's compiler resources.
func TestWave23NumericListeners(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE23_LISTENER_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	oracle := volumeOracle(h, "listeners-oracle", "oracle_wave_23_listeners.go")
	control := h.write("control.a", "function Widget(){return 1;}\nexport {};\n")
	h.write("prelude.d.ts", "")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","noEmit":true,"lib":["ESNext"]},"files":["prelude.d.ts"]}`)
	manifest := h.write("manifest", control+"\n")
	truth := h.must("go-listeners", exec.Command(oracle, config, manifest, "--listener-kinds"))
	rows := strings.Split(strings.TrimSpace(string(truth.stdout)), "\n")
	if len(rows) != 18 {
		t.Fatalf("expected eighteen production rule registrations, got %d", len(rows))
	}
	for _, row := range rows {
		parts := strings.Split(row, "\t")
		if len(parts) != 2 {
			t.Fatal("bad Go listener record", row)
		}
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/listeners-wave23", parts[0], "rule.json"))
		if err != nil {
			t.Fatal(err)
		}
		var declaration struct {
			Name  string `json:"name"`
			Kinds []int  `json:"kinds"`
		}
		if err := json.Unmarshal(data, &declaration); err != nil {
			t.Fatal(err)
		}
		var numbers []string
		for _, kind := range declaration.Kinds {
			numbers = append(numbers, fmt.Sprint(kind))
		}
		got := declaration.Name + "\t" + strings.Join(numbers, ",")
		if got != row {
			t.Fatalf("rule.json kinds differ from Go: got %s, want %s", got, row)
		}
	}
	t.Log("eighteen rule.json name/kinds declarations match the independent Go registry")
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_23_listener_probe.a")
	for _, sanitize := range []bool{false, true} {
		name := "native"
		if sanitize {
			name = "native-asan"
		}
		binary := filepath.Join(directory, name)
		args := []string{"build", entry, "-o", binary}
		if sanitize {
			args = append(args, "--sanitize")
		}
		h.must(name+"-build", exec.Command(stage0, args...))
		got := h.must(name+"-run", exec.Command(binary))
		if len(got.stderr) != 0 || !bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s listener bytes differ at %d: %s", name, firstDifference(got.stdout, truth.stdout), got.stderr)
		}
	}
	runtimeDirectory := filepath.Join(directory, "node_modules/adamic")
	if err := os.MkdirAll(runtimeDirectory, 0755); err != nil {
		t.Fatal(err)
	}
	runtime, err := os.ReadFile(filepath.Join(repository, "oracle/adamic.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runtimeDirectory, "index.mjs"), runtime, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runtimeDirectory, "package.json"), []byte(`{"type":"module","exports":"./index.mjs"}`), 0644); err != nil {
		t.Fatal(err)
	}
	js := h.must("javascript-build", exec.Command(stage0, "js", entry))
	javascript := h.write("listeners.mjs", string(js.stdout))
	node := h.must("node-run", exec.Command("node", javascript))
	if len(node.stderr) != 0 || !bytes.Equal(node.stdout, truth.stdout) {
		t.Fatal("emitted JavaScript listeners differ from Go")
	}
	metadata, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/wave_23_listener_kinds.a"))
	if err != nil {
		t.Fatal(err)
	}
	probe, err := os.ReadFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	for i, row := range rows {
		parts := strings.Split(row, "\t")
		if len(parts) != 2 {
			t.Fatal("bad Go listener record", row)
		}
		kinds := strings.Split(parts[1], ",")
		from := fmt.Sprintf("{ name: '%s', kinds: [%s", parts[0], kinds[0])
		// Drop the actual first registration. A valid but wrong numeric kind must survive compilation.
		to := fmt.Sprintf("{ name: '%s', kinds: [0", parts[0])
		if strings.Count(string(metadata), from) != 1 {
			t.Fatal("nonunique metadata mutant", parts[0])
		}
		mutantDirectory := filepath.Join(directory, fmt.Sprintf("mutant-%02d", i))
		if err := os.MkdirAll(mutantDirectory, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(mutantDirectory, "wave_23_listener_kinds.a"), []byte(strings.Replace(string(metadata), from, to, 1)), 0644); err != nil {
			t.Fatal(err)
		}
		mutantEntry := filepath.Join(mutantDirectory, "wave_23_listener_probe.a")
		if err := os.WriteFile(mutantEntry, probe, 0644); err != nil {
			t.Fatal(err)
		}
		binary := filepath.Join(mutantDirectory, "native")
		h.must(fmt.Sprintf("mutant-%02d-build", i), exec.Command(stage0, "build", mutantEntry, "-o", binary))
		got := h.must(fmt.Sprintf("mutant-%02d-run", i), exec.Command(binary))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("numeric listener mutant survived", parts[0])
		}
		t.Logf("%s listener mutant: exit 0, empty stderr, Go byte oracle catches byte %d", parts[0], firstDifference(got.stdout, truth.stdout))
	}
	t.Logf("eighteen production registrations: %d identical Go/native/sanitized/Node bytes", len(truth.stdout))
}
