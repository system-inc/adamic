package lint

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

// The old isolated sweep rebuilt four products and selected leaf 001 in
// one command. Each cold recipe and the isolated selection now has one owner.
var grain30IsolationCases = []struct{ test, product string }{
	{"TestProduct_LintFindsCompleteSuggestionGoOracle", "complete-suggestion-go-oracle_oracle-owner-v1"},
	{"TestProduct_CompleteSuggestionLowered", "complete-suggestion-lowered-false"},
	{"TestProduct_CompleteSuggestionMutantLowered", "complete-suggestion-lowered-true"},
	{"TestProduct_CompleteSuggestionNative", "complete-suggestion-native"},
	{"TestCompleteSuggestionSerialization_001", ""},
}

func grain30Probe(test *testing.T, name string, plant bool) ([]byte, error) {
	test.Helper()
	// This child may prepare products. No test-side deadline wraps that setup;
	// the selected leaf starts its own deadline after preparation, and Loom
	// bounds this whole top-level unit at 90 seconds.
	cancelContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	command := lintCommand(cancelContext, "", os.Args[0], "-test.run=^"+name+"$", "-test.timeout=0", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_COMPLETE_SUGGESTION_PLANT="+strconv.FormatBool(plant))
	if plant {
		command.Env = append(command.Env, "ADAMIC_COMPLETE_SUGGESTION_PLANT=1")
	}
	// Read through childguard.Capture for lintCommand's reason, stdout then stderr. Capture kills the child's group
	// once it exits, so no product build it started outlives it.
	stdout, stderr, captureError := childguard.Capture(cancelContext, command)
	return append(stdout, stderr...), captureError
}

func grain30IsolationResult(index int, output []byte, grain30ProbeError error) error {
	row := grain30IsolationCases[index]
	if grain30ProbeError != nil || !bytes.Contains(output, []byte("--- PASS: "+row.test+" ")) {
		return fmt.Errorf("isolated %s: %v\n%s", row.test, grain30ProbeError, output)
	}
	if row.product != "" && !productLookedUp(output, regexp.QuoteMeta(row.product)) {
		return fmt.Errorf("isolated %s never looked up its product %s itself:\n%s", row.test, row.product, output)
	}
	return nil
}

func grain30Isolation(test *testing.T, index int) {
	test.Helper()
	started := time.Now()
	// The child prepares and looks up everything its selected test reads; this process reads none of it.
	output, grain30ProbeError := grain30Probe(test, grain30IsolationCases[index].test, false)
	test.Logf("preparation: %.3fs", time.Since(started).Seconds())
	own := time.Now()
	defer func() { test.Logf("own work: %.3fs", time.Since(own).Seconds()) }()
	if failure := grain30IsolationResult(index, output, grain30ProbeError); failure != nil {
		test.Fatal(failure)
	}
	// Plant a failed selection into the real receipt. The same verifier must
	// reject it in every child, including the leaf whose cache is already warm.
	broken := bytes.ReplaceAll(output, []byte("--- PASS: "), []byte("--- FAIL: "))
	if grain30IsolationResult(index, broken, grain30ProbeError) == nil {
		test.Fatal("planted selection failure survived")
	}
	if product := grain30IsolationCases[index].product; product != "" {
		broken = bytes.ReplaceAll(output, []byte("build "+product+" "), []byte("built "+product+" "))
		if grain30IsolationResult(index, broken, grain30ProbeError) == nil {
			test.Fatal("planted missing product lookup survived")
		}
	}
	test.Logf("case %d exactly once; planted failure caught", index)
}

func grain30PlantResult(shard int, output []byte, grain30ProbeError error) error {
	name := fmt.Sprintf("TestCompleteSuggestionSerialization_%03d", shard)
	failures := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "--- FAIL:") {
			failures = append(failures, line)
		}
	}
	if shard == 1 {
		if grain30ProbeError == nil ||
			len(failures) != 1 ||
			!strings.HasPrefix(failures[0], "--- FAIL: "+name+" ") ||
			!bytes.Contains(output, []byte("planted suggestion disagreement")) {
			return fmt.Errorf("wrong planted failure: %v; failures %v\n%s", grain30ProbeError, failures, output)
		}
	} else if grain30ProbeError != nil || len(failures) != 0 || !bytes.Contains(output, []byte("--- PASS: "+name+" ")) {
		return fmt.Errorf("non-owner %s failed: %v\n%s", name, grain30ProbeError, output)
	}
	return nil
}

