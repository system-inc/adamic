package estree

import (
	"context"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/estree/oracles"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const testAcceptanceMutantsShards = 2

type acceptanceMutant struct{ name, file, from, to string }

func acceptanceMutantEnumeration() []acceptanceMutant {
	return []acceptanceMutant{
		{
			"catch-initializer",
			"convert.ts",
			"this.separated(this.child(declaration, 0), this.child(declaration, 1), 'ColonToken')",
			"true",
		},
		{
			"class-keyword-name",
			"sourceStatements.ts",
			"!(this.parser.peek() === 'Identifier' || this.parser.peek().endsWith('Keyword'))",
			"false",
		},
	}
}

// The pinned mutation table partitions into one complete manifest per shard.
// Comparing complete canonical bytes retains the original mutant-must-fail oracle.
func acceptanceMutantCheck(want, got []byte) error {
	if firstDifference(want, got) == "" {
		return fmt.Errorf("mutant survived")
	}
	return nil
}

func acceptanceMutantProducts(test *testing.T, item acceptanceMutant, compile bool) (string, string) {
	test.Helper()
	inputs := buildcache.Inputs{
		Name:  "estree-acceptance-mutant-lowered-" + item.name,
		Files: []string{"stage1/cohere/estree", "stage1/typescript", "internal", "cohere", "go.mod"},
		Flags: []string{
			item.file,
			item.from,
			item.to,
			"repository=" + root(test),
			"ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"),
		}, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}
	lowered := buildcache.Product(test, inputs, func(dir string) error {
		main := mutantPort(test, item.file, item.from, item.to)
		program, loadError := load.Load([]string{main})
		if loadError != nil {
			return loadError
		}
		lowered, lowerError := lower.Lower(context.Background(), program)
		if lowerError != nil {
			return lowerError
		}
		if writeError := os.WriteFile(
			filepath.Join(dir, "port.c"),
			[]byte(native.C(lowered)),
			0644,
		); writeError != nil {
			return writeError
		}
		// Source remains usable after the building test's temporary directories disappear.
		files, globError := filepath.Glob(filepath.Join(filepath.Dir(main), "*.ts"))
		if globError != nil {
			return globError
		}
		for _, file := range files {
			data, readError := os.ReadFile(file)
			if readError != nil {
				return readError
			}
			if writeError := os.WriteFile(filepath.Join(dir, filepath.Base(file)), data, 0644); writeError != nil {
				return writeError
			}
		}
		// mutantPort's imports name this checkout; the product names it <repository> (#tqrqx60).
		return buildcache.RelativeFiles(dir)
	})
	if !compile {
		return resolvedPort(test, lowered, "main.ts"), ""
	}
	inputs.Name = "estree-acceptance-mutant-native-" + item.name
	inputs.Flags = append(inputs.Flags, native.Flags(native.Options{Sanitize: true, Split: true})...)
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	product := buildcache.Product(test, inputs, func(dir string) error {
		data, readError := os.ReadFile(filepath.Join(lowered, "port.c"))
		if readError != nil {
			return readError
		}
		return native.Build(string(data), filepath.Join(dir, "port"), native.Options{Sanitize: true, Split: true})
	})
	return resolvedPort(test, lowered, "main.ts"), filepath.Join(product, "port")
}

// Each independently selected shard builds or fetches only its own mutant.
var acceptancePrepared [testAcceptanceMutantsShards]struct {
	once                 sync.Once
	main, binary, oracle string
}

func acceptanceMutantsOracle(test *testing.T) string {
	test.Helper()
	return oracles.AcceptanceMutantsOracle(test)
}

func acceptanceMutantsPrepareShard(test *testing.T, shard int) (string, string, string) {
	test.Helper()
	prepared := &acceptancePrepared[shard]
	prepared.once.Do(func() {
		prepared.oracle = acceptanceMutantsOracle(test)
		prepared.main, prepared.binary = acceptanceMutantProducts(test, acceptanceMutantEnumeration()[shard], true)
	})
	return prepared.main, prepared.binary, prepared.oracle
}

func TestProduct_EstreeFindsAcceptanceOracle(test *testing.T) {
	test.Parallel()
	acceptanceMutantsOracle(test)
}
func TestProduct_AcceptanceCatchLowered(test *testing.T) {
	test.Parallel()
	acceptanceMutantProducts(test, acceptanceMutantEnumeration()[0], false)
}
func TestProduct_AcceptanceCatchNative(test *testing.T) {
	test.Parallel()
	acceptanceMutantProducts(test, acceptanceMutantEnumeration()[0], true)
}
func TestProduct_AcceptanceClassLowered(test *testing.T) {
	test.Parallel()
	acceptanceMutantProducts(test, acceptanceMutantEnumeration()[1], false)
}
func TestProduct_AcceptanceClassNative(test *testing.T) {
	test.Parallel()
	acceptanceMutantProducts(test, acceptanceMutantEnumeration()[1], true)
}

// acceptanceMutantsCommand's output is read through childguard.Capture, to its end with no timer deciding when: exec's
// own copying under a five-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF five
// seconds after the child exited, so an oracle, Node, native or proof run that had finished could fail its shard under
// load (#n3gk22h's shape). Each caller's 90 s deadline ends a hung run.
func acceptanceMutantsCommand(ctx context.Context, executable string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, executable, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(killError, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return killError
	}
	return command
}

func acceptanceMutantsExecute(test *testing.T, ctx context.Context, name string, args ...string) []byte {
	test.Helper()
	output, stderr, captureError := childguard.Capture(ctx, acceptanceMutantsCommand(ctx, name, args...))
	if captureError != nil || len(stderr) != 0 {
		test.Fatalf("%s %v: %v\n%s", name, args, captureError, stderr)
	}
	return output
}

