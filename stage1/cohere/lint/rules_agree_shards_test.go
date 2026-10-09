package lint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

const testRulesAgreeShards = 16

type rulesAgreeProducts struct {
	directory, oracle, binary, module string
	rows                              []string
	assignments                       [][]int
}

var rulesAgreeOnce sync.Once
var rulesAgreeLowerOnce sync.Once
var rulesAgreeNativeOnce sync.Once
var rulesAgreePrepared *rulesAgreeProducts
var rulesAgreeLoweredDirectory, rulesAgreeBinaryDirectory string
var rulesAgreeBuildInputs buildcache.Inputs

func rulesAgreeLowered(t *testing.T) string {
	t.Helper()
	rulesAgreeLowerOnce.Do(func() {
		directory := packageDirectory
		prepareRegistry(t, directory)
		inputs := buildcache.Inputs{
			Name:      "lint-rules-agree-lowered",
			Files:     []string{"stage1/typescript", "internal/load", "internal/lower", "internal/javascript", "internal/native", "cohere", "go.mod", "go.work"},
			Flags:     []string{"native.C", "javascript.JavaScript"},
			Toolchain: []string{runtime.Version()},
		}
		for _, path := range portFiles(t) {
			inputs.Files = append(inputs.Files, filepath.ToSlash(filepath.Join("stage1/cohere/lint", path)))
		}
		rulesAgreeBuildInputs = inputs
		rulesAgreeLoweredDirectory = buildcache.Product(t, inputs, func(output string) error {
			program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
			if err != nil {
				return err
			}
			result, err := lower.Lower(context.Background(), program)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(output, "lint.c"), []byte(native.C(result)), 0644); err != nil {
				return err
			}
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
			"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"),
			"ADAMIC_NATIVE_JOBS="+os.Getenv("ADAMIC_NATIVE_JOBS"),
			"ADAMIC_GATE_UNCACHED="+os.Getenv("ADAMIC_GATE_UNCACHED"))
		inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
		rulesAgreeBinaryDirectory = buildcache.Product(t, inputs, func(output string) error {
			source, err := os.ReadFile(filepath.Join(lowered, "lint.c"))
			if err != nil {
				return err
			}
			return native.Build(string(source), filepath.Join(output, "scanner"), native.Options{Sanitize: true, Split: true, Jobs: 4})
		})

	})
	if rulesAgreeBinaryDirectory == "" {
		t.Fatal("native setup did not complete")
	}
	return rulesAgreeBinaryDirectory
}

// Setup leaves run sequentially; shards share process-lifetime products.
func TestRulesAgreeSetupNative(t *testing.T) {
	started := time.Now()
	rulesAgreeNative(t)
	t.Logf("TestRulesAgree (setup native): %.3fs", time.Since(started).Seconds())
	if time.Since(started) >= 60*time.Second {
		t.Fatal("cooked: native setup over 60s")
	}
}
func TestRulesAgreeSetupCorpus(t *testing.T) {
	started := time.Now()
	rulesAgreeCorpus(t)
	t.Logf("TestRulesAgree (setup corpus): %.3fs", time.Since(started).Seconds())
	if time.Since(started) >= 60*time.Second {
		t.Fatal("cooked: corpus setup over 60s")
	}
}

