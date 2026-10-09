package lint

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/system-inc/adamic/internal/native"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Keep build descriptions beside their builders until internal/buildcache lands
// on the lint seat. Existing shared helpers build each product once per run;
// each builder writes its output into directory and shards only read it.
type lintCoreInputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

func lintCoreProduct(t *testing.T, inputs lintCoreInputs, build func(directory string) error) (string, time.Duration) {
	t.Helper()
	directory := t.TempDir()
	started := time.Now()
	if err := build(directory); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(started)
	t.Logf("build input %s: %s (files=%d flags=%v toolchain=%s)", inputs.Name, elapsed, len(inputs.Files), inputs.Flags, inputs.Toolchain)
	return directory, elapsed
}

func lintCoreCopy(from, to string) error {
	input, err := os.Open(from)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(to, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

// Conservatively include every tracked repository and pinned submodule source:
// lowering and C runtime changes can alter the binary even when no port file
// changed. Generated dispatch is included by portFiles as well.
func lintCoreBuildFiles(t *testing.T) []string {
	t.Helper()
	command := exec.Command("git", "ls-files", "--recurse-submodules", "-z")
	command.Dir = repository
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	unique := map[string]bool{}
	for _, path := range strings.Split(string(output), "\x00") {
		if path != "" {
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			unique[absolute] = true
		}
	}
	for _, path := range portFiles(t) {
		absolute, err := filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		unique[absolute] = true
	}
	files := make([]string, 0, len(unique))
	for path := range unique {
		files = append(files, path)
	}
	sort.Strings(files)
	return files
}

func lintCoreToolchain(t *testing.T, clang bool) string {
	t.Helper()
	command := exec.Command("go", "env", "GOOS", "GOARCH", "CGO_ENABLED", "GOEXPERIMENT", "CC", "CXX", "CGO_CFLAGS", "CGO_LDFLAGS")
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	description := runtime.Version() + " " + strings.Join(strings.Fields(string(output)), " ")
	if clang {
		version, err := exec.Command("clang", "--version").Output()
		if err != nil {
			t.Fatal(err)
		}
		description += "; " + strings.SplitN(string(version), "\n", 2)[0]
	}
	return description
}

func lintCoreOracle(t *testing.T) (string, time.Duration) {
	t.Helper()
	inputs := lintCoreInputs{Name: "lint-go-oracle (local overlay)", Files: lintCoreBuildFiles(t), Flags: []string{"go build", "overlay of registered oracle sources"}, Toolchain: lintCoreToolchain(t, false)}
	directory, elapsed := lintCoreProduct(t, inputs, func(directory string) error {
		return lintCoreCopy(goOracle(t), filepath.Join(directory, "oracle"))
	})
	return filepath.Join(directory, "oracle"), elapsed
}

func lintCoreNative(t *testing.T, source string, sanitize bool) (string, time.Duration) {
	t.Helper()
	flags := append(native.Flags(native.Options{Sanitize: sanitize, Jobs: 1}), "go build -buildmode=c-archive ./bridge/tsgo/archive", "GOMAXPROCS=4")
	if sanitize {
		flags = append(flags, "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
	}
	inputs := lintCoreInputs{Name: "lint-native", Files: lintCoreBuildFiles(t), Flags: flags, Toolchain: lintCoreToolchain(t, true)}
	directory, elapsed := lintCoreProduct(t, inputs, func(directory string) error {
		return lintCoreCopy(buildPort(t, source, sanitize), filepath.Join(directory, "scanner"))
	})
	return filepath.Join(directory, "scanner"), elapsed
}

func lintCoreAssignments(cases, count int) [][]int {
	assignments := make([][]int, count)
	for index := 0; index < cases; index++ {
		assignments[index%count] = append(assignments[index%count], index)
	}
	return assignments
}

// Ordinal IDs identify every occurrence in the original enumeration, including
// two cases selecting the same file with different options. Check the entire
// union before applying the worker selector so a distributed run cannot hide a
// missing or repeated case in a shard assigned to another worker.
func lintCoreUnion(rows []string, assignments [][]int, count int) error {
	if len(assignments) != count {
		return fmt.Errorf("enumerated %d shards, want %d", len(assignments), count)
	}
	seen := make([]bool, len(rows))
	total := 0
	for shard, indices := range assignments {
		if len(indices) == 0 {
			return fmt.Errorf("shard-%03d is empty", shard)
		}
		for _, index := range indices {
			if index < 0 || index >= len(rows) {
				return fmt.Errorf("shard-%03d: unknown case %d", shard, index)
			}
			if seen[index] {
				return fmt.Errorf("shard-%03d: repeated case %d", shard, index)
			}
			seen[index] = true
			total++
		}
	}
	for index, found := range seen {
		if !found {
			return fmt.Errorf("missing case %d", index)
		}
	}
	if total != len(rows) {
		return fmt.Errorf("union count %d, want %d", total, len(rows))
	}
	return nil
}

func lintCoreRunShards(t *testing.T, rows []string, assignments [][]int, count int, run func(*testing.T, []int)) {
	t.Helper()
	if err := lintCoreUnion(rows, assignments, count); err != nil {
		t.Fatal(err)
	}
	t.Logf("shard union: %d cases, %d distinct ordinal case IDs, %d shards", len(rows), len(rows), count)
	worker, workers := 0, 1
	if selection := os.Getenv("ADAMIC_TEST_SHARD"); selection != "" {
		parts := strings.Split(selection, "/")
		if len(parts) != 2 {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q: want i/n", selection)
		}
		var err error
		worker, err = strconv.Atoi(parts[0])
		if err != nil {
			t.Fatal(err)
		}
		workers, err = strconv.Atoi(parts[1])
		if err != nil {
			t.Fatal(err)
		}
		if workers <= 0 || worker < 0 || worker >= workers {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q", selection)
		}
	}
	for shard, indices := range assignments {
		if shard%workers != worker {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", shard), func(t *testing.T) {
			t.Parallel()
			run(t, indices)
		})
	}
}

func lintCoreEqual(t *testing.T, label string, got, want []byte) {
	t.Helper()
	if diff := difference(got, want); diff != "" {
		t.Fatalf("%s: %s", label, diff)
	}
}

// The subprocess must fail exactly one leaf using the same assignment, union,
// selection and comparison path as the corpus test. A disagreement planted in
// case 7 belongs to shard-007 with eight shards.
func TestLintCorePlantedDisagreement(t *testing.T) {
	for _, selection := range []string{"", "0/2", "1/2"} {
		command := exec.Command(os.Args[0], "-test.run=^TestLintCoreDisagreementProbe$", "-test.v")
		command.Env = append(os.Environ(), "ADAMIC_LINT_CORE_PROBE=1", "ADAMIC_TEST_SHARD="+selection)
		output, err := command.CombinedOutput()
		if selection == "0/2" {
			if err != nil || bytes.Contains(output, []byte("--- FAIL:")) {
				t.Fatalf("wrong worker caught case 7: %s", output)
			}
			continue
		}
		if err == nil {
			t.Fatalf("planted disagreement survived: %s", output)
		}
		if bytes.Count(output, []byte("--- FAIL: TestLintCoreDisagreementProbe/shard-")) != 1 || !bytes.Contains(output, []byte("--- FAIL: TestLintCoreDisagreementProbe/shard-007")) || !bytes.Contains(output, []byte("planted disagreement case 7")) {
			t.Fatalf("wrong failing shard: %s", output)
		}
	}
	t.Log("planted disagreement case 7 caught only by shard-007, including worker 1/2; worker 0/2 passes")
}

func TestLintCoreDisagreementProbe(t *testing.T) {
	if os.Getenv("ADAMIC_LINT_CORE_PROBE") != "1" {
		return
	}
	rows := make([]string, 16)
	for index := range rows {
		rows[index] = fmt.Sprintf("case %d", index)
	}
	lintCoreRunShards(t, rows, lintCoreAssignments(len(rows), 8), 8, func(t *testing.T, indices []int) {
		for _, index := range indices {
			got, want := []byte(rows[index]), []byte(rows[index])
			if index == 7 {
				got = []byte("planted disagreement case 7")
			}
			lintCoreEqual(t, "planted disagreement case 7", got, want)
		}
	})
}

func TestLintCoreShardUnion(t *testing.T) {
	rows := make([]string, 16)
	good := lintCoreAssignments(len(rows), 8)
	if err := lintCoreUnion(rows, good, 8); err != nil {
		t.Fatal(err)
	}
	repeated := lintCoreAssignments(len(rows), 8)
	repeated[0] = append(repeated[0], 7)
	if err := lintCoreUnion(rows, repeated, 8); err == nil {
		t.Fatal("repeated case accepted")
	}
	missing := lintCoreAssignments(len(rows), 8)
	missing[7] = missing[7][1:]
	if err := lintCoreUnion(rows, missing, 8); err == nil {
		t.Fatal("missing case accepted")
	}
	if err := lintCoreUnion(rows, good[:7], 8); err == nil {
		t.Fatal("wrong shard count accepted")
	}
}
