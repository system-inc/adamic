package printer

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/testguard"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Only buildcache products survive a test's cleanup. Every selected shard
// fetches its own prerequisites, once per process, before its work deadline.
type printerAsGoOnce struct {
	once sync.Once
	path string
}

func (p *printerAsGoOnce) get(t *testing.T, build func() string) string {
	t.Helper()
	p.once.Do(func() { p.path = build() })
	if p.path == "" {
		t.Fatal("GraphQL product preparation failed")
	}
	return p.path
}

var asGoSource, asGoLowered, asGoSanitized, asGoRelease printerAsGoOnce

func printerAsGoSource(t *testing.T) string {
	t.Helper()
	return asGoSource.get(t, func() string {
		files := []string{"../token.ts", "../characterClasses.ts", "../blockString.ts", "../lexer.ts", "../parser.ts", "../../json/width.ts", "../../json/widthTables.ts", "doc.ts", "printer.ts", "main.ts", "printer_test.go", "grain_asgo_test.go"}
		return printerBuild(t, printerBuildInputs{Name: "GraphQL printer source", Files: printerInputFiles(t, files...), Toolchain: runtime.Version()}, func(dir string) error {
			printerDirectoryAt(t, dir, "", "", "")
			return nil
		}) + "/graphql/printer/main.ts"
	})
}

func printerAsGoLowered(t *testing.T) string {
	t.Helper()
	return asGoLowered.get(t, func() string { return printerLoweredProduct(t, printerAsGoSource(t)) })
}
func printerAsGoNative(t *testing.T, sanitize bool) string {
	t.Helper()
	once := &asGoRelease
	if sanitize {
		once = &asGoSanitized
	}
	return once.get(t, func() string {
		lowered := printerAsGoLowered(t)
		data, err := os.ReadFile(lowered + "/port.c")
		if err != nil {
			t.Fatal(err)
		}
		options := native.Options{Sanitize: sanitize, Split: true, Jobs: 4}
		flags := append(native.Flags(options), "split=true", "jobs=4")
		return printerBuild(t, printerBuildInputs{
			Name:  fmt.Sprintf("GraphQL printer split native sanitize=%t", sanitize),
			Files: append([]string{lowered + "/port.c"}, printerInputFiles(t, filepath.Join(repository, "internal/native"))...),
			Flags: flags, Toolchain: buildcache.Tool("clang", "--version"),
		}, func(dir string) error { return native.Build(string(data), dir+"/port", options) }) + "/port"
	})
}
func printerAsGoProducts(t *testing.T) printerProducts {
	t.Helper()
	product := printerProducts{source: printerAsGoSource(t), backend: printerAsGoLowered(t) + "/program.mjs", sanitized: printerAsGoNative(t, true)}
	if runtime.GOOS == "darwin" {
		product.release = printerAsGoNative(t, false)
	}
	return product
}

func TestProduct_GraphQLPrinterSource(t *testing.T)    { t.Parallel(); printerAsGoSource(t) }
func TestProduct_GraphQLPrinterLowered(t *testing.T)   { t.Parallel(); printerAsGoLowered(t) }
func TestProduct_GraphQLPrinterSanitized(t *testing.T) { t.Parallel(); printerAsGoNative(t, true) }
func TestProduct_GraphQLPrinterRelease(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin leak check only")
	}
	printerAsGoNative(t, false)
}
func TestProduct_GraphQLPrinterDefaults(t *testing.T) { t.Parallel(); printerCases(t, "defaults") }
func TestProduct_GraphQLPrinterNarrow(t *testing.T)   { t.Parallel(); printerCases(t, "narrow") }
func TestProduct_GraphQLPrinterTight(t *testing.T)    { t.Parallel(); printerCases(t, "tight") }
func TestProduct_GraphQLPrinterTabs(t *testing.T)     { t.Parallel(); printerCases(t, "tabs") }