func rulesAgreeCorpus(t *testing.T) *rulesAgreeProducts {
	t.Helper()
	rulesAgreeOnce.Do(func() {
		directory := packageDirectory
		// GoBuild has not landed on this main: preserve the original Go build.
		oracle := goOracle(t)
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
		captureInputs := buildcache.Inputs{
			Name:      "lint-rules-agree-capture",
			Files:     []string{"cohere", "stage1/cohere/lint", "go.mod", "go.work"},
			Flags:     []string{"asserted-cases-overlay", "-count=1", "-timeout=75s"},
			Toolchain: []string{runtime.Version()},
		}
		captureDirectory := buildcache.Product(t, captureInputs, func(output string) error {
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
func rulesAgreeSetup(t *testing.T) *rulesAgreeProducts {
	t.Helper()
	lowered := rulesAgreeLowered(t)
	binary := rulesAgreeNative(t)
	corpus := *rulesAgreeCorpus(t)
	corpus.binary = filepath.Join(binary, "scanner")
	corpus.module = filepath.Join(lowered, "lint.mjs")
	return &corpus
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
	t.Parallel()
	started := time.Now()
	deadline := time.AfterFunc(75*time.Second, func() { panic(fmt.Sprintf("%s cooked: over budget at 75s; split smaller", t.Name())) })
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
	products := rulesAgreeCorpus(t)
	t.Parallel()
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

// Plant one output disagreement on a real corpus case. Routing and the same
// byte comparator used by parity must detect it in exactly one owner shard.
func TestRulesAgreePlantedFailure(t *testing.T) {
	products := rulesAgreeSetup(t)
	t.Parallel()
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
	path := manifest(t, recoveryRows(t, products.oracle, []string{products.rows[planted]}))
	want := execute(t, "", products.oracle, "--manifest", path).output
	got := execute(t, "", products.binary, "--manifest", path).output
	if diff := difference(got, want); diff != "" {
		t.Fatal(diff)
	}
	got = append(append([]byte(nil), got...), []byte("planted disagreement\n")...)
	caught, owner := 0, -1
	for shard, indices := range products.assignments {
		for _, index := range indices {
			if index == planted && difference(got, want) != "" {
				caught++
				owner = shard
			}
		}
	}
	if caught != 1 {
		t.Fatalf("planted failure caught %d times", caught)
	}
	t.Logf("planted output disagreement: case %d caught exactly once by TestRulesAgree_%03d", planted, owner)
}

func TestRulesAgree_000(t *testing.T) { rulesAgreeRunShard(t, 0) }
func TestRulesAgree_001(t *testing.T) { rulesAgreeRunShard(t, 1) }
func TestRulesAgree_002(t *testing.T) { rulesAgreeRunShard(t, 2) }
func TestRulesAgree_003(t *testing.T) { rulesAgreeRunShard(t, 3) }
func TestRulesAgree_004(t *testing.T) { rulesAgreeRunShard(t, 4) }
func TestRulesAgree_005(t *testing.T) { rulesAgreeRunShard(t, 5) }
func TestRulesAgree_006(t *testing.T) { rulesAgreeRunShard(t, 6) }
func TestRulesAgree_007(t *testing.T) { rulesAgreeRunShard(t, 7) }
func TestRulesAgree_008(t *testing.T) { rulesAgreeRunShard(t, 8) }
func TestRulesAgree_009(t *testing.T) { rulesAgreeRunShard(t, 9) }
func TestRulesAgree_010(t *testing.T) { rulesAgreeRunShard(t, 10) }
func TestRulesAgree_011(t *testing.T) { rulesAgreeRunShard(t, 11) }
func TestRulesAgree_012(t *testing.T) { rulesAgreeRunShard(t, 12) }
func TestRulesAgree_013(t *testing.T) { rulesAgreeRunShard(t, 13) }
func TestRulesAgree_014(t *testing.T) { rulesAgreeRunShard(t, 14) }
func TestRulesAgree_015(t *testing.T) { rulesAgreeRunShard(t, 15) }

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
	// concurrent packages to the instance's four CPUs; each Go test has its
	// own 75-second deadline.
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
			_, failures[index] = run(root, environment, "go", "test", "-overlay="+overlayPath, "./internal/lint/rules/"+name, "-run", "^("+strings.Join(packages[name], "|")+")", "-count=1", "-timeout=75s")
			t.Logf("capture package %s: %.3fs cooked=%t", name, time.Since(started).Seconds(), time.Since(started) >= 60*time.Second)
			if failures[index] == nil && time.Since(started) >= 60*time.Second {
				failures[index] = fmt.Errorf("capture package %s cooked: over 60s; split smaller", name)
			}
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
