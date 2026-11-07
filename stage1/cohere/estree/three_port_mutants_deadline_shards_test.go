package estree

import (
	"bytes"
	"context"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/stage1/cohere/estree/oracles"
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
)

const testThreePortMutantsShards = 12
const threePortCasesPerShard = 20

type threePortMutation struct{ name, file, from, to, witness string }

func threePortMutations() []threePortMutation {
	return []threePortMutation{
		{
			"member-computed",
			"convert.ts",
			"boolValue(node.kind === 'ElementAccessExpression')",
			"boolValue(node.kind === 'PropertyAccessExpression')",
			"a.b; a[b];",
		},
		{
			"logical-rebalance",
			"postprocess.ts",
			"completed.set(id, this.rebalance(id));",
			"completed.set(id, id);",
			"a && (b && (c && d));",
		},
		{"merged-jsdoc-value", "postprocess.ts", "*//*", "*/ /*", "x; /**\n * one\n *//**\n * two\n */ y;"},
	}
}

type threePortCase struct {
	mutant, index int
	witness       bool
}
type threePortShard struct {
	name  string
	cases []threePortCase
}

func threePortShards(test *testing.T) []threePortShard {
	file, parseFileError := parser.ParseFile(token.NewFileSet(), "three_port_mutants_deadline_shards_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	names := map[string]bool{}
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok &&
			strings.HasPrefix(funcDecl.Name.Name, "TestThreePortMutants_") &&
			len(funcDecl.Name.Name) == len("TestThreePortMutants_000") {
			names[funcDecl.Name.Name] = true
		}
	}
	if len(names) != testThreePortMutantsShards {
		test.Fatalf("top-level enumeration: %d, declared %d", len(names), testThreePortMutantsShards)
	}
	for index := 0; index < testThreePortMutantsShards; index++ {
		if !names[fmt.Sprintf("TestThreePortMutants_%03d", index)] {
			test.Fatalf("missing shard %03d", index)
		}
	}
	return threePortShardPlan(test, generated(), threePortMutations(), testThreePortMutantsShards)
}

func threePortShardPlan(test *testing.T, cases []string, mutations []threePortMutation, declared int) []threePortShard {
	test.Helper()
	var shards []threePortShard
	var unsplit []threePortCase
	for m, mutation := range mutations {
		witnesses := 0
		var pairs []threePortCase
		for index, text := range cases {
			pair := threePortCase{m, index, strings.Contains(text, mutation.witness)}
			if pair.witness {
				witnesses++
			}
			pairs = append(pairs, pair)
			unsplit = append(unsplit, pair)
		}
		if witnesses != 1 {
			test.Fatalf("%s: expected one witness, got %d", mutation.name, witnesses)
		}
		for start := 0; start < len(pairs); start += threePortCasesPerShard {
			end := min(start+threePortCasesPerShard, len(pairs))
			shards = append(shards, threePortShard{fmt.Sprintf("shard-%03d", len(shards)), pairs[start:end]})
		}
	}
	if len(shards) != declared {
		test.Fatalf("enumerated %d shards, declared %d", len(shards), declared)
	}
	want := make(map[[2]int]bool)
	for _, pair := range unsplit {
		id := [2]int{pair.mutant, pair.index}
		if want[id] {
			test.Fatalf("duplicate unsplit case %v", id)
		}
		want[id] = true
	}
	seen := make(map[[2]int]bool)
	count := 0
	for _, shard := range shards {
		for _, pair := range shard.cases {
			id := [2]int{pair.mutant, pair.index}
			if !want[id] || seen[id] {
				test.Fatalf("%s: extra or repeated case %v", shard.name, id)
			}
			seen[id] = true
			count++
		}
	}
	if count != len(unsplit) || len(seen) != len(want) {
		test.Fatalf("shard union: %d/%d cases, %d/%d ids", count, len(unsplit), len(seen), len(want))
	}
	for id := range want {
		if !seen[id] {
			test.Fatalf("shard union missing %v", id)
		}
	}
	test.Logf(
		"shard union: %d pairs, %d distinct ids, %d mutants x %d cases",
		count,
		len(seen),
		len(mutations),
		len(cases),
	)
	return shards
}

