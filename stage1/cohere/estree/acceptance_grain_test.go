package estree

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const testAcceptanceMutantsShards = 2

type acceptanceMutant struct{ name, file, from, to string }

func acceptanceMutantEnumeration() []acceptanceMutant {
	return []acceptanceMutant{
		{"catch-initializer", "convert.ts", "this.separated(this.child(declaration, 0), this.child(declaration, 1), 'ColonToken')", "true"},
		{"class-keyword-name", "sourceStatements.ts", "!(this.parser.peek() === 'Identifier' || this.parser.peek().endsWith('Keyword'))", "false"},
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

func acceptanceMutantProducts(t *testing.T, item acceptanceMutant, compile bool) (string, string) {
	t.Helper()
	inputs := buildcache.Inputs{
		Name:  "estree-acceptance-mutant-lowered-" + item.name,
		Files: []string{"stage1/cohere/estree", "stage1/typescript", "internal", "cohere", "go.mod"},
		Flags: []string{item.file, item.from, item.to, "repository=" + root(t), "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS=" + os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED=" + os.Getenv("ADAMIC_GATE_UNCACHED")}, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}
	lowered := buildcache.Product(t, inputs, func(dir string) error {
		main := mutantPort(t, item.file, item.from, item.to)
		program, err := load.Load([]string{main})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "port.c"), []byte(native.C(lowered)), 0644); err != nil {
			return err
		}
		// Source remains usable after the building test's temporary directories disappear.
		files, err := filepath.Glob(filepath.Join(filepath.Dir(main), "*.ts"))
		if err != nil {
			return err
		}
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(dir, filepath.Base(file)), data, 0644); err != nil {
				return err
			}
		}
		return nil
	})
	if !compile {
		return filepath.Join(lowered, "main.ts"), ""
	}
	inputs.Name = "estree-acceptance-mutant-native-" + item.name
	inputs.Flags = append(inputs.Flags, native.Flags(native.Options{Sanitize: true, Split: true})...)
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	product := buildcache.Product(t, inputs, func(dir string) error {
		data, err := os.ReadFile(filepath.Join(lowered, "port.c"))
		if err != nil {
			return err
		}
		return native.Build(string(data), filepath.Join(dir, "port"), native.Options{Sanitize: true, Split: true})
	})
	return filepath.Join(lowered, "main.ts"), filepath.Join(product, "port")
}

// Each independently selected shard builds or fetches only its own mutant.
var acceptancePrepared [testAcceptanceMutantsShards]struct {
	once                 sync.Once
	main, binary, oracle string
}
var acceptanceOracleOnce sync.Once
var acceptanceOraclePath string

