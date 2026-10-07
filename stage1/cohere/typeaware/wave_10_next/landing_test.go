package wave10next_test

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

// This gate reuses stage 0 and bridge archives produced by the process gate.
// Not parallel: its normal, sanitized and mutant executables share an artifact directory.
func TestLandingNativeRulesAndMutants(t *testing.T) {
	directory := os.Getenv("ADAMIC_WAVE10_LANDING_ARTIFACTS")
	fixtures := os.Getenv("ADAMIC_WAVE10_LANDING_FIXTURES")
	if directory == "" || fixtures == "" {
		t.Skip("set landing artifacts and prepared 92-control fixture paths")
	}
	repository, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	owned := filepath.Join(repository, "stage1/cohere/typeaware/wave_10_next")
	stage0 := filepath.Join(directory, "adamic")
	archive := filepath.Join(directory, "checker.a")
	sanitized := filepath.Join(directory, "checker-asan.a")
	config := filepath.Join(fixtures, "tsconfig.json")
	manifest := filepath.Join(fixtures, "all-controls.manifest")
	for _, name := range []string{"process", "blocking"} {
		virtual := filepath.Join(repository, "cohere/adamic_wave10_"+name+"_oracle.go")
		data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(owned, "testdata/oracle_"+name+".go")}})
		if err != nil {
			t.Fatal(err)
		}
		oracle := filepath.Join(directory, "landing-"+name+"-oracle")
		command := exec.Command("go", "build", "-overlay", h.write("landing-"+name+"-oracle.json", string(data)), "-o", oracle, virtual)
		command.Dir = filepath.Join(repository, "cohere")
		h.must(name+"-oracle-build", command)
		binary := h.build(stage0, "landing-"+name, filepath.Join(owned, name+"_suite.a"), archive, false)
		truth := h.compare(name+"-controls92", oracle, binary, config, manifest)
		asan := h.build(stage0, "landing-"+name+"-asan", filepath.Join(owned, name+"_suite.a"), sanitized, true)
		h.compare(name+"-controls92-asan", oracle, asan, config, manifest)
		mutantDirectory := filepath.Join(directory, "landing-"+name+"-mutant-source")
		if err := os.MkdirAll(mutantDirectory, 0755); err != nil {
			t.Fatal(err)
		}
		sources, err := filepath.Glob(filepath.Join(owned, "*.a"))
		if err != nil {
			t.Fatal(err)
		}
		for _, source := range sources {
			data, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			text := strings.ReplaceAll(string(data), "'../", "'"+filepath.Join(repository, "stage1/cohere/typeaware")+"/")
			text = strings.ReplaceAll(text, filepath.Join(repository, "stage1/cohere/typeaware")+"/../../typescript/", filepath.Join(repository, "stage1/typescript")+"/")
			from := ""
			to := ""
			if name == "process" && filepath.Base(source) == "no_process_exit_after_output.a" {
				from = "next.push(chain);"
				to = "return [];"
			}
			if name == "blocking" && filepath.Base(source) == "require_blocking_standard_streams.a" {
				from = "this.ordered = this.held().canBlock.has(this.rules.path);"
				to = "this.ordered = false;"
			}
			if from != "" {
				if strings.Count(text, from) != 1 {
					t.Fatal("nonunique native rule mutant")
				}
				text = strings.Replace(text, from, to, 1)
			}
			if err := os.WriteFile(filepath.Join(mutantDirectory, filepath.Base(source)), []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
		}
		mutant := h.build(stage0, "landing-"+name+"-mutant", filepath.Join(mutantDirectory, name+"_suite.a"), archive, false)
		got := h.must(name+"-native-mutant-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("native rule mutant survived or failed outside comparison")
		}
		t.Logf("%s native mutant exits 0, empty stderr, Go byte comparison catches byte %d", name, firstDifference(got.stdout, truth.stdout))
		for _, corpus := range []struct{ name, config, manifest string }{{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE10_COMPILER_MANIFEST")}, {"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE10_REPOSITORY_MANIFEST")}} {
			if corpus.manifest == "" {
				t.Fatal("landing requires both frozen corpora")
			}
			h.compare(name+"-"+corpus.name, oracle, binary, corpus.config, corpus.manifest)
			h.compare(name+"-"+corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		}
		// Each timing includes loading and serialization; alternate process order.
		for _, dataset := range []struct{ name, config, manifest string }{{"controls", config, manifest}, {"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE10_COMPILER_MANIFEST")}, {"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE10_REPOSITORY_MANIFEST")}} {
			for round := 0; round < 3; round++ {
				var native, goRun result
				if round%2 == 0 {
					native = h.must(fmt.Sprintf("%s-%s-native-bench-%d", name, dataset.name, round), exec.Command(binary, dataset.config, dataset.manifest))
					goRun = h.must(fmt.Sprintf("%s-%s-go-bench-%d", name, dataset.name, round), exec.Command(oracle, dataset.config, dataset.manifest))
				} else {
					goRun = h.must(fmt.Sprintf("%s-%s-go-bench-%d", name, dataset.name, round), exec.Command(oracle, dataset.config, dataset.manifest))
					native = h.must(fmt.Sprintf("%s-%s-native-bench-%d", name, dataset.name, round), exec.Command(binary, dataset.config, dataset.manifest))
				}
				if len(native.stderr) != 0 || !bytes.Equal(native.stdout, goRun.stdout) {
					t.Fatal("timing run disagrees")
				}
				t.Logf("timing %s %s round %d native_ns=%d go_ns=%d", name, dataset.name, round, native.elapsed.Nanoseconds(), goRun.elapsed.Nanoseconds())
			}
		}
	}
}
