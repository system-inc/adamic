package lint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/testgrain"
	"github.com/system-inc/adamic/internal/testguard"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

const testRulesAgreeShards = 16

type rulesAgreeProducts struct {
	directory, oracle, binary, module string
	rows                              []string
	assignments                       [][]int
}

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
				if name == ".git" || name == "testdata" || name == "performance" || name == "evidence" {
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
	return testgrain.Setup(t, "wave1/rulesAgreeLowered", func() (string, error) {
		return rulesAgreeLoweredPrepare(t), nil
	})
}
func rulesAgreeLoweredPrepare(t *testing.T) string {
	t.Helper()
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

	if rulesAgreeLoweredDirectory == "" {
		t.Fatal("lowered setup did not complete")
	}
	return rulesAgreeLoweredDirectory
}

func rulesAgreeNative(t *testing.T) string {
	t.Helper()
	return testgrain.Setup(t, "wave1/rulesAgreeNative", func() (string, error) {
		return rulesAgreeNativePrepare(t), nil
	})
}
func rulesAgreeNativePrepare(t *testing.T) string {
	t.Helper()
	lowered := rulesAgreeLowered(t)
	inputs := rulesAgreeBuildInputs
	inputs.Name = "lint-rules-agree-native-sanitized"
	inputs.Flags = append(native.Flags(native.Options{Sanitize: true, Split: true, Jobs: 4}),
		"Split=true", "Jobs=4",
		"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"),
		"ADAMIC_NATIVE_JOBS="+os.Getenv("ADAMIC_NATIVE_JOBS"),
		"ADAMIC_GATE_UNCACHED="+os.Getenv("ADAMIC_GATE_UNCACHED"))
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

	if rulesAgreeBinaryDirectory == "" {
		t.Fatal("native setup did not complete")
	}
	return rulesAgreeBinaryDirectory
}

var rulesAgreeOracleDirectory, rulesAgreeCaptureDirectory string

func rulesAgreeOracle(t *testing.T) string {
	t.Helper()
	return testgrain.Setup(t, "wave1/rulesAgreeOracle", func() (string, error) {
		return rulesAgreeOraclePrepare(

			// GoBuild is not present on this main. Cache the original overlay build
			// as a Product so a filtered shard never recompiles everyone's oracle.
			t), nil
	})
}
func rulesAgreeOraclePrepare(t *testing.T) string {
	t.Helper()
	directory := packageDirectory

	oracleInputs := buildcache.Inputs{
		Name:      "lint-rules-agree-go-oracle",
		Files:     append(rulesAgreeSourceInputs(t), "stage1/cohere/lint/testdata/oracle.go"),
		Flags:     []string{"go build", "overlay registry and rule adapters", "GOFLAGS=" + os.Getenv("GOFLAGS"), "GOWORK=" + os.Getenv("GOWORK")},
		Toolchain: []string{runtime.Version()},
	}
	rulesAgreeOracleDirectory = buildcache.Product(t, oracleInputs, func(output string) error {
		_, err := rulesAgreeGoOracleIn(t, directory, output)
		return err
	})

	if rulesAgreeOracleDirectory == "" {
		t.Fatal("oracle preparation failed")
	}
	return rulesAgreeOracleDirectory
}

func rulesAgreeCapture(t *testing.T) string {
	t.Helper()
	return testgrain.Setup(t, "wave1/rulesAgreeCapture", func() (string, error) {
		return rulesAgreeCapturePrepare(t), nil
	})
}
func rulesAgreeCapturePrepare(t *testing.T) string {
	t.Helper()
	directory := packageDirectory
	captureInputs := buildcache.Inputs{
		Name:      "lint-rules-agree-capture",
		Files:     []string{"cohere", "stage1/cohere/lint", "go.mod", "go.work"},
		Flags:     []string{"asserted-cases-overlay", "-count=1", "-timeout=0"},
		Toolchain: []string{runtime.Version()},
	}
	rulesAgreeCaptureDirectory = buildcache.Product(t, captureInputs, func(output string) error {
		captured, err := rulesAgreeCaptureUpstream(t, directory, output)
		if err != nil {
			return err
		}
		for index, row := range captured {
			fields := strings.SplitN(row, "\t", 2)
			relative, err := filepath.Rel(output, fields[0])
			if err != nil {
				return err
			}
			captured[index] = relative + "\t" + fields[1]
		}
		data, err := json.Marshal(captured)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(output, "rows.json"), data, 0644)
	})

	if rulesAgreeCaptureDirectory == "" {
		t.Fatal("capture preparation failed")
	}
	return rulesAgreeCaptureDirectory
}

