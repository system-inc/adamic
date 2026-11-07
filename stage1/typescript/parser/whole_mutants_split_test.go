package parser

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/stage1/typescript/parser/oracles"
)

// Build products are inputs. No package-local cache: Product addresses each
// oracle, lowered program and sanitized executable by its complete inputs.
func wholeMutantBuildFiles(test *testing.T) []string {
	test.Helper()
	files := []string{"go.mod", "go.work"}
	for _, root := range []string{"internal", "cohere"} {
		// Reads: cohere
		walkError := filepath.WalkDir(filepath.Join(
			repository,
			root,
		), func(path string, entry fs.DirEntry, pathError error) error {
			if pathError != nil {
				return pathError
			}
			if entry.IsDir() {
				if entry.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			name := entry.Name()
			if strings.HasSuffix(name, ".go") ||
				strings.HasSuffix(name, ".c") ||
				strings.HasSuffix(name, ".h") ||
				strings.HasSuffix(name, ".d.ts") ||
				strings.HasSuffix(name, ".json.gz") ||
				name == "go.mod" ||
				name == "go.sum" {
				relative, relativePathError := filepath.Rel(repository, path)
				if relativePathError != nil {
					return relativePathError
				}
				files = append(files, filepath.ToSlash(relative))
			}
			return nil
		})
		if walkError != nil {
			test.Fatal(walkError)
		}
	}
	return files
}

func wholeMutantBuildFlags() []string {
	flags := []string{"GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH}
	for _, name := range []string{
		"GOFLAGS",
		"GOEXPERIMENT",
		"GOAMD64",
		"GOARM64",
		"CGO_ENABLED",
		"ADAMIC_NATIVE_SPLIT",
	} {
		flags = append(flags, name+"="+os.Getenv(name))
	}
	return flags
}

func wholeMutantBuildOracleProduct(test *testing.T) string {
	test.Helper()
	return filepath.Join(oracles.ParserOracleProduct(test), "oracle")
}

func wholeMutantPortInputs(test *testing.T, directory string) ([]string, []string) {
	test.Helper()
	files := wholeMutantBuildFiles(test)
	for _, name := range portFiles {
		files = append(files, "stage1/typescript/parser/"+name)
	}
	for _, name := range []string{"scanner.ts", "characters.ts", "tokens.ts"} {
		files = append(files, "stage1/typescript/scanner/"+name)
	}
	// Mutant copies live outside the repository. Their bytes, including each
	// changed site, are inputs too; normalize the scanner's relocation only.
	sourceHash := wholeMutantsSourceHash(test, directory)
	flags := append(wholeMutantBuildFlags(), sourceHash)
	return files, flags
}

func wholeMutantLowerProduct(test *testing.T, directory string) string {
	test.Helper()
	files, flags := wholeMutantPortInputs(test, directory)
	inputs := buildcache.Inputs{
		Name:      "typescript-parser-lowered",
		Files:     files,
		Flags:     flags,
		Toolchain: []string{runtime.Version()},
	}
	lowered := buildcache.Product(test, inputs, func(dir string) error {
		program, loadError := load.Load([]string{filepath.Join(directory, "main.ts")})
		if loadError != nil {
			return loadError
		}
		lowered, lowerError := lower.Lower(context.Background(), program)
		if lowerError != nil {
			return lowerError
		}
		return os.WriteFile(filepath.Join(dir, "main.c"), []byte(native.C(lowered)), 0644)
	})
	return lowered
}

func wholeMutantBuildPortProduct(test *testing.T, directory string, sanitize bool) string {
	test.Helper()
	lowered := wholeMutantLowerProduct(test, directory)
	files, flags := wholeMutantPortInputs(test, directory)
	source, readError := os.ReadFile(filepath.Join(lowered, "main.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	options := native.Options{Sanitize: sanitize}
	nativeInputs := buildcache.Inputs{
		Name:      "typescript-parser-native",
		Files:     files,
		Flags:     append(append(flags, native.Flags(options)...), fmt.Sprintf("C=%x", sha256.Sum256(source))),
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")},
	}
	product := buildcache.Product(test, nativeInputs, func(dir string) error {
		return native.Build(string(source), filepath.Join(dir, "scanner"), options)
	})
	return filepath.Join(product, "scanner")
}

// ADAMIC_TEST_SHARD=i/n selects indices congruent to i modulo n; unset runs
// every shard. Validate the complete union before applying that selection.
func wholeMutantShardSelected(test *testing.T, index int) bool {
	test.Helper()
	selection := os.Getenv("ADAMIC_TEST_SHARD")
	if selection == "" {
		return true
	}
	parts := strings.Split(selection, "/")
	if len(parts) != 2 {
		test.Fatalf("invalid ADAMIC_TEST_SHARD %q", selection)
	}
	i, firstError := strconv.Atoi(parts[0])
	n, secondError := strconv.Atoi(parts[1])
	if firstError != nil || secondError != nil || n <= 0 || i < 0 || i >= n {
		test.Fatalf("invalid ADAMIC_TEST_SHARD %q", selection)
	}
	return index%n == i
}

func wholeMutantShardUnion(test *testing.T, ids []string, groups [][]string, count int) {
	test.Helper()
	if len(groups) != count {
		test.Fatalf("enumerated %d shards, declared %d", len(groups), count)
	}
	expected := make(map[string]bool, len(ids))
	for _, id := range ids {
		if expected[id] {
			test.Fatalf("duplicate unsplit case %q", id)
		}
		expected[id] = true
	}
	seen := make(map[string]bool, len(ids))
	total := 0
	for _, group := range groups {
		for _, id := range group {
			if !expected[id] {
				test.Fatalf("unexpected sharded case %q", id)
			}
			if seen[id] {
				test.Fatalf("repeated sharded case %q", id)
			}
			seen[id] = true
			total++
		}
	}
	if total != len(ids) {
		test.Fatalf("union has %d cases, want %d", total, len(ids))
	}
	for _, id := range ids {
		if !seen[id] {
			test.Fatalf("missing sharded case %q", id)
		}
	}
	test.Logf("union: %d distinct case ids across %d shards", total, count)
}

func wholeMutantSetup(test *testing.T, started time.Time, products time.Duration) {
	test.Helper()
	total := time.Since(started)
	test.Logf(
		"setup: total=%.3fs products=%.3fs without-products=%.3fs",
		total.Seconds(),
		products.Seconds(),
		(total - products).Seconds(),
	)
}

type wholeMutantCase struct{ name, file, from, to string }

func wholeMutantCases() []wholeMutantCase {
	return []wholeMutantCase{
		{
			"for-of becomes for-in",
			"statements.ts",
			"of ? 'ForOfStatement' : 'ForInStatement'",
			"of ? 'ForInStatement' : 'ForInStatement'",
		},
		{
			"type-only import phase lost",
			"statements.ts",
			"this.parser.node(clause).semantic = phase;",
			"this.parser.node(clause).semantic = 'Unknown';",
		},
		{
			"keyof becomes readonly",
			"parser.ts",
			"this.node(left).operator = operator;",
			"this.node(left).operator = operator === 'KeyOfKeyword' ? 'ReadonlyKeyword' : operator;",
		},
	}
}

// Explicit top-level leaves are visible to go test -list and the gate without
// a children() entry. Their table below is also the complete union census.
func TestWholeMutants_000(test *testing.T) { test.Parallel(); wholeMutantsRunShard(test, 0) }
func TestWholeMutants_001(test *testing.T) { test.Parallel(); wholeMutantsRunShard(test, 1) }
func TestWholeMutants_002(test *testing.T) { test.Parallel(); wholeMutantsRunShard(test, 2) }

func TestWholeMutantsUnion(test *testing.T) {
	test.Parallel()
	cases := wholeMutantCases()
	ids := make([]string, len(cases))
	for i, c := range cases {
		ids[i] = c.name
	}
	leaves := []struct {
		index int
		test  func(*testing.T)
	}{
		{0, TestWholeMutants_000}, {1, TestWholeMutants_001}, {2, TestWholeMutants_002},
	}
	groups := make([][]string, len(leaves))
	for i, leaf := range leaves {
		if leaf.index < 0 || leaf.index >= len(cases) {
			test.Fatalf("invalid shard index %d", leaf.index)
		}
		groups[i] = []string{cases[leaf.index].name}
	}
	wholeMutantShardUnion(test, ids, groups, testWholeMutantsShards)
}

// wholeMutantCommandContext bounds our direct child commands without an
// external timeout utility, and also cancels their compiler descendants. Its
// output is read through childguard.Capture, to its end with no timer deciding
// when: exec's own copying under a five-second WaitDelay reported
// exec.ErrWaitDelay whenever the reads hadn't reached EOF five seconds after the
// child exited, so a leaf or survivor proof that had finished could fail under
// load (#n3gk22h's shape). Each caller's 90 s deadline ends a hung run.
func wholeMutantCommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	// A tool named bare (go) is toolchain's program, under toolchain's environment; a binary named by its path runs as
	// itself.
	if !strings.ContainsRune(name, filepath.Separator) {
		tool := toolchain.Command(name, args...)
		command.Path, command.Err, command.Env = tool.Path, tool.Err, tool.Env
	}
	// Cancellation cleans up the whole process group, including descendants.
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if killError == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return killError
	}
	return command
}

func TestWholeMutantsRejectsSurvivor(test *testing.T) {
	test.Parallel()
	environment := wholeMutantsChildEnvironment(test)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := wholeMutantCommandContext(
		ctx,
		os.Args[0],
		"-test.run=^TestWholeMutants_[0-9]{3}$",
		"-test.v",
		"-test.timeout=90s",
		"-test.parallel=4",
	)
	for _, variable := range environment {
		if strings.HasPrefix(variable, "ADAMIC_TEST_SHARD=") ||
			strings.HasPrefix(variable, "ADAMIC_WHOLE_MUTANT_SURVIVOR=") ||
			strings.HasPrefix(variable, wholeMutantsCaseChildEnv+"=") {
			continue
		}
		command.Env = append(command.Env, variable)
	}
	command.Env = append(
		command.Env,
		"ADAMIC_WHOLE_MUTANT_SURVIVOR=keyof becomes readonly",
		wholeMutantsCaseChildEnv+"=1",
	)
	// Read through childguard.Capture for wholeMutantCommandContext's reason, stdout then stderr.
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	output := append(stdout, stderr...)
	if ctx.Err() != nil {
		test.Fatal("cooked: survivor proof exceeded 90-second command deadline")
	}
	if captureError == nil {
		test.Fatalf("planted surviving mutant escaped\n%s", output)
	}
	var failed []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "--- FAIL: TestWholeMutants_") {
			failed = append(failed, strings.Fields(line)[2])
		}
	}
	const owner = "TestWholeMutants_002"
	if len(failed) != 1 ||
		failed[0] != owner ||
		!bytes.Contains(output, []byte("native planted mutant survived: keyof becomes readonly")) {
		test.Fatalf("plant must fail exactly %s, got %v: %v\n%s", owner, failed, captureError, output)
	}
	test.Logf("planted surviving keyof mutant caught only by %s", owner)
}
