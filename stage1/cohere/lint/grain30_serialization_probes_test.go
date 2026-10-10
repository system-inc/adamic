package lint

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// The old isolated sweep rebuilt four products and selected leaf 001 in
// one command. Each cold recipe and the isolated selection now has one owner.
var grain30IsolationCases = []struct{ test, product string }{
	{"TestProduct_CompleteSuggestionGoOracle", "complete-suggestion-go-oracle"},
	{"TestProduct_CompleteSuggestionLowered", "complete-suggestion-lowered-false"},
	{"TestProduct_CompleteSuggestionMutantLowered", "complete-suggestion-lowered-true"},
	{"TestProduct_CompleteSuggestionNative", "complete-suggestion-native"},
	{"TestCompleteSuggestionSerialization_001", ""},
}

func grain30Probe(t *testing.T, name string, plant bool) ([]byte, error) {
	t.Helper()
	// This child may prepare products. No test-side deadline wraps that setup;
	// the selected leaf starts its own deadline after preparation, and Loom
	// bounds this whole top-level unit at 90 seconds.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	command := completeSuggestionCommand(ctx, os.Args[0], "-test.run=^"+name+"$", "-test.timeout=0", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_COMPLETE_SUGGESTION_PLANT="+strconv.FormatBool(plant))
	if plant {
		command.Env = append(command.Env, "ADAMIC_COMPLETE_SUGGESTION_PLANT=1")
	}
	output, err := command.CombinedOutput()
	if command.Process != nil {
		_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	return output, err
}

func grain30IsolationResult(index int, output []byte, err error) error {
	row := grain30IsolationCases[index]
	if err != nil || !bytes.Contains(output, []byte("--- PASS: "+row.test+" ")) {
		return fmt.Errorf("isolated %s: %v\n%s", row.test, err, output)
	}
	if row.product != "" && !productLookedUp(output, regexp.QuoteMeta(row.product)) {
		return fmt.Errorf("isolated %s never looked up its product %s itself:\n%s", row.test, row.product, output)
	}
	return nil
}

func grain30Isolation(t *testing.T, index int) {
	t.Helper()
	started := time.Now()
	if index == 4 {
		completeSuggestionReady(t)
	}
	output, err := grain30Probe(t, grain30IsolationCases[index].test, false)
	t.Logf("preparation: %.3fs", time.Since(started).Seconds())
	own := time.Now()
	defer func() { t.Logf("own work: %.3fs", time.Since(own).Seconds()) }()
	if failure := grain30IsolationResult(index, output, err); failure != nil {
		t.Fatal(failure)
	}
	// Plant a failed selection into the real receipt. The same verifier must
	// reject it in every child, including the leaf whose cache is already warm.
	broken := bytes.ReplaceAll(output, []byte("--- PASS: "), []byte("--- FAIL: "))
	if grain30IsolationResult(index, broken, err) == nil {
		t.Fatal("planted selection failure survived")
	}
	if product := grain30IsolationCases[index].product; product != "" {
		broken = bytes.ReplaceAll(output, []byte("build "+product+" "), []byte("built "+product+" "))
		if grain30IsolationResult(index, broken, err) == nil {
			t.Fatal("planted missing product lookup survived")
		}
	}
	t.Logf("case %d exactly once; planted failure caught", index)
}

func grain30PlantResult(shard int, output []byte, err error) error {
	name := fmt.Sprintf("TestCompleteSuggestionSerialization_%03d", shard)
	failures := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "--- FAIL:") {
			failures = append(failures, line)
		}
	}
	if shard == 1 {
		if err == nil || len(failures) != 1 || !strings.HasPrefix(failures[0], "--- FAIL: "+name+" ") || !bytes.Contains(output, []byte("planted suggestion disagreement")) {
			return fmt.Errorf("wrong planted failure: %v; failures %v\n%s", err, failures, output)
		}
	} else if err != nil || len(failures) != 0 || !bytes.Contains(output, []byte("--- PASS: "+name+" ")) {
		return fmt.Errorf("non-owner %s failed: %v\n%s", name, err, output)
	}
	return nil
}

func grain30Plant(t *testing.T, shard int) {
	t.Helper()
	started := time.Now()
	completeSuggestionReady(t)
	t.Logf("preparation: %.3fs", time.Since(started).Seconds())
	own := time.Now()
	defer func() { t.Logf("own work: %.3fs", time.Since(own).Seconds()) }()
	output, err := grain30Probe(t, fmt.Sprintf("TestCompleteSuggestionSerialization_%03d", shard), true)
	if failure := grain30PlantResult(shard, output, err); failure != nil {
		t.Fatal(failure)
	}
	broken := bytes.ReplaceAll(output, []byte("--- PASS: "), []byte("--- FAIL: "))
	if shard == 1 {
		broken = bytes.ReplaceAll(output, []byte("planted suggestion disagreement"), []byte("missing disagreement"))
	}
	if grain30PlantResult(shard, broken, err) == nil {
		t.Fatal("planted probe-verdict failure survived")
	}
	t.Logf("leaf %d exactly once; planted failure caught", shard)
}

