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
	_ "github.com/system-inc/adamic/internal/ccjobserver"
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

func (once *printerAsGoOnce) get(test *testing.T, build func() string) string {
	test.Helper()
	once.once.Do(func() { once.path = build() })
	if once.path == "" {
		test.Fatal("GraphQL product preparation failed")
	}
	return once.path
}

var asGoSource, asGoLowered, asGoSanitized, asGoRelease printerAsGoOnce

func printerAsGoSource(test *testing.T) string {
	test.Helper()
	return asGoSource.get(test, func() string {
		files := []string{
			"../token.ts",
			"../characterClasses.ts",
			"../blockString.ts",
			"../lexer.ts",
			"../parser.ts",
			"../../json/width.ts",
			"../../json/widthTables.ts",
			"doc.ts",
			"printer.ts",
			"main.ts",
			"printer_test.go",
			"grain_asgo_test.go",
		}
		return printerBuild(test, printerBuildInputs{
			Name:      "GraphQL printer source",
			Files:     printerInputFiles(test, files...),
			Toolchain: runtime.Version(),
		}, func(dir string) error {
			printerDirectoryAt(test, dir, "", "", "")
			return nil
		}) + "/graphql/printer/main.ts"
	})
}

func printerAsGoLowered(test *testing.T) string {
	test.Helper()
	return asGoLowered.get(test, func() string { return printerLoweredProduct(test, printerAsGoSource(test)) })
}
func printerAsGoNative(test *testing.T, sanitize bool) string {
	test.Helper()
	once := &asGoRelease
	if sanitize {
		once = &asGoSanitized
	}
	return once.get(test, func() string {
		lowered := printerAsGoLowered(test)
		data, readError := os.ReadFile(lowered + "/port.c")
		if readError != nil {
			test.Fatal(readError)
		}
		options := native.Options{Sanitize: sanitize, Split: true, Jobs: 4}
		flags := append(native.Flags(options), "split=true", "jobs=4")
		return printerBuild(test, printerBuildInputs{
			Name: fmt.Sprintf("GraphQL printer split native sanitize=%t", sanitize),
			Files: append(
				[]string{lowered + "/port.c"},
				printerInputFiles(test, filepath.Join(repository, "internal/native"))...,
			),
			Flags: flags, Toolchain: buildcache.Tool("clang", "--version"),
		}, func(dir string) error { return native.Build(string(data), dir+"/port", options) }) + "/port"
	})
}
func printerAsGoProducts(test *testing.T) printerProducts {
	test.Helper()
	product := printerProducts{
		source:    printerAsGoSource(test),
		backend:   printerAsGoLowered(test) + "/program.mjs",
		sanitized: printerAsGoNative(test, true),
	}
	if runtime.GOOS == "darwin" {
		product.release = printerAsGoNative(test, false)
	}
	return product
}

func TestProduct_GraphQLPrinterSource(test *testing.T)  { test.Parallel(); printerAsGoSource(test) }
func TestProduct_GraphQLPrinterLowered(test *testing.T) { test.Parallel(); printerAsGoLowered(test) }
func TestProduct_GraphQLPrinterSanitized(test *testing.T) {
	test.Parallel()
	printerAsGoNative(test, true)
}
func TestProduct_GraphQLPrinterRelease(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "darwin" {
		test.Skip("Darwin leak check only")
	}
	printerAsGoNative(test, false)
}
func TestProduct_GraphQLPrinterFindsDefaults(test *testing.T) {
	test.Parallel()
	printerCases(test, "defaults")
}
func TestProduct_GraphQLPrinterFindsNarrow(test *testing.T) {
	test.Parallel()
	printerCases(test, "narrow")
}
func TestProduct_GraphQLPrinterFindsTight(test *testing.T) {
	test.Parallel()
	printerCases(test, "tight")
}
func TestProduct_GraphQLPrinterFindsTabs(test *testing.T) {
	test.Parallel()
	printerCases(test, "tabs")
}

