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

func TestOriginalArrayWitnessNode(t *testing.T) {
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
			t.Run(group+"/"+probe.Name, func(t *testing.T) {
				t.Parallel()
				if d := disagreement(run{stdout: []byte(probe.Source)}, execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), filepath.Join(directory, probe.Name+".ts"))); d != "" {
					t.Fatal(d)
				}
			})
		}
	}
}
