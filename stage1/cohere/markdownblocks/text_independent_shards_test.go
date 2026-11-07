package markdownblocks

import (
	"bytes"
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
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

// Each shard is killed at 90s, after shared setup and admission.
func textDeadline(test *testing.T) func() {
	return textPhaseDeadline(test, test.Name())
}

func textPhaseDeadline(test *testing.T, name string) func() {
	timer := time.AfterFunc(90*time.Second, func() {
		textActiveCommands.Range(func(key, value any) bool {
			if value == test {
				_ = key.(*exec.Cmd).Cancel()
			}
			return true
		})
		panic("cooked: " + name + " exceeded 90s hard deadline")
	})
	return func() { timer.Stop() }
}

// Each child has a context ceiling, and cancellation kills its compiler descendants. Each command comes with its
// context, for childguard.Capture to end the reads by.
var textActiveCommands sync.Map

func textBounded(test *testing.T, name string, arguments ...string) (*exec.Cmd, context.Context) {
	test.Helper()
	timeoutContext, cancel := context.WithTimeout(test.Context(), 90*time.Second)
	test.Cleanup(cancel)
	return textCommand(test, timeoutContext, name, arguments...), timeoutContext
}

func textSetupCommand(test *testing.T, name string, arguments ...string) (*exec.Cmd, context.Context) {
	ctx, cancel := markdownLayoutSetupContext(test.Context())
	test.Cleanup(cancel)
	return textCommand(test, ctx, name, arguments...), ctx
}

// textGoCommand is textSetupCommand for go as toolchain resolves it, run in directory under toolchain.Env, the
// environment whose compilers the Go builds' keys name.
func textGoCommand(test *testing.T, directory string, arguments ...string) (*exec.Cmd, context.Context) {
	test.Helper()
	resolved := toolchain.CommandIn(directory, "go", arguments...)
	if resolved.Err != nil {
		return resolved, test.Context()
	}
	command, ctx := textSetupCommand(test, resolved.Path, arguments...)
	command.Dir, command.Env = resolved.Dir, resolved.Env
	return command, ctx
}

// markdownCompilers are the C and C++ compilers go builds a Go product with, and the C compiler's report.
type markdownCompilers struct{ c, cPlusPlus, report string }

// markdownGoCompilers reads the compilers toolchain.Env gives go: the caller's CC and CXX, or clang and clang++ when
// it names neither.
func markdownGoCompilers() markdownCompilers {
	var compilers markdownCompilers
	for _, setting := range toolchain.Env() {
		switch name, value, _ := strings.Cut(setting, "="); name {
		case "CC":
			compilers.c = value
		case "CXX":
			compilers.cPlusPlus = value
		}
	}
	compilers.report = "no C compiler"
	if fields := strings.Fields(compilers.c); len(fields) > 0 {
		compilers.report = buildcache.Tool(fields[0], "--version")
	}
	return compilers
}

func textCommand(test *testing.T, callerContext context.Context, name string, arguments ...string) *exec.Cmd {
	command := exec.CommandContext(callerContext, name, arguments...)
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(killError, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return killError
	}
	textActiveCommands.Store(command, test)
	return command
}

// textCombinedOutput returns a run's stdout, then its stderr, read through childguard.Capture for textExecute's
// reason. The regeneration it runs starts the formatter as a child of its own, and Capture kills the group once the
// generator exits, so neither holds a pipe after it. The generator writes nothing until it fails, so of
// childguard.Run's windows only the 30-minute first-output one could have fired, and the unit's own ceiling ends a hung
// run sooner.
func textCombinedOutput(callerContext context.Context, command *exec.Cmd) ([]byte, error) {
	defer textActiveCommands.Delete(command)
	stdout, stderr, captureError := childguard.Capture(callerContext, command)
	return append(stdout, stderr...), captureError
}

// textExecute reads a run's output through childguard.Capture, to its end with no timer deciding when: exec's own
// copying under a one-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF a second after
// the child exited, so a run that had finished could fail its shard under load (#n3gk22h's shape). Capture has no stall
// guard, and the runs lose nothing by it: childguard.Run's windows (two minutes of silence, thirty before the first
// byte) are longer than the command's 90 s deadline, which ends a hung run sooner.
func textExecute(test *testing.T, environment []string, name string, arguments ...string) run {
	test.Helper()
	command, ctx := textBounded(test, name, arguments...)
	defer textActiveCommands.Delete(command)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	var exitError *exec.ExitError
	if captureError != nil && !errors.As(captureError, &exitError) {
		test.Fatalf("running %s: %v", name, captureError)
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}
}

func textOnNode(test *testing.T, path string, arguments ...string) run {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return textExecute(
		test,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...,
	)
}

func textShardFor(key string, count int) int {
	sum := sha256.Sum256([]byte(key))
	return int(binary.BigEndian.Uint64(sum[:8]) % uint64(count))
}

func textVerifyLeaves(test *testing.T) {
	test.Helper()
	file, parseFileError := parser.ParseFile(token.NewFileSet(), "text_independent_shards_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	seen := make(map[int]bool)
	for _, decl := range file.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		suffix, ok := strings.CutPrefix(function.Name.Name, "TestMarkdownTextSplitting_")
		if !ok {
			continue
		}
		index, parseError := strconv.Atoi(suffix)
		if parseError != nil ||
			len(suffix) != 3 ||
			index < 0 ||
			index >= testMarkdownTextSplittingShards ||
			seen[index] {
			test.Fatalf("invalid top-level shard %s", function.Name.Name)
		}
		if len(function.Body.List) != 2 {
			test.Fatalf("shard %s must have t.Parallel followed by one dispatcher", function.Name.Name)
		}
		expression, ok := function.Body.List[1].(*ast.ExprStmt)
		if !ok {
			test.Fatal("invalid shard dispatcher")
		}
		call, ok := expression.X.(*ast.CallExpr)
		if !ok || len(call.Args) != 2 {
			test.Fatal("invalid shard call")
		}
		callee, ok := call.Fun.(*ast.Ident)
		if !ok || callee.Name != "textSplittingShard" {
			test.Fatal("invalid shard helper")
		}
		literal, ok := call.Args[1].(*ast.BasicLit)
		if !ok || literal.Value != strconv.Itoa(index) {
			test.Fatal("shard name and slice differ")
		}
		seen[index] = true
	}
	if len(seen) != testMarkdownTextSplittingShards {
		test.Fatalf("enumerated %d top-level shards, want %d", len(seen), testMarkdownTextSplittingShards)
	}
}

func textVerifyCorpus(test *testing.T, inputs []auditInput) {
	test.Helper()
	repositoryFiles, cohereFiles, typeScriptFiles := 0, 0, 0
	for _, input := range inputs {
		switch {
		case strings.HasPrefix(input.Name, "generated/"):
		case strings.HasPrefix(input.Name, "cohere/TypeScript/"):
			typeScriptFiles++
		case strings.HasPrefix(input.Name, "cohere/"):
			cohereFiles++
		default:
			repositoryFiles++
		}
	}
	if repositoryFiles == 0 {
		test.Fatal("repository Markdown corpus must be non-empty")
	}
	// These upstream file counts are fixed by corpusfiles.CohereCommit and
	// corpusfiles.TypeScriptGoCommit; repository case totals always remain live.
	if census() && (cohereFiles != 786 || typeScriptFiles != 67) {
		test.Fatalf("pinned corpus lost files: cohere=%d TypeScript=%d", cohereFiles, typeScriptFiles)
	}
}

func textShardInputs(test *testing.T, inputs []auditInput, count int) [][]auditInput {
	test.Helper()
	if count != testMarkdownTextSplittingShards {
		test.Fatal("planner shard count differs from enumeration")
	}
	shards := make([][]auditInput, count)
	expected := make(map[string]bool, len(inputs))
	for _, input := range inputs {
		if expected[input.Name] {
			test.Fatalf("repeated unsplit case id %q", input.Name)
		}
		expected[input.Name] = true
		index := textShardFor(input.Name, count)
		shards[index] = append(shards[index], input)
	}
	seen := make(map[string]bool, len(inputs))
	total := 0
	for index, shard := range shards {
		for _, input := range shard {
			if !expected[input.Name] || seen[input.Name] || textShardFor(input.Name, count) != index {
				test.Fatalf("invalid union case %q in shard-%03d", input.Name, index)
			}
			seen[input.Name] = true
			total++
		}
	}
	if total != len(inputs) || len(seen) != len(expected) {
		test.Fatal("shard union differs from live enumeration")
	}
	if len(inputs) == 0 {
		test.Fatal("empty corpus")
	}
	for name := range expected {
		if !seen[name] {
			test.Fatalf("missing case %q", name)
		}
	}
	test.Logf("live union: %d unique cases, %d shards", total, count)
	return shards
}

func textShardSelected(test *testing.T, shard int) bool {
	test.Helper()
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return true
	}
	fields := strings.Split(value, "/")
	if len(fields) != 2 {
		test.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	i, parseError := strconv.Atoi(fields[0])
	n, f := strconv.Atoi(fields[1])
	if parseError != nil || f != nil || n < 1 || i < 0 || i >= n || n > testMarkdownTextSplittingShards {
		test.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	return shard%n == i
}

func textMutantWitness(name string) string {
	switch name {
	case "fake whitespace":
		return "sequence/a中a"
	case "Hangul kind":
		return "U+001100"
	case "variation selector":
		return "selector/a" + string(rune(0x2c7)) + string(rune(0xfe00))
	}
	panic("unknown text mutant")
}

type textSetupState struct {
	root     string
	inputs   []auditInput
	shards   [][]auditInput
	products textProducts
	ready    bool
}

var textSetupOnce sync.Once
var textSetupShared textSetupState

// Every shard prepares shared products once per process before its own deadline.
// Shared setup has no deadline; Loom bounds the complete unit.
func textReadySetup(test *testing.T) textSetupState {
	test.Helper()
	textSetupOnce.Do(func() {
		started := time.Now()
		root, inputs := textSplittingInputs(test)
		products := textBuildProducts(test, root)
		shards := textShardInputs(test, inputs, testMarkdownTextSplittingShards)
		textSetupShared = textSetupState{root: root, inputs: inputs, shards: shards, products: products, ready: true}
		test.Logf("shared text setup ready in %.3fs", time.Since(started).Seconds())
	})
	if !textSetupShared.ready {
		test.Fatal("shared text setup did not complete")
	}
	return textSetupShared
}

type textProducts struct{ main, goBinary, sanitized, release, backend string }

func textBuildProducts(test *testing.T, root string) textProducts {
	return textBuildProduct(test, root, "")
}

// Product declarations and shard setup use exactly the same recipes and keys.
func textBuildProduct(test *testing.T, root, target string) textProducts {
	test.Helper()
	started := time.Now()
	tools := []string{buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")}
	// The lowered port is keyed by what lowering reads (whitespaceLayoutBuildFiles, which
	// TestLoweringInputsCoverCompilerDependencies keeps complete); the native builds by its C.
	files := whitespaceLayoutBuildFiles()
	main := filepath.Join(root, "stage1/cohere/markdownblocks/testdata/text_probe.ts")
	products := textProducts{main: main}
	if target == "" || target == "lowered" || target == "sanitized" || target == "release" {
		lowerDir := buildcache.Product(test, buildcache.Inputs{
			Name:      "markdownblocks-text-lowered",
			Files:     files,
			Toolchain: tools,
		}, func(dir string) error {
			program, loweredResultError := loweredResult(main)
			if loweredResultError != nil {
				return loweredResultError
			}
			if writeError := os.WriteFile(
				filepath.Join(dir, "program.c"),
				[]byte(native.C(program)),
				0644,
			); writeError != nil {
				return writeError
			}
			return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
		})
		source, readError := os.ReadFile(filepath.Join(lowerDir, "program.c"))
		if readError != nil {
			test.Fatal(readError)
		}
		products.backend = filepath.Join(lowerDir, "program.mjs")
		for _, sanitize := range []bool{true, false} {
			if target != "" &&
				(target == "lowered" || (sanitize && target != "sanitized") || (!sanitize && target != "release")) {
				continue
			}
			options := native.Options{Sanitize: sanitize}
			inputs := buildcache.Inputs{
				Name:  fmt.Sprintf("markdownblocks-text-native-%t", sanitize),
				Files: []string{"internal/native"},
				Flags: append(
					native.Flags(options),
					fmt.Sprintf("source=%x", sha256.Sum256(source)),
					"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"),
				),
				Toolchain: tools,
			}
			dir := buildcache.Product(
				test,
				inputs,
				func(dir string) error { return native.Build(string(source), filepath.Join(dir, "port"), options) },
			)
			if sanitize {
				products.sanitized = filepath.Join(dir, "port")
			} else {
				products.release = filepath.Join(dir, "port")
			}
		}
	}
	if target == "" || target == "go" {
		products.goBinary = oracles.TextGoBuild(test, "go")
	}

	// Regeneration is independent of case partitioning and executes once, after its build inputs.
	var formatterDir, generatorDir string
	if target == "" || target == "formatter" {
		formatterDir = oracles.TextFormatterProduct(test, root)
	}
	if target == "" || target == "generator" {
		generatorDir = oracles.TextGeneratorProduct(test, root)
	}
	if target == "" {
		command, ctx := textSetupCommand(
			test,
			filepath.Join(generatorDir, "generate"),
			"-check",
			"-formatter",
			filepath.Join(formatterDir, "cohere"),
		)
		command.Dir = root
		if output, textCombinedOutputError := textCombinedOutput(ctx, command); textCombinedOutputError != nil {
			test.Fatalf("regeneration: %v\n%s", textCombinedOutputError, output)
		}
	}
	test.Logf("text input preparation including builds %.3fs", time.Since(started).Seconds())
	return products
}

// textGoProductInputs keys the text Go build target names: "go", "formatter" or "generator". Each runs go under
// toolchain.Env (textGoCommand), so its key names the C and C++ compilers that environment gives go, and the C compiler
// by its report: one the key didn't name could change a cgo dependency's bytes unseen.
func textGoProductInputs(target string, compilers markdownCompilers) buildcache.Inputs {
	return oracles.TextGoProductInputs(target, compilers.c, compilers.cPlusPlus, compilers.report)
}

func textLeakReport(test *testing.T, products textProducts, cases string) string {
	test.Helper()
	if runtime.GOOS == "darwin" {
		report := textExecute(test, nil, "leaks", "--atExit", "--", products.release, cases)
		if report.exitCode != 0 {
			return string(report.stdout)
		}
		return ""
	}
	report := textExecute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, products.sanitized, cases)
	if report.exitCode != 0 {
		return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
	}
	return ""
}

// textOutputDifference is the same byte comparison used by every execution side.
func textOutputDifference(inputs []auditInput, got, want []byte) error {
	if bytes.Equal(got, want) {
		return nil
	}
	a := bytes.Split(bytes.TrimSuffix(got, []byte("\n")), []byte("\n"))
	b := bytes.Split(bytes.TrimSuffix(want, []byte("\n")), []byte("\n"))
	for index, input := range inputs {
		if index >= len(a) || index >= len(b) || !bytes.Equal(a[index], b[index]) {
			return fmt.Errorf("mismatch in %q", input.Name)
		}
	}
	return fmt.Errorf("output count or framing differs: got %d lines, want %d", len(a), len(b))
}

// An injected output difference must be caught once, in the case's owning shard.
func TestMarkdownTextShardDisagreement(test *testing.T) {
	test.Parallel()
	key := "sequence/a中a"
	inputs := []auditInput{{Name: key, Text: "a中a"}, {Name: "sequence/aaa", Text: "aaa"}}
	shards := textShardInputs(test, inputs, testMarkdownTextSplittingShards)
	caught := []int{}
	for index, shard := range shards {
		for _, input := range shard {
			want := []byte("tokens\n")
			got := append([]byte(nil), want...)
			if input.Name == key {
				got = []byte("planted disagreement\n")
			}
			if textOutputDifference([]auditInput{input}, got, want) != nil {
				caught = append(caught, index)
			}
		}
	}
	owner := textShardFor(key, testMarkdownTextSplittingShards)
	if len(caught) != 1 || caught[0] != owner {
		test.Fatalf("planted disagreement caught by %v, want only shard-%03d", caught, owner)
	}
	test.Logf("planted disagreement caught only by shard-%03d", owner)
}

// Top-level leaves are visible to go test -list and need no gate planner changes.
func TestMarkdownTextSplitting_000(test *testing.T) { test.Parallel(); textSplittingShard(test, 0) }
func TestMarkdownTextSplitting_001(test *testing.T) { test.Parallel(); textSplittingShard(test, 1) }
func TestMarkdownTextSplitting_002(test *testing.T) { test.Parallel(); textSplittingShard(test, 2) }
func TestMarkdownTextSplitting_003(test *testing.T) { test.Parallel(); textSplittingShard(test, 3) }
func TestMarkdownTextSplitting_004(test *testing.T) { test.Parallel(); textSplittingShard(test, 4) }
func TestMarkdownTextSplitting_005(test *testing.T) { test.Parallel(); textSplittingShard(test, 5) }
func TestMarkdownTextSplitting_006(test *testing.T) { test.Parallel(); textSplittingShard(test, 6) }
func TestMarkdownTextSplitting_007(test *testing.T) { test.Parallel(); textSplittingShard(test, 7) }
func TestMarkdownTextSplitting_008(test *testing.T) { test.Parallel(); textSplittingShard(test, 8) }
func TestMarkdownTextSplitting_009(test *testing.T) { test.Parallel(); textSplittingShard(test, 9) }
func TestMarkdownTextSplitting_010(test *testing.T) { test.Parallel(); textSplittingShard(test, 10) }
func TestMarkdownTextSplitting_011(test *testing.T) { test.Parallel(); textSplittingShard(test, 11) }
func TestMarkdownTextSplitting_012(test *testing.T) { test.Parallel(); textSplittingShard(test, 12) }
func TestMarkdownTextSplitting_013(test *testing.T) { test.Parallel(); textSplittingShard(test, 13) }
func TestMarkdownTextSplitting_014(test *testing.T) { test.Parallel(); textSplittingShard(test, 14) }
func TestMarkdownTextSplitting_015(test *testing.T) { test.Parallel(); textSplittingShard(test, 15) }
func TestMarkdownTextSplitting_016(test *testing.T) { test.Parallel(); textSplittingShard(test, 16) }
func TestMarkdownTextSplitting_017(test *testing.T) { test.Parallel(); textSplittingShard(test, 17) }
func TestMarkdownTextSplitting_018(test *testing.T) { test.Parallel(); textSplittingShard(test, 18) }
func TestMarkdownTextSplitting_019(test *testing.T) { test.Parallel(); textSplittingShard(test, 19) }
func TestMarkdownTextSplitting_020(test *testing.T) { test.Parallel(); textSplittingShard(test, 20) }
func TestMarkdownTextSplitting_021(test *testing.T) { test.Parallel(); textSplittingShard(test, 21) }
func TestMarkdownTextSplitting_022(test *testing.T) { test.Parallel(); textSplittingShard(test, 22) }
func TestMarkdownTextSplitting_023(test *testing.T) { test.Parallel(); textSplittingShard(test, 23) }
func TestMarkdownTextSplitting_024(test *testing.T) { test.Parallel(); textSplittingShard(test, 24) }
func TestMarkdownTextSplitting_025(test *testing.T) { test.Parallel(); textSplittingShard(test, 25) }
func TestMarkdownTextSplitting_026(test *testing.T) { test.Parallel(); textSplittingShard(test, 26) }
func TestMarkdownTextSplitting_027(test *testing.T) { test.Parallel(); textSplittingShard(test, 27) }
func TestMarkdownTextSplitting_028(test *testing.T) { test.Parallel(); textSplittingShard(test, 28) }
func TestMarkdownTextSplitting_029(test *testing.T) { test.Parallel(); textSplittingShard(test, 29) }
func TestMarkdownTextSplitting_030(test *testing.T) { test.Parallel(); textSplittingShard(test, 30) }
func TestMarkdownTextSplitting_031(test *testing.T) { test.Parallel(); textSplittingShard(test, 31) }
func TestMarkdownTextSplitting_032(test *testing.T) { test.Parallel(); textSplittingShard(test, 32) }
func TestMarkdownTextSplitting_033(test *testing.T) { test.Parallel(); textSplittingShard(test, 33) }
func TestMarkdownTextSplitting_034(test *testing.T) { test.Parallel(); textSplittingShard(test, 34) }
func TestMarkdownTextSplitting_035(test *testing.T) { test.Parallel(); textSplittingShard(test, 35) }
func TestMarkdownTextSplitting_036(test *testing.T) { test.Parallel(); textSplittingShard(test, 36) }
func TestMarkdownTextSplitting_037(test *testing.T) { test.Parallel(); textSplittingShard(test, 37) }
func TestMarkdownTextSplitting_038(test *testing.T) { test.Parallel(); textSplittingShard(test, 38) }
func TestMarkdownTextSplitting_039(test *testing.T) { test.Parallel(); textSplittingShard(test, 39) }
func TestMarkdownTextSplitting_040(test *testing.T) { test.Parallel(); textSplittingShard(test, 40) }
func TestMarkdownTextSplitting_041(test *testing.T) { test.Parallel(); textSplittingShard(test, 41) }
func TestMarkdownTextSplitting_042(test *testing.T) { test.Parallel(); textSplittingShard(test, 42) }
func TestMarkdownTextSplitting_043(test *testing.T) { test.Parallel(); textSplittingShard(test, 43) }
func TestMarkdownTextSplitting_044(test *testing.T) { test.Parallel(); textSplittingShard(test, 44) }
func TestMarkdownTextSplitting_045(test *testing.T) { test.Parallel(); textSplittingShard(test, 45) }
func TestMarkdownTextSplitting_046(test *testing.T) { test.Parallel(); textSplittingShard(test, 46) }
func TestMarkdownTextSplitting_047(test *testing.T) { test.Parallel(); textSplittingShard(test, 47) }
func TestMarkdownTextSplitting_048(test *testing.T) { test.Parallel(); textSplittingShard(test, 48) }
func TestMarkdownTextSplitting_049(test *testing.T) { test.Parallel(); textSplittingShard(test, 49) }
func TestMarkdownTextSplitting_050(test *testing.T) { test.Parallel(); textSplittingShard(test, 50) }
func TestMarkdownTextSplitting_051(test *testing.T) { test.Parallel(); textSplittingShard(test, 51) }
func TestMarkdownTextSplitting_052(test *testing.T) { test.Parallel(); textSplittingShard(test, 52) }
func TestMarkdownTextSplitting_053(test *testing.T) { test.Parallel(); textSplittingShard(test, 53) }
func TestMarkdownTextSplitting_054(test *testing.T) { test.Parallel(); textSplittingShard(test, 54) }
func TestMarkdownTextSplitting_055(test *testing.T) { test.Parallel(); textSplittingShard(test, 55) }
func TestMarkdownTextSplitting_056(test *testing.T) { test.Parallel(); textSplittingShard(test, 56) }
func TestMarkdownTextSplitting_057(test *testing.T) { test.Parallel(); textSplittingShard(test, 57) }
func TestMarkdownTextSplitting_058(test *testing.T) { test.Parallel(); textSplittingShard(test, 58) }
func TestMarkdownTextSplitting_059(test *testing.T) { test.Parallel(); textSplittingShard(test, 59) }
func TestMarkdownTextSplitting_060(test *testing.T) { test.Parallel(); textSplittingShard(test, 60) }
func TestMarkdownTextSplitting_061(test *testing.T) { test.Parallel(); textSplittingShard(test, 61) }
func TestMarkdownTextSplitting_062(test *testing.T) { test.Parallel(); textSplittingShard(test, 62) }
func TestMarkdownTextSplitting_063(test *testing.T) { test.Parallel(); textSplittingShard(test, 63) }
