package lint

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func positionMutantCheck(t *testing.T, name string, got, want []byte) {
	t.Helper()
	if diff := difference(got, want); diff == "" {
		t.Fatalf("position mutant survived on %s", name)
	} else {
		t.Logf("position mutant caught on %s: %s", name, diff)
	}
}

func TestPositionShardFailureProbe(t *testing.T) {
	if os.Getenv("ADAMIC_POSITION_FAILURE_PROBE") != "1" {
		return
	}
	rows := []string{"case-000", "case-001"}
	for i, cases := range profileAssignments(t, rows, 2) {
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			for _, id := range cases {
				got := []byte("killed mutant")
				if id == "case-001" {
					got = []byte(id)
				}
				positionMutantCheck(t, "planted survivor", got, []byte(id))
			}
		})
	}
}

func TestPositionShardCatchesPlantedSurvivor(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "-test.run=^TestPositionShardFailureProbe$", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_POSITION_FAILURE_PROBE=1")
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("planted surviving mutant was accepted")
	}
	text := string(output)
	if strings.Count(text, "--- FAIL: TestPositionShardFailureProbe/shard-") != 1 || !strings.Contains(text, "--- FAIL: TestPositionShardFailureProbe/shard-001") || !strings.Contains(text, "--- PASS: TestPositionShardFailureProbe/shard-000") {
		t.Fatalf("wrong failure attribution:\n%s", text)
	}
	t.Log("planted case-001 surviving mutant caught only by shard-001")
}
