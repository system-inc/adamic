package markdownblocks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
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
)

// Each independent phase is killed at 90s; leaf callers start after setup and admission.
func textDeadline(t *testing.T) func() {
	return textPhaseDeadline(t, t.Name())
}

func textPhaseDeadline(t *testing.T, name string) func() {
	timer := time.AfterFunc(90*time.Second, func() {
		textActiveCommands.Range(func(key, value any) bool {
			if value == t {
				_ = key.(*exec.Cmd).Cancel()
			}
			return true
		})
		panic("cooked: " + name + " exceeded 90s hard deadline")
	})
	return func() { timer.Stop() }
}

// Each child has a context ceiling, and cancellation kills its compiler descendants.
var textActiveCommands sync.Map

func textBounded(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	textActiveCommands.Store(command, t)
	return command
}

func textCombinedOutput(command *exec.Cmd) ([]byte, error) {
	defer textActiveCommands.Delete(command)
	return childguard.CombinedOutput(command, childguard.Options{})
}

func textExecute(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	command := textBounded(t, name, arguments...)
	defer textActiveCommands.Delete(command)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := childguard.Run(command, childguard.Options{})
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}

func textOnNode(t *testing.T, path string, arguments ...string) run {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return textExecute(t, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...)
}

func textShardFor(key string, count int) int {
	sum := sha256.Sum256([]byte(key))
	return int(binary.BigEndian.Uint64(sum[:8]) % uint64(count))
}

func textVerifyLeaves(t *testing.T) {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "text_shards_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
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
		index, err := strconv.Atoi(suffix)
		if err != nil || len(suffix) != 3 || index < 0 || index >= testMarkdownTextSplittingShards || seen[index] {
			t.Fatalf("invalid top-level shard %s", function.Name.Name)
		}
		if len(function.Body.List) != 2 {
			t.Fatalf("shard %s must have t.Parallel followed by one dispatcher", function.Name.Name)
		}
		expression, ok := function.Body.List[1].(*ast.ExprStmt)
		if !ok {
			t.Fatal("invalid shard dispatcher")
		}
		call, ok := expression.X.(*ast.CallExpr)
		if !ok || len(call.Args) != 2 {
			t.Fatal("invalid shard call")
		}
		callee, ok := call.Fun.(*ast.Ident)
		if !ok || callee.Name != "textSplittingShard" {
			t.Fatal("invalid shard helper")
		}
		literal, ok := call.Args[1].(*ast.BasicLit)
		if !ok || literal.Value != strconv.Itoa(index) {
			t.Fatal("shard name and slice differ")
		}
		seen[index] = true
	}
	if len(seen) != testMarkdownTextSplittingShards {
		t.Fatalf("enumerated %d top-level shards, want %d", len(seen), testMarkdownTextSplittingShards)
	}
}

func textVerifyCorpus(t *testing.T, inputs []auditInput) {
	t.Helper()
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
		t.Fatal("repository Markdown corpus must be non-empty")
	}
	// These upstream file counts are fixed by corpusfiles.CohereCommit and
	// corpusfiles.TypeScriptGoCommit; repository case totals always remain live.
	if census() && (cohereFiles != 786 || typeScriptFiles != 67) {
		t.Fatalf("pinned corpus lost files: cohere=%d TypeScript=%d", cohereFiles, typeScriptFiles)
	}
}

func textShardInputs(t *testing.T, inputs []auditInput, count int) [][]auditInput {
	t.Helper()
	if count != testMarkdownTextSplittingShards {
		t.Fatal("planner shard count differs from enumeration")
	}
	shards := make([][]auditInput, count)
	expected := make(map[string]bool, len(inputs))
	for _, input := range inputs {
		if expected[input.Name] {
			t.Fatalf("repeated unsplit case id %q", input.Name)
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
				t.Fatalf("invalid union case %q in shard-%03d", input.Name, index)
			}
			seen[input.Name] = true
			total++
		}
	}
	if total != len(inputs) || len(seen) != len(expected) {
		t.Fatal("shard union differs from live enumeration")
	}
	if len(inputs) == 0 {
		t.Fatal("empty corpus")
	}
	for name := range expected {
		if !seen[name] {
			t.Fatalf("missing case %q", name)
		}
	}
	t.Logf("live union: %d unique cases, %d shards", total, count)
	return shards
}

