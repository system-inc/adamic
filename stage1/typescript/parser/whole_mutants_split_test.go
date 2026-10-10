package parser

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
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
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Build products are inputs. No package-local cache: Product addresses each
// oracle, lowered program and sanitized executable by its complete inputs.
func wholeMutantBuildFiles(t *testing.T) []string {
	t.Helper()
	files := []string{"go.mod", "go.work"}
	for _, root := range []string{"internal", "cohere"} {
		err := filepath.WalkDir(filepath.Join(repository, root), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			name := entry.Name()
			if strings.HasSuffix(name, ".go") || strings.HasSuffix(name, ".c") || strings.HasSuffix(name, ".h") || strings.HasSuffix(name, ".d.ts") || strings.HasSuffix(name, ".json.gz") || name == "go.mod" || name == "go.sum" {
				relative, err := filepath.Rel(repository, path)
				if err != nil {
					return err
				}
				files = append(files, filepath.ToSlash(relative))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return files
}

func wholeMutantBuildFlags() []string {
	flags := []string{"GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH}
	for _, name := range []string{"GOFLAGS", "GOEXPERIMENT", "GOAMD64", "GOARM64", "CGO_ENABLED", "GOTOOLCHAIN", "ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "ADAMIC_GATE_UNCACHED"} {
		flags = append(flags, name+"="+os.Getenv(name))
	}
	return flags
}

func wholeMutantBuildOracleProduct(t *testing.T) string {
	t.Helper()
	files := append(wholeMutantBuildFiles(t), "stage1/typescript/parser/testdata/oracle.go")
	inputs := buildcache.Inputs{Name: "typescript-parser-oracle", Files: files, Flags: wholeMutantBuildFlags(), Toolchain: []string{buildcache.Tool("go", "version")}}
	directory := buildcache.Product(t, inputs, func(dir string) error {
		root, err := filepath.Abs(filepath.Join(repository, "cohere/TypeScript/tsc"))
		if err != nil {
			return err
		}
		side, err := filepath.Abs("testdata/oracle.go")
		if err != nil {
			return err
		}
		virtual := filepath.Join(root, "adamic_parser_oracle.go")
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
		if err != nil {
			return err
		}
		path := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		command := wholeMutantCommandContext(ctx, "go", "build", "-overlay="+path, "-o", filepath.Join(dir, "oracle"), virtual)
		command.Dir = root
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("oracle: %w\n%s", err, output)
		}
		return nil
	})
	return filepath.Join(directory, "oracle")
}

func wholeMutantPortInputs(t *testing.T, directory string) ([]string, []string) {
	t.Helper()
	files := wholeMutantBuildFiles(t)
	for _, name := range portFiles {
		files = append(files, "stage1/typescript/parser/"+name)
	}
	for _, name := range []string{"scanner.ts", "characters.ts", "tokens.ts"} {
		files = append(files, "stage1/typescript/scanner/"+name)
	}
	// Mutant copies live outside the repository. Their bytes, including each
	// changed site, are inputs too; normalize the scanner's relocation only.
	var sources strings.Builder
	scanner, err := filepath.Abs("../scanner/scanner.ts")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range portFiles {
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		normalized := strings.ReplaceAll(string(data), scanner, "../scanner/scanner.ts")
		fmt.Fprintf(&sources, "%s %d\n%s", name, len(normalized), normalized)
	}
	sourceHash := fmt.Sprintf("source=%x", sha256.Sum256([]byte(sources.String())))
	flags := append(wholeMutantBuildFlags(), sourceHash)
	return files, flags
}

func wholeMutantLowerProduct(t *testing.T, directory string) string {
	t.Helper()
	files, flags := wholeMutantPortInputs(t, directory)
	inputs := buildcache.Inputs{Name: "typescript-parser-lowered", Files: files, Flags: flags, Toolchain: []string{runtime.Version()}}
	lowered := buildcache.Product(t, inputs, func(dir string) error {
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "main.c"), []byte(native.C(lowered)), 0644)
	})
	return lowered
}

func wholeMutantBuildPortProduct(t *testing.T, directory string, sanitize bool) string {
	t.Helper()
	lowered := wholeMutantLowerProduct(t, directory)
	files, flags := wholeMutantPortInputs(t, directory)
	source, err := os.ReadFile(filepath.Join(lowered, "main.c"))
	if err != nil {
		t.Fatal(err)
	}
	options := native.Options{Sanitize: sanitize}
	nativeInputs := buildcache.Inputs{Name: "typescript-parser-native", Files: files, Flags: append(append(flags, native.Flags(options)...), fmt.Sprintf("C=%x", sha256.Sum256(source))), Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")}}
	product := buildcache.Product(t, nativeInputs, func(dir string) error {
		return native.Build(string(source), filepath.Join(dir, "scanner"), options)
	})
	return filepath.Join(product, "scanner")
}

// ADAMIC_TEST_SHARD=i/n selects indices congruent to i modulo n; unset runs
// every shard. Validate the complete union before applying that selection.
func wholeMutantShardSelected(t *testing.T, index int) bool {
	t.Helper()
	selection := os.Getenv("ADAMIC_TEST_SHARD")
	if selection == "" {
		return true
	}
	parts := strings.Split(selection, "/")
	if len(parts) != 2 {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q", selection)
	}
	i, firstErr := strconv.Atoi(parts[0])
	n, secondErr := strconv.Atoi(parts[1])
	if firstErr != nil || secondErr != nil || n <= 0 || i < 0 || i >= n {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q", selection)
	}
	return index%n == i
}

func wholeMutantShardUnion(t *testing.T, ids []string, groups [][]string, count int) {
	t.Helper()
	if len(groups) != count {
		t.Fatalf("enumerated %d shards, declared %d", len(groups), count)
	}
	expected := make(map[string]bool, len(ids))
	for _, id := range ids {
		if expected[id] {
			t.Fatalf("duplicate unsplit case %q", id)
		}
		expected[id] = true
	}
	seen := make(map[string]bool, len(ids))
	total := 0
	for _, group := range groups {
		for _, id := range group {
			if !expected[id] {
				t.Fatalf("unexpected sharded case %q", id)
			}
			if seen[id] {
				t.Fatalf("repeated sharded case %q", id)
			}
			seen[id] = true
			total++
		}
	}
	if total != len(ids) {
		t.Fatalf("union has %d cases, want %d", total, len(ids))
	}
	for _, id := range ids {
		if !seen[id] {
			t.Fatalf("missing sharded case %q", id)
		}
	}
	t.Logf("union: %d distinct case ids across %d shards", total, count)
}

func wholeMutantSetup(t *testing.T, started time.Time, products time.Duration) {
	t.Helper()
	total := time.Since(started)
	t.Logf("setup: total=%.3fs products=%.3fs without-products=%.3fs", total.Seconds(), products.Seconds(), (total - products).Seconds())
}

type wholeMutantCase struct{ name, file, from, to string }

func wholeMutantCases() []wholeMutantCase {
	return []wholeMutantCase{
		{"for-of becomes for-in", "statements.ts", "of ? 'ForOfStatement' : 'ForInStatement'", "of ? 'ForInStatement' : 'ForInStatement'"},
		{"type-only import phase lost", "statements.ts", "this.parser.node(clause).semantic = phase;", "this.parser.node(clause).semantic = 'Unknown';"},
		{"keyof becomes readonly", "parser.ts", "this.node(left).operator = operator;", "this.node(left).operator = operator === 'KeyOfKeyword' ? 'ReadonlyKeyword' : operator;"},
	}
}

// Explicit top-level leaves are visible to go test -list and the gate without
// a children() entry. Their table below is also the complete union census.
func TestWholeMutants_000(t *testing.T) { t.Parallel(); wholeMutantsRunShard(t, 0) }
func TestWholeMutants_001(t *testing.T) { t.Parallel(); wholeMutantsRunShard(t, 1) }
func TestWholeMutants_002(t *testing.T) { t.Parallel(); wholeMutantsRunShard(t, 2) }

func TestWholeMutantsUnion(t *testing.T) {
	t.Parallel()
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
			t.Fatalf("invalid shard index %d", leaf.index)
		}
		groups[i] = []string{cases[leaf.index].name}
	}
	wholeMutantShardUnion(t, ids, groups, testWholeMutantsShards)
}