func runAcceptanceMutantShard(test *testing.T, shard int) {
	test.Helper()
	setup := time.Now()
	main, binary, oracle := acceptanceMutantsPrepareShard(test, shard)
	test.Logf("setup wall=%.3fs", time.Since(setup).Seconds())
	started := time.Now()
	defer func() { test.Logf("own work wall=%.3fs", time.Since(started).Seconds()) }()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	items := acceptanceMutantEnumeration()
	if len(items) != testAcceptanceMutantsShards {
		test.Fatalf("enumerated %d mutations, declared %d shards", len(items), testAcceptanceMutantsShards)
	}
	list := manifest(test, acceptanceGrammar())
	want := acceptanceMutantsExecute(test, ctx, oracle, "--manifest", list)
	for name, got := range map[string][]byte{
		"Node": acceptanceMutantsExecute(
			test,
			ctx,
			"node",
			"--disable-warning=ExperimentalWarning",
			filepath.Join(root(test), "oracle/node.mjs"),
			main,
			"--manifest",
			list,
		),
		"native": acceptanceMutantsExecute(test, ctx, binary, "--manifest", list),
	} {
		if os.Getenv("ADAMIC_ACCEPTANCE_MUTANTS_PROOF") == "1" && shard == 0 && name == "native" {
			// Only the catch-initializer witness output is repaired: every
			// other case retains its actual sanitized native bytes.
			wantCases, gotCases := acceptanceMutantCanonicalCases(test, want), acceptanceMutantCanonicalCases(test, got)
			witness := -1
			for index, source := range acceptanceGrammar() {
				if source == "try {} catch(e=1){}" {
					if witness != -1 {
						test.Fatal("duplicate planted witness")
					}
					witness = index
				}
			}
			if witness < 0 {
				test.Fatal("missing planted witness")
			}
			gotCases[witness] = wantCases[witness]
			got = []byte(strings.Join(gotCases, ""))
		}
		if acceptanceMutantCheckError := acceptanceMutantCheck(want, got); acceptanceMutantCheckError != nil {
			test.Fatal(name + " " + acceptanceMutantCheckError.Error())
		}
		test.Log(name + ": " + firstDifference(want, got))
	}
}

func TestAcceptanceMutants_000(test *testing.T) { test.Parallel(); runAcceptanceMutantShard(test, 0) }
func TestAcceptanceMutants_001(test *testing.T) { test.Parallel(); runAcceptanceMutantShard(test, 1) }

func TestAcceptanceMutantsUnion(test *testing.T) {
	test.Parallel()
	items := acceptanceMutantEnumeration()
	runners := []int{0, 1}
	if len(runners) != testAcceptanceMutantsShards || len(items) != testAcceptanceMutantsShards {
		test.Fatal("shard enumeration mismatch")
	}
	seen := make(map[int]int)
	for _, index := range runners {
		seen[index]++
	}
	for index := range items {
		if seen[index] != 1 {
			test.Fatalf("mutation %d covered %d times", index, seen[index])
		}
	}
	test.Logf(
		"union: %d mutations x %d corpus cases = %d pairs, each exactly once",
		len(items),
		len(acceptanceGrammar()),
		len(items)*len(acceptanceGrammar()),
	)
}

// Each proof runs one real shard in its own process, retaining all corpus bytes.
// Separate proof units avoid aggregating two cold mutant builds into one grain.
// The two assertions together require exactly shard 000 to catch the planted
// surviving native witness, and shard 001 to retain its original passing verdict.
func proveAcceptanceMutantShard(test *testing.T, index int) {
	test.Helper()
	executable, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	name := fmt.Sprintf("TestAcceptanceMutants_%03d", index)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := acceptanceMutantsCommand(ctx, executable, "-test.run=^"+name+"$", "-test.timeout=90s", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_ACCEPTANCE_MUTANTS_PROOF=1")
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	output := append(stdout, stderr...)
	test.Logf("isolated planted proof:\n%s", output)
	if ctx.Err() != nil {
		test.Fatalf("%s exceeded child deadline: %v", name, ctx.Err())
	}
	if index == 0 {
		exit, ok := captureError.(*exec.ExitError)
		if !ok ||
			exit.ExitCode() != 1 ||
			!strings.Contains(string(output), "native mutant survived") ||
			!strings.Contains(string(output), "--- FAIL: "+name) {
			test.Fatalf("planted surviving mutant must fail only %s: %v\n%s", name, captureError, output)
		}
		test.Log("planted native witness caught by " + name)
	} else if captureError != nil {
		test.Fatalf("unplanted shard %s must pass: %v\n%s", name, captureError, output)
	} else {
		test.Log("planted witness has no ownership in " + name)
	}
}

func TestAcceptanceMutantsPlanted_000(test *testing.T) {
	test.Parallel()
	proveAcceptanceMutantShard(test, 0)
}
func TestAcceptanceMutantsPlanted_001(test *testing.T) {
	test.Parallel()
	proveAcceptanceMutantShard(test, 1)
}

// Each canonical dump terminates with a stripped-source line, whose source
// newlines are escaped. Preserve every byte while identifying corpus cases.
func acceptanceMutantCanonicalCases(test *testing.T, data []byte) []string {
	test.Helper()
	var cases []string
	var current strings.Builder
	for _, line := range strings.SplitAfter(string(data), "\n") {
		current.WriteString(line)
		if strings.HasPrefix(line, "stripped ") {
			cases = append(cases, current.String())
			current.Reset()
		}
	}
	if current.Len() != 0 || len(cases) != len(acceptanceGrammar()) {
		test.Fatalf(
			"canonical case count %d, want %d; trailing bytes %d",
			len(cases),
			len(acceptanceGrammar()),
			current.Len(),
		)
	}
	return cases
}