func TestPrinterAsGoCohere_Union(test *testing.T) { test.Parallel(); printerAsGoUnit(test, -1) }
func TestPrinterAsGoCohere_000(test *testing.T)   { test.Parallel(); printerAsGoUnit(test, 0) }
func TestPrinterAsGoCohere_001(test *testing.T)   { test.Parallel(); printerAsGoUnit(test, 1) }
func TestPrinterAsGoCohere_002(test *testing.T)   { test.Parallel(); printerAsGoUnit(test, 2) }
func TestPrinterAsGoCohere_003(test *testing.T)   { test.Parallel(); printerAsGoUnit(test, 3) }
func TestPrinterAsGoCohere_004(test *testing.T)   { test.Parallel(); printerAsGoUnit(test, 4) }
func TestPrinterAsGoCohere_005(test *testing.T)   { test.Parallel(); printerAsGoUnit(test, 5) }
func TestPrinterAsGoCohere_006(test *testing.T)   { test.Parallel(); printerAsGoUnit(test, 6) }
func TestPrinterAsGoCohere_007(test *testing.T)   { test.Parallel(); printerAsGoUnit(test, 7) }
func TestPrinterAsGoCohere_008(test *testing.T)   { test.Parallel(); printerAsGoUnit(test, 8) }
func TestPrinterAsGoCohere_009(test *testing.T)   { test.Parallel(); printerAsGoUnit(test, 9) }
func TestPrinterAsGoCohere_010(test *testing.T)   { test.Parallel(); printerAsGoUnit(test, 10) }
func TestPrinterAsGoCohere_011(test *testing.T)   { test.Parallel(); printerAsGoUnit(test, 11) }