func rulesAgreeCorpus(t *testing.T) *rulesAgreeProducts {
	t.Helper()
	return testgrain.Setup(t, "wave1/rulesAgreeCorpus", func() (*rulesAgreeProducts, error) {
		return rulesAgreeCorpusPrepare(t), nil
	})
}
func rulesAgreeCorpusPrepare(t *testing.T) *rulesAgreeProducts {
	t.Helper()
	directory := packageDirectory
	oracleDirectory := rulesAgreeOracle(t)
	oracle := filepath.Join(oracleDirectory, "oracle")
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
	if rulesAgreePrepared == nil {
		t.Fatal("corpus setup did not complete")
	}
	return rulesAgreePrepared
}

// Shared setup has no deadline; Loom still covers the whole unit.
// Filtered single-shard runs use this same preparation path and persistent cache.
func rulesAgreeSetup(t *testing.T) *rulesAgreeProducts {
	t.Helper()
	return testgrain.Setup(t, "wave1/rulesAgreeSetup", func() (*rulesAgreeProducts, error) {
		return rulesAgreeSetupPrepare(t), nil
	})
}

func rulesAgreeSetupPrepare(t *testing.T) *rulesAgreeProducts {
	t.Helper()
	if path := os.Getenv("ADAMIC_RULES_AGREE_PROBE"); path != "" {
		return rulesAgreeProbeProducts(t, path)
	}

	// Setup stages sharing t run sequentially: one stage cannot clean up another's children.
	corpus := rulesAgreeCorpus(t)
	lowered := rulesAgreeLowered(t)
	binary := rulesAgreeNative(t)
	ready := *corpus
	ready.binary = filepath.Join(binary, "scanner")
	ready.module = filepath.Join(lowered, "lint.mjs")
	return &ready
}

// The hash includes source, fixture basename, and all options. Temporary
// capture roots and corpus ordering never affect assignment.
func rulesAgreeAssignments(t *testing.T, rows []string) [][]int {
	t.Helper()
	identities := make([]string, len(rows))
	for index, row := range rows {
		fields := strings.SplitN(row, "\t", 2)
		source, err := os.ReadFile(fields[0])
		if err != nil {
			t.Fatal(err)
		}
		identity := fmt.Sprintf("%q\n%q\n", filepath.Base(fields[0]), source)
		if len(fields) == 2 {
			identity += fmt.Sprintf("%q", fields[1])
		}
		identities[index] = identity
	}
	return testgrain.Assign(identities, testRulesAgreeShards)
}

func rulesAgreeRunShard(t *testing.T, shard int) {
	t.Helper()
	products := rulesAgreeSetup(t)
	testgrain.Unit(t)
	t.Logf("shard-%03d: cases=%d", shard, len(products.assignments[shard]))
	var rows []string
	for _, index := range products.assignments[shard] {
		row := products.rows[index]
		if strings.HasSuffix(row, "\tunsupported-recovery") {
			grainLintCheckRecoveryRefusal(t, products.oracle, products.binary, products.directory, row)
		} else {
			rows = append(rows, row)
		}
	}
	if len(rows) > 0 {
		grainLintCompareWithJavaScript(t, products.oracle, products.binary, products.directory, manifest(t, grainLintRecoveryRows(t, products.oracle, rows)), products.module)
	}
}