// Canonical output ends every whole-file record with one escaped "stripped"
// line. Do not divide a file's AST, comments, or stripped-source checks.
func threePortCanonicalRecords(test *testing.T, data []byte, count int) [][]byte {
	test.Helper()
	var records [][]byte
	start, offset := 0, 0
	for _, line := range bytes.SplitAfter(data, []byte("\n")) {
		offset += len(line)
		if bytes.HasPrefix(line, []byte("stripped ")) {
			records = append(records, data[start:offset])
			start = offset
		}
	}
	if start != len(data) || len(records) != count {
		test.Fatalf("canonical records: %d, want %d; trailing bytes %d", len(records), count, len(data)-start)
	}
	return records
}

func checkThreePortCase(pair threePortCase, want, source, native []byte) error {
	sourceDifference := firstDifference(want, source)
	nativeDifference := firstDifference(want, native)
	if pair.witness {
		if sourceDifference == "" {
			return fmt.Errorf("case %03d: source Node mutant survived", pair.index)
		}
		if nativeDifference == "" {
			return fmt.Errorf("case %03d: sanitized native mutant survived", pair.index)
		}
	}
	if (sourceDifference == "") != (nativeDifference == "") {
		return fmt.Errorf("case %03d: mutant ports differ in agreement with Go", pair.index)
	}
	if diff := firstDifference(source, native); diff != "" {
		return fmt.Errorf("case %03d: mutant ports disagree: %s", pair.index, diff)
	}
	return nil
}

// Products are addressed by live source content and built once across shard processes.
func threePortLoweredProduct(test *testing.T, mutation threePortMutation) string {
	test.Helper()
	inputs := buildcache.Inputs{Name: "three-port-mutant-lowered-" + mutation.name,
		Files: []string{
			"stage1/cohere/estree",
			"stage1/typescript",
			"internal",
			"cohere/TypeScript/tsc",
			"cohere/TypeScript-shim",
			"go.mod",
			"go.work",
		},
		Flags:     []string{mutation.file, mutation.from, mutation.to, "repository=" + root(test)},
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH}}
	lowered := buildcache.Product(test, inputs, func(dir string) error {
		source := filepath.Join(dir, "source")
		if createDirectoryError := os.Mkdir(source, 0755); createDirectoryError != nil {
			return createDirectoryError
		}
		files, globError := filepath.Glob("*.ts")
		if globError != nil {
			return globError
		}
		for _, name := range files {
			data, readError := os.ReadFile(name)
			if readError != nil {
				return readError
			}
			text := string(data)
			if name == mutation.file {
				if strings.Count(text, mutation.from) != 1 {
					return fmt.Errorf("mutant anchor count changed")
				}
				text = strings.Replace(text, mutation.from, mutation.to, 1)
			}
			text = strings.ReplaceAll(
				text,
				"'../../typescript/",
				"'"+filepath.ToSlash(filepath.Join(root(test), "stage1/typescript"))+"/",
			)
			if writeError := os.WriteFile(filepath.Join(source, name), []byte(text), 0644); writeError != nil {
				return writeError
			}
		}
		program, loadError := load.Load([]string{filepath.Join(source, "main.ts")})
		if loadError != nil {
			return loadError
		}
		ir, lowerError := lower.Lower(context.Background(), program)
		if lowerError != nil {
			return lowerError
		}
		if writeError := os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(ir)), 0644); writeError != nil {
			return writeError
		}
		// The source's imports name this checkout for the lowering above; the product names it <repository> (#tqrqx60).
		return buildcache.RelativeFiles(source)
	})
	return lowered
}

