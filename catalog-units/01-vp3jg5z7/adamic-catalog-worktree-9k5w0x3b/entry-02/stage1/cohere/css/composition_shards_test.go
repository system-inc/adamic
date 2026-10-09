package css

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

const testCompositionMatchesGoShards = 16

func compositionShard(key string) int {
	sum := sha256.Sum256([]byte(key))
	return int(binary.LittleEndian.Uint64(sum[:8]) % testCompositionMatchesGoShards)
}

func compositionPartition(t *testing.T, lines []string) [][]string {
	t.Helper()
	shards := make([][]string, testCompositionMatchesGoShards)
	seen := make([]int, len(lines))
	// Occurrences remain separate cases, including duplicate corpus entries.
	indices := make([][]int, testCompositionMatchesGoShards)
	for i, line := range lines {
		shard := compositionShard(line)
		shards[shard] = append(shards[shard], line)
		indices[shard] = append(indices[shard], i)
	}
	for _, shard := range indices {
		for _, i := range shard {
			seen[i]++
		}
	}
	for i, count := range seen {
		if count != 1 {
			t.Fatalf("case %d occurs %d times", i, count)
		}
	}
	if len(shards) != testCompositionMatchesGoShards {
		t.Fatal("shard enumeration differs from constant")
	}
	t.Logf("union: %d live cases, each exactly once across %d shards", len(lines), len(shards))
	return shards
}

type compositionProducts struct {
	shards                             [][]string
	source, product, sanitized, oracle string
}

var compositionOnce sync.Once
var compositionBuilt compositionProducts

