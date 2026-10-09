package estree

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const testScalarEdgesShards = 4

// Contiguous ranges are recomputed from the live pinned corpus, including additions.
func scalarEdgeRanges(count int) [testScalarEdgesShards][2]int {
	var ranges [testScalarEdgesShards][2]int
	for i := range ranges {
		ranges[i] = [2]int{i * count / testScalarEdgesShards, (i + 1) * count / testScalarEdgesShards}
	}
	return ranges
}

func scalarEdgeInputs() buildcache.Inputs {
	return buildcache.Inputs{
		Name:      "scalar-edges-lowered",
		Files:     []string{"stage1/cohere/estree", "stage1/typescript", "internal", "cohere", "go.mod"},
		Toolchain: []string{runtime.Version()},
	}
}

func scalarEdgeLowered(t *testing.T, main string) string {
	t.Helper()
	return buildcache.Product(t, scalarEdgeInputs(), func(dir string) error {
		program, err := load.Load([]string{main})
		if err != nil {
			return err
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "port.c"), []byte(native.C(ir)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "port.mjs"), []byte(javascript.JavaScript(ir)), 0644)
	})
}

func scalarEdgeNative(t *testing.T, lowered string) string {
	t.Helper()
	inputs := scalarEdgeInputs()
	inputs.Name = "scalar-edges-sanitized-native-v2"
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	inputs.Flags = append(native.Flags(native.Options{Sanitize: true, Split: true}), []string{"Sanitize=true", "Split=true", "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS=" + os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED=" + os.Getenv("ADAMIC_GATE_UNCACHED")}...)
	product := buildcache.Product(t, inputs, func(dir string) error {
		source, err := os.ReadFile(filepath.Join(lowered, "port.c"))
		if err != nil {
			return err
		}
		return native.Build(string(source), filepath.Join(dir, "port"), native.Options{Sanitize: true, Split: true})
	})
	return filepath.Join(product, "port")
}

func scalarEdgeProducts(t *testing.T, main string) (string, string) {
	t.Helper()
	lowered := scalarEdgeLowered(t, main)
	return scalarEdgeNative(t, lowered), filepath.Join(lowered, "port.mjs")
}

func scalarEdgeOracle(t *testing.T) string {
	t.Helper()
	repo := root(t)
	source := filepath.Join(repo, "stage1/cohere/estree/testdata/oracle.go")
	virtual := filepath.Join(repo, "cohere/adamic_estree_oracle.go")
	product := buildcache.Product(t, buildcache.Inputs{
		Name: "scalar-edges-go-oracle", Files: []string{"cohere", "stage1/cohere/estree/testdata/oracle.go"},
		Flags: []string{"go build overlay oracle"}, Toolchain: []string{buildcache.Tool("go", "version")},
	}, func(dir string) error {
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source}})
		if err != nil {
			return err
		}
		path := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		command := exec.CommandContext(context.Background(), "go", "build", "-overlay="+path, "-o", filepath.Join(dir, "oracle"), virtual)
		command.Dir = filepath.Join(repo, "cohere")
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("oracle: %w\n%s", err, output)
		}
		return nil
	})
	return filepath.Join(product, "oracle")
}

type scalarEdgeReady struct{ Oracle, Binary, Script string }

var scalarEdgePreparation struct {
	once  sync.Once
	ready scalarEdgeReady
}

// Every selected test prepares its products before starting its case deadline.
// Product owns cross-process caching; Once shares the products within this process.
func scalarEdgePrepare(t *testing.T) scalarEdgeReady {
	t.Helper()
	scalarEdgePreparation.once.Do(func() {
		main, err := filepath.Abs("main.ts")
		if err != nil {
			t.Fatal(err)
		}
		ready := scalarEdgeReady{Oracle: scalarEdgeOracle(t)}
		ready.Binary, ready.Script = scalarEdgeProducts(t, main)
		scalarEdgePreparation.ready = ready
	})
	ready := scalarEdgePreparation.ready
	if ready.Oracle == "" || ready.Binary == "" || ready.Script == "" {
		t.Fatal("scalar-edge preparation failed")
	}
	return ready
}

func scalarEdgeCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	return command
}

func TestScalarEdges_Setup(t *testing.T) {
	t.Parallel()
	scalarEdgePrepare(t)
}

// Retain the original selection as a preparation-only compatibility entry point.
func TestScalarEdges(t *testing.T) {
	t.Parallel()
	scalarEdgePrepare(t)
}

