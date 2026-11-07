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
func computeCorpusProduct(test *testing.T, mutant, build bool) string {
	test.Helper()
	inputs, goBuildInInputsError := buildcache.GoBuildInInputs("", "adamic", "./cmd/adamic", nil, nil)
	if goBuildInInputsError != nil {
		test.Fatal(goBuildInInputsError)
	}
	inputs.Name = fmt.Sprintf("worker compute corpus sieve-mutant=%t", mutant)
	inputs.Files = slices.Clone(inputs.Files)
	inputs.Flags = slices.Clone(inputs.Flags)
	inputs.Toolchain = slices.Clone(inputs.Toolchain)
	inputs.Files = append(
		inputs.Files,
		"internal/worker/compute_corpus_test.go",
		"workers/compute/handler.a",
		"workers/compute/native-driver.a",
	)
	inputs.Flags = append(inputs.Flags, fmt.Sprintf("sieve-mutant=%t", mutant))
	request, runtimeRequestForError := native.RuntimeRequestFor(native.Options{Release: true})
	if runtimeRequestForError != nil {
		test.Fatal(runtimeRequestForError)
	}
	inputs.Flags = append(inputs.Flags, request.Portable...)
	inputs.Toolchain = append(
		inputs.Toolchain,
		buildcache.Tool(request.Compiler, "--version"),
		buildcache.Tool(request.Archiver, "--version"),
	)
	if !build {
		return findWorkerProduct(test, inputs)
	}
	compiler := workerCompilerProduct(test, nil, false)
	root, absolutePathError := filepath.Abs("../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		source, readError := os.ReadFile(filepath.Join(root, "workers/compute/handler.a"))
		if readError != nil {
			return readError
		}
		if mutant {
			before := "for (let value = 2; value <= limit; value += 1)"
			if strings.Count(string(source), before) != 1 {
				return fmt.Errorf("sieve mutant target changed")
			}
			source = []byte(
				strings.Replace(string(source), before, "for (let value = 2; value < limit; value += 1)", 1),
			)
		}
		entry := filepath.Join(directory, "handler.a")
		if writeError := os.WriteFile(entry, source, 0600); writeError != nil {
			return writeError
		}
		driver, readError := os.ReadFile(filepath.Join(root, "workers/compute/native-driver.a"))
		if readError != nil {
			return readError
		}
		driverPath := filepath.Join(directory, "native-driver.a")
		if writeError := os.WriteFile(driverPath, driver, 0600); writeError != nil {
			return writeError
		}
		for _, args := range [][]string{
			{"worker", entry, "--out", filepath.Join(directory, "worker")},
			{"build", driverPath, "-o", filepath.Join(directory, "native")},
		} {
			command := exec.CommandContext(test.Context(), compiler, args...)
			command.Dir = root
			if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
				return fmt.Errorf("compute corpus build: %w\n%s", combinedOutputError, output)
			}
		}
		return nil
	})
}

// Not parallel: compiles the shared native and JavaScript corpus inputs ahead.
func TestProduct_WorkerComputeCorpus(test *testing.T) { computeCorpusProduct(test, false, true) }

// Not parallel: compiles the one-bound sieve mutant ahead of its behavior unit.
func TestProduct_WorkerComputeSieveMutant(test *testing.T) { computeCorpusProduct(test, true, true) }

func runComputeCorpus(test *testing.T, mutant bool) {
	test.Helper()
	directory := computeCorpusProduct(test, mutant, false)
	mode := "good"
	if mutant {
		mode = "mutant"
	}
	command := exec.CommandContext(
		test.Context(),
		"node",
		"--disable-warning=ExperimentalWarning",
		"compute-corpus.mjs",
		directory,
		test.TempDir(),
		mode,
	)
	command.Env = append(os.Environ(), "ADAMIC_JSON_TYPES="+jsonTypes(test))
	if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("compute corpus: %v\n%s", combinedOutputError, output)
	} else {
		test.Logf("%s", output)
	}
}
func TestWorkerComputeCorpus(test *testing.T)      { test.Parallel(); runComputeCorpus(test, false) }
func TestWorkerComputeSieveMutant(test *testing.T) { test.Parallel(); runComputeCorpus(test, true) }
