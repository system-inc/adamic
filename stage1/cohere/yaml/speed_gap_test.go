package yaml

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestSpeedCostProbes(t *testing.T) {
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
		t.Run(probe.file, func(t *testing.T) {
			entry, err := filepath.Abs(filepath.Join("gaps", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{entry})
			if err != nil {
				t.Fatal(err)
			}
			lowered, err := lower.Lower(context.Background(), program)
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(t.TempDir(), "probe")
			if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			emitted := filepath.Join(t.TempDir(), "probe.mjs")
			if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
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