func TestCompleteSuggestionSerialization_IsolatedShard_000(t *testing.T) {
	t.Parallel()
	grain30Isolation(t, 0)
}
func TestCompleteSuggestionSerialization_IsolatedShard_001(t *testing.T) {
	t.Parallel()
	grain30Isolation(t, 1)
}
func TestCompleteSuggestionSerialization_IsolatedShard_002(t *testing.T) {
	t.Parallel()
	grain30Isolation(t, 2)
}
func TestCompleteSuggestionSerialization_IsolatedShard_003(t *testing.T) {
	t.Parallel()
	grain30Isolation(t, 3)
}
func TestCompleteSuggestionSerialization_IsolatedShard_004(t *testing.T) {
	t.Parallel()
	grain30Isolation(t, 4)
}
func TestCompleteSuggestionSerialization_PlantedFailure_000(t *testing.T) {
	t.Parallel()
	grain30Plant(t, 0)
}
func TestCompleteSuggestionSerialization_PlantedFailure_001(t *testing.T) {
	t.Parallel()
	grain30Plant(t, 1)
}
func TestCompleteSuggestionSerialization_PlantedFailure_002(t *testing.T) {
	t.Parallel()
	grain30Plant(t, 2)
}
func TestCompleteSuggestionSerialization_PlantedFailure_003(t *testing.T) {
	t.Parallel()
	grain30Plant(t, 3)
}
func TestCompleteSuggestionSerialization_PlantedFailure_004(t *testing.T) {
	t.Parallel()
	grain30Plant(t, 4)
}
func TestCompleteSuggestionSerialization_PlantedFailure_005(t *testing.T) {
	t.Parallel()
	grain30Plant(t, 5)
}

func grain30ProbeUnion(groups [][]int, count int) error {
	seen := make([]int, count)
	for _, group := range groups {
		for _, id := range group {
			if id < 0 || id >= count {
				return fmt.Errorf("unknown case %d", id)
			}
			seen[id]++
		}
	}
	for id, n := range seen {
		if n != 1 {
			return fmt.Errorf("case %d occurs %d times", id, n)
		}
	}
	return nil
}

func TestCompleteSuggestionSerialization_ProbeUnion(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("grain30_serialization_probes_test.go")
	if err != nil {
		t.Fatal(err)
	}
	pattern := regexp.MustCompile(`(?s)func TestCompleteSuggestionSerialization_(IsolatedShard|PlantedFailure)_([0-9]+)\(t \*testing.T\) \{\s*t\.Parallel\(\)\s*(grain30Isolation|grain30Plant)\(t, ([0-9]+)\)\s*\}`)
	matches := pattern.FindAllSubmatch(source, -1)
	if len(matches) != len(grain30IsolationCases)+len(completeSuggestionCases) {
		t.Fatalf("enumerated %d probes", len(matches))
	}
	for _, kind := range []string{"IsolatedShard", "PlantedFailure"} {
		count, helper := len(grain30IsolationCases), "grain30Isolation"
		if kind == "PlantedFailure" {
			count, helper = len(completeSuggestionCases), "grain30Plant"
		}
		groups := make([][]int, count)
		for _, match := range matches {
			if string(match[1]) != kind {
				continue
			}
			owner, e1 := strconv.Atoi(string(match[2]))
			id, e2 := strconv.Atoi(string(match[4]))
			if e1 != nil || e2 != nil || owner < 0 || owner >= count || owner != id || string(match[3]) != helper {
				t.Fatalf("invalid probe %s", match[0])
			}
			groups[owner] = append(groups[owner], id)
		}
		// Empty future groups are legal. Mutation checks must neither index an
		// empty group nor claim that removing nothing caused missing coverage.
		groups = append(groups, nil)
		if err := grain30ProbeUnion(groups, count); err != nil {
			t.Fatal(err)
		}
		for owner, group := range groups {
			for _, fault := range []string{"unknown", "missing", "duplicate"} {
				broken := make([][]int, len(groups))
				for i := range groups {
					broken[i] = append([]int(nil), groups[i]...)
				}
				switch fault {
				case "unknown":
					broken[owner] = append(broken[owner], count)
				case "missing":
					if len(group) == 0 {
						if err := grain30ProbeUnion(broken, count); err != nil {
							t.Fatal(err)
						}
						continue
					}
					broken[owner] = broken[owner][1:]
				case "duplicate":
					id := 0
					if len(group) != 0 {
						id = group[0]
					}
					broken[owner] = append(broken[owner], id)
				}
				if grain30ProbeUnion(broken, count) == nil {
					t.Fatalf("%s group %d: %s survived", kind, owner, fault)
				}
			}
		}
		t.Logf("%s: %d cases exactly once, including empty-group self-check", kind, count)
	}
}