func TestPrinterAsGoCohere_Union(t *testing.T) { t.Parallel(); printerAsGoUnit(t, -1) }
func TestPrinterAsGoCohere_000(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 0) }
func TestPrinterAsGoCohere_001(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 1) }
func TestPrinterAsGoCohere_002(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 2) }
func TestPrinterAsGoCohere_003(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 3) }
func TestPrinterAsGoCohere_004(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 4) }
func TestPrinterAsGoCohere_005(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 5) }
func TestPrinterAsGoCohere_006(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 6) }
func TestPrinterAsGoCohere_007(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 7) }
func TestPrinterAsGoCohere_008(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 8) }
func TestPrinterAsGoCohere_009(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 9) }
func TestPrinterAsGoCohere_010(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 10) }
func TestPrinterAsGoCohere_011(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 11) }

// The gate runs each shard by its wrapper's name, so a wrapper deleted, renumbered or pointed at another shard would
// drop that shard's cases with nothing red. This file must declare exactly testPrinterAsGoCohereShards wrappers,
// TestPrinterAsGoCohere_000 on without a gap, each running its own shard.
func TestPrinterAsGoCohereWrappers(t *testing.T) {
	t.Parallel()
	file, err := parser.ParseFile(token.NewFileSet(), "grain_asgo_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	wrapped := map[int]string{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "TestPrinterAsGoCohere_") || function.Name.Name == "TestPrinterAsGoCohere_Union" {
			continue
		}
		suffix := strings.TrimPrefix(function.Name.Name, "TestPrinterAsGoCohere_")
		number, err := strconv.Atoi(suffix)
		if err != nil || fmt.Sprintf("%03d", number) != suffix {
			t.Fatalf("%s: not a three-digit shard wrapper", function.Name.Name)
		}
		runs := -1
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) != 2 {
				return true
			}
			if called, ok := call.Fun.(*ast.Ident); ok && called.Name == "printerAsGoUnit" {
				if literal, ok := call.Args[1].(*ast.BasicLit); ok && literal.Kind == token.INT {
					runs, _ = strconv.Atoi(literal.Value)
				}
			}
			return true
		})
		if runs != number {
			t.Fatalf("%s runs shard %d, not its own", function.Name.Name, runs)
		}
		if previous, exists := wrapped[number]; exists {
			t.Fatalf("shard-%03d has two wrappers: %s and %s", number, previous, function.Name.Name)
		}
		wrapped[number] = function.Name.Name
	}
	for number := range testPrinterAsGoCohereShards {
		if _, exists := wrapped[number]; !exists {
			t.Errorf("shard-%03d has no TestPrinterAsGoCohere_%03d: its cases would never run", number, number)
		}
	}
	if len(wrapped) != testPrinterAsGoCohereShards {
		t.Errorf("%d shard wrappers, %d shards", len(wrapped), testPrinterAsGoCohereShards)
	}
}

type printerAsGoJob struct {
	name, command     string
	args, environment []string
	leakReport        bool
}
type printerAsGoResult struct {
	job  printerAsGoJob
	run  run
	err  error
	wall time.Duration
}

func printerAsGoRun(ctx context.Context, job printerAsGoJob) printerAsGoResult {
	start := time.Now()
	command := exec.CommandContext(ctx, job.command, job.args...)
	if job.environment != nil {
		command.Env = append(os.Environ(), job.environment...)
	}
	if !strings.ContainsRune(job.command, filepath.Separator) {
		// A bare name (node, macOS's leaks) is a tool: the one internal/toolchain resolves, with a tool's environment
		// plus the job's settings, not PATH's, as toolchain.Command runs one under a context it has no form for. A
		// program a test built (a port) runs as named, with this process's environment. An error resolving it fails the
		// command as it starts.
		if path, err := toolchain.Path(job.command); err != nil {
			command.Err = err
		} else {
			command = exec.CommandContext(ctx, path, job.args...)
			command.Env = toolchain.Env(job.environment...)
		}
	}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	// Its own group, its output read to the end once the group is gone: no delay decides when the reads are done, so a
	// slow finish under load is never red (#n3gk22h, a leak pass's "WaitDelay expired before I/O complete" on Home).
	stdout, stderr, err := childguard.Capture(ctx, command)
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		err = nil
	}
	result := run{stdout: stdout, stderr: stderr, exitCode: -1}
	if command.ProcessState != nil {
		result.exitCode = command.ProcessState.ExitCode()
	}
	return printerAsGoResult{job, result, err, time.Since(start)}
}

