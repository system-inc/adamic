package lint

import (
	"bytes"
	"context"
	"encoding/gob"
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
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/testgrain"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

const testNodeTableIsLinkOnlyShards = 8

var nodeTableLowerDirectory, nodeTableEmitDirectory, nodeTableNativePath string

var nodeTableRows []string
var nodeTableAssignments [][]int

func nodeTableSourceInputs(t *testing.T) []string {
	t.Helper()
	return nodeTableSourceInputsAt(t, filepath.Clean(repository))
}

func nodeTableSourceInputsAt(t *testing.T, root string) []string {
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

func nodeTableInputs(t *testing.T) buildcache.Inputs {
	return buildcache.Inputs{Name: "lint-node-table-lowered-ir", Files: nodeTableSourceInputs(t), Flags: []string{"load.Load default", "lower.Lower default", "gob IR"}, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH}}
}
func nodeTableLowered(t *testing.T) string {
	return testgrain.Setup(t, "node-table/nodeTableLowered", func() (string, error) { return nodeTableLoweredPrepare(t), nil })
}
func nodeTableLoweredPrepare(t *testing.T) string {
	t.Helper()
	inputs := buildcache.Inputs{Name: "lint-node-table-lowered-ir", Files: nodeTableSourceInputs(t), Flags: []string{"load.Load default", "lower.Lower default", "gob IR"}, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH}}
	lowered := buildcache.Product(t, inputs, func(out string) error {
		program, err := load.Load([]string{filepath.Join(packageDirectory, "main.ts")})
		if err != nil {
			return err
		}
		result, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		file, err := os.Create(filepath.Join(out, "program.gob"))
		if err != nil {
			return err
		}
		defer file.Close()
		return gob.NewEncoder(file).Encode(result)
	})
	nodeTableLowerDirectory = lowered
	if nodeTableLowerDirectory == "" {
		t.Fatal("nodeTableLowered setup incomplete")
	}
	return nodeTableLowerDirectory
}
func nodeTableEmitted(t *testing.T) string {
	return testgrain.Setup(t, "node-table/nodeTableEmitted", func() (string, error) { return nodeTableEmittedPrepare(t), nil })
}
func nodeTableEmittedPrepare(t *testing.T) string {
	t.Helper()
	lowered := nodeTableLowered(t)
	inputs := nodeTableInputs(t)
	inputs.Name = "lint-node-table-emitted-c-serial-v2"
	inputs.Flags = append(inputs.Flags, "native.C")
	emitted := buildcache.Product(t, inputs, func(out string) error {
		file, err := os.Open(filepath.Join(lowered, "program.gob"))
		if err != nil {
			return err
		}
		defer file.Close()
		var program ir.Program
		if err := gob.NewDecoder(file).Decode(&program); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(out, "lint.c"), []byte(native.C(&program)), 0644)
	})
	nodeTableEmitDirectory = emitted
	if nodeTableEmitDirectory == "" {
		t.Fatal("nodeTableEmitted setup incomplete")
	}
	return nodeTableEmitDirectory
}
func nodeTableNative(t *testing.T) string {
	return testgrain.Setup(t, "node-table/nodeTableNative", func() (string, error) { return nodeTableNativePrepare(t), nil })
}
func nodeTableNativePrepare(t *testing.T) string {
	t.Helper()
	emitted := nodeTableEmitted(t)
	inputs := nodeTableInputs(t)
	inputs.Name = "lint-node-table-native"
	inputs.Flags = append(native.Flags(native.Options{Split: true, Jobs: 4}), "Split=true", "Jobs=4", "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS="+os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED="+os.Getenv("ADAMIC_GATE_UNCACHED"))
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	binary := buildcache.Product(t, inputs, func(out string) error {
		source, err := os.ReadFile(filepath.Join(emitted, "lint.c"))
		if err != nil {
			return err
		}
		return native.Build(string(source), filepath.Join(out, "scanner"), native.Options{Split: true, Jobs: 4})
	})
	nodeTableNativePath = filepath.Join(binary, "scanner")
	if nodeTableNativePath == "" {
		t.Fatal("nodeTableNative setup incomplete")
	}
	return nodeTableNativePath
}

// TestNodeTableIsLinkOnly requires identical output with every node-table row
// copied and attached to nothing. The live corpus and all original options and
// recovery classifications are retained; only manifest batching changes.
// Not parallel: prepares process-lived corpus and build products before parallel shards.
func TestNodeTableIsLinkOnly(t *testing.T) { nodeTableSetup(t) }

func nodeTableSetup(t *testing.T) {
	t.Helper()
	testgrain.Setup(t, "wave1/nodeTableSetup", func() (bool, error) {
		nodeTableSetupPrepare(t)
		return true, nil
	})
}
func nodeTableSetupPrepare(t *testing.T) {
	t.Helper()

	prepareRegistry(t, packageDirectory)
	// GoBuild is absent on this main: retain the original overlay builder.
	oracleInputs := buildcache.Inputs{Name: "lint-node-table-oracle", Files: append(nodeTableSourceInputs(t), "stage1/cohere/lint/testdata/oracle.go"), Flags: []string{"go build overlay registry and rule adapters", "GOFLAGS=" + os.Getenv("GOFLAGS"), "GOWORK=" + os.Getenv("GOWORK")}, Toolchain: []string{runtime.Version()}}
	oracleProduct := buildcache.Product(t, oracleInputs, func(out string) error { _, err := rulesAgreeGoOracleIn(t, packageDirectory, out); return err })
	oracle := filepath.Join(oracleProduct, "oracle")
	// Copy generated cases out of the setup test's TempDir: top-level parallel
	// shards outlive that test. Captured products already have cache lifetime.
	corpus, err := os.MkdirTemp(sharedDirectory, "node-table-")
	if err != nil {
		t.Fatal(err)
	}
	for index, row := range generated(t) {
		fields := strings.SplitN(row, "\t", 2)
		data, err := os.ReadFile(fields[0])
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(corpus, fmt.Sprintf("case-%03d", index), filepath.Base(fields[0]))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
		if len(fields) == 2 {
			path += "\t" + fields[1]
		}
		nodeTableRows = append(nodeTableRows, path)
	}
	captureInputs := buildcache.Inputs{Name: "lint-node-table-capture", Files: []string{"cohere", "stage1/cohere/lint/testdata", "go.mod", "go.work"}, Flags: []string{"captureUpstream all asserted cases", "three workers", "go test -p=1 -count=1 -timeout=0"}, Toolchain: []string{runtime.Version()}}
	// Registry descriptors decide which upstream assertions are in the live corpus.
	for _, file := range portFiles(t) {
		if strings.HasSuffix(file, "rule.json") {
			captureInputs.Files = append(captureInputs.Files, "stage1/cohere/lint/"+file)
		}
	}
	captured := buildcache.Product(t, captureInputs, func(out string) error {
		rows, err := nodeTableCaptureUpstream(t, packageDirectory, out)
		if err != nil {
			return err
		}
		for i, row := range rows {
			fields := strings.SplitN(row, "\t", 2)
			relative, err := filepath.Rel(out, fields[0])
			if err != nil {
				return err
			}
			rows[i] = relative + "\t" + fields[1]
		}
		data, err := json.Marshal(rows)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(out, "rows.json"), data, 0644)
	})
	data, err := os.ReadFile(filepath.Join(captured, "rows.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []string
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if strings.HasSuffix(row, "\tunsupported-recovery") {
			continue
		}
		fields := strings.SplitN(row, "\t", 2)
		nodeTableRows = append(nodeTableRows, filepath.Join(captured, fields[0])+"\t"+fields[1])
	}
	nodeTableRows = grainLintRecoveryRows(t, oracle, nodeTableRows)
	identities := make([]string, len(nodeTableRows))
	for index, row := range nodeTableRows {
		fields := strings.SplitN(row, "\t", 2)
		source, err := os.ReadFile(fields[0])
		if err != nil {
			t.Fatal(err)
		}
		key := fmt.Sprintf("%q\n%q\n", filepath.Base(fields[0]), source)
		if len(fields) == 2 {
			key += fields[1]
		}
		identities[index] = key
	}
	nodeTableAssignments = testgrain.Assign(identities, testNodeTableIsLinkOnlyShards)
}

func nodeTableShard(key string) int {
	for shard, indices := range testgrain.Assign([]string{key}, testNodeTableIsLinkOnlyShards) {
		if len(indices) > 0 {
			return shard
		}
	}
	panic("unassigned node table identity")
}

func nodeTableEqual(got, want []byte) error {
	if diff := difference(got, want); diff != "" {
		return fmt.Errorf("output changed with unattached rows in the node table: %s", diff)
	}
	return nil
}

func nodeTableRunShard(t *testing.T, shard int) {
	t.Helper()
	if os.Getenv("ADAMIC_NODE_TABLE_PROBE") == "1" {
		testgrain.Unit(t)
		if shard == nodeTableShard("case 7") {
			if err := nodeTableEqual([]byte("planted disagreement"), []byte("case 7")); err != nil {
				t.Fatal(err)
			}
		}
		return
	}

	// Setup stages sharing t run sequentially: each owns and finishes its child groups.
	nodeTableSetup(t)
	binary := nodeTableNative(t)
	testgrain.Unit(t)

	rows := make([]string, 0, len(nodeTableAssignments[shard]))
	for _, index := range nodeTableAssignments[shard] {
		rows = append(rows, nodeTableRows[index])
	}
	path := manifest(t, rows)
	plain := grainLintExecute(t, "", binary, "--manifest", path)
	t.Logf("shard %03d: cases=%d bytes=%d", shard, len(rows), len(plain.output))
	junk := grainLintExecute(t, "", binary, "--manifest", path, "--junk-rows")
	if err := nodeTableEqual(junk.output, plain.output); err != nil {
		t.Fatal(err)
	}

}
func TestNodeTableIsLinkOnly_000(t *testing.T) { t.Parallel(); nodeTableRunShard(t, 0) }
func TestNodeTableIsLinkOnly_001(t *testing.T) { t.Parallel(); nodeTableRunShard(t, 1) }
func TestNodeTableIsLinkOnly_002(t *testing.T) { t.Parallel(); nodeTableRunShard(t, 2) }
func TestNodeTableIsLinkOnly_003(t *testing.T) { t.Parallel(); nodeTableRunShard(t, 3) }
func TestNodeTableIsLinkOnly_004(t *testing.T) { t.Parallel(); nodeTableRunShard(t, 4) }
func TestNodeTableIsLinkOnly_005(t *testing.T) { t.Parallel(); nodeTableRunShard(t, 5) }
func TestNodeTableIsLinkOnly_006(t *testing.T) { t.Parallel(); nodeTableRunShard(t, 6) }
func TestNodeTableIsLinkOnly_007(t *testing.T) { t.Parallel(); nodeTableRunShard(t, 7) }

// The probe uses the actual top-level enumeration, assignment and byte oracle.
// One changed case must be caught by exactly one owner; union corruption fails.
func TestNodeTableIsLinkOnlyUnionAndPlantedFailure(t *testing.T) {
	t.Parallel()
	nodeTableSetup(t)
	testgrain.Union(t, "TestNodeTableIsLinkOnly", testNodeTableIsLinkOnlyShards, nodeTableAssignments, len(nodeTableRows))
	owner := nodeTableShard("case 7")
	caught := make(map[int]bool)
	for shard := 0; shard < testNodeTableIsLinkOnlyShards; shard++ {
		name := fmt.Sprintf("TestNodeTableIsLinkOnly_%03d", shard)
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		command := testgrain.CommandContext(t, ctx, os.Args[0], "-test.run=^"+name+"$", "-test.timeout=10s", "-test.v")
		command.Env = append(os.Environ(), "ADAMIC_NODE_TABLE_PROBE=1")
		output, err := command.CombinedOutput()
		cancel()
		if err != nil {
			caught[shard] = true
			if bytes.Count(output, []byte("--- FAIL:")) != 1 || !bytes.Contains(output, []byte("planted disagreement")) {
				t.Fatalf("wrong failure: %s", output)
			}
		}
	}
	testgrain.CaughtByExactly(t, caught, owner)
	t.Logf("planted disagreement case 7 caught only by TestNodeTableIsLinkOnly_%03d", owner)
}
func nodeTableCaptureUpstream(t *testing.T, sourceRoot, directory string) ([]string, error) {
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
	environment := []string{"COHERE_DOCS_CAPTURE=" + capture}
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
	// Reserve one of the instance's four CPUs for independent native emission.
	jobs := make(chan string)
	failures := make(chan error, len(names))
	var workers sync.WaitGroup
	for worker := 0; worker < 3; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for name := range jobs {
				if _, err := rulesAgreeCaptureRun(t, root, environment, "test", "-p=1", "-overlay="+overlayPath, "./internal/lint/rules/"+name, "-run", "^("+strings.Join(packages[name], "|")+")", "-count=1", "-timeout=0"); err != nil {
					failures <- err
				}
			}
		}()
	}
	for _, name := range names {
		jobs <- name
	}
	close(jobs)
	workers.Wait()
	close(failures)
	for err := range failures {
		return nil, err
	}
	files, err := filepath.Glob(filepath.Join(capture, "*.jsonl"))
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
