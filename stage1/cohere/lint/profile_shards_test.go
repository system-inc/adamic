package lint

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

// Validate the whole plan even when this process runs only a selected shard.
func profileUnion(rows []string, assignments [][]string, count int) error {
	if len(assignments) != count {
		return fmt.Errorf("enumerated %d shards, want %d", len(assignments), count)
	}
	want := make(map[string]bool, len(rows))
	for _, row := range rows {
		if want[row] {
			return fmt.Errorf("duplicate unsplit case %q", row)
		}
		want[row] = true
	}
	seen := make(map[string]bool, len(rows))
	total := 0
	for _, cases := range assignments {
		for _, row := range cases {
			total++
			if !want[row] || seen[row] {
				return fmt.Errorf("unknown or repeated case %q", row)
			}
			seen[row] = true
		}
	}
	if total != len(rows) || len(seen) != len(want) {
		return fmt.Errorf("union has %d cases, want %d", total, len(rows))
	}
	return nil
}

func profileAssignments(t *testing.T, rows []string, count int) [][]string {
	t.Helper()
	if count < 1 {
		t.Fatal("shard count must be positive")
	}
	assignments := make([][]string, count)
	for i, row := range rows {
		assignments[i*count/len(rows)] = append(assignments[i*count/len(rows)], row)
	}
	if err := profileUnion(rows, assignments, count); err != nil {
		t.Fatal(err)
	}
	t.Logf("shard union: %d cases, %d unique case IDs, %d shards", len(rows), len(rows), count)
	return assignments
}

func profileSelected(t *testing.T, index, count int) bool {
	t.Helper()
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return true
	}
	fields := strings.Split(value, "/")
	if len(fields) != 2 {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q, want i/n", value)
	}
	i, err := strconv.Atoi(fields[0])
	n, other := strconv.Atoi(fields[1])
	if err != nil || other != nil || n != count || i < 0 || i >= n {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q for %d shards", value, count)
	}
	return index == i
}

func profileCheck(t *testing.T, got, want []byte) {
	t.Helper()
	if diff := difference(got, want); diff != "" {
		t.Fatal(diff)
	}
}

func TestProfileShardFailureProbe(t *testing.T) {
	if os.Getenv("ADAMIC_PROFILE_FAILURE_PROBE") != "1" {
		return
	}
	rows := []string{"case-000", "case-001"}
	for i, cases := range profileAssignments(t, rows, 2) {
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			for _, id := range cases {
				got := []byte(id)
				if id == "case-001" {
					got = []byte("planted disagreement")
				}
				profileCheck(t, got, []byte(id))
			}
		})
	}
}

func TestProfileShardCatchesPlantedDisagreement(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "-test.run=^TestProfileShardFailureProbe$", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_PROFILE_FAILURE_PROBE=1")
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("planted disagreement survived")
	}
	text := string(output)
	if strings.Count(text, "--- FAIL: TestProfileShardFailureProbe/shard-") != 1 || !strings.Contains(text, "--- FAIL: TestProfileShardFailureProbe/shard-001") || !strings.Contains(text, "--- PASS: TestProfileShardFailureProbe/shard-000") {
		t.Fatalf("wrong failure attribution:\n%s", text)
	}
	t.Log("planted case-001 disagreement caught only by shard-001")
}

func TestProfileShardUnion(t *testing.T) {
	rows := []string{"a", "b"}
	for _, plan := range [][][]string{{{"a"}}, {{"a"}, {"a"}}, {{"a"}, {"c"}}} {
		if profileUnion(rows, plan, 2) == nil {
			t.Fatalf("invalid union accepted: %v", plan)
		}
	}
	if err := profileUnion(rows, [][]string{{"a"}, {"b"}}, 2); err != nil {
		t.Fatal(err)
	}
}