func compositionSetup(t *testing.T) compositionProducts {
	t.Helper()
	compositionOnce.Do(func() { compositionBuilt = buildCompositionProducts(t) })
	if compositionBuilt.oracle == "" {
		t.Fatal("composition setup did not complete")
	}
	return compositionBuilt
}
func buildCompositionProducts(t *testing.T) compositionProducts {
	started := time.Now()
	setupTimer := time.AfterFunc(75*time.Second, func() { panic("cooked: TestCompositionMatchesGo (setup) exceeded 75s") })
	defer setupTimer.Stop()
	cases := compositionCases(t)
	data, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	shards := compositionPartition(t, lines)
	// Shared products must outlive the first top-level shard's cleanup.
	directory, err := os.MkdirTemp("", "composition-oracle-")
	if err != nil {
		t.Fatal(err)
	}
	repo, _ := filepath.Abs(repository)
	side, _ := filepath.Abs("testdata/compose_side_test.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(repo, "cohere", "internal", "format", "css", "adamic_compose_side_test.go"): side}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(directory, "oracle")
	command := bounded(t, "go", "test", "-c", "-overlay="+overlayPath, "-o", oracle, "./internal/format/css")
	command.Dir = filepath.Join(repo, "cohere")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build Go composition oracle: %v\n%s", err, output)
	}
	source, _ := filepath.Abs("compose_main.ts")
	inputs := buildcache.Inputs{Name: "css-composition-lowered", Files: []string{"stage1/cohere/css", "stage1/cohere/selector", "stage1/cohere/values", "stage1/cohere/mediaquery", "stage1/cohere/cssstrings", "stage1/cohere/cssnumbers", "internal", "cohere", "go.mod", "go.work"}, Toolchain: []string{runtime.Version()}, Flags: []string{"C and JavaScript"}}
	product := buildcache.Product(t, inputs, func(dir string) error {
		program := lowered(t, source)
		if err := os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(program)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	c, err := os.ReadFile(filepath.Join(product, "program.c"))
	if err != nil {
		t.Fatal(err)
	}
	inputs.Name = "css-composition-native"
	inputs.Flags = append(native.Flags(native.Options{Sanitize: true}), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	nativeProduct := buildcache.Product(t, inputs, func(dir string) error {
		return native.Build(string(c), filepath.Join(dir, "port"), native.Options{Sanitize: true})
	})
	sanitized := filepath.Join(nativeProduct, "port")

	setupTimer.Stop()
	t.Logf("TestCompositionMatchesGo (setup): %.3fs", time.Since(started).Seconds())
	return compositionProducts{shards, source, product, sanitized, oracle}
}

func compositionRunShard(t *testing.T, index int) {
	t.Parallel()
	started := time.Now()
	timer := time.AfterFunc(75*time.Second, func() { panic("cooked: composition shard exceeded 75s") })
	defer timer.Stop()
	state := compositionSetup(t)
	shard := state.shards[index]
	dir := t.TempDir()
	cases := filepath.Join(dir, "cases.txt")
	answers := filepath.Join(dir, "answers.txt")
	if err := os.WriteFile(cases, []byte(strings.Join(shard, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	request, _ := json.Marshal(map[string]string{"Cases": cases, "Answers": answers})
	requestPath := filepath.Join(dir, "request.json")
	if err := os.WriteFile(requestPath, request, 0644); err != nil {
		t.Fatal(err)
	}
	result := execute(t, []string{"ADAMIC_PORT_REQUEST=" + requestPath}, state.oracle, "-test.run=^TestAdamicCompositionCases$", "-test.timeout=75s", "-test.v")
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("Go composition oracle: %d %s %s", result.exitCode, result.stdout, result.stderr)
	}
	data, err := os.ReadFile(answers)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(data), "\n") != len(shard)*2 {
		t.Fatal("Go answer count differs from live cases")
	}
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native ASan/UBSan", execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, state.sanitized, cases)},
		{"Node", onNode(t, state.source, cases)},
		{"JavaScript backend", onNode(t, filepath.Join(state.product, "program.mjs"), cases)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
			t.Fatalf("%s: %d %s", side.name, side.result.exitCode, side.result.stderr)
		}
		if difference := firstDifference(string(side.result.stdout), string(data)); difference != "" {
			t.Fatalf("%s: %s", side.name, difference)
		}
	}
	if report := leaks(t, nil, state.sanitized, cases); report != "" {
		t.Fatal(report)
	}
	t.Logf("shard-%03d: %.3fs cooked=false cases=%d (includes setup wait)", index, time.Since(started).Seconds(), len(shard))
}

func TestCompositionMatchesGoUnion(t *testing.T) {
	started := time.Now()
	timer := time.AfterFunc(75*time.Second, func() { panic("cooked: composition union exceeded 75s") })
	defer timer.Stop()
	cases := compositionCases(t)
	data, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(compositionShardFunctions) != testCompositionMatchesGoShards {
		t.Fatal("top-level shard enumeration differs from constant")
	}
	shards := compositionPartition(t, lines)
	// A single occurrence has a deliberately wrong answer. Run the same byte
	// comparator used by agreement on each partition, retaining duplicate cases.
	caught := 0
	owner := compositionShard(lines[0])
	offset := 0
	for i, shard := range shards {
		var got, want strings.Builder
		for j := range shard {
			fmt.Fprintf(&got, "case %d\nanswer\n", j)
			answer := "answer"
			if i == owner && j == 0 {
				answer = "planted disagreement"
			}
			fmt.Fprintf(&want, "case %d\n%s\n", j, answer)
			offset++
		}
		if firstDifference(got.String(), want.String()) != "" {
			caught++
			t.Logf("planted failure: one case disagrees; shard-%03d caught it", i)
		}
	}
	if caught != 1 || offset != len(lines) {
		t.Fatalf("planted failure caught by %d shards, union %d/%d", caught, offset, len(lines))
	}
	// Mutant-must-fail remains a whole-corpus condition, as before sharding.
	source, _ := filepath.Abs("compose_main.ts")
	baseline := onNode(t, source, cases)
	if baseline.exitCode != 0 || len(baseline.stderr) != 0 {
		t.Fatalf("Node baseline: %d %s", baseline.exitCode, baseline.stderr)
	}
	if keep := os.Getenv("ADAMIC_CSS_KEEP_COMPOSED"); keep != "" {
		if err := os.WriteFile(keep, baseline.stdout, 0644); err != nil {
			t.Fatal(err)
		}
	}
	for _, mutation := range mutants {
		mutated := portDirectory(t, &mutation)
		result := onNode(t, filepath.Join(mutated, "compose_main.ts"), cases)
		if result.exitCode != 0 || len(result.stderr) != 0 {
			t.Fatalf("composed mutant must terminate: %s", result.stderr)
		}
		if difference := firstDifference(string(result.stdout), string(baseline.stdout)); difference == "" {
			t.Fatalf("composition comparison missed mutant %s", mutation.name)
		} else {
			t.Logf("Node composition caught %s: %s", mutation.name, difference)
		}
	}
	t.Logf("union leaf: %.3fs cooked=false", time.Since(started).Seconds())
}

// Enumerate the original live corpus without its unrelated raw-parser benchmark.
func compositionCases(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cases := filepath.Join(dir, "cases.txt")
	repo, _ := filepath.Abs(repository)
	side, _ := filepath.Abs("testdata/composition_corpus_side_test.go")
	request, _ := json.Marshal(map[string]string{"cases": cases, "repository": repo, "fixtures": os.Getenv("ADAMIC_CSS_FIXTURES")})
	requestPath := filepath.Join(dir, "request.json")
	if err := os.WriteFile(requestPath, request, 0644); err != nil {
		t.Fatal(err)
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(repo, "cohere", "internal", "format", "css", "postcss", "adamic_port_side_test.go"): side}})
	overlayPath := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	command := bounded(t, "go", "test", "-timeout=75s", "-v", "-count=1", "-overlay="+overlayPath, "-run=^TestAdamicPortCases$", "./internal/format/css/postcss")
	command.Dir = filepath.Join(repo, "cohere")
	command.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("composition corpus: %v\n%s", err, output)
	} else {
		t.Logf("%s", output)
	}
	return cases
}