// A job whose process leaves a descendant holding its output (a shell's background sleep, standing in for a helper a
// sanitizer starts) still returns what the job wrote, with no error, as soon as the job exits. exec's own pipes with the
// one-second WaitDelay that printerAsGoRun had reported "WaitDelay expired before I/O complete" here instead (#n3gk22h).
func TestPrinterAsGoRunReadsPastADescendant(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result := printerAsGoRun(ctx, printerAsGoJob{"descendant", "/bin/sh", []string{"-c", "sleep 30 & printf ok; printf warned >&2"}, nil, false})
	if result.err != nil || string(result.run.stdout) != "ok" || string(result.run.stderr) != "warned" || result.run.exitCode != 0 {
		t.Fatalf("err %v, exit %d, stdout %q, stderr %q", result.err, result.run.exitCode, result.run.stdout, result.run.stderr)
	}
	if result.wall > 10*time.Second {
		t.Fatalf("the job's reads waited %.3fs on the descendant's sleep", result.wall.Seconds())
	}
	t.Logf("read past the descendant in %.3fs", result.wall.Seconds())
}

// Each of the four fixed option modes deals its cases to three buckets by a hash of the case's input, so no shard runs
// a whole mode's four processes over 3,564 texts (about 20 s of one core's work, Workshop Oct 11), a case keeps its
// bucket as the corpus grows or reorders, and bucket b holds the same inputs in every mode. Shard mode*3+b owns them.
var asGoModes = [...]string{"defaults", "narrow", "tight", "tabs"}

const testPrinterAsGoCohereBuckets = 3
const testPrinterAsGoCohereShards = len(asGoModes) * testPrinterAsGoCohereBuckets

func asGoBucket(input string) int {
	hash := sha256.Sum256([]byte(input))
	return int(binary.BigEndian.Uint64(hash[:8]) % testPrinterAsGoCohereBuckets)
}

// printerAsGoShards deals each mode's enumeration, given in asGoModes order, to its buckets, keeping corpus order
// within a bucket. A shard's path is set by the unit that runs it.
func printerAsGoShards(enumerations [][]printerCase) []printerShard {
	shards := make([]printerShard, testPrinterAsGoCohereShards)
	for index, mode := range asGoModes {
		for bucket := range testPrinterAsGoCohereBuckets {
			shards[index*testPrinterAsGoCohereBuckets+bucket].mode = mode
		}
		for _, item := range enumerations[index] {
			number := index*testPrinterAsGoCohereBuckets + asGoBucket(item.input)
			shards[number].cases = append(shards[number].cases, item)
		}
	}
	return shards
}

