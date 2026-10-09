package parser

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	wholeMutantsInputsEnv     = "ADAMIC_WHOLE_MUTANTS_INPUTS"
	wholeMutantsSetupChildEnv = "ADAMIC_WHOLE_MUTANTS_SETUP_CHILD"
	wholeMutantsCaseChildEnv  = "ADAMIC_WHOLE_MUTANTS_CASE_CHILD"
)

type wholeMutantsPreparedInputs struct {
	Oracle   string
	Binaries map[string]string
}

// Assigned eagerly before m.Run. A leaf can only read ready inputs; it never
// runs a builder, sync.Once, lazy initialization or a shared-cache lock.
var wholeMutantsPrepared *wholeMutantsPreparedInputs

// TestMain prepares our selected leaves before testing starts their deadlines.
// An independently selected leaf runs the same named setup test in a bounded
// child. The descriptor is transient coordination, not a second product cache.
func TestMain(m *testing.M) {
	flag.Parse()
	if os.Getenv(wholeMutantsSetupChildEnv) == "1" {
		os.Exit(m.Run())
	}
	if descriptor := os.Getenv(wholeMutantsInputsEnv); descriptor != "" {
		data, err := os.ReadFile(descriptor)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err = json.Unmarshal(data, &wholeMutantsPrepared); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(m.Run())
	}
	// Listing and unrelated selections never fetch these products.
	if flag.Lookup("test.list").Value.String() != "" {
		os.Exit(m.Run())
	}
	selection, err := regexp.Compile(flag.Lookup("test.run").Value.String())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	needsInputs := selection.MatchString("TestWholeMutantsRejectsSurvivor")
	for i := 0; i < testWholeMutantsShards; i++ {
		needsInputs = needsInputs || selection.MatchString(fmt.Sprintf("TestWholeMutants_%03d", i))
	}
	if !needsInputs {
		os.Exit(m.Run())
	}
	directory, err := os.MkdirTemp("", "whole-mutants-inputs-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	descriptor := filepath.Join(directory, "inputs.json")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	command := wholeMutantCommandContext(ctx, os.Args[0], "-test.run=^TestWholeMutants_Setup$", "-test.v", "-test.timeout=90s")
	command.Env = append(os.Environ(), wholeMutantsSetupChildEnv+"=1", wholeMutantsInputsEnv+"="+descriptor)
	output, err := command.CombinedOutput()
	cooked := ctx.Err() != nil
	cancel()
	if err != nil {
		if cooked {
			fmt.Fprintln(os.Stderr, "cooked: TestWholeMutants_Setup exceeded 90 seconds")
		}
		fmt.Fprintf(os.Stderr, "TestWholeMutants_Setup: %v\n%s", err, output)
		os.RemoveAll(directory)
		os.Exit(1)
	}
	// Expose the setup's build census without replaying nested test frames.
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "build ") || strings.Contains(line, "setup:") {
			fmt.Fprintln(os.Stdout, "TestWholeMutants_Setup:", strings.TrimSpace(line))
		}
	}
	data, err := os.ReadFile(descriptor)
	if err == nil {
		err = json.Unmarshal(data, &wholeMutantsPrepared)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.RemoveAll(directory)
		os.Exit(1)
	}
	if err := os.Setenv(wholeMutantsInputsEnv, descriptor); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.RemoveAll(directory)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(directory)
	os.Exit(code)
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
	if wholeMutantsPrepared != nil {
		wholeMutantsCheckInputs(t)
		return
	}
	// An independently selected setup unit also owns a bounded process group.
	if os.Getenv(wholeMutantsSetupChildEnv) != "1" {
		descriptor := filepath.Join(t.TempDir(), "inputs.json")
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		command := wholeMutantCommandContext(ctx, os.Args[0], "-test.run=^TestWholeMutants_Setup$", "-test.v", "-test.timeout=90s")
		command.Env = append(os.Environ(), wholeMutantsSetupChildEnv+"=1", wholeMutantsInputsEnv+"="+descriptor)
		output, err := command.CombinedOutput()
		if ctx.Err() != nil {
			t.Fatal("cooked: TestWholeMutants_Setup exceeded 90 seconds")
		}
		if err != nil {
			t.Fatalf("setup: %v\n%s", err, output)
		}
		data, err := os.ReadFile(descriptor)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &wholeMutantsPrepared); err != nil {
			t.Fatal(err)
		}
		wholeMutantsCheckInputs(t)
		t.Logf("%s", output)
		return
	}
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
	if descriptor := os.Getenv(wholeMutantsInputsEnv); descriptor != "" {
		data, err := json.Marshal(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(descriptor, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
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
		t.Fatal("mutant binary was not prepared by TestWholeMutants_Setup")
	}
	return binary
}

func wholeMutantsRunShard(t *testing.T, index int) {
	t.Helper()
	if !wholeMutantShardSelected(t, index) {
		t.Skip("assigned to another ADAMIC_TEST_SHARD")
	}
	wholeMutantsCheckInputs(t)
	if os.Getenv(wholeMutantsCaseChildEnv) == "1" {
		wholeMutantsShard(t, index)
		return
	}
	// Shared setup is already ready: only this case is inside the deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := wholeMutantCommandContext(ctx, os.Args[0], "-test.run=^"+t.Name()+"$", "-test.v", "-test.timeout=90s")
	command.Env = append(os.Environ(), wholeMutantsCaseChildEnv+"=1")
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
