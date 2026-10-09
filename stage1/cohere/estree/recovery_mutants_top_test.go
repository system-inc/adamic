package estree

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

// ADAMIC_TEST_SHARD=i/n selects top-level shards whose index modulo n is i.
// The fixed shard count has headroom: growing the live grammar does not move a
// mutant. Each mutant checks the complete grammar on Node and sanitized native.
func TestRecoveryMutants_000(t *testing.T) { t.Parallel(); runRecoveryMutantTop(t, 0) }
func TestRecoveryMutants_001(t *testing.T) { t.Parallel(); runRecoveryMutantTop(t, 1) }
func TestRecoveryMutants_002(t *testing.T) { t.Parallel(); runRecoveryMutantTop(t, 2) }

func recoveryMutantOwner(name string) int {
	key := "stage1/cohere/estree/recovery_test.go/TestRecoveryMutants/" + name
	sum := sha256.Sum256([]byte(key))
	return int(binary.BigEndian.Uint64(sum[:8]) % uint64(testRecoveryMutantsShards))
}
func recoveryMutantLiveSlices(t *testing.T, sources []string) [][]recoveryCase {
	t.Helper()
	if len(sources) == 0 || len(recoveryMutations) == 0 {
		t.Fatal("empty recovery mutant corpus")
	}
	shards := make([][]recoveryCase, testRecoveryMutantsShards)
	seen := map[string]bool{}
	total := 0
	for _, m := range recoveryMutations {
		occurrences := map[[32]byte]int{}
		for _, source := range sources {
			sum := sha256.Sum256([]byte(source))
			id := fmt.Sprintf("%s/%x/%d", m.name, sum, occurrences[sum])
			occurrences[sum]++
			if seen[id] {
				t.Fatalf("duplicate case %s", id)
			}
			seen[id] = true
			owner := recoveryMutantOwner(m.name)
			shards[owner] = append(shards[owner], recoveryCase{id: id, source: source})
			total++
		}
	}
	if total != len(sources)*len(recoveryMutations) {
		t.Fatal("live enumeration mismatch")
	}
	return shards
}
func checkRecoveryMutantUnion(t *testing.T) {
	t.Helper()
	declared := [...]func(*testing.T){TestRecoveryMutants_000, TestRecoveryMutants_001, TestRecoveryMutants_002}
	if len(declared) != testRecoveryMutantsShards {
		t.Fatal("top-level shard enumeration changed")
	}
	sources := recoveredGrammar()
	shards := recoveryMutantLiveSlices(t, sources)
	if len(shards) != testRecoveryMutantsShards {
		t.Fatal("shard count mismatch")
	}
	// Independent live enumeration: no fixed corpus total or positional keys.
	expected := map[string]bool{}
	for _, m := range recoveryMutations {
		occurrences := map[[32]byte]int{}
		for _, source := range sources {
			sum := sha256.Sum256([]byte(source))
			expected[fmt.Sprintf("%s/%x/%d", m.name, sum, occurrences[sum])] = true
			occurrences[sum]++
		}
	}
	seen := map[string]bool{}
	for i, cases := range shards {
		for _, c := range cases {
			name := strings.SplitN(c.id, "/", 2)[0]
			if !expected[c.id] || seen[c.id] || recoveryMutantOwner(name) != i {
				t.Fatalf("bad union case %s in shard %03d", c.id, i)
			}
			seen[c.id] = true
		}
	}
	if len(seen) != len(expected) {
		t.Fatalf("union %d != live enumeration %d", len(seen), len(expected))
	}
	t.Logf("union: %d cases, %d unique IDs, %d shards", len(seen), len(expected), len(shards))
	// Adding a live source must preserve the assignment of every old case.
	grown := recoveryMutantLiveSlices(t, append(append([]string{}, sources...), "new corpus growth sentinel"))
	for i, cases := range shards {
		ids := map[string]bool{}
		for _, c := range grown[i] {
			ids[c.id] = true
		}
		for _, c := range cases {
			if !ids[c.id] {
				t.Fatalf("growth moved %s", c.id)
			}
		}
	}
}
func TestRecoveryMutantsUnion(t *testing.T) { checkRecoveryMutantUnion(t) }

func runRecoveryMutantTop(t *testing.T, shard int) {
	selected, n := recoveryShardSelection(t)
	if shard%n != selected {
		t.Skip("ADAMIC_TEST_SHARD selects another shard")
	}
	sources := recoveredGrammar()
	cases := recoveryMutantLiveSlices(t, sources)[shard]
	planted := os.Getenv("ADAMIC_RECOVERY_MUTANT_SURVIVOR")
	if planted != "" {
		for _, m := range recoveryMutations {
			if recoveryMutantOwner(m.name) == shard {
				got := []byte("disagreement")
				if m.name == planted {
					got = []byte("oracle")
				}
				if failure := recoveryComparison([]byte("oracle"), got, true); failure != "" {
					t.Fatal(m.name + ": " + failure)
				}
			}
		}
		return
	}
	setup := beginRecoverySetup(t)
	defer setup.report(t)
	if len(cases) == 0 {
		return
	}
	list := manifest(t, sources)
	want := recoveryAnswer(t, recoveryOracle(t, setup), list, "--manifest")
	for _, m := range recoveryMutations {
		if recoveryMutantOwner(m.name) != shard {
			continue
		}
		count := 0
		for _, c := range cases {
			if strings.HasPrefix(c.id, m.name+"/") {
				count++
			}
		}
		if count != len(sources) {
			t.Fatalf("%s lost grammar cases: %d/%d", m.name, count, len(sources))
		}
		main := mutantPort(t, m.file, m.from, m.to)
		binary, _ := recoveryPort(t, setup, main)
		for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list)} {
			if failure := recoveryComparison(want, got, true); failure != "" {
				t.Fatal(m.name + " " + name + ": " + failure)
			}
			t.Log(m.name + " " + name + ": " + firstDifference(want, got))
		}
	}
}
func TestRecoveryMutantsTopSurvivor(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	planted := recoveryMutations[0].name
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, "-test.v", "-test.timeout=75s", "-test.run=^TestRecoveryMutants_[0-9]{3}$")
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
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("planted survivor subprocess cooked: %v", ctx.Err())
	}
	owner := fmt.Sprintf("TestRecoveryMutants_%03d", recoveryMutantOwner(planted))
	prefix := "--- FAIL: TestRecoveryMutants_"
	if err == nil || strings.Count(string(output), prefix) != 1 || !strings.Contains(string(output), "--- FAIL: "+owner+" ") || !strings.Contains(string(output), "mutant survived") {
		t.Fatalf("planted survivor must fail only %s: %v\n%s", owner, err, output)
	}
	t.Logf("planted surviving mutant %s caught only by %s", planted, owner)
}
