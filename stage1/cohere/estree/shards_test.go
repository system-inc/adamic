package estree

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Verify the complete plan before applying the distributed selection, including
// on instances which run only one partition. IDs refer to the unsplit enumeration.
func estreeShardPlan(t *testing.T, count, cases int, groups [][]int) []bool {
	t.Helper()
	if len(groups) != count {
		t.Fatalf("enumerated %d shards, declared %d", len(groups), count)
	}
	seen := make(map[int]bool)
	total := 0
	for shard, group := range groups {
		if len(group) == 0 {
			t.Fatalf("shard-%03d is empty", shard)
		}
		for _, id := range group {
			if id < 0 || id >= cases || seen[id] {
				t.Fatalf("invalid or repeated case %d in shard-%03d", id, shard)
			}
			seen[id] = true
			total++
		}
	}
	if total != cases || len(seen) != cases {
		t.Fatalf("union has %d cases and %d IDs; want %d", total, len(seen), cases)
	}
	for id := 0; id < cases; id++ {
		if !seen[id] {
			t.Fatalf("missing case %d", id)
		}
	}
	selected := make([]bool, count)
	i, n := 0, 1
	if value := os.Getenv("ADAMIC_TEST_SHARD"); value != "" {
		parts := strings.Split(value, "/")
		if len(parts) != 2 {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
		}
		var err error
		i, err = strconv.Atoi(parts[0])
		if err != nil {
			t.Fatal(err)
		}
		n, err = strconv.Atoi(parts[1])
		if err != nil || n <= 0 || i < 0 || i >= n {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
		}
	}
	for shard := range selected {
		selected[shard] = shard%n == i
	}
	t.Logf("union verified: %d unique case IDs in %d shards", total, count)
	return selected
}

func estreeTimedOracle(t *testing.T) string {
	t.Helper()
	start := time.Now()
	cpu := estreeCPU()
	result := goOracle(t)
	t.Logf("build Go oracle: %.3fs CPU %.3fs", time.Since(start).Seconds(), estreeCPU()-cpu)
	return result
}
func estreeTimedBuild(t *testing.T, path string, sanitize bool) (string, string) {
	t.Helper()
	start := time.Now()
	cpu := estreeCPU()
	binary, script := build(t, path, sanitize)
	t.Logf("build lowered/native/emitted: %.3fs CPU %.3fs", time.Since(start).Seconds(), estreeCPU()-cpu)
	return binary, script
}
func estreeMutantVerdict(want, got []byte) error {
	if firstDifference(want, got) == "" {
		return fmt.Errorf("mutant survived")
	}
	return nil
}

// Plant an oracle-identical (surviving) mutant at case 4 in the exact production
// plan. Run the ordinary t.Fatal verdict in a child so the proof itself passes.
func TestSyntaxMutantsShardFailure(t *testing.T) {
	if os.Getenv("ADAMIC_ESTREE_SHARD_PROOF") == "syntax-mutants" {
		groups, cases := syntaxMutantGroups()
		selected := estreeShardPlan(t, testSyntaxMutantsShards, cases, groups)
		for i, group := range groups {
			if !selected[i] {
				continue
			}
			t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
				t.Parallel()
				for _, id := range group {
					want, got := []byte("Go"), []byte("mutant")
					if id == 4 {
						got = want
					}
					if err := estreeMutantVerdict(want, got); err != nil {
						t.Fatalf("case %d: %v", id, err)
					}
				}
			})
		}
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "-test.run=^TestSyntaxMutantsShardFailure$", "-test.v")
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "ADAMIC_TEST_SHARD=") && !strings.HasPrefix(entry, "ADAMIC_ESTREE_SHARD_PROOF=") {
			command.Env = append(command.Env, entry)
		}
	}
	command.Env = append(command.Env, "ADAMIC_ESTREE_SHARD_PROOF=syntax-mutants")
	output, err := command.CombinedOutput()
	text := string(output)
	if err == nil || strings.Count(text, "--- FAIL: TestSyntaxMutantsShardFailure/shard-") != 1 || !strings.Contains(text, "--- FAIL: TestSyntaxMutantsShardFailure/shard-000") || !strings.Contains(text, "case 4: mutant survived") {
		t.Fatalf("planted survivor was not caught by exactly shard-000: %v\n%s", err, text)
	}
	t.Log("case 4 planted survivor caught by exactly shard-000")
}

func estreeCPU() float64 {
	var self, children syscall.Rusage
	_ = syscall.Getrusage(syscall.RUSAGE_SELF, &self)
	_ = syscall.Getrusage(syscall.RUSAGE_CHILDREN, &children)
	return float64(self.Utime.Sec+self.Stime.Sec+children.Utime.Sec+children.Stime.Sec) + float64(self.Utime.Usec+self.Stime.Usec+children.Utime.Usec+children.Stime.Usec)/1e6
}
func estreeAccounting(t *testing.T) {
	t.Helper()
	before := estreeCPU()
	t.Cleanup(func() { t.Logf("total test CPU: %.3fs", estreeCPU()-before) })
}