func threePortNativeProduct(test *testing.T, mutation threePortMutation) string {
	test.Helper()
	lowered := threePortLoweredProduct(test, mutation)
	inputs := buildcache.Inputs{Name: "three-port-mutant-native-" + mutation.name,
		Files: []string{
			"stage1/cohere/estree",
			"stage1/typescript",
			"internal",
			"cohere/TypeScript/tsc",
			"cohere/TypeScript-shim",
			"go.mod",
			"go.work",
		},
		Flags:     []string{mutation.file, mutation.from, mutation.to, "repository=" + root(test)},
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH}}
	inputs.Flags = append(inputs.Flags, "split=true")
	inputs.Flags = append(inputs.Flags, native.Flags(native.Options{Sanitize: true, Split: true})...)
	inputs.Flags = append(inputs.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	binary := buildcache.Product(test, inputs, func(dir string) error {
		content, readError := os.ReadFile(filepath.Join(lowered, "program.c"))
		if readError != nil {
			return readError
		}
		return native.Build(string(content), filepath.Join(dir, "port"), native.Options{Sanitize: true, Split: true})
	})
	return filepath.Join(binary, "port")
}

type threePortPreparedPort struct{ Source, Native string }
type threePortPrepared struct {
	Oracle string
	Ports  []threePortPreparedPort
}

// Shared preparation has no deadline. Every selected shard prepares once per
// process before its own clock starts; Loom still bounds the whole unit.
var threePortShared struct {
	once     sync.Once
	prepared *threePortPrepared
}

func threePortOracleInputs(test *testing.T) buildcache.Inputs {
	test.Helper()
	return buildcache.Inputs{Name: "three-port-mutants-go-oracle-v1",
		Files: []string{
			"stage1/cohere/estree",
			"stage1/typescript",
			"internal",
			"cohere/internal",
			"cohere/TypeScript/tsc",
			"cohere/TypeScript-shim",
			"cohere/go.mod",
			"cohere/go.sum",
			"go.mod",
			"go.work",
		},
		Flags: []string{
			"repository=" + root(test),
			"go build " + strings.Join(buildcache.ReproducibleGoFlags(), " "),
			"overlay=stage1/cohere/estree/testdata/oracle.go",
			"virtual=cohere/adamic_estree_oracle.go",
			"sanitize=true",
			"split=true",
			"ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"),
			"GOFLAGS=" + os.Getenv("GOFLAGS"),
			"CGO_ENABLED=" + os.Getenv("CGO_ENABLED"),
		},
		Toolchain: []string{
			runtime.Version(),
			runtime.GOOS,
			runtime.GOARCH,
			buildcache.Tool("go", "version"),
			buildcache.Tool("clang", "--version"),
		}}
}

func threePortSetupContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

func threePortOracleProduct(test *testing.T) string {
	test.Helper()
	return oracles.ThreePortMutantsOracle(test)
}

func threePortReady(test *testing.T) *threePortPrepared {
	test.Helper()
	threePortShared.once.Do(func() {
		prepared := &threePortPrepared{Oracle: threePortOracleProduct(test)}
		for _, mutation := range threePortMutations() {
			lowered := threePortLoweredProduct(test, mutation)
			prepared.Ports = append(prepared.Ports, threePortPreparedPort{
				resolvedPort(test, lowered, "source/main.ts"), threePortNativeProduct(test, mutation),
			})
		}
		threePortShared.prepared = prepared
	})
	if threePortShared.prepared == nil {
		test.Fatal("shared product build failed")
	}
	return threePortShared.prepared
}

// threePortCommand's output is read through childguard.Capture, to its end with no timer deciding when: exec's own
// copying under a two-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF two seconds
// after the child exited, so an oracle, Node, native or proof run that had finished could fail its shard under load
// (#n3gk22h's shape). Each caller's deadline ends a hung run.
func threePortCommand(callerContext context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(callerContext, name, args...)
	// A tool named bare (go, node) is toolchain's program, under toolchain's environment; a binary named by its path
	// runs as itself.
	if !strings.ContainsRune(name, filepath.Separator) {
		tool := toolchain.Command(name, args...)
		command.Path, command.Err, command.Env = tool.Path, tool.Err, tool.Env
	}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if killError == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return killError
	}
	return command
}

