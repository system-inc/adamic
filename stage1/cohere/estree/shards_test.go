package estree

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	result := estreeOracleProduct(t)
	t.Logf("build Go oracle: %.3fs CPU %.3fs", time.Since(start).Seconds(), estreeCPU()-cpu)
	return result
}
func estreeTimedBuild(t *testing.T, path string, sanitize bool) (string, string) {
	t.Helper()
	start := time.Now()
	cpu := estreeCPU()
	binary, script := estreePortProduct(t, path, sanitize)
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
	groups, cases := syntaxMutantGroups()
	estreeShardFailure(t, testSyntaxMutantsShards, cases, groups, "mutant", 4)
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

func estreeSingles(cases int) [][]int {
	groups := make([][]int, cases)
	for id := range groups {
		groups[id] = []int{id}
	}
	return groups
}
func estreeAgreementVerdict(want, got []byte) error {
	if diff := firstDifference(want, got); diff != "" {
		return fmt.Errorf("%s", diff)
	}
	return nil
}
func estreeRefusalVerdict(err error, timeout bool, size int64, stderr, diagnostic string) error {
	if timeout || err == nil || size != 0 || !strings.Contains(stderr, diagnostic) {
		return fmt.Errorf("timeout=%v exit=%v stdout=%d stderr=%s", timeout, err, size, stderr)
	}
	return nil
}
func estreeRefused(t *testing.T, argv []string, diagnostic string) {
	t.Helper()
	output, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	runErr, timeout := runWithCPUBudget(t, argv, output, &stderr, 2*time.Second)
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	if err := estreeRefusalVerdict(runErr, timeout, info.Size(), stderr.String(), diagnostic); err != nil {
		t.Fatalf("%v: %v", argv, err)
	}
}

// Use the production plan and verdict in a child which really fails. Exactly
// one shard must report the planted wrong answer, acceptance, or survivor.
func estreeShardFailure(t *testing.T, count, cases int, groups [][]int, mode string, planted int) {
	t.Helper()
	if os.Getenv("ADAMIC_ESTREE_SHARD_PROOF") == t.Name() {
		selected := estreeShardPlan(t, count, cases, groups)
		for i, group := range groups {
			if !selected[i] {
				continue
			}
			t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
				t.Parallel()
				if mode == "mutant" {
					// Production compares the whole manifest for a mutant,
					// rather than requiring every control input to differ.
					want := []byte(fmt.Sprintf("Go answers for %v", group))
					got := []byte("mutant answers")
					for _, id := range group {
						if id == planted {
							got = want
						}
					}
					if err := estreeMutantVerdict(want, got); err != nil {
						t.Fatalf("case %d: %v", planted, err)
					}
					return
				}
				for _, id := range group {
					var err error
					switch mode {
					case "refusal":
						exit := fmt.Errorf("refused")
						if id == planted {
							exit = nil
						}
						err = estreeRefusalVerdict(exit, false, 0, "expected diagnostic", "expected diagnostic")
					case "agreement":
						got := []byte("Go")
						if id == planted {
							got = []byte("planted disagreement")
						}
						err = estreeAgreementVerdict([]byte("Go"), got)
					default:
						t.Fatal("unknown proof mode")
					}
					if err != nil {
						t.Fatalf("case %d: %v", id, err)
					}
				}
			})
		}
		return
	}
	owner := -1
	for i, group := range groups {
		for _, id := range group {
			if id == planted {
				owner = i
			}
		}
	}
	if owner < 0 {
		t.Fatal("planted ID absent")
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "-test.run=^"+regexp.QuoteMeta(t.Name())+"$", "-test.v")
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "ADAMIC_TEST_SHARD=") && !strings.HasPrefix(entry, "ADAMIC_ESTREE_SHARD_PROOF=") {
			command.Env = append(command.Env, entry)
		}
	}
	command.Env = append(command.Env, "ADAMIC_ESTREE_SHARD_PROOF="+t.Name())
	output, err := command.CombinedOutput()
	text := string(output)
	prefix := "--- FAIL: " + t.Name() + "/shard-"
	expected := fmt.Sprintf("%s%03d", prefix, owner)
	if err == nil || strings.Count(text, prefix) != 1 || !strings.Contains(text, expected) || !strings.Contains(text, fmt.Sprintf("case %d:", planted)) {
		t.Fatalf("planted %s did not fail exactly shard-%03d: %v\n%s", mode, owner, err, text)
	}
	t.Logf("case %d planted %s caught by exactly shard-%03d", planted, mode, owner)
}

func estreeAgreementShards(t *testing.T, count int, sources []string) {
	t.Helper()
	estreeAccounting(t)
	started := time.Now()
	selected := estreeShardPlan(t, count, len(sources), estreeSingles(len(sources)))
	oracle := estreeTimedOracle(t)
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	binary, script := estreeTimedBuild(t, main, true)
	t.Logf("setup including builds: %.3fs", time.Since(started).Seconds())
	for i, source := range sources {
		if !selected[i] {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			list := manifest(t, []string{source})
			want := estreeOracleOutput(t, oracle, "--manifest", list)
			for name, got := range map[string][]byte{"source Node": onNode(t, main, "--manifest", list), "sanitized native": execute(t, "", binary, "--manifest", list), "emitted JS": onNode(t, script, "--manifest", list)} {
				if err := estreeAgreementVerdict(want, got); err != nil {
					t.Fatalf("case %d %s: %v", i, name, err)
				}
			}
		})
	}
}