// printerShardFile writes a shard's inputs in the corpus protocol, one per line, for the runners' --cases.
func printerShardFile(t *testing.T, shard printerShard) string {
	t.Helper()
	var protocol strings.Builder
	for _, item := range shard.cases {
		protocol.WriteString(item.input + "\n")
	}
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(protocol.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// ADAMIC_TEST_SHARD=i/n selects shard indices modulo n equal to i; unset runs all. The gate can instead select
// TestPrinterAsGoCohere_NNN directly. Products are prepared before case timing.
func printerAsGoUnit(t *testing.T, unit int) {
	setupStart := time.Now()
	var products printerProducts
	if unit >= 0 {
		products = printerAsGoProducts(t)
	}
	var whole []printerCase
	var enumerations [][]printerCase
	for _, mode := range asGoModes {
		cases, want := printerCases(t, mode)
		enumeration := enumeratePrinter(t, mode, cases, want)
		whole = append(whole, enumeration...)
		enumerations = append(enumerations, enumeration)
	}
	shards := printerAsGoShards(enumerations)
	for number, shard := range shards {
		if len(shard.cases) == 0 {
			t.Fatalf("shard-%03d (%s) holds no cases", number, shard.mode)
		}
	}
	if len(shards) != testPrinterAsGoCohereShards {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), testPrinterAsGoCohereShards)
	}
	if err := printerShardUnion(whole, shards); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d unique mode/case ids across %d shards", len(whole), len(shards))
	if unit < 0 {
		return
	}
	selected, err := printerShardSelection(os.Getenv("ADAMIC_TEST_SHARD"), len(shards))
	if err != nil {
		t.Fatal(err)
	}
	for number, shard := range shards {
		if number != unit || !selected[number] {
			continue
		}
		{
			shard.path = printerShardFile(t, shard)
			t.Logf("setup wall %.3fs", time.Since(setupStart).Seconds())
			start := time.Now()
			t.Cleanup(func() {
				t.Logf("own work wall %.3fs", time.Since(start).Seconds())
				if elapsed := time.Since(start); elapsed > 30*time.Second {
					t.Errorf("invalid test unit: %.3fs exceeds 30s", elapsed.Seconds())
				}
			})
			t.Logf("mode %s, bucket %d of %d: %d cases, %s to %s", shard.mode, number%testPrinterAsGoCohereBuckets, testPrinterAsGoCohereBuckets, len(shard.cases), shard.cases[0].id, shard.cases[len(shard.cases)-1].id)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
			if err != nil {
				t.Fatal(err)
			}
			args := []string{"--cases", shard.path, shard.mode}
			jobs := []printerAsGoJob{
				{"Node", "node", append([]string{"--disable-warning=ExperimentalWarning", runner, products.source}, args...), nil, false},
				{"native", products.sanitized, args, []string{"ASAN_OPTIONS=detect_leaks=0"}, false},
				{"JS backend", "node", append([]string{"--disable-warning=ExperimentalWarning", runner, products.backend}, args...), nil, false},
			}
			switch runtime.GOOS {
			case "linux":
				jobs = append(jobs, printerAsGoJob{"leaks", products.sanitized, args, []string{"ASAN_OPTIONS=detect_leaks=1"}, false})
			case "darwin":
				jobs = append(jobs, printerAsGoJob{"leaks", "leaks", append([]string{"--atExit", "--", products.release}, args...), nil, true})
			default:
				t.Fatalf("no leak check for %s", runtime.GOOS)
			}
			// All four processes read the same immutable corpus and products. At most the test's share of the unit's
			// CPUs run at once (testguard.Share), so the unit's other tests keep theirs; every backend and the leak
			// check still run.
			slots := make(chan struct{}, testguard.Share())
			results := make(chan printerAsGoResult, len(jobs))
			for _, job := range jobs {
				go func(job printerAsGoJob) {
					slots <- struct{}{}
					defer func() { <-slots }()
					results <- printerAsGoRun(ctx, job)
				}(job)
			}
			t.Logf("%d processes, at most %d at once", len(jobs), testguard.Share())
			for range jobs {
				result := <-results
				t.Logf("%s wall %.3fs", result.job.name, result.wall.Seconds())
				if result.err != nil {
					t.Errorf("%s: %v", result.job.name, result.err)
					continue
				}
				if result.job.leakReport {
					if result.run.exitCode != 0 {
						t.Errorf("leaks: exit %d stdout %s stderr %s", result.run.exitCode, result.run.stdout, result.run.stderr)
					}
				} else if err := printerShardDisagreement(number, shard, result.run); err != nil {
					t.Errorf("%s: %v", result.job.name, err)
				}
			}
		}
	}
}