func scalarEdgeExecute(t *testing.T, ctx context.Context, name string, args ...string) []byte {
	t.Helper()
	command := scalarEdgeCommand(ctx, name, args...)
	var output, stderr bytes.Buffer
	command.Stdout = &output
	command.Stderr = &stderr
	if err := command.Run(); err != nil || stderr.Len() != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	return output.Bytes()
}

func scalarEdgeNode(t *testing.T, ctx context.Context, path string, args ...string) []byte {
	t.Helper()
	argv := []string{"--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), path}
	return scalarEdgeExecute(t, ctx, "node", append(argv, args...)...)
}

func TestScalarEdgesUnion(t *testing.T) {
	t.Parallel()
	cases := scalarCases()
	ranges := scalarEdgeRanges(len(cases))
	seen := make([]int, len(cases))
	for _, span := range ranges {
		for i := span[0]; i < span[1]; i++ {
			seen[i]++
		}
	}
	for i, n := range seen {
		if n != 1 {
			t.Fatalf("case %d appears %d times", i, n)
		}
	}
	// This table is also the top-level test enumeration, so changing N alone fails.
	tests := []func(*testing.T){TestScalarEdges_000, TestScalarEdges_001, TestScalarEdges_002, TestScalarEdges_003}
	if len(tests) != testScalarEdgesShards {
		t.Fatalf("%d tests for %d shards", len(tests), testScalarEdgesShards)
	}
	t.Logf("union: %d cases, each exactly once; %d shards", len(cases), len(tests))
}

func TestScalarEdgesPlantedFailure(t *testing.T) {
	t.Parallel()
	ready := scalarEdgePrepare(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	oracle := ready.Oracle
	cases := scalarCases()
	caught := -1
	catches := 0
	for shard, span := range scalarEdgeRanges(len(cases)) {
		list := manifest(t, cases[span[0]:span[1]])
		want := scalarEdgeExecute(t, ctx, oracle, "--manifest", list)
		if span[0] == 0 && span[1] > 0 {
			listing, err := os.ReadFile(list)
			if err != nil {
				t.Fatal(err)
			}
			path := strings.Split(string(listing), "\n")[0]
			if err := os.WriteFile(path, []byte("0;"), 0644); err != nil {
				t.Fatal(err)
			}
		}
		got := scalarEdgeExecute(t, ctx, oracle, "--manifest", list)
		if firstDifference(want, got) != "" {
			catches++
			caught = shard
		}
	}
	if catches != 1 || caught != 0 {
		t.Fatalf("planted case caught %d times in shard %d", catches, caught)
	}
	t.Log("planted failure: case 0 changed to 0; caught only by TestScalarEdges_000")
}

func scalarEdgeShard(t *testing.T, shard int) {
	t.Helper()
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	ready := scalarEdgePrepare(t)
	// Only case work below is charged to the shard deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	started := time.Now()
	defer func() { t.Logf("shard-%03d cases: %.3fs", shard, time.Since(started).Seconds()) }()
	oracle, binary, script := ready.Oracle, ready.Binary, ready.Script
	cases := scalarCases()
	span := scalarEdgeRanges(len(cases))[shard]
	list := manifest(t, cases[span[0]:span[1]])
	want := scalarEdgeExecute(t, ctx, oracle, "--manifest", list)
	for name, got := range map[string][]byte{
		"source Node":      scalarEdgeNode(t, ctx, main, "--manifest", list),
		"sanitized native": scalarEdgeExecute(t, ctx, binary, "--manifest", list),
		"emitted JS":       scalarEdgeNode(t, ctx, script, "--manifest", list),
	} {
		if diff := firstDifference(want, got); diff != "" {
			t.Fatalf("%s: %s", name, diff)
		}
	}
	t.Logf("%d scalar edge files, %d bytes identical on Go, source Node, sanitized native and emitted JS", span[1]-span[0], len(want))
}

func TestScalarEdges_000(t *testing.T) {
	t.Parallel()
	scalarEdgeShard(t, 0)
}
func TestScalarEdges_001(t *testing.T) {
	t.Parallel()
	scalarEdgeShard(t, 1)
}
func TestScalarEdges_002(t *testing.T) {
	t.Parallel()
	scalarEdgeShard(t, 2)
}
func TestScalarEdges_003(t *testing.T) {
	t.Parallel()
	scalarEdgeShard(t, 3)
}