func grain30Plant(test *testing.T, shard int) {
	test.Helper()
	// The child prepares and looks up everything its leaf reads; this process reads none of it.
	own := time.Now()
	defer func() { test.Logf("own work: %.3fs", time.Since(own).Seconds()) }()
	output, grain30ProbeError := grain30Probe(
		test,
		fmt.Sprintf("TestCompleteSuggestionSerialization_%03d", shard),
		true,
	)
	if failure := grain30PlantResult(shard, output, grain30ProbeError); failure != nil {
		test.Fatal(failure)
	}
	broken := bytes.ReplaceAll(output, []byte("--- PASS: "), []byte("--- FAIL: "))
	if shard == 1 {
		broken = bytes.ReplaceAll(output, []byte("planted suggestion disagreement"), []byte("missing disagreement"))
	}
	if grain30PlantResult(shard, broken, grain30ProbeError) == nil {
		test.Fatal("planted probe-verdict failure survived")
	}
	test.Logf("leaf %d exactly once; planted failure caught", shard)
}

func TestCompleteSuggestionSerialization_IsolatedShard_000(test *testing.T) {
	test.Parallel()
	grain30Isolation(test, 0)
}
func TestCompleteSuggestionSerialization_IsolatedShard_001(test *testing.T) {
	test.Parallel()
	grain30Isolation(test, 1)
}
func TestCompleteSuggestionSerialization_IsolatedShard_002(test *testing.T) {
	test.Parallel()
	grain30Isolation(test, 2)
}
func TestCompleteSuggestionSerialization_IsolatedShard_003(test *testing.T) {
	test.Parallel()
	grain30Isolation(test, 3)
}
func TestCompleteSuggestionSerialization_IsolatedShard_004(test *testing.T) {
	test.Parallel()
	grain30Isolation(test, 4)
}
func TestCompleteSuggestionSerialization_PlantedFailure_000(test *testing.T) {
	test.Parallel()
	grain30Plant(test, 0)
}
func TestCompleteSuggestionSerialization_PlantedFailure_001(test *testing.T) {
	test.Parallel()
	grain30Plant(test, 1)
}
func TestCompleteSuggestionSerialization_PlantedFailure_002(test *testing.T) {
	test.Parallel()
	grain30Plant(test, 2)
}
func TestCompleteSuggestionSerialization_PlantedFailure_003(test *testing.T) {
	test.Parallel()
	grain30Plant(test, 3)
}
func TestCompleteSuggestionSerialization_PlantedFailure_004(test *testing.T) {
	test.Parallel()
	grain30Plant(test, 4)
}
func TestCompleteSuggestionSerialization_PlantedFailure_005(test *testing.T) {
	test.Parallel()
	grain30Plant(test, 5)
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

func TestCompleteSuggestionSerialization_ProbeUnion(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("grain30_serialization_probes_test.go")
	if readError != nil {
		test.Fatal(readError)
	}
	pattern := regexp.MustCompile(
		`(?s)func TestCompleteSuggestionSerialization_(IsolatedShard|PlantedFailure)_([0-9]+)\(t \*testing.T\) \{\s*t\.Parallel\(\)\s*(grain30Isolation|grain30Plant)\(t, ([0-9]+)\)\s*\}`,
	)
	matches := pattern.FindAllSubmatch(source, -1)
	if len(matches) != len(grain30IsolationCases)+len(completeSuggestionCases) {
		test.Fatalf("enumerated %d probes", len(matches))
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
				test.Fatalf("invalid probe %s", match[0])
			}
			groups[owner] = append(groups[owner], id)
		}
		// Empty future groups are legal. Mutation checks must neither index an
		// empty group nor claim that removing nothing caused missing coverage.
		groups = append(groups, nil)
		if grain30ProbeUnionError := grain30ProbeUnion(groups, count); grain30ProbeUnionError != nil {
			test.Fatal(grain30ProbeUnionError)
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
						if grain30ProbeUnionError := grain30ProbeUnion(broken, count); grain30ProbeUnionError != nil {
							test.Fatal(grain30ProbeUnionError)
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
					test.Fatalf("%s group %d: %s survived", kind, owner, fault)
				}
			}
		}
		test.Logf("%s: %d cases exactly once, including empty-group self-check", kind, count)
	}
}
