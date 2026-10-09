package lint

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"hash/fnv"
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
)

type dotARenameRuntime struct{ Directory, Native, JavaScript string }
type dotARenameProducts struct {
	Original, Changed dotARenameRuntime
	Oracle            string
	Rows              []string
	Before, After     []byte
}

func dotARenameCases() []string {
	return []string{"stage1/cohere/lint/rules/no-var/rule.a:rename-to-ts"}
}

func TestDotARenameUnion(t *testing.T) { t.Parallel(); dotARenameUnion(t) }
func dotARenameUnion(t *testing.T) {
	t.Helper()
	ids := dotARenameCases()
	if len(ids) == 0 {
		t.Fatal("empty regression enumeration")
	}
	shards := [][]string{nil} // The written top-level unit TestDotARename_000.
	if len(shards) != testDotARenameShards {
		t.Fatal("shard declaration differs from enumeration")
	}
	expected := map[string]bool{}
	for _, id := range ids {
		if expected[id] {
			t.Fatalf("repeated unsplit case %q", id)
		}
		expected[id] = true
		h := fnv.New64a()
		h.Write([]byte(id))
		index := int(h.Sum64() % uint64(testDotARenameShards))
		shards[index] = append(shards[index], id)
	}
	if len(shards) != testDotARenameShards {
		t.Fatal("shard declaration differs from enumeration")
	}
	seen := map[string]bool{}
	total := 0
	for _, shard := range shards {
		for _, id := range shard {
			if !expected[id] || seen[id] {
				t.Fatalf("missing or repeated shard case %q", id)
			}
			seen[id] = true
			total++
		}
	}
	if total != len(ids) || len(seen) != len(expected) {
		t.Fatal("shard union differs from live enumeration")
	}
	for id := range expected {
		if !seen[id] {
			t.Fatalf("missing case %q", id)
		}
	}
	t.Logf("union: %d cases, %d unique ids, %d shards", total, len(seen), len(shards))
}
func dotARenameSelected(t *testing.T) bool {
	t.Helper()
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return true
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		t.Fatal("ADAMIC_TEST_SHARD requires zero-based i/n")
	}
	i, e1 := strconv.Atoi(parts[0])
	n, e2 := strconv.Atoi(parts[1])
	if e1 != nil || e2 != nil || n < 1 || i < 0 || i >= n {
		t.Fatal("invalid ADAMIC_TEST_SHARD")
	}
	return i == 0
}
func dotARenameSupplied(t *testing.T) (dotARenameProducts, bool) {
	t.Helper()
	var p dotARenameProducts
	path := os.Getenv("ADAMIC_LINT_DOT_A_PRODUCTS")
	if path == "" {
		return p, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatal(err)
	}
	return p, true
}

// The temporary port is hashed by logical path and content. Only the checkout
// prefixes that copyPort inserts into imports are normalized. The compiler and
// its bundled checker libraries, imported Stage 1 sources, and this recipe are
// also inputs; no compiled product or test result is written into the package.
func dotARenameInputs(t *testing.T, directory string) buildcache.Inputs {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	port, err := filepath.Abs(directory)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.New()
	err = filepath.WalkDir(port, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".ts" && ext != ".a" && ext != ".json" {
			return nil
		}
		relative, err := filepath.Rel(port, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := strings.ReplaceAll(string(data), root, "$REPOSITORY")
		text = strings.ReplaceAll(text, port, "$PORT")
		fmt.Fprintf(hash, "%d:%s:%d:%s", len(relative), relative, len(text), text)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return buildcache.Inputs{Name: "lint-dot-a-lowered", Files: []string{
		"internal", "bridge", "stage1/typescript", "stage1/cohere/typeaware", "stage1/cohere/lint/registry",
		"stage1/cohere/lint/harness_test.go", "stage1/cohere/lint/harness_dot_a_units_test.go",
		"cohere",
		"go.mod", "go.work", "cohere/go.mod", "cohere/go.sum", "CohereSettings.json"},
		Flags:     []string{fmt.Sprintf("port-sha256=%x", hash.Sum(nil)), runtime.GOOS, runtime.GOARCH},
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")}}
}
func dotARenameBuild(t *testing.T, directory string) dotARenameRuntime {
	t.Helper()
	prepareRegistry(t, directory)
	inputs := dotARenameInputs(t, directory)
	lowered := buildcache.Product(t, inputs, func(dir string) error {
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			return err
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "lint.c"), []byte(native.C(ir)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "lint.mjs"), []byte(javascript.JavaScript(ir)), 0644)
	})
	data, err := os.ReadFile(filepath.Join(lowered, "lint.c"))
	if err != nil {
		t.Fatal(err)
	}
	options := native.Options{Sanitize: true, Split: true, Jobs: 4}
	nativeInputs := buildcache.Inputs{Name: "lint-dot-a-sanitized-native", Files: []string{"internal/native", "stage1/cohere/lint/harness_dot_a_units_test.go"},
		Flags: append(native.Flags(options), fmt.Sprintf("C-sha256=%x", sha256.Sum256(data)), "Split=true", "Jobs=4"), Toolchain: inputs.Toolchain}
	product := buildcache.Product(t, nativeInputs, func(dir string) error { return native.Build(string(data), filepath.Join(dir, "scanner"), options) })
	return dotARenameRuntime{directory, filepath.Join(product, "scanner"), filepath.Join(lowered, "lint.mjs")}
}
func dotARenamePlanted(t *testing.T, p dotARenameProducts) {
	t.Helper()
	data, err := os.ReadFile(p.Changed.JavaScript)
	if err != nil {
		t.Fatal(err)
	}
	p.Changed.JavaScript = filepath.Join(t.TempDir(), "planted.mjs")
	if err := os.WriteFile(p.Changed.JavaScript, append(data, []byte("\nconsole.log('planted rename disagreement');\n")...), 0644); err != nil {
		t.Fatal(err)
	}
	data, err = json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "products.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	for seat := 0; seat < 2; seat++ {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDotARename_000$", "-test.v", "-test.timeout=90s")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Cancel = func() error {
			if cmd.Process == nil {
				return os.ErrProcessDone
			}
			err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			if err == syscall.ESRCH {
				return os.ErrProcessDone
			}
			return err
		}
		cmd.Env = append(os.Environ(), "ADAMIC_LINT_DOT_A_PRODUCTS="+path, fmt.Sprintf("ADAMIC_TEST_SHARD=%d/2", seat))
		output, err := cmd.CombinedOutput()
		cancel()
		text := string(output)
		if seat == 0 {
			if err == nil || strings.Count(text, "--- FAIL: TestDotARename_000") != 1 || !strings.Contains(text, "emitted JavaScript:") || !strings.Contains(text, "planted rename disagreement") {
				t.Fatalf("wrong planted failure: %v\n%s", err, text)
			}
			t.Log("planted disagreement caught exactly by TestDotARename_000")
		} else if err != nil || strings.Contains(text, "--- FAIL:") || strings.Contains(text, "case ids:") || !strings.Contains(text, "--- SKIP: TestDotARename_000") {
			t.Fatalf("non-owning seat checked planted case: %v\n%s", err, text)
		}
	}
}

func dotARenameSetup(t *testing.T) func() {
	t.Helper()
	started := time.Now()
	return func() { t.Helper(); t.Logf("setup: %s", time.Since(started)) }
}
