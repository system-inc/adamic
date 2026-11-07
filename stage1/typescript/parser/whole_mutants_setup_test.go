package parser

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
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
func TestMain(suite *testing.M) {
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
	os.Exit(suite.Run())
}

// wholeMutantsTreeScanner is the scanner's absolute path as another checkout wrote it into a mutant copy: copyPort
// names this checkout's scanner, and a copy that is a product fetched from Workshop names Workshop's, so the hash
// reads either as the tree spells it (#t37sw0f; the copy's own bytes are #tqrqx60's).
var wholeMutantsTreeScanner = regexp.MustCompile(`/[^'"\s]*/stage1/typescript/scanner/scanner\.ts`)

func wholeMutantsSourceHash(test *testing.T, directory string) string {
	test.Helper()
	scanner, absolutePathError := filepath.Abs("../scanner/scanner.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	var sources strings.Builder
	for _, name := range portFiles {
		data, readError := os.ReadFile(filepath.Join(directory, name))
		if readError != nil {
			test.Fatal(readError)
		}
		normalized := wholeMutantsTreeScanner.ReplaceAllString(
			strings.ReplaceAll(string(data), scanner, "../scanner/scanner.ts"),
			"../scanner/scanner.ts",
		)
		// A copy with a scanner of its own (a scanner mutant) imports it by the copy's temporary path.
		normalized = strings.ReplaceAll(normalized, directory, "$PORT")
		fmt.Fprintf(&sources, "%s %d\n%s", name, len(normalized), normalized)
	}
	// That scanner's bytes, including its changed site, are inputs like the port's.
	for _, name := range []string{"scanner.ts", "characters.ts", "tokens.ts"} {
		data, readError := os.ReadFile(filepath.Join(directory, "scanner", name))
		if errors.Is(readError, fs.ErrNotExist) {
			continue
		}
		if readError != nil {
			test.Fatal(readError)
		}
		fmt.Fprintf(&sources, "scanner/%s %d\n%s", name, len(data), data)
	}
	return fmt.Sprintf("source=%x", sha256.Sum256([]byte(sources.String())))
}

func TestWholeMutants_Setup(test *testing.T) {
	test.Parallel()
	wholeMutantsPrepare(test)
}

func wholeMutantsPrepare(test *testing.T) {
	test.Helper()
	wholeMutantsPrepareOnce.Do(func() {
		// Setup has no deadline; Loom's kill still covers the whole unit.
		if wholeMutantsPrepared != nil {
			return
		}
		wholeMutantsBuildInputs(test)
	})
	wholeMutantsCheckInputs(test)
}

func wholeMutantsBuildInputs(test *testing.T) {
	test.Helper()
	started := time.Now()
	directory, absolutePathError := filepath.Abs(".")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	inputs := &wholeMutantsPreparedInputs{
		Oracle:   wholeMutantBuildOracleProduct(test),
		Binaries: make(map[string]string, testWholeMutantsShards+1),
	}
	directories := []string{directory}
	for _, mutation := range wholeMutantCases() {
		directories = append(directories, copyPort(test, mutation.file, mutation.from, mutation.to))
	}
	binaries := make([]string, len(directories))
	var builds sync.WaitGroup
	for index, source := range directories {
		builds.Add(1)
		go func() {
			defer builds.Done()
			binaries[index] = wholeMutantBuildPortProduct(test, source, true)
		}()
	}
	builds.Wait()
	for index, source := range directories {
		if binaries[index] == "" {
			test.Fatal("setup product build did not complete")
		}
		inputs.Binaries[wholeMutantsSourceHash(test, source)] = binaries[index]
	}
	wholeMutantsPrepared = inputs
	wholeMutantsCheckInputs(test)

	test.Logf("setup: %.3fs, all hash-addressed inputs ready", time.Since(started).Seconds())
}

func wholeMutantsCheckInputs(test *testing.T) {
	test.Helper()
	if wholeMutantsPrepared == nil || len(wholeMutantsPrepared.Binaries) != testWholeMutantsShards+1 {
		test.Fatal("shared setup must finish before a whole-mutant leaf runs")
	}
	for _, path := range append([]string{wholeMutantsPrepared.Oracle}, wholeMutantsBinaryPaths()...) {
		if _, statError := os.Stat(path); statError != nil {
			test.Fatalf("prepared input %s: %v", path, statError)
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

func wholeMutantOracleProduct(test *testing.T) string {
	test.Helper()
	wholeMutantsCheckInputs(test)
	return wholeMutantsPrepared.Oracle
}

func wholeMutantPortProduct(test *testing.T, directory string, sanitize bool) string {
	test.Helper()
	wholeMutantsCheckInputs(test)
	if !sanitize {
		test.Fatal("whole-mutant inputs must retain sanitizers")
	}
	binary, ok := wholeMutantsPrepared.Binaries[wholeMutantsSourceHash(test, directory)]
	if !ok {
		test.Fatal("mutant binary was not prepared before the shard deadline")
	}
	return binary
}

func wholeMutantsRunShard(test *testing.T, index int) {
	test.Helper()
	if !wholeMutantShardSelected(test, index) {
		test.Skip("assigned to another ADAMIC_TEST_SHARD")
	}
	wholeMutantsPrepare(test)
	if os.Getenv(wholeMutantsCaseChildEnv) == "1" {
		wholeMutantsShard(test, index)
		return
	}
	environment := wholeMutantsChildEnvironment(test)
	// Shared setup is already ready: only this case is inside the deadline.
	timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := wholeMutantCommandContext(
		timeoutContext,
		os.Args[0],
		"-test.run=^"+test.Name()+"$",
		"-test.v",
		"-test.timeout=90s",
	)
	command.Env = append(environment, wholeMutantsCaseChildEnv+"=1")
	// Read through childguard.Capture for wholeMutantCommandContext's reason, stdout then stderr.
	stdout, stderr, captureError := childguard.Capture(timeoutContext, command)
	output := append(stdout, stderr...)
	if timeoutContext.Err() != nil {
		test.Fatalf("cooked: %s exceeded 90 seconds", test.Name())
	}
	if captureError != nil {
		test.Fatalf("%s: %v\n%s", test.Name(), captureError, output)
	}
	if strings.Contains(string(output), "build typescript-parser-") {
		test.Fatalf("leaf attempted a product build after its deadline started\n%s", output)
	}
	test.Logf("%s", output)
}

// Transient coordination for a child process, not another build cache.
func wholeMutantsChildEnvironment(test *testing.T) []string {
	test.Helper()
	wholeMutantsPrepare(test)
	data, encodeError := json.Marshal(wholeMutantsPrepared)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	descriptor := filepath.Join(test.TempDir(), "inputs.json")
	if writeError := os.WriteFile(descriptor, data, 0600); writeError != nil {
		test.Fatal(writeError)
	}
	var environment []string
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, wholeMutantsInputsEnv+"=") {
			environment = append(environment, entry)
		}
	}
	return append(environment, wholeMutantsInputsEnv+"="+descriptor)
}