func TestCompositionMatchesGo_000(t *testing.T) { compositionRunShard(t, 0) }

func TestCompositionMatchesGo_001(t *testing.T) { compositionRunShard(t, 1) }

func TestCompositionMatchesGo_002(t *testing.T) { compositionRunShard(t, 2) }

func TestCompositionMatchesGo_003(t *testing.T) { compositionRunShard(t, 3) }

func TestCompositionMatchesGo_004(t *testing.T) { compositionRunShard(t, 4) }

func TestCompositionMatchesGo_005(t *testing.T) { compositionRunShard(t, 5) }

func TestCompositionMatchesGo_006(t *testing.T) { compositionRunShard(t, 6) }

func TestCompositionMatchesGo_007(t *testing.T) { compositionRunShard(t, 7) }

func TestCompositionMatchesGo_008(t *testing.T) { compositionRunShard(t, 8) }

func TestCompositionMatchesGo_009(t *testing.T) { compositionRunShard(t, 9) }

func TestCompositionMatchesGo_010(t *testing.T) { compositionRunShard(t, 10) }

func TestCompositionMatchesGo_011(t *testing.T) { compositionRunShard(t, 11) }

func TestCompositionMatchesGo_012(t *testing.T) { compositionRunShard(t, 12) }

func TestCompositionMatchesGo_013(t *testing.T) { compositionRunShard(t, 13) }

func TestCompositionMatchesGo_014(t *testing.T) { compositionRunShard(t, 14) }

func TestCompositionMatchesGo_015(t *testing.T) { compositionRunShard(t, 15) }

var compositionShardFunctions = [...]func(*testing.T){
	TestCompositionMatchesGo_000,
	TestCompositionMatchesGo_001,
	TestCompositionMatchesGo_002,
	TestCompositionMatchesGo_003,
	TestCompositionMatchesGo_004,
	TestCompositionMatchesGo_005,
	TestCompositionMatchesGo_006,
	TestCompositionMatchesGo_007,
	TestCompositionMatchesGo_008,
	TestCompositionMatchesGo_009,
	TestCompositionMatchesGo_010,
	TestCompositionMatchesGo_011,
	TestCompositionMatchesGo_012,
	TestCompositionMatchesGo_013,
	TestCompositionMatchesGo_014,
	TestCompositionMatchesGo_015,
}
