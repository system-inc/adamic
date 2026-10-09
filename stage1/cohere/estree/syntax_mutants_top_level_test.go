package estree

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Fixed headroom for unpinned repository fixtures. Each top-level test is visible
// to go test -list; ADAMIC_TEST_SHARD=i/n selects indices modulo n, unset runs all.
func TestSyntaxMutants_000(t *testing.T) { t.Parallel(); syntaxMutantsShard(t, 0) }
func TestSyntaxMutants_001(t *testing.T) { t.Parallel(); syntaxMutantsShard(t, 1) }
func TestSyntaxMutants_002(t *testing.T) { t.Parallel(); syntaxMutantsShard(t, 2) }
func TestSyntaxMutants_003(t *testing.T) { t.Parallel(); syntaxMutantsShard(t, 3) }
func TestSyntaxMutants_004(t *testing.T) { t.Parallel(); syntaxMutantsShard(t, 4) }
func TestSyntaxMutants_005(t *testing.T) { t.Parallel(); syntaxMutantsShard(t, 5) }
func TestSyntaxMutants_006(t *testing.T) { t.Parallel(); syntaxMutantsShard(t, 6) }
func TestSyntaxMutants_007(t *testing.T) { t.Parallel(); syntaxMutantsShard(t, 7) }
func TestSyntaxMutants_008(t *testing.T) { t.Parallel(); syntaxMutantsShard(t, 8) }
func TestSyntaxMutants_009(t *testing.T) { t.Parallel(); syntaxMutantsShard(t, 9) }
func TestSyntaxMutants_010(t *testing.T) { t.Parallel(); syntaxMutantsShard(t, 10) }
func TestSyntaxMutants_011(t *testing.T) { t.Parallel(); syntaxMutantsShard(t, 11) }
func TestSyntaxMutants_012(t *testing.T) { t.Parallel(); syntaxMutantsShard(t, 12) }
func TestSyntaxMutants_013(t *testing.T) { t.Parallel(); syntaxMutantsShard(t, 13) }
func TestSyntaxMutants_014(t *testing.T) { t.Parallel(); syntaxMutantsShard(t, 14) }
func TestSyntaxMutants_015(t *testing.T) { t.Parallel(); syntaxMutantsShard(t, 15) }

var syntaxMutantsTopLevel = [...]func(*testing.T){
	TestSyntaxMutants_000,
	TestSyntaxMutants_001,
	TestSyntaxMutants_002,
	TestSyntaxMutants_003,
	TestSyntaxMutants_004,
	TestSyntaxMutants_005,
	TestSyntaxMutants_006,
	TestSyntaxMutants_007,
	TestSyntaxMutants_008,
	TestSyntaxMutants_009,
	TestSyntaxMutants_010,
	TestSyntaxMutants_011,
	TestSyntaxMutants_012,
	TestSyntaxMutants_013,
	TestSyntaxMutants_014,
	TestSyntaxMutants_015,
}

func TestSyntaxMutantsUnion(t *testing.T) {
	if len(syntaxMutantsTopLevel) != testSyntaxMutantsShards {
		t.Fatal("top-level shard enumeration differs from declared count")
	}
	groups, cases := syntaxMutantGroups()
	estreeShardPlan(t, testSyntaxMutantsShards, cases, groups)
}

// Plant a surviving mapped mutant in the actual production top-level functions.
// Exactly its owning top-level leaf must fail; other leaves run their real checks.
func TestSyntaxMutantsTopLevelFailure(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, "-test.run=^TestSyntaxMutants_[0-9]{3}$", "-test.timeout=75s", "-test.v")
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	command.WaitDelay = time.Second
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "ADAMIC_TEST_SHARD=") && !strings.HasPrefix(entry, "ADAMIC_SYNTAX_MUTANTS_PLANT=") {
			command.Env = append(command.Env, entry)
		}
	}
	command.Env = append(command.Env, "ADAMIC_SYNTAX_MUTANTS_PLANT=mapped-constraint")
	output, runErr := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatal("cooked: planted-failure child exceeded 75 seconds")
	}
	frames := regexp.MustCompile(`(?m)^--- FAIL: (TestSyntaxMutants_[0-9]{3}) `).FindAllStringSubmatch(string(output), -1)
	expected := ""
	for _, mutant := range syntaxMutants() {
		if mutant.name == "mapped-constraint" {
			expected = fmt.Sprintf("TestSyntaxMutants_%03d", syntaxMutantShard(mutant))
		}
	}
	if runErr == nil || len(frames) != 1 || frames[0][1] != expected || !strings.Contains(string(output), "mutant survived") {
		t.Fatalf("planted survivor did not fail exactly %s: %v\n%s", expected, runErr, output)
	}
	t.Logf("planted survivor caught by exactly %s", expected)
}
