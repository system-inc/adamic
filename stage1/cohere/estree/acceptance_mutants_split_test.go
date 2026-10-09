package estree

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

func acceptanceMutantProducts(t *testing.T, item acceptanceMutant) (string, string) {
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
	inputs.Name = "estree-acceptance-mutant-native-" + item.name
	inputs.Flags = append(inputs.Flags, native.Flags(native.Options{Sanitize: true})...)
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	product := buildcache.Product(t, inputs, func(dir string) error {
		data, err := os.ReadFile(filepath.Join(lowered, "port.c"))
		if err != nil {
			return err
		}
		return native.Build(string(data), filepath.Join(dir, "port"), native.Options{Sanitize: true})
	})
	return filepath.Join(lowered, "main.ts"), filepath.Join(product, "port")
}

func runAcceptanceMutantShard(t *testing.T, shard int) {
	t.Helper()
	started := time.Now()
	items := acceptanceMutantEnumeration()
	if len(items) != testAcceptanceMutantsShards {
		t.Fatalf("enumerated %d mutations, declared %d shards", len(items), testAcceptanceMutantsShards)
	}
	list := manifest(t, acceptanceGrammar())
	want := execute(t, "", goOracle(t), "--manifest", list)
	main, binary := acceptanceMutantProducts(t, items[shard])
	t.Logf("TestAcceptanceMutants (setup): %.3fs", time.Since(started).Seconds())
	for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list)} {
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
	// Plant one surviving mutant and exercise each real top-level shard separately.
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	caught := []int{}
	for _, index := range runners {
		name := fmt.Sprintf("TestAcceptanceMutants_%03d", index)
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		command := exec.CommandContext(ctx, executable, "-test.run=^"+name+"$", "-test.timeout=90s", "-test.v")
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		command.Cancel = func() error {
			// Kill the entire group so compiler descendants cannot outlive the proof.
			err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
			if errors.Is(err, syscall.ESRCH) {
				return os.ErrProcessDone
			}
			return err
		}
		command.WaitDelay = 5 * time.Second
		command.Env = append(os.Environ(), "ADAMIC_ACCEPTANCE_MUTANTS_PROOF=1")
		output, err := command.CombinedOutput()
		contextErr := ctx.Err()
		cancel()
		if contextErr != nil {
			t.Fatalf("%s exceeded child deadline: %v\n%s", name, contextErr, output)
		}
		if err != nil {
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "mutant survived") || !strings.Contains(string(output), "--- FAIL: "+name) {
				t.Fatalf("unexpected planted failure: %v\n%s", err, output)
			}
			caught = append(caught, index)
		}
	}
	if len(caught) != 1 || caught[0] != 0 {
		t.Fatalf("planted failure caught by %v", caught)
	}
	t.Logf("union: %d mutations x %d corpus cases = %d pairs, each exactly once; planted surviving mutant caught only by TestAcceptanceMutants_000", len(items), len(acceptanceGrammar()), len(items)*len(acceptanceGrammar()))
}

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
