package lint

import (
	"bytes"
	"context"
	"encoding/json"

	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/lint/oracles"
	"hash/fnv"
	"os"

	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const testRulesAgreeShards = 16
const rulesAgreeKill = 90 * time.Second

type rulesAgreeProducts struct {
	directory, oracle, binary, module string
	rows                              []string
	assignments                       [][]int
}

var rulesAgreeSetupOnce sync.Once
var rulesAgreeReady *rulesAgreeProducts

var rulesAgreeOnce sync.Once
var rulesAgreeLowerOnce sync.Once
var rulesAgreeNativeOnce sync.Once
var rulesAgreePrepared *rulesAgreeProducts
var rulesAgreeLoweredDirectory, rulesAgreeBinaryDirectory string
var rulesAgreeBuildInputs buildcache.Inputs

// Hash production sources and embedded inputs, including the port's generated
// registry and compiler dependencies. Test moves and captured test corpora are
// not inputs to lowering and must not turn a warm fetch into a cold rebuild.
func rulesAgreeSourceInputs(t *testing.T) []string {
	t.Helper()
	return rulesAgreeSourceInputsAt(t, filepath.Clean(repository))
}

func rulesAgreeSourceInputsAt(t *testing.T, root string) []string {
	t.Helper()
	files := []string{"go.mod", "go.work"}
	for _, directory := range []string{"internal", "bridge", "stage1/typescript", "stage1/cohere/lint", "cohere"} {
		err := filepath.WalkDir(filepath.Join(root, directory), func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			name := entry.Name()
			if entry.IsDir() {
				if name == ".git" || name == "testdata" || name == "performance" || name == "evidence" || lintRuleSubdirectory(path) {
					return filepath.SkipDir
				}
				return nil
			}
			if name == ".git" || strings.HasSuffix(name, "_test.go") {
				return nil
			}
			switch filepath.Ext(name) {
			case ".go", ".a", ".ts", ".json", ".c", ".h", ".inc", ".mod", ".sum":
				relative, err := filepath.Rel(root, path)
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
	sort.Strings(files)
	return files
}

// Protect the warm-fetch contract: test-only moves preserve the key, while
// actual port, registry, lowering, and embedded runtime edits invalidate it.
func TestRulesAgreeLoweringCacheKey(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, directory := range []string{"internal", "bridge", "stage1/typescript", "stage1/cohere/lint", "cohere"} {
		if err := os.MkdirAll(filepath.Join(root, directory), 0755); err != nil {
			t.Fatal(err)
		}
	}
	sources := []string{"go.mod", "go.work", "stage1/cohere/lint/main.a", "stage1/cohere/lint/rules/probe/rule.ts", "stage1/cohere/lint/.generated/registry.ts", "stage1/cohere/lint/rules/probe/rule.json", "internal/lower/lower.go", "internal/native/runtime/heap.h", "internal/load/prelude.d.ts", "cohere/TypeScript/tsc/checker.go"}
	write := func(name, content string) {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range sources {
		write(name, "original")
	}
	key := func() string {
		result, err := buildcache.Key(root, buildcache.Inputs{Name: "lowering-cache-proof", Files: rulesAgreeSourceInputsAt(t, root), Flags: []string{"default lowering"}, Toolchain: []string{runtime.Version()}})
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	original := key()
	write("internal/lower/new_shards_test.go", "package lower // a moved test")
	write("stage1/cohere/lint/new_shards_test.go", "package lint // a moved test")
	if key() != original {
		t.Fatal("test-only moves invalidated the lowered product")
	}
	// A rule's evidence, validation and completion notes are not its sources: their churn keeps the key (#s57yq6y).
	write("stage1/cohere/lint/rules/probe/evidence/report-mutant.json", "{}")
	write("stage1/cohere/lint/rules/probe/validation/run.json", "{}")
	write("stage1/cohere/lint/rules/probe/evidence-completion/notes.json", "{}")
	if key() != original {
		t.Fatal("a rule's evidence, validation or completion notes invalidated the lowered product")
	}
	for _, name := range sources {
		write(name, "changed")
		if key() == original {
			t.Fatalf("lowering input %s did not invalidate the product", name)
		}
		write(name, "original")
	}
	write("stage1/cohere/lint/new_module.a", "new production source")
	if key() == original {
		t.Fatal("new port source did not invalidate the product")
	}
}

func rulesAgreeLowered(t *testing.T) string {
	t.Helper()
	rulesAgreeLowerOnce.Do(func() {
		directory := packageDirectory
		prepareRegistry(t, directory)
		inputs := buildcache.Inputs{
			Name:      "lint-rules-agree-lowered",
			Files:     rulesAgreeSourceInputs(t),
			Flags:     []string{"load.Load: default checker mode", "lower.Lower: default options", "native.C", "javascript.JavaScript"},
			Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
		}
		rulesAgreeBuildInputs = inputs
		rulesAgreeLoweredDirectory = buildcache.Product(t, inputs, func(output string) error {
			phase := time.Now()
			program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
			if err != nil {
				return err
			}
			t.Logf("rules load: %.3fs", time.Since(phase).Seconds())
			phase = time.Now()
			result, err := lower.Lower(context.Background(), program)
			if err != nil {
				return err
			}
			t.Logf("rules lower: %.3fs", time.Since(phase).Seconds())
			phase = time.Now()
			if err := os.WriteFile(filepath.Join(output, "lint.c"), []byte(native.C(result)), 0644); err != nil {
				return err
			}
			t.Logf("rules native emit: %.3fs", time.Since(phase).Seconds())
			phase = time.Now()
			defer func() { t.Logf("rules JavaScript emit: %.3fs", time.Since(phase).Seconds()) }()
			return os.WriteFile(filepath.Join(output, "lint.mjs"), []byte(javascript.JavaScript(result)), 0644)
		})

	})
	if rulesAgreeLoweredDirectory == "" {
		t.Fatal("lowered setup did not complete")
	}
	return rulesAgreeLoweredDirectory
}

func rulesAgreeNative(t *testing.T) string {
	t.Helper()
	lowered := rulesAgreeLowered(t)
	rulesAgreeNativeOnce.Do(func() {
		inputs := rulesAgreeBuildInputs
		inputs.Name = "lint-rules-agree-native-sanitized"
		inputs.Flags = append(native.Flags(native.Options{Sanitize: true, Split: true, Jobs: 4}),
			"Split=true", "Jobs=4",
			"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
		inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
		rulesAgreeBinaryDirectory = buildcache.Product(t, inputs, func(output string) error {
			source, err := os.ReadFile(filepath.Join(lowered, "lint.c"))
			if err != nil {
				return err
			}
			started := time.Now()
			defer func() { t.Logf("rules clang: %.3fs", time.Since(started).Seconds()) }()
			return native.Build(string(source), filepath.Join(output, "scanner"), native.Options{Sanitize: true, Split: true, Jobs: 4})
		})

	})
	if rulesAgreeBinaryDirectory == "" {
		t.Fatal("native setup did not complete")
	}
	return rulesAgreeBinaryDirectory
}

func rulesAgreeCapture(t *testing.T) string { t.Helper(); return oracles.RulesAgreeCapture(t) }

func rulesAgreeCorpus(t *testing.T) *rulesAgreeProducts {
	t.Helper()
	rulesAgreeOnce.Do(func() {
		directory := packageDirectory
		oracle := lintGoOracleProduct(t)
		corpus, err := os.MkdirTemp(sharedDirectory, "rules-agree-")
		if err != nil {
			t.Fatal(err)
		}
		var rows []string
		for index, row := range generated(t) {
			fields := strings.SplitN(row, "\t", 2)
			source, err := os.ReadFile(fields[0])
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(corpus, fmt.Sprintf("generated-%03d", index), filepath.Base(fields[0]))
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, source, 0644); err != nil {
				t.Fatal(err)
			}
			if len(fields) == 2 {
				path += "\t" + fields[1]
			}
			rows = append(rows, path)
		}
		captureDirectory := rulesAgreeCapture(t)
		data, err := os.ReadFile(filepath.Join(captureDirectory, "rows.json"))
		if err != nil {
			t.Fatal(err)
		}
		var captured []string
		if err := json.Unmarshal(data, &captured); err != nil {
			t.Fatal(err)
		}
		for _, row := range captured {
			fields := strings.SplitN(row, "\t", 2)
			rows = append(rows, filepath.Join(captureDirectory, fields[0])+"\t"+fields[1])
		}

		assignments := rulesAgreeAssignments(t, rows)
		rulesAgreePrepared = &rulesAgreeProducts{directory: directory, oracle: oracle, rows: rows, assignments: assignments}
	})
	if rulesAgreePrepared == nil {
		t.Fatal("corpus setup did not complete")
	}
	return rulesAgreePrepared
}

// Shared setup has no deadline; Loom still covers the whole unit.
// Filtered single-shard runs use this same preparation path and persistent cache.
func rulesAgreeSetup(t *testing.T) *rulesAgreeProducts {
	t.Helper()
	if path := os.Getenv("ADAMIC_RULES_AGREE_PROBE"); path != "" {
		return rulesAgreeProbeProducts(t, path)
	}
	rulesAgreeSetupOnce.Do(func() {
		started := time.Now()
		var corpus *rulesAgreeProducts
		var binary, lowered string
		var workers sync.WaitGroup
		workers.Add(2)
		go func() {
			defer workers.Done()
			lowered = rulesAgreeLowered(t)
			binary = rulesAgreeNative(t)
		}()
		go func() {
			defer workers.Done()
			corpus = rulesAgreeCorpus(t)
		}()
		workers.Wait()
		if t.Failed() || corpus == nil {
			t.Fatal("RulesAgree preparation failed")
		}
		ready := *corpus
		ready.binary = filepath.Join(binary, "scanner")
		ready.module = filepath.Join(lowered, "lint.mjs")
		rulesAgreeReady = &ready
		t.Logf("TestRulesAgree_Setup: %.3fs", time.Since(started).Seconds())
	})
	if rulesAgreeReady == nil {
		t.Fatal("shared RulesAgree setup did not complete")
	}
	return rulesAgreeReady
}

// The hash includes source, fixture basename, and all options. Temporary
// capture roots and corpus ordering never affect assignment.
func rulesAgreeAssignments(t *testing.T, rows []string) [][]int {
	t.Helper()
	assignments := make([][]int, testRulesAgreeShards)
	for index, row := range rows {
		fields := strings.SplitN(row, "\t", 2)
		source, err := os.ReadFile(fields[0])
		if err != nil {
			t.Fatal(err)
		}
		hash := fnv.New64a()
		fmt.Fprintf(hash, "%q\n%q\n", filepath.Base(fields[0]), source)
		if len(fields) == 2 {
			fmt.Fprintf(hash, "%q", fields[1])
		}
		shard := int(hash.Sum64() % testRulesAgreeShards)
		assignments[shard] = append(assignments[shard], index)
	}
	return assignments
}

func rulesAgreeRunShard(t *testing.T, shard int) {
	t.Helper()
	products := rulesAgreeSetup(t)
	started := time.Now()
	deadline := time.AfterFunc(rulesAgreeKill, func() { panic(fmt.Sprintf("%s cooked: over budget at 90s; split smaller", t.Name())) })
	defer deadline.Stop()
	defer func() {
		t.Logf("shard-%03d: %.3fs cooked=%t cases=%d", shard, time.Since(started).Seconds(), time.Since(started) >= 60*time.Second, len(products.assignments[shard]))
	}()
	var rows []string
	for _, index := range products.assignments[shard] {
		row := products.rows[index]
		if strings.HasSuffix(row, "\tunsupported-recovery") {
			checkRecoveryRefusal(t, products.oracle, products.binary, products.directory, row)
		} else {
			rows = append(rows, row)
		}
	}
	if len(rows) > 0 {
		compareWithJavaScript(t, products.oracle, products.binary, products.directory, manifest(t, recoveryRows(t, products.oracle, rows)), products.module)
	}
	if time.Since(started) >= 60*time.Second {
		t.Fatal("cooked: shard over 60s budget; split smaller")
	}
}

func TestRulesAgreeUnion(t *testing.T) {
	t.Parallel()
	started := time.Now()
	products := rulesAgreeCorpus(t)
	t.Logf("union (setup corpus): %.3fs", time.Since(started).Seconds())
	started = time.Now()
	defer func() { t.Logf("union (own work): %.3fs", time.Since(started).Seconds()) }()
	functions := []func(*testing.T){TestRulesAgree_000, TestRulesAgree_001, TestRulesAgree_002, TestRulesAgree_003, TestRulesAgree_004, TestRulesAgree_005, TestRulesAgree_006, TestRulesAgree_007, TestRulesAgree_008, TestRulesAgree_009, TestRulesAgree_010, TestRulesAgree_011, TestRulesAgree_012, TestRulesAgree_013, TestRulesAgree_014, TestRulesAgree_015}
	if len(functions) != testRulesAgreeShards || len(products.assignments) != testRulesAgreeShards {
		t.Fatal("shard enumeration differs from const")
	}
	seen := make([]int, len(products.rows))
	for _, indices := range products.assignments {
		for _, index := range indices {
			seen[index]++
		}
	}
	for index, count := range seen {
		if count != 1 {
			t.Fatalf("case %d covered %d times", index, count)
		}
	}
	if len(seen) == 0 {
		t.Fatal("empty live corpus")
	}
	t.Logf("union: %d cases, each exactly once, %d top-level shards", len(seen), testRulesAgreeShards)
}

// Probe children run the ordinary shard selection and comparator on one real
// corpus case. The original hash still assigns it to exactly one owner.
type rulesAgreeProbe struct {
	Directory, Oracle, Binary, Module string
	Rows                              []string
}

func rulesAgreeProbeProducts(t *testing.T, path string) *rulesAgreeProducts {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var probe rulesAgreeProbe
	if err := json.Unmarshal(data, &probe); err != nil {
		t.Fatal(err)
	}
	return &rulesAgreeProducts{directory: probe.Directory, oracle: probe.Oracle, binary: probe.Binary, module: probe.Module, rows: probe.Rows, assignments: rulesAgreeAssignments(t, probe.Rows)}
}

func TestRulesAgreePlantedFailure(t *testing.T) {
	t.Parallel()
	products := rulesAgreeSetup(t)
	started := time.Now()
	defer func() { t.Logf("planted failure (own work): %.3fs", time.Since(started).Seconds()) }()
	planted := -1
	for index, row := range products.rows {
		if !strings.HasSuffix(row, "\tunsupported-recovery") {
			planted = index
			break
		}
	}
	if planted < 0 {
		t.Fatal("no supported case to plant")
	}
	data, err := os.ReadFile(products.module)
	if err != nil {
		t.Fatal(err)
	}
	module := filepath.Join(t.TempDir(), "planted.mjs")
	marker := "planted RulesAgree disagreement"
	if err := os.WriteFile(module, append(data, []byte(fmt.Sprintf("\nconsole.log(%q);\n", marker))...), 0644); err != nil {
		t.Fatal(err)
	}
	data, err = json.Marshal(rulesAgreeProbe{products.directory, products.oracle, products.binary, module, []string{products.rows[planted]}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "products.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	caught, owner := 0, -1
	for shard := 0; shard < testRulesAgreeShards; shard++ {
		name := fmt.Sprintf("TestRulesAgree_%03d", shard)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		command := dotARenameChild(ctx, "-test.run=^"+name+"$", "-test.v", "-test.timeout=10s")
		command.Env = append(os.Environ(), "ADAMIC_RULES_AGREE_PROBE="+path)
		// Read through childguard.Capture for dotARenameChild's reason, stdout then stderr.
		stdout, stderr, err := childguard.Capture(ctx, command)
		output := append(stdout, stderr...)
		expired := ctx.Err()
		cancel()
		if expired != nil {
			t.Fatalf("%s probe timed out: %v\n%s", name, expired, output)
		}
		if err != nil {
			if bytes.Count(output, []byte("--- FAIL: "+name)) != 1 || !bytes.Contains(output, []byte("emitted JavaScript:")) || !bytes.Contains(output, []byte(marker)) {
				t.Fatalf("wrong planted failure in %s: %v\n%s", name, err, output)
			}
			caught++
			owner = shard
		}
	}
	if caught != 1 {
		t.Fatalf("planted failure caught %d times", caught)
	}
	t.Logf("planted output disagreement: case %d caught exactly once by TestRulesAgree_%03d", planted, owner)
}

func TestRulesAgree_000(t *testing.T) {
	t.Parallel()
	rulesAgreeRunShard(t, 0)
}
func TestRulesAgree_001(t *testing.T) {
	t.Parallel()
	rulesAgreeRunShard(t, 1)
}
func TestRulesAgree_002(t *testing.T) {
	t.Parallel()
	rulesAgreeRunShard(t, 2)
}
func TestRulesAgree_003(t *testing.T) {
	t.Parallel()
	rulesAgreeRunShard(t, 3)
}
func TestRulesAgree_004(t *testing.T) {
	t.Parallel()
	rulesAgreeRunShard(t, 4)
}
func TestRulesAgree_005(t *testing.T) {
	t.Parallel()
	rulesAgreeRunShard(t, 5)
}
func TestRulesAgree_006(t *testing.T) {
	t.Parallel()
	rulesAgreeRunShard(t, 6)
}
func TestRulesAgree_007(t *testing.T) {
	t.Parallel()
	rulesAgreeRunShard(t, 7)
}
func TestRulesAgree_008(t *testing.T) {
	t.Parallel()
	rulesAgreeRunShard(t, 8)
}
func TestRulesAgree_009(t *testing.T) {
	t.Parallel()
	rulesAgreeRunShard(t, 9)
}
func TestRulesAgree_010(t *testing.T) {
	t.Parallel()
	rulesAgreeRunShard(t, 10)
}
func TestRulesAgree_011(t *testing.T) {
	t.Parallel()
	rulesAgreeRunShard(t, 11)
}
func TestRulesAgree_012(t *testing.T) {
	t.Parallel()
	rulesAgreeRunShard(t, 12)
}
func TestRulesAgree_013(t *testing.T) {
	t.Parallel()
	rulesAgreeRunShard(t, 13)
}
func TestRulesAgree_014(t *testing.T) {
	t.Parallel()
	rulesAgreeRunShard(t, 14)
}
func TestRulesAgree_015(t *testing.T) {
	t.Parallel()
	rulesAgreeRunShard(t, 15)
}