func TestRulesAgreeUnion(t *testing.T) {
	t.Parallel()
	products := rulesAgreeCorpus(t)
	testgrain.Union(t, "TestRulesAgree", testRulesAgreeShards, products.assignments, len(products.rows))
	if len(products.rows) == 0 {
		t.Fatal("empty live corpus")
	}
	t.Logf("union: %d cases, each exactly once, %d top-level shards", len(products.rows), testRulesAgreeShards)
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
	testgrain.Unit(t)
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
	caught := make(map[int]bool)
	owner := -1
	for shard, indices := range products.assignments {
		for _, index := range indices {
			if index == planted {
				owner = shard
			}
		}
	}
	for shard := 0; shard < testRulesAgreeShards; shard++ {
		name := fmt.Sprintf("TestRulesAgree_%03d", shard)
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		command := testgrain.CommandContext(t, ctx, os.Args[0], "-test.run=^"+name+"$", "-test.v", "-test.timeout=10s")
		command.Env = append(os.Environ(), "ADAMIC_RULES_AGREE_PROBE="+path)
		output, err := command.CombinedOutput()
		expired := ctx.Err()
		cancel()
		if expired != nil {
			t.Fatalf("%s probe timed out: %v\n%s", name, expired, output)
		}
		if err != nil {
			if bytes.Count(output, []byte("--- FAIL: "+name)) != 1 || !bytes.Contains(output, []byte("emitted JavaScript:")) || !bytes.Contains(output, []byte(marker)) {
				t.Fatalf("wrong planted failure in %s: %v\n%s", name, err, output)
			}
			caught[shard] = true
		}
	}
	testgrain.CaughtByExactly(t, caught, owner)
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

func rulesAgreeGoOracleIn(t *testing.T, sourceRoot, directory string) (string, error) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return "", err
	}
	side, err := filepath.Abs(filepath.Join(packageDirectory, "testdata/oracle.go"))
	if err != nil {
		return "", err
	}
	descriptors, err := registry.Generate(sourceRoot)
	if err != nil {
		return "", err
	}
	replacements := map[string]string{}
	var virtualFiles []string
	var failure error
	add := func(name, source string) {
		virtual := filepath.Join(root, "adamic_lint_"+name+".go")
		absolute, err := filepath.Abs(source)
		if err != nil {
			failure = err
			return
		}
		replacements[virtual] = absolute
		virtualFiles = append(virtualFiles, virtual)
	}
	add("oracle", side)
	add("registry", filepath.Join(sourceRoot, ".generated/registry.go"))
	for _, d := range descriptors {
		add(strings.ReplaceAll(d.Slug, "-", "_"), filepath.Join(sourceRoot, "rules", d.Slug, "oracle.go"))
	}
	if failure != nil {
		return "", failure
	}
	overlay, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		return "", err
	}
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		return "", err
	}
	binary := filepath.Join(directory, "oracle")
	args := append([]string{"build", "-overlay=" + path, "-o", binary}, virtualFiles...)
	if _, err := rulesAgreeCaptureRun(t, root, nil, args...); err != nil {
		return "", err
	}
	return binary, nil
}

