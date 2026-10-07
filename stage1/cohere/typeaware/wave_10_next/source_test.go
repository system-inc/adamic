package wave10next_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Source execution is useful for the port's logic. It deliberately does not
// claim native equivalence, emitted-JavaScript parity, sanitizer safety or timing.
func TestSourceAgreementAndMutants(t *testing.T) {
	directory := os.Getenv("ADAMIC_WAVE10_SOURCE_ARTIFACTS")
	records := os.Getenv("ADAMIC_WAVE10_SOURCE_RECORDS")
	if directory == "" || records == "" {
		t.Skip("set both source artifact and raw-fact paths for the independent source comparison")
	}
	repository, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	owned := filepath.Join(repository, "stage1/cohere/typeaware/wave_10_next")
	config := filepath.Join(directory, "tsconfig.json")
	manifest := filepath.Join(directory, "all-controls.manifest")
	for _, name := range []string{"process", "blocking"} {
		virtual := filepath.Join(repository, "cohere/adamic_wave10_"+name+"_oracle.go")
		data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(owned, "testdata/oracle_"+name+".go")}})
		if err != nil {
			t.Fatal(err)
		}
		oracle := filepath.Join(directory, "source-"+name+"-oracle")
		command := exec.Command("go", "build", "-overlay", h.write("source-"+name+"-overlay.json", string(data)), "-o", oracle, virtual)
		command.Dir = filepath.Join(repository, "cohere")
		h.must(name+"-oracle-build", command)
		truth := h.must(name+"-oracle", exec.Command(oracle, config, manifest))
		node := func(mutant string) *exec.Cmd {
			command := exec.Command("node", "--max-old-space-size=8192", "--import", filepath.Join(owned, "testdata/node_loader.mjs"), filepath.Join(owned, name+"_suite.a"), config, manifest)
			command.Env = append(os.Environ(), "NODE_NO_WARNINGS=1", "ADAMIC_CONTEXT_RECORDS="+records, "ADAMIC_SOURCE_MUTANT="+mutant)
			return command
		}
		got := h.must(name+"-source", node(""))
		if len(got.stderr) != 0 || !bytes.Equal(truth.stdout, got.stdout) {
			t.Fatalf("%s source mismatch at byte %d: %s", name, firstDifference(truth.stdout, got.stdout), got.stderr)
		}
		mutant := h.must(name+"-source-mutant", node(name))
		if len(mutant.stderr) != 0 {
			t.Fatalf("mutant failed outside comparison: %s", mutant.stderr)
		}
		if bytes.Equal(truth.stdout, mutant.stdout) {
			t.Fatal("source mutant survived")
		}
		t.Logf("%s source: %d identical bytes; mutant exits 0, comparison catches byte %d", name, len(truth.stdout), firstDifference(truth.stdout, mutant.stdout))
	}
}
