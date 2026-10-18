package oracle

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These are upstream TypeScript checker-profile witnesses, not standalone
// Adamic admissions. Their source bytes and Node goldens survive the rename.
func TestOriginalArrayWitnessExtensions(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane2/originalB/witness-extensions.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct{ From, To, SHA256 string }
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 473 {
		t.Fatalf("renamed %d witnesses, expected 473", len(rows))
	}
	for _, row := range rows {
		if !strings.HasSuffix(row.To, ".ts") || !strings.HasSuffix(row.From, ".a") {
			t.Fatalf("wrong extension: %#v", row)
		}
		bytes, err := os.ReadFile(filepath.Join(repository, row.To))
		if err != nil {
			t.Fatal(err)
		}
		if actual := fmt.Sprintf("%x", sha256.Sum256(bytes)); actual != row.SHA256 {
			t.Fatalf("witness bytes changed: %s", row.To)
		}
		if _, err = os.Stat(filepath.Join(repository, row.From)); !os.IsNotExist(err) {
			t.Fatalf("old Adamic witness remains: %s", row.From)
		}
	}
}

// Each top-level shard holds the same source-profile probes under the leaf budget.
func originalArrayWitnessNodeShard(t *testing.T, shard int) {
	t.Helper()
	ordinal := 0
	for _, group := range []string{"B", "B2", "B3", "B4"} {
		directory := filepath.Join(repository, "stage3/interface-downcasts/lane2/original"+group)
		directory, err := filepath.Abs(directory)
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(directory, "probes.json"))
		if err != nil {
			t.Fatal(err)
		}
		var probes []struct{ Name, Source string }
		if err = json.Unmarshal(data, &probes); err != nil {
			t.Fatal(err)
		}
		for _, probe := range probes {
			selected := ordinal%16 == shard
			ordinal++
			if !selected {
				continue
			}
			if d := disagreement(run{stdout: []byte(probe.Source)}, execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), filepath.Join(directory, probe.Name+".ts"))); d != "" {
				t.Fatalf("%s/%s: %s", group, probe.Name, d)
			}
		}
	}
}

func TestOriginalArrayWitnessNode_0000(t *testing.T) {
	t.Parallel()
	originalArrayWitnessNodeShard(t, 0)
}

func TestOriginalArrayWitnessNode_0001(t *testing.T) {
	t.Parallel()
	originalArrayWitnessNodeShard(t, 1)
}

func TestOriginalArrayWitnessNode_0002(t *testing.T) {
	t.Parallel()
	originalArrayWitnessNodeShard(t, 2)
}

func TestOriginalArrayWitnessNode_0003(t *testing.T) {
	t.Parallel()
	originalArrayWitnessNodeShard(t, 3)
}

func TestOriginalArrayWitnessNode_0004(t *testing.T) {
	t.Parallel()
	originalArrayWitnessNodeShard(t, 4)
}

func TestOriginalArrayWitnessNode_0005(t *testing.T) {
	t.Parallel()
	originalArrayWitnessNodeShard(t, 5)
}

func TestOriginalArrayWitnessNode_0006(t *testing.T) {
	t.Parallel()
	originalArrayWitnessNodeShard(t, 6)
}

func TestOriginalArrayWitnessNode_0007(t *testing.T) {
	t.Parallel()
	originalArrayWitnessNodeShard(t, 7)
}

func TestOriginalArrayWitnessNode_0008(t *testing.T) {
	t.Parallel()
	originalArrayWitnessNodeShard(t, 8)
}

func TestOriginalArrayWitnessNode_0009(t *testing.T) {
	t.Parallel()
	originalArrayWitnessNodeShard(t, 9)
}

func TestOriginalArrayWitnessNode_0010(t *testing.T) {
	t.Parallel()
	originalArrayWitnessNodeShard(t, 10)
}

func TestOriginalArrayWitnessNode_0011(t *testing.T) {
	t.Parallel()
	originalArrayWitnessNodeShard(t, 11)
}

func TestOriginalArrayWitnessNode_0012(t *testing.T) {
	t.Parallel()
	originalArrayWitnessNodeShard(t, 12)
}

func TestOriginalArrayWitnessNode_0013(t *testing.T) {
	t.Parallel()
	originalArrayWitnessNodeShard(t, 13)
}

func TestOriginalArrayWitnessNode_0014(t *testing.T) {
	t.Parallel()
	originalArrayWitnessNodeShard(t, 14)
}

func TestOriginalArrayWitnessNode_0015(t *testing.T) {
	t.Parallel()
	originalArrayWitnessNodeShard(t, 15)
}
