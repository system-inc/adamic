package typeaware

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: native archives and sanitizer subprocesses share the worker machine.
func TestWave11ListenerDeclarationsAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE_11_LISTENER_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_11_listener_probe.a")
	binary := h.build(stage0, "listeners", entry, archive, false)
	oracle := volumeOracle(h, "listeners-oracle", "oracle_wave_11_listeners.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	source := h.write("listener-source.ts", "export {};\n")
	truth := h.must("listeners-go", exec.Command(oracle, config, source))
	wave11CheckManifests(h, truth.stdout)
	got := h.must("listeners-native", exec.Command(binary))
	if len(truth.stderr) != 0 || len(got.stderr) != 0 || !bytes.Equal(truth.stdout, got.stdout) || bytes.Count(truth.stdout, []byte("\n")) != 15 {
		t.Fatalf("listener declarations differ from production Go registrations: Go %q native %q", truth.stdout, got.stdout)
	}
	// Node executes the original .a sources, independently of Adamic lowering.
	onNode := h.must("listeners-source-node", exec.Command("node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), entry))
	if len(onNode.stderr) != 0 || !bytes.Equal(onNode.stdout, truth.stdout) {
		t.Fatal("source Node listener declarations differ from Go")
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "listeners-asan", entry, sanitized, true)
	clean := h.must("listeners-asan-run", exec.Command(asan))
	if len(clean.stderr) != 0 || !bytes.Equal(clean.stdout, truth.stdout) {
		t.Fatal("sanitized listener declarations differ")
	}
	for _, change := range []struct{ name, file, from, to string }{
		{"wrong-kind", "wave_11_syntax_kinds.a", "export const CallExpression = 215;", "export const CallExpression = 216;"},
		{"missing-listener", "no_new_wrappers.a", "export const listenerKinds: readonly number[] = [NewExpression];", "export const listenerKinds: readonly number[] = [];"},
	} {
		mutantDirectory := filepath.Join(directory, change.name+"-source")
		if err := os.MkdirAll(mutantDirectory, 0755); err != nil {
			t.Fatal(err)
		}
		entries, err := os.ReadDir(filepath.Join(repository, "stage1/cohere/typeaware"))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() || (filepath.Ext(entry.Name()) != ".a" && filepath.Ext(entry.Name()) != ".ts") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware", entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			text := string(data)
			if entry.Name() == change.file {
				if strings.Count(text, change.from) != 1 {
					t.Fatalf("nonunique mutant %s", change.name)
				}
				text = strings.Replace(text, change.from, change.to, 1)
			}
			text = strings.ReplaceAll(text, "../../typescript/", filepath.Join(repository, "stage1/typescript")+"/")
			text = strings.ReplaceAll(text, "../lint/", filepath.Join(repository, "stage1/cohere/lint")+"/")
			if err := os.WriteFile(filepath.Join(mutantDirectory, entry.Name()), []byte(text), 0644); err != nil {
				t.Fatal(err)
			}
		}
		mutant := h.build(stage0, change.name, filepath.Join(mutantDirectory, "wave_11_listener_probe.a"), archive, false)
		changed := h.must(change.name+"-run", exec.Command(mutant))
		if len(changed.stderr) != 0 || bytes.Equal(changed.stdout, truth.stdout) {
			t.Fatalf("%s survived", change.name)
		}
		t.Logf("%s: exit 0, empty stderr, Go listener comparison catches byte %d", change.name, firstDifference(changed.stdout, truth.stdout))
	}
	t.Log("15 numeric listener sets match production Go defaults; ASAN/UBSAN/LSAN clean")
}
