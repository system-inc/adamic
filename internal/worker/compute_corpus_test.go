package worker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
)

// Products hold compiled inputs only. Both stock-runner units replay all 600
// observations afresh, without workerd or any environment opt-in.
func computeCorpusProduct(t *testing.T, mutant, build bool) string {
	t.Helper()
	inputs, err := buildcache.GoBuildInInputs("", "adamic", "./cmd/adamic", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	inputs.Name = fmt.Sprintf("worker compute corpus sieve-mutant=%t", mutant)
	inputs.Files = slices.Clone(inputs.Files)
	inputs.Flags = slices.Clone(inputs.Flags)
	inputs.Toolchain = slices.Clone(inputs.Toolchain)
	inputs.Files = append(inputs.Files, "internal/worker/compute_corpus_test.go", "workers/compute/handler.a", "workers/compute/native-driver.a")
	inputs.Flags = append(inputs.Flags, fmt.Sprintf("sieve-mutant=%t", mutant))
	request, err := native.RuntimeRequestFor(native.Options{Release: true})
	if err != nil {
		t.Fatal(err)
	}
	inputs.Flags = append(inputs.Flags, request.Portable...)
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool(request.Compiler, "--version"), buildcache.Tool(request.Archiver, "--version"))
	if !build {
		return findWorkerProduct(t, inputs)
	}
	compiler := workerCompilerProduct(t, nil, false)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return buildcache.Product(t, inputs, func(directory string) error {
		source, err := os.ReadFile(filepath.Join(root, "workers/compute/handler.a"))
		if err != nil {
			return err
		}
		if mutant {
			before := "for (let value = 2; value <= limit; value += 1)"
			if strings.Count(string(source), before) != 1 {
				return fmt.Errorf("sieve mutant target changed")
			}
			source = []byte(strings.Replace(string(source), before, "for (let value = 2; value < limit; value += 1)", 1))
		}
		entry := filepath.Join(directory, "handler.a")
		if err := os.WriteFile(entry, source, 0600); err != nil {
			return err
		}
		driver, err := os.ReadFile(filepath.Join(root, "workers/compute/native-driver.a"))
		if err != nil {
			return err
		}
		driverPath := filepath.Join(directory, "native-driver.a")
		if err := os.WriteFile(driverPath, driver, 0600); err != nil {
			return err
		}
		for _, args := range [][]string{{"worker", entry, "--out", filepath.Join(directory, "worker")}, {"build", driverPath, "-o", filepath.Join(directory, "native")}} {
			command := exec.CommandContext(t.Context(), compiler, args...)
			command.Dir = root
			if output, err := command.CombinedOutput(); err != nil {
				return fmt.Errorf("compute corpus build: %w\n%s", err, output)
			}
		}
		return nil
	})
}

// Not parallel: compiles the shared native and JavaScript corpus inputs ahead.
func TestProduct_WorkerComputeCorpus(t *testing.T) { computeCorpusProduct(t, false, true) }

// Not parallel: compiles the one-bound sieve mutant ahead of its behavior unit.
func TestProduct_WorkerComputeSieveMutant(t *testing.T) { computeCorpusProduct(t, true, true) }

func runComputeCorpus(t *testing.T, mutant bool) {
	t.Helper()
	directory := computeCorpusProduct(t, mutant, false)
	mode := "good"
	if mutant {
		mode = "mutant"
	}
	command := exec.CommandContext(t.Context(), "node", "--disable-warning=ExperimentalWarning", "compute-corpus.mjs", directory, t.TempDir(), mode)
	command.Env = append(os.Environ(), "ADAMIC_JSON_TYPES="+jsonTypes(t))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("compute corpus: %v\n%s", err, output)
	} else {
		t.Logf("%s", output)
	}
}
func TestWorkerComputeCorpus(t *testing.T)      { t.Parallel(); runComputeCorpus(t, false) }
func TestWorkerComputeSieveMutant(t *testing.T) { t.Parallel(); runComputeCorpus(t, true) }