func threePortExecute(test *testing.T, callerContext context.Context, name string, args ...string) []byte {
	test.Helper()
	output, stderr, captureError := childguard.Capture(callerContext, threePortCommand(callerContext, name, args...))
	if callerContext.Err() != nil {
		test.Fatalf("cooked: %s exceeded leaf's 90s deadline", test.Name())
	}
	if captureError != nil || len(stderr) != 0 {
		test.Fatalf("%s %v: %v\n%s", name, args, captureError, stderr)
	}
	return output
}

func TestThreePortMutants_Setup(test *testing.T) {
	test.Parallel()
	threePortReady(test)
}

func TestThreePortMutants_SetupDeadlineProof(test *testing.T) {
	test.Parallel()
	ctx, cancel := threePortSetupContext()
	defer cancel()
	if _, ok := ctx.Deadline(); ok {
		test.Fatal("shared setup has a deadline")
	}
	own, stop := threePortOwnWorkContext(100 * time.Millisecond)
	defer stop()
	command := threePortCommand(own, "node", "-e", "setInterval(() => {}, 1000)")
	if runError := command.Run(); runError == nil {
		test.Fatal("own-work deadline did not kill its child")
	}
	if own.Err() != context.DeadlineExceeded {
		test.Fatalf("own work deadline did not fire: %v", own.Err())
	}
	if ctx.Err() != nil {
		test.Fatalf("own work deadline cancelled setup: %v", ctx.Err())
	}
}

func threePortOwnWorkContext(budget time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), budget)
}

func TestProduct_EstreeFindsThreePortOracle(test *testing.T) {
	test.Parallel()
	threePortOracleProduct(test)
}
func TestProduct_ThreePortLowered0(test *testing.T) {
	test.Parallel()
	threePortLoweredProduct(test, threePortMutations()[0])
}
func TestProduct_ThreePortNative0(test *testing.T) {
	test.Parallel()
	threePortNativeProduct(test, threePortMutations()[0])
}
func TestProduct_ThreePortLowered1(test *testing.T) {
	test.Parallel()
	threePortLoweredProduct(test, threePortMutations()[1])
}
func TestProduct_ThreePortNative1(test *testing.T) {
	test.Parallel()
	threePortNativeProduct(test, threePortMutations()[1])
}
func TestProduct_ThreePortLowered2(test *testing.T) {
	test.Parallel()
	threePortLoweredProduct(test, threePortMutations()[2])
}
func TestProduct_ThreePortNative2(test *testing.T) {
	test.Parallel()
	threePortNativeProduct(test, threePortMutations()[2])
}

func runThreePortShard(test *testing.T, index int) {
	test.Helper()
	shards := threePortShards(test)
	shard := shards[index]
	if planted := os.Getenv("ADAMIC_THREE_PORT_PROOF"); planted != "" {
		number, parseError := strconv.Atoi(planted)
		if parseError != nil {
			test.Fatal(parseError)
		}
		for _, pair := range shard.cases {
			want, source, native := []byte("oracle"), []byte("oracle"), []byte("oracle")
			if pair.witness {
				source, native = []byte("mutant"), []byte("mutant")
			}
			if pair.mutant == number && pair.witness {
				native = want
			}
			if checkThreePortCaseError := checkThreePortCase(
				pair,
				want,
				source,
				native,
			); checkThreePortCaseError != nil {
				test.Fatal(checkThreePortCaseError)
			}
		}
		return
	}
	// Prepare shared products before starting this shard's own-work deadline.
	prepared := threePortReady(test)
	ctx, cancel := threePortOwnWorkContext(90 * time.Second)
	defer cancel()
	m := shard.cases[0].mutant
	mutation := threePortMutations()[m]
	oracle := prepared.Oracle
	sourcePath, binary := prepared.Ports[m].Source, prepared.Ports[m].Native
	texts := make([]string, len(shard.cases))
	for index, pair := range shard.cases {
		texts[index] = generated()[pair.index]
	}
	list := manifest(test, texts)
	want := threePortCanonicalRecords(test, threePortExecute(test, ctx, oracle, "--manifest", list), len(texts))
	source := threePortCanonicalRecords(test, threePortExecute(
		test,
		ctx,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(root(test), "oracle/node.mjs"),
		sourcePath,
		"--manifest",
		list,
	), len(texts))
	native := threePortCanonicalRecords(test, threePortExecute(test, ctx, binary, "--manifest", list), len(texts))
	for index, pair := range shard.cases {
		if checkThreePortCaseError := checkThreePortCase(
			pair,
			want[index],
			source[index],
			native[index],
		); checkThreePortCaseError != nil {
			test.Fatal(checkThreePortCaseError)
		}
	}
	test.Logf("%s: %d cases", mutation.name, len(texts))
}

