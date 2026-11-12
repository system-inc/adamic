package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// statsFieldFixtures are a Stats' every field (#6xfkdwj), held to Node by the ordinary families: native
// under the sanitizers and the leak check, the JavaScript backend, release, gob and counted, and
// wasm32-wasi, which refuses fields.a and mode.a by name at their read of mode (wasiRefusedFixtures)
// and agrees with Node on wasi_held.a, the fields its filestat holds. fields.a's mutants, run by hand:
// node_fs_file.c's file_stat_value writing mtime into the ctimeMs slot, and its change-time row reads
// false; union.c's keyed_stats making a fresh Date for each read of one, and its by-key twice rows do.
var statsFieldFixtures = []string{"fields", "wasi_held", "mode"}

func init() {
	for _, name := range statsFieldFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/stats_fields/" + name + ".a", true, false})
	}
}

func statsFieldPath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/stats_fields", name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// keyed_mode.a reads a Stats' mode by a run-time key, which answers as Node's natively, released and on
// the JavaScript backend (#6xfkdwj). It isn't one of fixtures, since on wasm32-wasi it stops
// (TestStatsFieldByKeyStopsOnWASI).
func TestStatsFieldByKeyAnswers(t *testing.T) {
	t.Parallel()
	path := statsFieldPath(t, "keyed_mode.a")
	truth := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("directory true\nmode number\n")}, truth); difference != "" {
		t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	nativeRun, _ := natively(t, program)
	for name, got := range map[string]run{"native": nativeRun, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
		}
	}
}

// On wasm32-wasi a run-time key read of a field its filestat has no answer for stops by name, in the
// StatsField form, before any slot is read (union.c's keyed_stats), after what Node printed before it.
// Mutant: that stop dropped, and the wasm32 leg reads the slot and prints mode number.
func TestStatsFieldByKeyStopsOnWASI(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	program, err := lowered(t, statsFieldPath(t, "keyed_mode.a"))
	if err != nil {
		t.Fatal(err)
	}
	got := onWASI(t, native.C(program))
	want := reflectGetStop(t, "directory true\n", "StatsField", "mode")
	if difference := disagreement(want, got); difference != "" {
		t.Fatalf("%s: exit %d, stdout %q, stderr %q", difference, got.exitCode, got.stdout, got.stderr)
	}
}

// mode.a's wasm32 refusal is what keeps wasi-libc's mode from a program: built with the read's name
// pasted together, so the build's token check doesn't see it, the binary runs and prints the file type
// alone, which only the comparison with Node catches (WebAssembly's pin on #6xfkdwj). The refusal
// itself is held by wasiRefusedFixtures and TestWASIRefusesStatsFieldsByName.
func TestWASIStatsModeWithoutItsRefusalIsTypeOnly(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path := statsFieldPath(t, "mode.a")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	if strings.Count(code, "adamic_fs_file_stats_mode(") != 1 {
		t.Fatal("mode.a doesn't read mode through adamic_fs_file_stats_mode once")
	}
	code = strings.Replace(code, "#include \"adamic.h\"", "#include \"adamic.h\"\n#define ADAMIC_STATS_MODE_PASTED adamic_fs_file_stats_ ## mode", 1)
	code = strings.Replace(code, "adamic_fs_file_stats_mode(", "ADAMIC_STATS_MODE_PASTED(", 1)
	got := onWASI(t, code)
	truth := onNode(t, path)
	if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) != "stdout differs" || !strings.Contains(string(got.stdout), "mode 16384\n") {
		t.Fatalf("want the type-only mode 16384 that only Node's stdout catches: node %q, wasi %+v", truth.stdout, got)
	}
}