// The first four oracle cases of two modes, dealt as the units deal them, are a real Node control; one planted
// disagreement must be caught by exactly the shard asGoBucket names for it.
func TestPrinterShardPlantedDisagreement(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs("main.ts")
	enumerations := make([][]printerCase, len(asGoModes))
	for index, mode := range asGoModes {
		if mode != "defaults" && mode != "tabs" {
			continue
		}
		cases, want := printerCases(t, mode)
		enumerations[index] = append([]printerCase(nil), enumeratePrinter(t, mode, cases, want)[:4]...)
	}
	shards := printerAsGoShards(enumerations)
	results := make([]run, len(shards))
	for number, shard := range shards {
		if len(shard.cases) == 0 {
			continue
		}
		results[number] = onNode(t, path, "--cases", printerShardFile(t, shard), shard.mode)
		if err := printerShardDisagreement(number, shard, results[number]); err != nil {
			t.Fatal(err)
		}
	}
	planted := enumerations[0][1]
	owner := asGoBucket(planted.input)
	for index := range shards[owner].cases {
		if shards[owner].cases[index].id == planted.id {
			shards[owner].cases[index].want += "planted disagreement"
		}
	}
	caught := 0
	for number, shard := range shards {
		if len(shard.cases) == 0 {
			continue
		}
		if err := printerShardDisagreement(number, shard, results[number]); err != nil {
			caught++
			if number != owner || !strings.Contains(err.Error(), fmt.Sprintf("shard-%03d %s", owner, planted.id)) {
				t.Fatalf("wrong owner: %v", err)
			}
			t.Logf("planted disagreement caught: %v", err)
		}
	}
	if caught != 1 {
		t.Fatalf("planted disagreement caught by %d shards, want exactly one", caught)
	}
}

// A case's shard is its mode's and its input's: cases added anywhere, even ahead of it, never move it, and the same
// input takes the same bucket in every mode.
func TestPrinterAsGoShardGrowth(t *testing.T) {
	t.Parallel()
	original := make([]printerCase, 12)
	for index := range original {
		original[index] = printerCase{id: fmt.Sprintf("defaults/case-%06d", index), input: fmt.Sprintf(">query{field%d}", index)}
	}
	grown := append([]printerCase{{id: "defaults/case-added", input: ">query{added}"}}, original...)
	tabs := make([]printerCase, len(original))
	for index, item := range original {
		tabs[index] = printerCase{id: "tabs" + strings.TrimPrefix(item.id, "defaults"), input: item.input}
	}
	before := printerAsGoShards([][]printerCase{original, nil, nil, tabs})
	after := printerAsGoShards([][]printerCase{grown, nil, nil, tabs})
	if len(before) != testPrinterAsGoCohereShards || len(after) != len(before) {
		t.Fatal("growth changed shard count")
	}
	if err := printerShardUnion(append(append([]printerCase(nil), original...), tabs...), before); err != nil {
		t.Fatal(err)
	}
	if err := printerShardUnion(append(append([]printerCase(nil), grown...), tabs...), after); err != nil {
		t.Fatal(err)
	}
	tabsIndex := (len(asGoModes) - 1) * testPrinterAsGoCohereBuckets
	for number := range testPrinterAsGoCohereBuckets {
		if len(before[number].cases) != len(before[tabsIndex+number].cases) {
			t.Fatalf("bucket %d: %d defaults inputs, %d tabs inputs", number, len(before[number].cases), len(before[tabsIndex+number].cases))
		}
		for index, item := range before[number].cases {
			if before[tabsIndex+number].cases[index].input != item.input {
				t.Fatalf("bucket %d holds %q in defaults, %q in tabs", number, item.input, before[tabsIndex+number].cases[index].input)
			}
		}
	}
	for number, shard := range before {
		for _, item := range shard.cases {
			found := false
			for _, candidate := range after[number].cases {
				found = found || candidate == item
			}
			if !found {
				t.Fatalf("growth moved %s from shard-%03d", item.id, number)
			}
		}
	}
}