// wholeMutantCommandContext bounds our direct child commands without an
// external timeout utility, and also cancels their compiler descendants.
func wholeMutantCommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	// Cancellation cleans up the whole process group, including descendants.
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = 5 * time.Second
	return command
}

func TestWholeMutantsRejectsSurvivor(t *testing.T) {
	t.Parallel()
	environment := wholeMutantsChildEnvironment(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := wholeMutantCommandContext(ctx, os.Args[0], "-test.run=^TestWholeMutants_[0-9]{3}$", "-test.v", "-test.timeout=90s", "-test.parallel=4")
	for _, variable := range environment {
		if strings.HasPrefix(variable, "ADAMIC_TEST_SHARD=") || strings.HasPrefix(variable, "ADAMIC_WHOLE_MUTANT_SURVIVOR=") || strings.HasPrefix(variable, wholeMutantsCaseChildEnv+"=") {
			continue
		}
		command.Env = append(command.Env, variable)
	}
	command.Env = append(command.Env, "ADAMIC_WHOLE_MUTANT_SURVIVOR=keyof becomes readonly", wholeMutantsCaseChildEnv+"=1")
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatal("cooked: survivor proof exceeded 90-second command deadline")
	}
	if err == nil {
		t.Fatalf("planted surviving mutant escaped\n%s", output)
	}
	var failed []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "--- FAIL: TestWholeMutants_") {
			failed = append(failed, strings.Fields(line)[2])
		}
	}
	const owner = "TestWholeMutants_002"
	if len(failed) != 1 || failed[0] != owner || !bytes.Contains(output, []byte("native planted mutant survived: keyof becomes readonly")) {
		t.Fatalf("plant must fail exactly %s, got %v: %v\n%s", owner, failed, err, output)
	}
	t.Logf("planted surviving keyof mutant caught only by %s", owner)
}
