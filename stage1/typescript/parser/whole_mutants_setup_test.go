package parser

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	wholeMutantsInputsEnv    = "ADAMIC_WHOLE_MUTANTS_INPUTS"
	wholeMutantsCaseChildEnv = "ADAMIC_WHOLE_MUTANTS_CASE_CHILD"
)

type wholeMutantsPreparedInputs struct {
	Oracle   string
	Binaries map[string]string
}

// Products are prepared once per process before a shard starts its own clock.
var wholeMutantsPrepared *wholeMutantsPreparedInputs
var wholeMutantsPrepareOnce sync.Once

// Children receive the parent's ready products, never a setup-test prerequisite.
func TestMain(m *testing.M) {
	if descriptor := os.Getenv(wholeMutantsInputsEnv); descriptor != "" {
		data, err := os.ReadFile(descriptor)
		if err == nil {
			err = json.Unmarshal(data, &wholeMutantsPrepared)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}

func wholeMutantsSourceHash(t *testing.T, directory string) string {
	t.Helper()
	scanner, err := filepath.Abs("../scanner/scanner.ts")
	if err != nil {
		t.Fatal(err)
	}
	var sources strings.Builder
	for _, name := range portFiles {
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		normalized := strings.ReplaceAll(string(data), scanner, "../scanner/scanner.ts")
		fmt.Fprintf(&sources, "%s %d\n%s", name, len(normalized), normalized)
	}
	return fmt.Sprintf("source=%x", sha256.Sum256([]byte(sources.String())))
}

func TestWholeMutants_Setup(t *testing.T) {
	t.Parallel()
	wholeMutantsPrepare(t)
}

func wholeMutantsPrepare(t *testing.T) {
	t.Helper()
	wholeMutantsPrepareOnce.Do(func() {
		// Setup has no deadline; Loom's kill still covers the whole unit.
		if wholeMutantsPrepared != nil {
			return
		}
		wholeMutantsBuildInputs(t)
	})
	wholeMutantsCheckInputs(t)
}

func wholeMutantsBuildInputs(t *testing.T) {
	t.Helper()
	started := time.Now()
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	inputs := &wholeMutantsPreparedInputs{
		Oracle:   wholeMutantBuildOracleProduct(t),
		Binaries: make(map[string]string, testWholeMutantsShards+1),
	}
	directories := []string{directory}
	for _, mutation := range wholeMutantCases() {
		directories = append(directories, copyPort(t, mutation.file, mutation.from, mutation.to))
	}
	binaries := make([]string, len(directories))
	var builds sync.WaitGroup
	for i, source := range directories {
		builds.Add(1)
		go func() {
			defer builds.Done()
			binaries[i] = wholeMutantBuildPortProduct(t, source, true)
		}()
	}
	builds.Wait()
	for i, source := range directories {
		if binaries[i] == "" {
			t.Fatal("setup product build did not complete")
		}
		inputs.Binaries[wholeMutantsSourceHash(t, source)] = binaries[i]
	}
	wholeMutantsPrepared = inputs
	wholeMutantsCheckInputs(t)

	t.Logf("setup: %.3fs, all hash-addressed inputs ready", time.Since(started).Seconds())
}

func wholeMutantsCheckInputs(t *testing.T) {
	t.Helper()
	if wholeMutantsPrepared == nil || len(wholeMutantsPrepared.Binaries) != testWholeMutantsShards+1 {
		t.Fatal("shared setup must finish before a whole-mutant leaf runs")
	}
	for _, path := range append([]string{wholeMutantsPrepared.Oracle}, wholeMutantsBinaryPaths()...) {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("prepared input %s: %v", path, err)
		}
	}
}

func wholeMutantsBinaryPaths() []string {
	var paths []string
	for _, path := range wholeMutantsPrepared.Binaries {
		paths = append(paths, path)
	}
	return paths
}

func wholeMutantOracleProduct(t *testing.T) string {
	t.Helper()
	wholeMutantsCheckInputs(t)
	return wholeMutantsPrepared.Oracle
}

func wholeMutantPortProduct(t *testing.T, directory string, sanitize bool) string {
	t.Helper()
	wholeMutantsCheckInputs(t)
	if !sanitize {
		t.Fatal("whole-mutant inputs must retain sanitizers")
	}
	binary, ok := wholeMutantsPrepared.Binaries[wholeMutantsSourceHash(t, directory)]
	if !ok {
		t.Fatal("mutant binary was not prepared before the shard deadline")
	}
	return binary
}

func wholeMutantsRunShard(t *testing.T, index int) {
	t.Helper()
	if !wholeMutantShardSelected(t, index) {
		t.Skip("assigned to another ADAMIC_TEST_SHARD")
	}
	wholeMutantsPrepare(t)
	if os.Getenv(wholeMutantsCaseChildEnv) == "1" {
		wholeMutantsShard(t, index)
		return
	}
	environment := wholeMutantsChildEnvironment(t)
	// Shared setup is already ready: only this case is inside the deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := wholeMutantCommandContext(ctx, os.Args[0], "-test.run=^"+t.Name()+"$", "-test.v", "-test.timeout=90s")
	command.Env = append(environment, wholeMutantsCaseChildEnv+"=1")
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("cooked: %s exceeded 90 seconds", t.Name())
	}
	if err != nil {
		t.Fatalf("%s: %v\n%s", t.Name(), err, output)
	}
	if strings.Contains(string(output), "build typescript-parser-") {
		t.Fatalf("leaf attempted a product build after its deadline started\n%s", output)
	}
	t.Logf("%s", output)
}

// Transient coordination for a child process, not another build cache.
func wholeMutantsChildEnvironment(t *testing.T) []string {
	t.Helper()
	wholeMutantsPrepare(t)
	data, err := json.Marshal(wholeMutantsPrepared)
	if err != nil {
		t.Fatal(err)
	}
	descriptor := filepath.Join(t.TempDir(), "inputs.json")
	if err := os.WriteFile(descriptor, data, 0600); err != nil {
		t.Fatal(err)
	}
	var environment []string
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, wholeMutantsInputsEnv+"=") {
			environment = append(environment, entry)
		}
	}
	return append(environment, wholeMutantsInputsEnv+"="+descriptor)
}