func TestThreePortMutants_000(test *testing.T) { test.Parallel(); runThreePortShard(test, 0) }
func TestThreePortMutants_001(test *testing.T) { test.Parallel(); runThreePortShard(test, 1) }
func TestThreePortMutants_002(test *testing.T) { test.Parallel(); runThreePortShard(test, 2) }
func TestThreePortMutants_003(test *testing.T) { test.Parallel(); runThreePortShard(test, 3) }
func TestThreePortMutants_004(test *testing.T) { test.Parallel(); runThreePortShard(test, 4) }
func TestThreePortMutants_005(test *testing.T) { test.Parallel(); runThreePortShard(test, 5) }
func TestThreePortMutants_006(test *testing.T) { test.Parallel(); runThreePortShard(test, 6) }
func TestThreePortMutants_007(test *testing.T) { test.Parallel(); runThreePortShard(test, 7) }
func TestThreePortMutants_008(test *testing.T) { test.Parallel(); runThreePortShard(test, 8) }
func TestThreePortMutants_009(test *testing.T) { test.Parallel(); runThreePortShard(test, 9) }
func TestThreePortMutants_010(test *testing.T) { test.Parallel(); runThreePortShard(test, 10) }
func TestThreePortMutants_011(test *testing.T) { test.Parallel(); runThreePortShard(test, 11) }

// Plant a surviving native witness and execute each actual top-level shard
// independently. This proof performs no builds and never measures the whole test.
func TestThreePortMutants_ShardProof(test *testing.T) {
	test.Parallel()
	shards := threePortShards(test)
	executable, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	for m, mutation := range threePortMutations() {
		failed, caught, expected, caseID := 0, -1, -1, -1
		for s, shard := range shards {
			for _, pair := range shard.cases {
				if pair.mutant == m && pair.witness {
					expected = s
					caseID = pair.index
				}
			}
		}
		for s := range shards {
			name := fmt.Sprintf("TestThreePortMutants_%03d", s)
			timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			command := threePortCommand(
				timeoutContext,
				executable,
				"-test.run=^"+name+"$",
				"-test.timeout=90s",
				"-test.v",
			)
			command.Env = append(os.Environ(), fmt.Sprintf("ADAMIC_THREE_PORT_PROOF=%d", m))
			stdout, stderr, captureError := childguard.Capture(timeoutContext, command)
			output := append(stdout, stderr...)
			deadlineError := timeoutContext.Err()
			cancel()
			if deadlineError != nil {
				test.Fatalf("proof child exceeded 90s: %v", deadlineError)
			}
			if captureError != nil {
				exit, ok := captureError.(*exec.ExitError)
				if !ok || exit.ExitCode() != 1 {
					test.Fatalf("proof child: %v: %s", captureError, output)
				}
				failed++
				caught = s
				if s != expected || !strings.Contains(
					string(output),
					fmt.Sprintf("case %03d: sanitized native mutant survived", caseID),
				) {
					test.Fatalf("unexpected failure in %s: %s", name, output)
				}
			}
		}
		if failed != 1 || caught != expected {
			test.Fatalf("planted %s: %d failing shards, caught %d want %d", mutation.name, failed, caught, expected)
		}
		test.Logf(
			"planted surviving native %s case %03d caught by TestThreePortMutants_%03d; exactly one shard",
			mutation.name,
			caseID,
			caught,
		)
	}
}