// Shared Go compilation and capture have no setup deadline.
// Preserve the child CPU guard and kill the entire compiler process group.
func rulesAgreeCaptureRun(t *testing.T, directory string, environment []string, args ...string) ([]byte, error) {
	command := testgrain.Command(t, "go", args...)
	command.Dir = directory
	command.Env = append(os.Environ(), environment...)
	command.Env = append(command.Env, "PWD="+directory)
	output, err := os.CreateTemp(sharedDirectory, "rules-agree-command-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(output.Name())
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	err = testguard.Run(command, testguard.Budget, testguard.Ceiling)
	if err != nil || len(commandDiagnostics("go", stderr.Bytes())) != 0 {
		return nil, fmt.Errorf("go %v: %v\n%s", args, err, &stderr)
	}
	return os.ReadFile(output.Name())
}

func rulesAgreeCaptureUpstream(t *testing.T, sourceRoot, directory string) ([]string, error) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return nil, err
	}
	harness := filepath.Join(root, "internal/lint/testing/rule_testing.go")
	data, err := os.ReadFile(harness)
	if err != nil {
		return nil, err
	}
	original := "return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}"
	replacement := "result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result"
	if strings.Count(string(data), original) != 1 {
		return nil, fmt.Errorf("capture overlay anchor changed")
	}
	side := filepath.Join(directory, "rule_testing.go")
	if err := os.WriteFile(side, []byte(strings.Replace(string(data), original, replacement, 1)), 0644); err != nil {
		return nil, err
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{harness: side}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		return nil, err
	}
	capture := filepath.Join(directory, "capture")
	descriptors, err := registry.Generate(sourceRoot)
	if err != nil {
		return nil, err
	}
	discovered := map[string]bool{}
	packages := map[string][]string{}
	for _, d := range descriptors {
		discovered[d.Name] = true
		packages[d.UpstreamPackage] = append(packages[d.UpstreamPackage], d.UpstreamTest)
	}
	var names []string
	for name := range packages {
		names = append(names, name)
	}
	sort.Strings(names)
	// Independent upstream packages capture into separate destinations. Bound
	// concurrent packages to the instance's four CPUs; capture is shared setup.
	var workers sync.WaitGroup
	slots := make(chan struct{}, 4)
	failures := make([]error, len(names))
	for index, name := range names {
		workers.Add(1)
		go func() {
			defer workers.Done()
			slots <- struct{}{}
			defer func() { <-slots }()
			destination := filepath.Join(capture, name)
			environment := []string{"COHERE_DOCS_CAPTURE=" + destination}
			started := time.Now()
			_, failures[index] = rulesAgreeCaptureRun(t, root, environment, "test", "-overlay="+overlayPath, "./internal/lint/rules/"+name, "-run", "^("+strings.Join(packages[name], "|")+")", "-count=1", "-timeout=0")
			t.Logf("capture package %s: %.3fs", name, time.Since(started).Seconds())
		}()
	}
	workers.Wait()
	for _, failure := range failures {
		if failure != nil {
			return nil, failure
		}
	}

	var files []string
	err = filepath.WalkDir(capture, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".jsonl") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	type record struct {
		Rule, File, Source, Outcome, FixedSource string
		Options                                  json.RawMessage
	}
	unique := map[string]record{}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		for _, line := range bytes.Split(data, []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var row record
			if err := json.Unmarshal(line, &row); err != nil {
				return nil, err
			}
			if !discovered[row.Rule] {
				continue
			}
			key := fmt.Sprintf("%s\t%s\t%+v\t%s", row.Rule, row.File, row.Options, row.Source)
			unique[key] = row
		}
	}
	var keys []string
	for key := range unique {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var rows []string
	for i, key := range keys {
		row := unique[key]
		// The case keeps its file name's directories, not only its base name: a rule that judges a
		// path (a utils folder, a page directory) reads them, and Go's capture recorded them.
		name := filepath.Clean(strings.TrimLeft(strings.ReplaceAll(row.File, "\\", "/"), "/"))
		if name == "." || name == "" || strings.HasPrefix(name, "..") {
			name = filepath.Base(name)
		}
		if name == "." || name == "" || name == ".." {
			name = "source.ts"
		}
		caseDirectory := filepath.Join(directory, fmt.Sprintf("case-%03d", i))
		path := filepath.Join(caseDirectory, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(row.Source), 0644); err != nil {
			return nil, err
		}
		var legacy struct {
			Mode, Null      string
			AllowEmptyCatch bool
		}
		if len(row.Options) > 0 && row.Options[0] == '{' {
			if err := json.Unmarshal(row.Options, &legacy); err != nil {
				return nil, err
			}
		}
		mode := ""
		if row.Rule == "@typescript-eslint/method-signature-style" {
			switch row.Source {
			case "type T = { m: => void };":
				mode = "recovery"
			case "interface I", "interface I { m(a: string): void;", "interface I { m<(a: string): void; }", "interface I { m<T(a: T): T; }":
				mode = "unsupported-recovery"
			}
		}
		if row.Rule == "no-div-regex" && (row.Source == "var a = /;" || row.Source == "var a = /" || row.Source == "var a = [/];" || row.Source == "if (/) {}" || row.Source == "var a = /=") {
			mode = "recovery"
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%t\t%s\t%s", path, row.Rule, legacy.Mode, legacy.Null, legacy.AllowEmptyCatch, string(row.Options), mode))
	}
	if len(rows) < 150 {
		return nil, fmt.Errorf("capture unexpectedly small: %d cases", len(rows))
	}
	return rows, nil
}

func grainLintExecute(t *testing.T, directory, name string, args ...string) execution {
	t.Helper()
	command := testgrain.Command(t, name, args...)
	command.Dir = directory
	output, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	started := time.Now()
	err = testguard.Run(command, testguard.Budget, testguard.Ceiling)
	duration := time.Since(started)
	if err != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return execution{data, duration}
}

func grainLintRecoveryRows(t *testing.T, oracle string, rows []string) []string {
	t.Helper()
	if len(rows) == 0 {
		return rows
	}
	answer := grainLintExecute(t, "", oracle, "--manifest", manifest(t, rows), "--diagnostics")
	flags := strings.Fields(string(answer.output))
	if len(flags) != len(rows) {
		t.Fatalf("diagnostics answered %d rows of %d", len(flags), len(rows))
	}
	result := make([]string, len(rows))
	for index, row := range rows {
		result[index] = row
		if flags[index] != "1" {
			continue
		}
		fields := strings.Split(row, "\t")
		for len(fields) < 7 {
			fields = append(fields, "")
		}
		if fields[6] == "" {
			fields[6] = "recovery"
		}
		result[index] = strings.Join(fields, "\t")
	}
	return result
}

func grainLintNode(t *testing.T, directory, manifest string, count bool) execution {
	t.Helper()
	prepareRegistry(t, directory)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", manifest}
	if count {
		args = append(args, "--count")
	}
	return grainLintExecute(t, "", "node", args...)
}

func grainLintRunJavaScript(t *testing.T, module, manifest string, count bool) execution {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"--disable-warning=ExperimentalWarning", runner, module, "--manifest", manifest}
	if count {
		args = append(args, "--count")
	}
	return grainLintExecute(t, "", "node", args...)
}

