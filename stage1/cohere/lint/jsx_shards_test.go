package lint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Each build has its inputs beside its closure. Until internal/buildcache is
// available on this base, the existing harness shares it within the run. This
// helper never creates a persistent package cache.
type jsxProductInputs struct {
	Name      string
	Files     []string
	Flags     []string
	Toolchain string
}

// Include source dependencies as files, not just the entry point. A future
// content-keyed Product must invalidate when an imported source or tool changes.
func jsxInputFiles(t *testing.T, roots ...string) []string {
	t.Helper()
	files := map[string]bool{}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == ".git" || entry.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			switch filepath.Ext(path) {
			case ".go", ".ts", ".a", ".c", ".h", ".mod", ".sum", ".work", ".json":
				if strings.HasSuffix(path, "_test.go") {
					return nil
				}
				absolute, err := filepath.Abs(path)
				if err != nil {
					return err
				}
				files[absolute] = true
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	result := make([]string, 0, len(files))
	for path := range files {
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}

func jsxProduct(t *testing.T, inputs jsxProductInputs, build func(dir string) error) string {
	t.Helper()
	value := shared(fmt.Sprintf("jsx product %+v", inputs), func(value *sharedValue) {
		value.path, value.err = os.MkdirTemp(sharedDirectory, "jsx-product-")
		if value.err != nil {
			return
		}
		started := time.Now()
		value.err = build(value.path)
		t.Logf("build %s cold wall: %s", inputs.Name, time.Since(started))
	})
	if value.err != nil {
		t.Fatal(value.err)
	}
	return value.path
}

func jsxGoOracle(t *testing.T, name, root, side, virtualName string) string {
	t.Helper()
	inputs := jsxProductInputs{Name: name, Files: jsxInputFiles(t, side, filepath.Join(repository, "cohere")), Flags: []string{"build", "-overlay"}, Toolchain: runtime.Version()}
	directory := jsxProduct(t, inputs, func(dir string) error {
		virtual := filepath.Join(root, virtualName)
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
		if err != nil {
			return err
		}
		overlayPath := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
			return err
		}
		command := exec.Command("go", "build", "-overlay="+overlayPath, "-o", filepath.Join(dir, "oracle"), virtual)
		command.Dir = root
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		if err := command.Run(); err != nil {
			return fmt.Errorf("build %s: %w\n%s\n%s", name, err, &stdout, &stderr)
		}
		if diagnostics := commandDiagnostics("go", stderr.Bytes()); len(diagnostics) != 0 {
			return fmt.Errorf("build %s: %s", name, diagnostics)
		}
		return nil
	})
	return filepath.Join(directory, "oracle")
}

func jsxTreeShards(paths []string) ([][]string, error) {
	ordered := append([]string(nil), paths...)
	sort.Strings(ordered)
	shards := make([][]string, testJsxLintTreesShards)
	for i, path := range ordered {
		shards[i%len(shards)] = append(shards[i%len(shards)], path)
	}
	if err := jsxValidateUnion(paths, shards); err != nil {
		return nil, err
	}
	return shards, nil
}

func jsxValidateUnion(paths []string, shards [][]string) error {
	if len(shards) != testJsxLintTreesShards {
		return fmt.Errorf("enumerated %d shards, want %d", len(shards), testJsxLintTreesShards)
	}
	expected := map[string]bool{}
	for _, path := range paths {
		if path == "" || expected[path] {
			return fmt.Errorf("empty or repeated unsplit case id %q", path)
		}
		expected[path] = true
	}
	seen := map[string]bool{}
	count := 0
	for i, cases := range shards {
		if len(cases) == 0 {
			return fmt.Errorf("shard-%03d is empty", i)
		}
		for _, path := range cases {
			if !expected[path] {
				return fmt.Errorf("shard-%03d has unknown case id %q", i, path)
			}
			if seen[path] {
				return fmt.Errorf("shard-%03d repeats case id %q", i, path)
			}
			seen[path] = true
			count++
		}
	}
	if count != len(paths) || len(seen) != len(expected) {
		return fmt.Errorf("shard union has %d cases and %d ids, want %d", count, len(seen), len(paths))
	}
	for path := range expected {
		if !seen[path] {
			return fmt.Errorf("shard union missing case id %q", path)
		}
	}
	return nil
}

func jsxShardSelection(value string) (func(int) bool, error) {
	if value == "" {
		return func(int) bool { return true }, nil
	}
	fields := strings.Split(value, "/")
	if len(fields) != 2 {
		return nil, fmt.Errorf("invalid ADAMIC_TEST_SHARD %q: want i/n", value)
	}
	index, first := strconv.Atoi(fields[0])
	count, second := strconv.Atoi(fields[1])
	if first != nil || second != nil || count <= 0 || count > testJsxLintTreesShards || index < 0 || index >= count {
		return nil, fmt.Errorf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	return func(shard int) bool { return shard%count == index }, nil
}

func jsxPlantDisagreement(output []byte) []byte {
	// Keep case numbering and change only the first case's tree.
	newline := bytes.IndexByte(output, '\n')
	if newline < 0 {
		panic("tree output has no case header")
	}
	result := append([]byte(nil), output[:newline+1]...)
	result = append(result, []byte("planted tree disagreement\n")...)
	return append(result, output[newline+1:]...)
}

func jsxCheckTree(t *testing.T, side string, got, want []byte) {
	t.Helper()
	if diff := difference(got, want); diff != "" {
		t.Fatalf("%s: %s", side, diff)
	}
}

func TestJsxLintTreesShardCoverage(t *testing.T) {
	paths := make([]string, 2*testJsxLintTreesShards+3)
	for i := range paths {
		paths[i] = fmt.Sprintf("case-%03d.tsx", i)
	}
	shards, err := jsxTreeShards(paths)
	if err != nil {
		t.Fatal(err)
	}
	for _, fault := range []string{"missing", "repeated", "unknown", "shard count"} {
		t.Run(fault, func(t *testing.T) {
			broken := make([][]string, len(shards))
			for i := range shards {
				broken[i] = append([]string(nil), shards[i]...)
			}
			switch fault {
			case "missing":
				broken[0] = broken[0][1:]
			case "repeated":
				broken[0] = append(broken[0], broken[1][0])
			case "unknown":
				broken[0][0] = "unknown.tsx"
			case "shard count":
				broken = broken[:len(broken)-1]
			}
			if err := jsxValidateUnion(paths, broken); err == nil {
				t.Fatalf("%s union fault survived", fault)
			}
		})
	}
	for n := 1; n <= testJsxLintTreesShards; n++ {
		counts := make([]int, len(shards))
		for i := 0; i < n; i++ {
			selectShard, err := jsxShardSelection(fmt.Sprintf("%d/%d", i, n))
			if err != nil {
				t.Fatal(err)
			}
			for shard := range shards {
				if selectShard(shard) {
					counts[shard]++
				}
			}
		}
		for shard, count := range counts {
			if count != 1 {
				t.Fatalf("%d boxes select shard-%03d %d times", n, shard, count)
			}
		}
	}
	for _, invalid := range []string{"1", "-1/16", "16/16", "0/0", "0/17", "x/16", "0/16/1"} {
		if _, err := jsxShardSelection(invalid); err == nil {
			t.Fatalf("invalid selection %q accepted", invalid)
		}
	}
}

// Run the actual comparison helper in a child so this test requires an observed
// failing leaf, rather than treating an error returned by a mock as detection.
func TestJsxLintTreesShardDisagreement(t *testing.T) {
	if os.Getenv("ADAMIC_JSX_SHARD_CHILD") == "1" {
		paths := make([]string, testJsxLintTreesShards)
		for i := range paths {
			paths[i] = fmt.Sprintf("case-%03d.tsx", i)
		}
		shards, err := jsxTreeShards(paths)
		if err != nil {
			t.Fatal(err)
		}
		for i, cases := range shards {
			t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
				t.Parallel()
				want := []byte("case 0\nJsxElement " + cases[0] + "\n")
				got := want
				if cases[0] == "case-003.tsx" {
					got = jsxPlantDisagreement(got)
				}
				jsxCheckTree(t, "native", got, want)
			})
		}
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "-test.run=^TestJsxLintTreesShardDisagreement$", "-test.v", "-test.timeout=20s")
	command.Env = append(os.Environ(), "ADAMIC_JSX_SHARD_CHILD=1")
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatalf("planted disagreement survived:\n%s", output)
	}
	failed := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "--- FAIL: TestJsxLintTreesShardDisagreement/shard-") {
			failed = append(failed, strings.Fields(line)[2])
		}
	}
	if len(failed) != 1 || failed[0] != "TestJsxLintTreesShardDisagreement/shard-003" || !bytes.Contains(output, []byte("planted tree disagreement")) {
		t.Fatalf("wanted only shard-003 to catch case-003, got %v:\n%s", failed, output)
	}
	t.Log("planted disagreement in case-003.tsx caught only by shard-003")
}