// The gate runs each shard by its wrapper's name, so a wrapper deleted, renumbered or pointed at another shard would
// drop that shard's cases with nothing red. This file must declare exactly testPrinterAsGoCohereShards wrappers,
// TestPrinterAsGoCohere_000 on without a gap, each running its own shard.
func TestPrinterAsGoCohereWrappers(test *testing.T) {
	test.Parallel()
	file, parseFileError := parser.ParseFile(token.NewFileSet(), "grain_asgo_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	wrapped := map[int]string{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok ||
			!strings.HasPrefix(function.Name.Name, "TestPrinterAsGoCohere_") ||
			function.Name.Name == "TestPrinterAsGoCohere_Union" {
			continue
		}
		suffix := strings.TrimPrefix(function.Name.Name, "TestPrinterAsGoCohere_")
		number, parseError := strconv.Atoi(suffix)
		if parseError != nil || fmt.Sprintf("%03d", number) != suffix {
			test.Fatalf("%s: not a three-digit shard wrapper", function.Name.Name)
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
			test.Fatalf("%s runs shard %d, not its own", function.Name.Name, runs)
		}
		if previous, exists := wrapped[number]; exists {
			test.Fatalf("shard-%03d has two wrappers: %s and %s", number, previous, function.Name.Name)
		}
		wrapped[number] = function.Name.Name
	}
	for number := range testPrinterAsGoCohereShards {
		if _, exists := wrapped[number]; !exists {
			test.Errorf("shard-%03d has no TestPrinterAsGoCohere_%03d: its cases would never run", number, number)
		}
	}
	if len(wrapped) != testPrinterAsGoCohereShards {
		test.Errorf("%d shard wrappers, %d shards", len(wrapped), testPrinterAsGoCohereShards)
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

func printerAsGoRun(callerContext context.Context, job printerAsGoJob) printerAsGoResult {
	start := time.Now()
	command := exec.CommandContext(callerContext, job.command, job.args...)
	if job.environment != nil {
		command.Env = append(os.Environ(), job.environment...)
	}
	if !strings.ContainsRune(job.command, filepath.Separator) {
		// A bare name (node, macOS's leaks) is a tool: the one internal/toolchain resolves, with a tool's environment
		// plus the job's settings, not PATH's, as toolchain.Command runs one under a context it has no form for. A
		// program a test built (a port) runs as named, with this process's environment. An error resolving it fails the
		// command as it starts.
		if path, pathError := toolchain.Path(job.command); pathError != nil {
			command.Err = pathError
		} else {
			command = exec.CommandContext(callerContext, path, job.args...)
			command.Env = toolchain.Env(job.environment...)
		}
	}
	command.Cancel = func() error {
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(killError, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return killError
	}
	// Its own group, its output read to the end once the group is gone: no delay decides when the reads are done, so a
	// slow finish under load is never red (#n3gk22h, a leak pass's "WaitDelay expired before I/O complete" on Home).
	stdout, stderr, captureError := childguard.Capture(callerContext, command)
	var exitError *exec.ExitError
	if errors.As(captureError, &exitError) {
		captureError = nil
	}
	result := run{stdout: stdout, stderr: stderr, exitCode: -1}
	if command.ProcessState != nil {
		result.exitCode = command.ProcessState.ExitCode()
	}
	return printerAsGoResult{job, result, captureError, time.Since(start)}
}

// A job whose process leaves a descendant holding its output (a shell's background sleep, standing in for a helper a
// sanitizer starts) still returns what the job wrote, with no error, as soon as the job exits. exec's own pipes with
// the one-second WaitDelay that printerAsGoRun had reported "WaitDelay expired before I/O complete" here instead
// (#n3gk22h).
func TestPrinterAsGoRunReadsPastADescendant(test *testing.T) {
	test.Parallel()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result := printerAsGoRun(
		timeoutContext,
		printerAsGoJob{"descendant", "/bin/sh", []string{"-c", "sleep 30 & printf ok; printf warned >&2"}, nil, false},
	)
	if result.err != nil ||
		string(result.run.stdout) != "ok" ||
		string(result.run.stderr) != "warned" ||
		result.run.exitCode != 0 {
		test.Fatalf(
			"err %v, exit %d, stdout %q, stderr %q",
			result.err,
			result.run.exitCode,
			result.run.stdout,
			result.run.stderr,
		)
	}
	if result.wall > 10*time.Second {
		test.Fatalf("the job's reads waited %.3fs on the descendant's sleep", result.wall.Seconds())
	}
	test.Logf("read past the descendant in %.3fs", result.wall.Seconds())
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
func printerShardFile(test *testing.T, shard printerShard) string {
	test.Helper()
	var protocol strings.Builder
	for _, item := range shard.cases {
		protocol.WriteString(item.input + "\n")
	}
	path := filepath.Join(test.TempDir(), "cases.txt")
	if writeError := os.WriteFile(path, []byte(protocol.String()), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	return path
}

// ADAMIC_TEST_SHARD=i/n selects shard indices modulo n equal to i; unset runs all. The gate can instead select
// TestPrinterAsGoCohere_NNN directly. Products are prepared before case timing.
func printerAsGoUnit(test *testing.T, unit int) {
	setupStart := time.Now()
	var products printerProducts
	if unit >= 0 {
		products = printerAsGoProducts(test)
	}
	var whole []printerCase
	var enumerations [][]printerCase
	for _, mode := range asGoModes {
		cases, want := printerCases(test, mode)
		enumeration := enumeratePrinter(test, mode, cases, want)
		whole = append(whole, enumeration...)
		enumerations = append(enumerations, enumeration)
	}
	shards := printerAsGoShards(enumerations)
	for number, shard := range shards {
		if len(shard.cases) == 0 {
			test.Fatalf("shard-%03d (%s) holds no cases", number, shard.mode)
		}
	}
	if len(shards) != testPrinterAsGoCohereShards {
		test.Fatalf("enumerated %d shards, declared %d", len(shards), testPrinterAsGoCohereShards)
	}
	if printerShardUnionError := printerShardUnion(whole, shards); printerShardUnionError != nil {
		test.Fatal(printerShardUnionError)
	}
	test.Logf("union: %d unique mode/case ids across %d shards", len(whole), len(shards))
	if unit < 0 {
		return
	}
	selected, printerShardSelectionError := printerShardSelection(os.Getenv("ADAMIC_TEST_SHARD"), len(shards))
	if printerShardSelectionError != nil {
		test.Fatal(printerShardSelectionError)
	}
	for number, shard := range shards {
		if number != unit || !selected[number] {
			continue
		}
		{
			shard.path = printerShardFile(test, shard)
			test.Logf("setup wall %.3fs", time.Since(setupStart).Seconds())
			start := time.Now()
			test.Cleanup(func() {
				test.Logf("own work wall %.3fs", time.Since(start).Seconds())
				if elapsed := time.Since(start); elapsed > 30*time.Second {
					test.Errorf("invalid test unit: %.3fs exceeds 30s", elapsed.Seconds())
				}
			})
			test.Logf(
				"mode %s, bucket %d of %d: %d cases, %s to %s",
				shard.mode,
				number%testPrinterAsGoCohereBuckets,
				testPrinterAsGoCohereBuckets,
				len(shard.cases),
				shard.cases[0].id,
				shard.cases[len(shard.cases)-1].id,
			)
			timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			args := []string{"--cases", shard.path, shard.mode}
			jobs := []printerAsGoJob{
				{
					"Node",
					"node",
					append([]string{"--disable-warning=ExperimentalWarning", runner, products.source}, args...),
					nil,
					false,
				},
				{"native", products.sanitized, args, []string{"ASAN_OPTIONS=detect_leaks=0"}, false},
				{
					"JS backend",
					"node",
					append([]string{"--disable-warning=ExperimentalWarning", runner, products.backend}, args...),
					nil,
					false,
				},
			}
			switch runtime.GOOS {
			case "linux":
				jobs = append(
					jobs,
					printerAsGoJob{"leaks", products.sanitized, args, []string{"ASAN_OPTIONS=detect_leaks=1"}, false},
				)
			case "darwin":
				jobs = append(jobs, printerAsGoJob{
					"leaks",
					"leaks",
					append([]string{"--atExit", "--", products.release}, args...),
					nil,
					true,
				})
			default:
				test.Fatalf("no leak check for %s", runtime.GOOS)
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
					results <- printerAsGoRun(timeoutContext, job)
				}(job)
			}
			test.Logf("%d processes, at most %d at once", len(jobs), testguard.Share())
			for range jobs {
				result := <-results
				test.Logf("%s wall %.3fs", result.job.name, result.wall.Seconds())
				if result.err != nil {
					test.Errorf("%s: %v", result.job.name, result.err)
					continue
				}
				if result.job.leakReport {
					if result.run.exitCode != 0 {
						test.Errorf(
							"leaks: exit %d stdout %s stderr %s",
							result.run.exitCode,
							result.run.stdout,
							result.run.stderr,
						)
					}
				} else if printerShardDisagreementError := printerShardDisagreement(
					number,
					shard,
					result.run,
				); printerShardDisagreementError != nil {
					test.Errorf("%s: %v", result.job.name, printerShardDisagreementError)
				}
			}
		}
	}
}

// The first four oracle cases of two modes, dealt as the units deal them, are a real Node control; one planted
// disagreement must be caught by exactly the shard asGoBucket names for it.
func TestPrinterShardPlantedDisagreement(test *testing.T) {
	test.Parallel()
	path, _ := filepath.Abs("main.ts")
	enumerations := make([][]printerCase, len(asGoModes))
	for index, mode := range asGoModes {
		if mode != "defaults" && mode != "tabs" {
			continue
		}
		cases, want := printerCases(test, mode)
		enumerations[index] = append([]printerCase(nil), enumeratePrinter(test, mode, cases, want)[:4]...)
	}
	shards := printerAsGoShards(enumerations)
	results := make([]run, len(shards))
	for number, shard := range shards {
		if len(shard.cases) == 0 {
			continue
		}
		results[number] = onNode(test, path, "--cases", printerShardFile(test, shard), shard.mode)
		if printerShardDisagreementError := printerShardDisagreement(
			number,
			shard,
			results[number],
		); printerShardDisagreementError != nil {
			test.Fatal(printerShardDisagreementError)
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
		if printerShardDisagreementError := printerShardDisagreement(
			number,
			shard,
			results[number],
		); printerShardDisagreementError != nil {
			caught++
			if number != owner || !strings.Contains(
				printerShardDisagreementError.Error(),
				fmt.Sprintf("shard-%03d %s", owner, planted.id),
			) {
				test.Fatalf("wrong owner: %v", printerShardDisagreementError)
			}
			test.Logf("planted disagreement caught: %v", printerShardDisagreementError)
		}
	}
	if caught != 1 {
		test.Fatalf("planted disagreement caught by %d shards, want exactly one", caught)
	}
}

// A case's shard is its mode's and its input's: cases added anywhere, even ahead of it, never move it, and the same
// input takes the same bucket in every mode.
func TestPrinterAsGoShardGrowth(test *testing.T) {
	test.Parallel()
	original := make([]printerCase, 12)
	for index := range original {
		original[index] = printerCase{
			id:    fmt.Sprintf("defaults/case-%06d", index),
			input: fmt.Sprintf(">query{field%d}", index),
		}
	}
	grown := append([]printerCase{{id: "defaults/case-added", input: ">query{added}"}}, original...)
	tabs := make([]printerCase, len(original))
	for index, item := range original {
		tabs[index] = printerCase{id: "tabs" + strings.TrimPrefix(item.id, "defaults"), input: item.input}
	}
	before := printerAsGoShards([][]printerCase{original, nil, nil, tabs})
	after := printerAsGoShards([][]printerCase{grown, nil, nil, tabs})
	if len(before) != testPrinterAsGoCohereShards || len(after) != len(before) {
		test.Fatal("growth changed shard count")
	}
	if printerShardUnionError := printerShardUnion(
		append(append([]printerCase(nil), original...), tabs...),
		before,
	); printerShardUnionError != nil {
		test.Fatal(printerShardUnionError)
	}
	if printerShardUnionError := printerShardUnion(
		append(append([]printerCase(nil), grown...), tabs...),
		after,
	); printerShardUnionError != nil {
		test.Fatal(printerShardUnionError)
	}
	tabsIndex := (len(asGoModes) - 1) * testPrinterAsGoCohereBuckets
	for number := range testPrinterAsGoCohereBuckets {
		if len(before[number].cases) != len(before[tabsIndex+number].cases) {
			test.Fatalf(
				"bucket %d: %d defaults inputs, %d tabs inputs",
				number,
				len(before[number].cases),
				len(before[tabsIndex+number].cases),
			)
		}
		for index, item := range before[number].cases {
			if before[tabsIndex+number].cases[index].input != item.input {
				test.Fatalf(
					"bucket %d holds %q in defaults, %q in tabs",
					number,
					item.input,
					before[tabsIndex+number].cases[index].input,
				)
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
				test.Fatalf("growth moved %s from shard-%03d", item.id, number)
			}
		}
	}
}