func textShardSelected(t *testing.T, shard int) bool {
	t.Helper()
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return true
	}
	fields := strings.Split(value, "/")
	if len(fields) != 2 {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	i, e := strconv.Atoi(fields[0])
	n, f := strconv.Atoi(fields[1])
	if e != nil || f != nil || n < 1 || i < 0 || i >= n || n > testMarkdownTextSplittingShards {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
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

// The serial TestMarkdownTextSplitting_Setup initializes this before parallel
// leaves resume. A filtered run that omits Setup uses the same separately bounded
// setup phase; its leaf deadline is still started only after readiness.
func textReadySetup(t *testing.T) textSetupState {
	t.Helper()
	textSetupOnce.Do(func() {
		stop := textPhaseDeadline(t, "TestMarkdownTextSplitting_Setup")
		defer stop()
		started := time.Now()
		root, inputs := textSplittingInputs(t)
		products := textBuildProducts(t, root)
		shards := textShardInputs(t, inputs, testMarkdownTextSplittingShards)
		textSetupShared = textSetupState{root: root, inputs: inputs, shards: shards, products: products, ready: true}
		t.Logf("shared text setup ready in %.3fs", time.Since(started).Seconds())
	})
	if !textSetupShared.ready {
		t.Fatal("shared text setup did not complete")
	}
	return textSetupShared
}

type textProducts struct{ main, goBinary, sanitized, release, backend string }

func textBuildProducts(t *testing.T, root string) textProducts {
	t.Helper()
	started := time.Now()
	tools := []string{buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")}
	goFlags := []string{"GOFLAGS=" + os.Getenv("GOFLAGS"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED"), "GOOS=" + os.Getenv("GOOS"), "GOARCH=" + os.Getenv("GOARCH"), "CC=" + os.Getenv("CC"), "CXX=" + os.Getenv("CXX")}
	files := []string{"go.mod", "cohere", "internal", "stage1/cohere/markdownblocks", "oracle"}
	main := filepath.Join(root, "stage1/cohere/markdownblocks/testdata/text_probe.ts")
	lowerDir := buildcache.Product(t, buildcache.Inputs{Name: "markdownblocks-text-lowered", Files: files, Toolchain: tools}, func(dir string) error {
		program, err := loweredResult(main)
		if err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(program)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	source, err := os.ReadFile(filepath.Join(lowerDir, "program.c"))
	if err != nil {
		t.Fatal(err)
	}
	products := textProducts{main: main, backend: filepath.Join(lowerDir, "program.mjs")}
	for _, sanitize := range []bool{true, false} {
		options := native.Options{Sanitize: sanitize}
		inputs := buildcache.Inputs{Name: fmt.Sprintf("markdownblocks-text-native-%t", sanitize), Files: []string{"internal/native"}, Flags: append(native.Flags(options), fmt.Sprintf("source=%x", sha256.Sum256(source)), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS="+os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED="+os.Getenv("ADAMIC_GATE_UNCACHED")), Toolchain: tools}
		dir := buildcache.Product(t, inputs, func(dir string) error { return native.Build(string(source), filepath.Join(dir, "port"), options) })
		if sanitize {
			products.sanitized = filepath.Join(dir, "port")
		} else {
			products.release = filepath.Join(dir, "port")
		}
	}
	goDir := buildcache.Product(t, buildcache.Inputs{Name: "markdownblocks-text-go", Files: []string{"cohere", "stage1/cohere/markdownblocks/testdata/text_go.go", "stage1/cohere/markdownblocks/testdata/text_bridge.go"}, Flags: append(goFlags, "go build", "overlay text_go.go,text_bridge.go"), Toolchain: tools[:1]}, func(dir string) error {
		cohere := filepath.Join(root, "cohere")
		mainPath := filepath.Join(cohere, "cmd/adamic_text/main.go")
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{mainPath: filepath.Join(root, "stage1/cohere/markdownblocks/testdata/text_go.go"), filepath.Join(cohere, "internal/format/markdown/adamic_text.go"): filepath.Join(root, "stage1/cohere/markdownblocks/testdata/text_bridge.go")}})
		if err != nil {
			return err
		}
		overlayPath := filepath.Join(dir, "overlay.json")
		if err = os.WriteFile(overlayPath, overlay, 0644); err != nil {
			return err
		}
		command := textBounded(t, "go", "build", "-overlay="+overlayPath, "-o", filepath.Join(dir, "go-text"), mainPath)
		command.Dir = cohere
		output, err := textCombinedOutput(command)
		if err != nil {
			return fmt.Errorf("Go splitText: %w\n%s", err, output)
		}
		return nil
	})
	products.goBinary = filepath.Join(goDir, "go-text")
	// Regeneration is independent of case partitioning and executes once, after its build inputs.
	formatterDir := buildcache.Product(t, buildcache.Inputs{Name: "markdownblocks-text-formatter", Files: []string{"cohere"}, Flags: goFlags, Toolchain: tools[:1]}, func(dir string) error {
		command := textBounded(t, "go", "build", "-o", filepath.Join(dir, "cohere"), "./command/cohere")
		command.Dir = filepath.Join(root, "cohere")
		output, err := textCombinedOutput(command)
		if err != nil {
			return fmt.Errorf("formatter: %w\n%s", err, output)
		}
		return nil
	})
	generatorDir := buildcache.Product(t, buildcache.Inputs{Name: "markdownblocks-text-generator", Files: []string{"go.mod", "stage1/cohere/markdownblocks/tools/generate_classes"}, Flags: goFlags, Toolchain: tools[:1]}, func(dir string) error {
		command := textBounded(t, "go", "build", "-o", filepath.Join(dir, "generate"), "./stage1/cohere/markdownblocks/tools/generate_classes")
		command.Dir = root
		output, err := textCombinedOutput(command)
		if err != nil {
			return fmt.Errorf("generator: %w\n%s", err, output)
		}
		return nil
	})
	command := textBounded(t, filepath.Join(generatorDir, "generate"), "-check", "-formatter", filepath.Join(formatterDir, "cohere"))
	command.Dir = root
	if output, err := textCombinedOutput(command); err != nil {
		t.Fatalf("regeneration: %v\n%s", err, output)
	}
	t.Logf("text input preparation including builds %.3fs", time.Since(started).Seconds())
	return products
}

func textLeakReport(t *testing.T, products textProducts, cases string) string {
	t.Helper()
	if runtime.GOOS == "darwin" {
		report := textExecute(t, nil, "leaks", "--atExit", "--", products.release, cases)
		if report.exitCode != 0 {
			return string(report.stdout)
		}
		return ""
	}
	report := textExecute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, products.sanitized, cases)
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
func TestMarkdownTextShardDisagreement(t *testing.T) {
	t.Parallel()
	key := "sequence/a中a"
	inputs := []auditInput{{Name: key, Text: "a中a"}, {Name: "sequence/aaa", Text: "aaa"}}
	shards := textShardInputs(t, inputs, testMarkdownTextSplittingShards)
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
		t.Fatalf("planted disagreement caught by %v, want only shard-%03d", caught, owner)
	}
	t.Logf("planted disagreement caught only by shard-%03d", owner)
}

// Top-level leaves are visible to go test -list and need no gate planner changes.
func TestMarkdownTextSplitting_000(t *testing.T) { t.Parallel(); textSplittingShard(t, 0) }
func TestMarkdownTextSplitting_001(t *testing.T) { t.Parallel(); textSplittingShard(t, 1) }
func TestMarkdownTextSplitting_002(t *testing.T) { t.Parallel(); textSplittingShard(t, 2) }
func TestMarkdownTextSplitting_003(t *testing.T) { t.Parallel(); textSplittingShard(t, 3) }
func TestMarkdownTextSplitting_004(t *testing.T) { t.Parallel(); textSplittingShard(t, 4) }
func TestMarkdownTextSplitting_005(t *testing.T) { t.Parallel(); textSplittingShard(t, 5) }
func TestMarkdownTextSplitting_006(t *testing.T) { t.Parallel(); textSplittingShard(t, 6) }
func TestMarkdownTextSplitting_007(t *testing.T) { t.Parallel(); textSplittingShard(t, 7) }
func TestMarkdownTextSplitting_008(t *testing.T) { t.Parallel(); textSplittingShard(t, 8) }
func TestMarkdownTextSplitting_009(t *testing.T) { t.Parallel(); textSplittingShard(t, 9) }
func TestMarkdownTextSplitting_010(t *testing.T) { t.Parallel(); textSplittingShard(t, 10) }
func TestMarkdownTextSplitting_011(t *testing.T) { t.Parallel(); textSplittingShard(t, 11) }
func TestMarkdownTextSplitting_012(t *testing.T) { t.Parallel(); textSplittingShard(t, 12) }
func TestMarkdownTextSplitting_013(t *testing.T) { t.Parallel(); textSplittingShard(t, 13) }
func TestMarkdownTextSplitting_014(t *testing.T) { t.Parallel(); textSplittingShard(t, 14) }
func TestMarkdownTextSplitting_015(t *testing.T) { t.Parallel(); textSplittingShard(t, 15) }
func TestMarkdownTextSplitting_016(t *testing.T) { t.Parallel(); textSplittingShard(t, 16) }
func TestMarkdownTextSplitting_017(t *testing.T) { t.Parallel(); textSplittingShard(t, 17) }
func TestMarkdownTextSplitting_018(t *testing.T) { t.Parallel(); textSplittingShard(t, 18) }
func TestMarkdownTextSplitting_019(t *testing.T) { t.Parallel(); textSplittingShard(t, 19) }
func TestMarkdownTextSplitting_020(t *testing.T) { t.Parallel(); textSplittingShard(t, 20) }
func TestMarkdownTextSplitting_021(t *testing.T) { t.Parallel(); textSplittingShard(t, 21) }
func TestMarkdownTextSplitting_022(t *testing.T) { t.Parallel(); textSplittingShard(t, 22) }
func TestMarkdownTextSplitting_023(t *testing.T) { t.Parallel(); textSplittingShard(t, 23) }
func TestMarkdownTextSplitting_024(t *testing.T) { t.Parallel(); textSplittingShard(t, 24) }
func TestMarkdownTextSplitting_025(t *testing.T) { t.Parallel(); textSplittingShard(t, 25) }
func TestMarkdownTextSplitting_026(t *testing.T) { t.Parallel(); textSplittingShard(t, 26) }
func TestMarkdownTextSplitting_027(t *testing.T) { t.Parallel(); textSplittingShard(t, 27) }
func TestMarkdownTextSplitting_028(t *testing.T) { t.Parallel(); textSplittingShard(t, 28) }
func TestMarkdownTextSplitting_029(t *testing.T) { t.Parallel(); textSplittingShard(t, 29) }
func TestMarkdownTextSplitting_030(t *testing.T) { t.Parallel(); textSplittingShard(t, 30) }
func TestMarkdownTextSplitting_031(t *testing.T) { t.Parallel(); textSplittingShard(t, 31) }
func TestMarkdownTextSplitting_032(t *testing.T) { t.Parallel(); textSplittingShard(t, 32) }
func TestMarkdownTextSplitting_033(t *testing.T) { t.Parallel(); textSplittingShard(t, 33) }
func TestMarkdownTextSplitting_034(t *testing.T) { t.Parallel(); textSplittingShard(t, 34) }
func TestMarkdownTextSplitting_035(t *testing.T) { t.Parallel(); textSplittingShard(t, 35) }
func TestMarkdownTextSplitting_036(t *testing.T) { t.Parallel(); textSplittingShard(t, 36) }
func TestMarkdownTextSplitting_037(t *testing.T) { t.Parallel(); textSplittingShard(t, 37) }
func TestMarkdownTextSplitting_038(t *testing.T) { t.Parallel(); textSplittingShard(t, 38) }
func TestMarkdownTextSplitting_039(t *testing.T) { t.Parallel(); textSplittingShard(t, 39) }
func TestMarkdownTextSplitting_040(t *testing.T) { t.Parallel(); textSplittingShard(t, 40) }
func TestMarkdownTextSplitting_041(t *testing.T) { t.Parallel(); textSplittingShard(t, 41) }
func TestMarkdownTextSplitting_042(t *testing.T) { t.Parallel(); textSplittingShard(t, 42) }
func TestMarkdownTextSplitting_043(t *testing.T) { t.Parallel(); textSplittingShard(t, 43) }
func TestMarkdownTextSplitting_044(t *testing.T) { t.Parallel(); textSplittingShard(t, 44) }
func TestMarkdownTextSplitting_045(t *testing.T) { t.Parallel(); textSplittingShard(t, 45) }
func TestMarkdownTextSplitting_046(t *testing.T) { t.Parallel(); textSplittingShard(t, 46) }
func TestMarkdownTextSplitting_047(t *testing.T) { t.Parallel(); textSplittingShard(t, 47) }
func TestMarkdownTextSplitting_048(t *testing.T) { t.Parallel(); textSplittingShard(t, 48) }
func TestMarkdownTextSplitting_049(t *testing.T) { t.Parallel(); textSplittingShard(t, 49) }
func TestMarkdownTextSplitting_050(t *testing.T) { t.Parallel(); textSplittingShard(t, 50) }
func TestMarkdownTextSplitting_051(t *testing.T) { t.Parallel(); textSplittingShard(t, 51) }
func TestMarkdownTextSplitting_052(t *testing.T) { t.Parallel(); textSplittingShard(t, 52) }
func TestMarkdownTextSplitting_053(t *testing.T) { t.Parallel(); textSplittingShard(t, 53) }
func TestMarkdownTextSplitting_054(t *testing.T) { t.Parallel(); textSplittingShard(t, 54) }
func TestMarkdownTextSplitting_055(t *testing.T) { t.Parallel(); textSplittingShard(t, 55) }
func TestMarkdownTextSplitting_056(t *testing.T) { t.Parallel(); textSplittingShard(t, 56) }
func TestMarkdownTextSplitting_057(t *testing.T) { t.Parallel(); textSplittingShard(t, 57) }
func TestMarkdownTextSplitting_058(t *testing.T) { t.Parallel(); textSplittingShard(t, 58) }
func TestMarkdownTextSplitting_059(t *testing.T) { t.Parallel(); textSplittingShard(t, 59) }
func TestMarkdownTextSplitting_060(t *testing.T) { t.Parallel(); textSplittingShard(t, 60) }
func TestMarkdownTextSplitting_061(t *testing.T) { t.Parallel(); textSplittingShard(t, 61) }
func TestMarkdownTextSplitting_062(t *testing.T) { t.Parallel(); textSplittingShard(t, 62) }
func TestMarkdownTextSplitting_063(t *testing.T) { t.Parallel(); textSplittingShard(t, 63) }
