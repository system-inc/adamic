package estree

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Each top-level shard prepares its own products before its work deadline.
func TestRecoveryMutants(t *testing.T) {
	t.Parallel()
	// Compatibility enumeration only; execution lives in the top-level shards.
	checkRecoveryMutantUnion(t)
}

func TestRecoveryMutantsShardSurvivor(t *testing.T) {
	t.Parallel()
	proveRecoveryShard(t, recoveryMutantCases(), testRecoveryMutantsShards, true)
}

// ADAMIC_TEST_SHARD=i/n selects top-level shards whose index modulo n is i.
// Each mutant retains the full live grammar on Node and sanitized native.
func TestRecoveryMutants_000(t *testing.T) { t.Parallel(); runRecoveryMutantTop(t, 0) }

func TestRecoveryMutants_001(t *testing.T) { t.Parallel(); runRecoveryMutantTop(t, 1) }

func TestRecoveryMutants_002(t *testing.T) { t.Parallel(); runRecoveryMutantTop(t, 2) }

func TestRecoveryMutantsUnion(t *testing.T) { t.Parallel(); checkRecoveryMutantUnion(t) }

func TestRecoveryMutantsTopSurvivor(t *testing.T) {
	t.Parallel()
	for _, mutation := range recoveryMutations {
		recoveryMutantPrepared(t, mutation.name)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	planted := recoveryMutantLiveSlices(t, recoveredGrammar())[recoveryMutantOwner(recoveryMutations[0].name)][0].id
	var output []byte
	err = nil
	for shard := 0; shard < testRecoveryMutantsShards; shard++ {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, executable, "-test.v", "-test.timeout=90s", fmt.Sprintf("-test.run=^TestRecoveryMutants_%03d$", shard))
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		command.Cancel = func() error {
			err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
			if err == syscall.ESRCH {
				return os.ErrProcessDone
			}
			return err
		}
		command.WaitDelay = time.Second
		for _, entry := range os.Environ() {
			if !strings.HasPrefix(entry, "ADAMIC_TEST_SHARD=") && !strings.HasPrefix(entry, "ADAMIC_RECOVERY_MUTANT_SURVIVOR=") {
				command.Env = append(command.Env, entry)
			}
		}
		command.Env = append(command.Env, "ADAMIC_RECOVERY_MUTANT_SURVIVOR="+planted)
		part, childErr := command.CombinedOutput()
		output = append(output, part...)
		if childErr != nil {
			err = childErr
		}
		if ctx.Err() != nil {
			t.Fatalf("planted survivor subprocess cooked: %v", ctx.Err())
		}

	}
	owner := fmt.Sprintf("TestRecoveryMutants_%03d", recoveryMutantOwner(strings.SplitN(planted, "/", 2)[0]))
	prefix := "--- FAIL: TestRecoveryMutants_"
	if err == nil || strings.Count(string(output), prefix) != 1 || !strings.Contains(string(output), "--- FAIL: "+owner+" ") || !strings.Contains(string(output), "mutant survived") {
		t.Fatalf("planted survivor must fail only %s: %v\n%s", owner, err, output)
	}
	t.Logf("planted surviving mutant %s caught only by %s", planted, owner)
}

// Optional prewarming uses exactly the preparation path used by each shard.
func TestRecoveryMutants_Setup(t *testing.T) {
	t.Parallel()
	for _, mutation := range recoveryMutations {
		recoveryMutantPrepared(t, mutation.name)
	}
}
