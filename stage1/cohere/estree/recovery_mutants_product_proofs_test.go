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

	"github.com/system-inc/adamic/internal/childguard"
)

// Each top-level shard prepares its own products before its work deadline.
func TestRecoveryMutants(test *testing.T) {
	test.Parallel()
	// Compatibility enumeration only; execution lives in the top-level shards.
	checkRecoveryMutantUnion(test)
}

func TestRecoveryMutantsShardSurvivor(test *testing.T) {
	test.Parallel()
	proveRecoveryShard(test, recoveryMutantCases(), testRecoveryMutantsShards, true)
}

// ADAMIC_TEST_SHARD=i/n selects top-level shards whose index modulo n is i.
// Each mutant retains the full live grammar on Node and sanitized native.
func TestRecoveryMutants_000(test *testing.T) { test.Parallel(); runRecoveryMutantTop(test, 0) }

func TestRecoveryMutants_001(test *testing.T) { test.Parallel(); runRecoveryMutantTop(test, 1) }

func TestRecoveryMutants_002(test *testing.T) { test.Parallel(); runRecoveryMutantTop(test, 2) }

func TestRecoveryMutantsUnion(test *testing.T) { test.Parallel(); checkRecoveryMutantUnion(test) }

func TestRecoveryMutantsTopSurvivor(test *testing.T) {
	test.Parallel()
	// The planted verdict path uses synthetic bytes; it requires no build products.
	executable, err := os.Executable()
	if err != nil {
		test.Fatal(err)
	}
	planted := recoveryMutantLiveSlices(test, recoveredGrammar())[recoveryMutantOwner(recoveryMutations[0].name)][0].id
	var output []byte
	err = nil
	for shard := 0; shard < testRecoveryMutantsShards; shard++ {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		command := exec.CommandContext(
			ctx,
			executable,
			"-test.v",
			"-test.timeout=90s",
			fmt.Sprintf("-test.run=^TestRecoveryMutants_%03d$", shard),
		)
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		command.Cancel = func() error {
			killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
			if killError == syscall.ESRCH {
				return os.ErrProcessDone
			}
			return killError
		}
		for _, entry := range os.Environ() {
			if !strings.HasPrefix(entry, "ADAMIC_TEST_SHARD=") &&
				!strings.HasPrefix(entry, "ADAMIC_RECOVERY_MUTANT_SURVIVOR=") {
				command.Env = append(command.Env, entry)
			}
		}
		command.Env = append(command.Env, "ADAMIC_RECOVERY_MUTANT_SURVIVOR="+planted)
		// Read through childguard.Capture, to its end with no timer deciding when: exec's copying under a one-second
		// WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF a second after the shard exited,
		// which counted as a failing shard here and cut the output the verdict reads (#n3gk22h's shape).
		stdout, stderr, childError := childguard.Capture(ctx, command)
		output = append(append(output, stdout...), stderr...)
		if childError != nil {
			err = childError
		}
		if ctx.Err() != nil {
			test.Fatalf("planted survivor subprocess cooked: %v", ctx.Err())
		}

	}
	owner := fmt.Sprintf("TestRecoveryMutants_%03d", recoveryMutantOwner(strings.SplitN(planted, "/", 2)[0]))
	prefix := "--- FAIL: TestRecoveryMutants_"
	if err == nil ||
		strings.Count(string(output), prefix) != 1 ||
		!strings.Contains(string(output), "--- FAIL: "+owner+" ") ||
		!strings.Contains(string(output), "mutant survived") {
		test.Fatalf("planted survivor must fail only %s: %v\n%s", owner, err, output)
	}
	test.Logf("planted surviving mutant %s caught only by %s", planted, owner)
}

// Optional prewarming uses exactly the preparation path used by each shard.
func TestRecoveryMutants_Setup(test *testing.T) {
	test.Parallel()
	for _, mutation := range recoveryMutations {
		recoveryMutantPrepared(test, mutation.name)
	}
}