func grainLintCompareWithJavaScript(t *testing.T, oracle, binary, directory, path, module string) []byte {
	t.Helper()
	want := grainLintExecute(t, "", oracle, "--manifest", path)
	for _, side := range []struct {
		name string
		run  execution
	}{{"Node", grainLintNode(t, directory, path, false)}, {"emitted JavaScript", grainLintRunJavaScript(t, module, path, false)}, {"native", grainLintExecute(t, "", binary, "--manifest", path)}} {
		if diff := difference(side.run.output, want.output); diff != "" {
			t.Fatalf("%s: %s", side.name, diff)
		}
	}
	t.Logf("Go, Node, emitted JavaScript, native identical: %d bytes", len(want.output))
	return want.output
}

func grainLintCheckRecoveryRefusal(t *testing.T, oracle, binary, directory, row string) {
	t.Helper()
	recovered := manifest(t, []string{strings.TrimSuffix(row, "unsupported-recovery") + "recovery"})
	answer := grainLintExecute(t, "", oracle, "--manifest", recovered)
	t.Logf("Go recovered output: %s", answer.output)
	path := manifest(t, []string{row})
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name string
		args []string
	}{
		{binary, []string{"--manifest", path}},
		{"node", []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", path}},
	} {
		command := testgrain.Command(t, side.name, side.args...)
		output, err := os.CreateTemp(t.TempDir(), "recovery-refusal-")
		if err != nil {
			t.Fatal(err)
		}
		command.Stdout = output
		var stderr bytes.Buffer
		command.Stderr = &stderr
		err = testguard.Run(command, 2*time.Second, testguard.Ceiling)
		cpuGuard := strings.Contains(fmt.Sprint(err), "child CPU hang guard exceeded")
		output.Close()
		if err == nil || (!cpuGuard && !strings.Contains(stderr.String(), "adamic: panic:")) {
			t.Fatalf("expected parser refusal from %s, got %v: %s", side.name, err, stderr.String())
		}
		t.Logf("explicit unsupported recovery: %s: CPU-guard=%t: %v: %s", side.name, cpuGuard, err, stderr.String())
	}
}