func acceptanceMutantsOracle(t *testing.T) string {
	t.Helper()
	acceptanceOracleOnce.Do(func() {
		inputs := buildcache.Inputs{Name: "estree-acceptance-mutants-go-oracle",
			Files:     []string{"stage1/cohere/estree", "cohere", "go.mod", "go.work"},
			Flags:     []string{"repository=" + root(t), "GOTOOLCHAIN=" + os.Getenv("GOTOOLCHAIN"), "GOFLAGS=" + os.Getenv("GOFLAGS"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED")},
			Toolchain: []string{runtime.Version(), buildcache.Tool("go", "version")}}
		directory := buildcache.Product(t, inputs, func(directory string) error {
			// GoBuild is absent on this base; preserve the existing overlay recipe.
			binary := goOracle(t)
			data, err := os.ReadFile(binary)
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(directory, "oracle"), data, 0755)
		})
		acceptanceOraclePath = filepath.Join(directory, "oracle")
	})
	return acceptanceOraclePath
}

func acceptanceMutantsPrepareShard(t *testing.T, shard int) (string, string, string) {
	t.Helper()
	prepared := &acceptancePrepared[shard]
	prepared.once.Do(func() {
		prepared.oracle = acceptanceMutantsOracle(t)
		prepared.main, prepared.binary = acceptanceMutantProducts(t, acceptanceMutantEnumeration()[shard], true)
	})
	return prepared.main, prepared.binary, prepared.oracle
}

func TestProduct_AcceptanceOracle(t *testing.T) { t.Parallel(); acceptanceMutantsOracle(t) }
func TestProduct_AcceptanceCatchLowered(t *testing.T) {
	t.Parallel()
	acceptanceMutantProducts(t, acceptanceMutantEnumeration()[0], false)
}
func TestProduct_AcceptanceCatchNative(t *testing.T) {
	t.Parallel()
	acceptanceMutantProducts(t, acceptanceMutantEnumeration()[0], true)
}
func TestProduct_AcceptanceClassLowered(t *testing.T) {
	t.Parallel()
	acceptanceMutantProducts(t, acceptanceMutantEnumeration()[1], false)
}
func TestProduct_AcceptanceClassNative(t *testing.T) {
	t.Parallel()
	acceptanceMutantProducts(t, acceptanceMutantEnumeration()[1], true)
}

func acceptanceMutantsCommand(ctx context.Context, executable string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, executable, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = 5 * time.Second
	return command
}

func acceptanceMutantsExecute(t *testing.T, ctx context.Context, name string, args ...string) []byte {
	t.Helper()
	command := acceptanceMutantsCommand(ctx, name, args...)
	var output, stderr bytes.Buffer
	command.Stdout, command.Stderr = &output, &stderr
	if err := command.Run(); err != nil || stderr.Len() != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	return output.Bytes()
}

func runAcceptanceMutantShard(t *testing.T, shard int) {
	t.Helper()
	setup := time.Now()
	main, binary, oracle := acceptanceMutantsPrepareShard(t, shard)
	t.Logf("setup wall=%.3fs", time.Since(setup).Seconds())
	started := time.Now()
	defer func() { t.Logf("own work wall=%.3fs", time.Since(started).Seconds()) }()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	items := acceptanceMutantEnumeration()
	if len(items) != testAcceptanceMutantsShards {
		t.Fatalf("enumerated %d mutations, declared %d shards", len(items), testAcceptanceMutantsShards)
	}
	list := manifest(t, acceptanceGrammar())
	want := acceptanceMutantsExecute(t, ctx, oracle, "--manifest", list)
	for name, got := range map[string][]byte{
		"Node":   acceptanceMutantsExecute(t, ctx, "node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, "--manifest", list),
		"native": acceptanceMutantsExecute(t, ctx, binary, "--manifest", list),
	} {
		if os.Getenv("ADAMIC_ACCEPTANCE_MUTANTS_PROOF") == "1" && shard == 0 && name == "native" {
			// Only the catch-initializer witness output is repaired: every
			// other case retains its actual sanitized native bytes.
			wantCases, gotCases := acceptanceMutantCanonicalCases(t, want), acceptanceMutantCanonicalCases(t, got)
			witness := -1
			for index, source := range acceptanceGrammar() {
				if source == "try {} catch(e=1){}" {
					if witness != -1 {
						t.Fatal("duplicate planted witness")
					}
					witness = index
				}
			}
			if witness < 0 {
				t.Fatal("missing planted witness")
			}
			gotCases[witness] = wantCases[witness]
			got = []byte(strings.Join(gotCases, ""))
		}
		if err := acceptanceMutantCheck(want, got); err != nil {
			t.Fatal(name + " " + err.Error())
		}
		t.Log(name + ": " + firstDifference(want, got))
	}
}

func TestAcceptanceMutants_000(t *testing.T) { t.Parallel(); runAcceptanceMutantShard(t, 0) }
func TestAcceptanceMutants_001(t *testing.T) { t.Parallel(); runAcceptanceMutantShard(t, 1) }

func TestAcceptanceMutantsUnion(t *testing.T) {
	t.Parallel()
	items := acceptanceMutantEnumeration()
	runners := []int{0, 1}
	if len(runners) != testAcceptanceMutantsShards || len(items) != testAcceptanceMutantsShards {
		t.Fatal("shard enumeration mismatch")
	}
	seen := make(map[int]int)
	for _, index := range runners {
		seen[index]++
	}
	for index := range items {
		if seen[index] != 1 {
			t.Fatalf("mutation %d covered %d times", index, seen[index])
		}
	}
	t.Logf("union: %d mutations x %d corpus cases = %d pairs, each exactly once", len(items), len(acceptanceGrammar()), len(items)*len(acceptanceGrammar()))
}

// Each proof runs one real shard in its own process, retaining all corpus bytes.
// Separate proof units avoid aggregating two cold mutant builds into one grain.
// The two assertions together require exactly shard 000 to catch the planted
// surviving native witness, and shard 001 to retain its original passing verdict.
func proveAcceptanceMutantShard(t *testing.T, index int) {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("TestAcceptanceMutants_%03d", index)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := acceptanceMutantsCommand(ctx, executable, "-test.run=^"+name+"$", "-test.timeout=90s", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_ACCEPTANCE_MUTANTS_PROOF=1")
	output, err := command.CombinedOutput()
	t.Logf("isolated planted proof:\n%s", output)
	if ctx.Err() != nil {
		t.Fatalf("%s exceeded child deadline: %v", name, ctx.Err())
	}
	if index == 0 {
		exit, ok := err.(*exec.ExitError)
		if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "native mutant survived") || !strings.Contains(string(output), "--- FAIL: "+name) {
			t.Fatalf("planted surviving mutant must fail only %s: %v\n%s", name, err, output)
		}
		t.Log("planted native witness caught by " + name)
	} else if err != nil {
		t.Fatalf("unplanted shard %s must pass: %v\n%s", name, err, output)
	} else {
		t.Log("planted witness has no ownership in " + name)
	}
}

func TestAcceptanceMutantsPlanted_000(t *testing.T) { t.Parallel(); proveAcceptanceMutantShard(t, 0) }
func TestAcceptanceMutantsPlanted_001(t *testing.T) { t.Parallel(); proveAcceptanceMutantShard(t, 1) }

// Each canonical dump terminates with a stripped-source line, whose source
// newlines are escaped. Preserve every byte while identifying corpus cases.
func acceptanceMutantCanonicalCases(t *testing.T, data []byte) []string {
	t.Helper()
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
		t.Fatalf("canonical case count %d, want %d; trailing bytes %d", len(cases), len(acceptanceGrammar()), current.Len())
	}
	return cases
}
