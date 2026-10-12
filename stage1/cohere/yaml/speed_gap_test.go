package yaml

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSpeedCostProbes(t *testing.T) {
	t.Parallel()
	input := filepath.Join(t.TempDir(), "units.txt")
	if err := os.WriteFile(input, []byte(strings.Repeat("abc中😀", 4000)), 0644); err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	runner := filepath.Join(root, "oracle/node.mjs")
	for _, probe := range []struct{ file, first, second, answer string }{
		{"stringUnitScan.ts", "slice", "numeric", "52998400000\n"},
		{"arenaNodeRead.ts", "accessor", "direct", "57600000000\n"},
		{"numericMapLookup.ts", "integer", "fractional", "1702000\n"},
	} {
		// Not parallel: one probe at a time keeps this test to one child at a time, its share.
		t.Run(probe.file, func(t *testing.T) {
			binary := yamlSpeedProbesBuild.binary(t, "gaps/"+probe.file)
			emitted := yamlSpeedProbesBuild.emitted(t, "gaps/"+probe.file)
			entry, err := filepath.Abs(filepath.Join("gaps", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			argument := input
			if probe.file == "numericMapLookup.ts" {
				argument = "185"
			}
			for _, mode := range []string{probe.first, probe.second} {
				for _, side := range []struct {
					name string
					out  []byte
				}{
					{"native ASan/UBSan/LSan", run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, argument, mode, "100")},
					{"Node", run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, argument, mode, "100")},
					{"emitted JavaScript", run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, emitted, argument, mode, "100")},
				} {
					if !bytes.Equal(side.out, []byte(probe.answer)) {
						t.Fatalf("%s %s: got %q, want %q", mode, side.name, side.out, probe.answer)
					}
				}
			}
			t.Logf("both modes match Node's checksum %s", strings.TrimSpace(probe.answer))
		})
	}
}
