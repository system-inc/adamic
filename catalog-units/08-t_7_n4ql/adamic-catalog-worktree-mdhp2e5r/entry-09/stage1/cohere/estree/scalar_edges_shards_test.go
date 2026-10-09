package estree

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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

func scalarEdgeProducts(t *testing.T, main string) (string, string) {
	t.Helper()
	inputs := buildcache.Inputs{
		Name:      "scalar-edges-lowered",
		Files:     []string{"stage1/cohere/estree", "stage1/typescript", "internal", "cohere", "go.mod"},
		Toolchain: []string{runtime.Version()},
	}
	lowered := buildcache.Product(t, inputs, func(dir string) error {
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
	inputs.Name = "scalar-edges-sanitized-native"
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	inputs.Flags = append(native.Flags(native.Options{Sanitize: true}), []string{"Sanitize=true", "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS=" + os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED=" + os.Getenv("ADAMIC_GATE_UNCACHED")}...)
	product := buildcache.Product(t, inputs, func(dir string) error {
		source, err := os.ReadFile(filepath.Join(lowered, "port.c"))
		if err != nil {
			return err
		}
		return native.Build(string(source), filepath.Join(dir, "port"), native.Options{Sanitize: true})
	})
	return filepath.Join(product, "port"), filepath.Join(lowered, "port.mjs")
}

var scalarEdgeShared struct {
	once                   sync.Once
	oracle, binary, script string
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
		command := exec.Command("go", "build", "-overlay="+path, "-o", filepath.Join(dir, "oracle"), virtual)
		command.Dir = filepath.Join(repo, "cohere")
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("oracle: %w\n%s", err, output)
		}
		return nil
	})
	return filepath.Join(product, "oracle")
}

func scalarEdgeSetup(t *testing.T, main string) (string, string, string) {
	t.Helper()
	scalarEdgeShared.once.Do(func() {
		scalarEdgeShared.oracle = scalarEdgeOracle(t)
		scalarEdgeShared.binary, scalarEdgeShared.script = scalarEdgeProducts(t, main)
	})
	return scalarEdgeShared.oracle, scalarEdgeShared.binary, scalarEdgeShared.script
}

func runScalarEdges(t *testing.T, main string) {
	started := time.Now()
	scalarEdgeSetup(t, main)
	t.Logf("TestScalarEdges (setup): %.3fs", time.Since(started).Seconds())
}

func TestScalarEdgesUnion(t *testing.T) {
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
	oracle := scalarEdgeOracle(t)
	cases := scalarCases()
	caught := -1
	catches := 0
	for shard, span := range scalarEdgeRanges(len(cases)) {
		list := manifest(t, cases[span[0]:span[1]])
		want := execute(t, "", oracle, "--manifest", list)
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
		got := execute(t, "", oracle, "--manifest", list)
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
	t.Parallel()
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	oracle, binary, script := scalarEdgeSetup(t, main)
	cases := scalarCases()
	span := scalarEdgeRanges(len(cases))[shard]
	list := manifest(t, cases[span[0]:span[1]])
	want := execute(t, "", oracle, "--manifest", list)
	for name, got := range map[string][]byte{
		"source Node":      onNode(t, main, "--manifest", list),
		"sanitized native": execute(t, "", binary, "--manifest", list),
		"emitted JS":       onNode(t, script, "--manifest", list),
	} {
		if diff := firstDifference(want, got); diff != "" {
			t.Fatalf("%s: %s", name, diff)
		}
	}
	t.Logf("%d scalar edge files, %d bytes identical on Go, source Node, sanitized native and emitted JS", span[1]-span[0], len(want))
}

func TestScalarEdges_000(t *testing.T) { scalarEdgeShard(t, 0) }
func TestScalarEdges_001(t *testing.T) { scalarEdgeShard(t, 1) }
func TestScalarEdges_002(t *testing.T) { scalarEdgeShard(t, 2) }
func TestScalarEdges_003(t *testing.T) { scalarEdgeShard(t, 3) }
