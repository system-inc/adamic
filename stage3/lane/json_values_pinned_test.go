package lane

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// A public-any exemption remains justified only while its pinned API sample still reads the unchecked value.
// This test runs git to read the pinned object, never Go or a compiler. Entries may be removed, never added here.
func TestJSONValues52PinnedAnySampleReadsRemain(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "stage3/adapt/52-json-values/public-any-pinned.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Path     []string `json:"path"`
		Kind     string   `json:"kind"`
		Baseline string   `json:"baseline"`
		Read     string   `json:"read"`
	}
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) > 2 {
		t.Fatal("the reviewed public-any exemption list grew")
	}
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Fatal("ADAMIC_TYPESCRIPT_SOURCE must name the pinned upstream git input")
	}
	command := toolchain.GitCommandIn(source, "show", typescriptPin(t, root)+":tests/cases/compiler/APISample_parseConfig.ts")
	sample, err := command.Output()
	if err != nil {
		t.Fatalf("read pinned API sample: %v", err)
	}
	seen := map[string]bool{}
	for _, row := range rows {
		if len(row.Path) != 2 || row.Path[1] != "config" ||
			(row.Path[0] != "readConfigFile" && row.Path[0] != "parseConfigFileTextToJson") ||
			row.Kind != "PropertySignature" || row.Baseline != "APISample_parseConfig.errors.txt" ||
			row.Read != `config.config["compilerOptions"]` {
			t.Fatalf("unreviewed public-any exemption: %+v", row)
		}
		name := strings.Join(row.Path, ".")
		if seen[name] {
			t.Fatalf("duplicate exemption %s", name)
		}
		seen[name] = true
		if !strings.Contains(string(sample), row.Read) {
			t.Errorf("%s is no longer pinned by %s: remove its exemption; missing read %q", name, row.Baseline, row.Read)
		}
	}
}
