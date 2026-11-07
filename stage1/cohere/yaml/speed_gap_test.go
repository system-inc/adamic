package yaml

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSpeedCostProbes(test *testing.T) {
	test.Parallel()
	input := filepath.Join(test.TempDir(), "units.txt")
	if writeError := os.WriteFile(input, []byte(strings.Repeat("abc中😀", 4000)), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	runner := filepath.Join(root, "oracle/node.mjs")
	for _, probe := range []struct{ file, first, second, answer string }{
		{"stringUnitScan.ts", "slice", "numeric", "52998400000\n"},
		{"arenaNodeRead.ts", "accessor", "direct", "57600000000\n"},
		{"numericMapLookup.ts", "integer", "fractional", "1702000\n"},
	} {
		// Not parallel: one probe at a time keeps this test to one child at a time, its share.
		test.Run(probe.file, func(test *testing.T) {
			binary := yamlSpeedProbesBuild.binary(test, "gaps/"+probe.file)
			emitted := yamlSpeedProbesBuild.emitted(test, "gaps/"+probe.file)
			entry, absolutePathError := filepath.Abs(filepath.Join("gaps", probe.file))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
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
					{
						"native ASan/UBSan/LSan",
						run(test, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, argument, mode, "100"),
					},
					{"Node", run(
						test,
						"",
						nil,
						"node",
						"--disable-warning=ExperimentalWarning",
						runner,
						entry,
						argument,
						mode,
						"100",
					)},
					{"emitted JavaScript", run(
						test,
						"",
						nil,
						"node",
						"--disable-warning=ExperimentalWarning",
						runner,
						emitted,
						argument,
						mode,
						"100",
					)},
				} {
					if !bytes.Equal(side.out, []byte(probe.answer)) {
						test.Fatalf("%s %s: got %q, want %q", mode, side.name, side.out, probe.answer)
					}
				}
			}
			test.Logf("both modes match Node's checksum %s", strings.TrimSpace(probe.answer))
		})
	}
}
