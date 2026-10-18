package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// These programs use only the schemas main can prove, without runtime providers.
var jsonEncoderFixtures = []string{"scalars", "values", "options", "escapes", "numbers", "undefined", "indent", "keys", "replacer", "census_append_overload"}

// One fixture and one runtime configuration per unit, including cold archives.
func jsonEncoderMainFixture(t *testing.T, index, backend int) {
	t.Helper()
	name := jsonEncoderFixtures[index]
	fixture := "internal/oracle/testdata/json_stringify_" + name + ".a"
	if name == "census_append_overload" {
		fixture = "internal/oracle/testdata/" + name + ".a"
	}
	path, err := filepath.Abs(filepath.Join(repository, fixture))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNodeWith(t, inputRun{}, path)
	code := native.C(program)
	var actual run
	if backend == 2 {
		script := filepath.Join(t.TempDir(), "program.mjs")
		if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0600); err != nil {
			t.Fatal(err)
		}
		actual = onNodeWith(t, inputRun{}, script)
	} else {
		options := native.Options{Sanitize: backend == 0, Count: backend == 4}
		if backend == 3 {
			if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
				t.Skip("set ADAMIC_ORACLE_WASI=1")
			}
			options.Target = "wasm32-wasi"
		}
		binary := filepath.Join(t.TempDir(), "program")
		if err := native.Build(code, binary, options); err != nil {
			t.Fatal(err)
		}
		if backend == 3 {
			actual = execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
		} else {
			actual = executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
		}
		if backend == 4 {
			match := countsLine.FindSubmatch(actual.stderr)
			if match == nil {
				t.Fatalf("missing counts: %s", actual.stderr)
			}
			numbers := []string{}
			for _, value := range match[1:] {
				numbers = append(numbers, string(value))
			}
			row := "| " + fixture + " | " + strings.Join(numbers, " | ") + " |"
			recorded, err := recordedCounts()
			if err != nil {
				t.Fatal(err)
			}
			if recorded[fixture] != row {
				t.Fatalf("counts differ: got %s; recorded %s", row, recorded[fixture])
			}
			actual.stderr = countsLine.ReplaceAll(actual.stderr, nil)
		}
	}
	if difference := disagreement(want, actual); difference != "" {
		t.Fatalf("backend %d: %s", backend, difference)
	}
}

// Production callers supply no providers; missing metadata must never be approximated.
func jsonEncoderMissingProviders(t *testing.T, backend int) {
	code := jsonEncoderSource(t)
	for _, target := range []string{"native", "wasm32-wasi"}[backend : backend+1] {
		t.Run(target, func(t *testing.T) {
			t.Parallel()
			options := native.Options{}
			if target == "wasm32-wasi" {
				if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
					t.Skip("set ADAMIC_ORACLE_WASI=1")
				}
				options.Target = target
			}
			binary := filepath.Join(t.TempDir(), "probe")
			if err := native.Build(code, binary, options); err != nil {
				t.Fatal(err)
			}
			for _, probe := range []struct{ arg, reason string }{
				{"empty_array", "JSON.stringify runtime array without complete element descriptors"},
				{"array", "JSON.stringify runtime array without complete element descriptors"},
				{"union", "JSON union lacks proven container metadata"},
				{"missing_hook", "JSON toJSON lacks a typed runtime result provider"},
			} {
				t.Run(probe.arg, func(t *testing.T) {
					var actual run
					if target == "native" {
						actual = execute(t, binary, probe.arg)
					} else {
						actual = execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "wasi.mjs"), binary, probe.arg)
					}
					if actual.exitCode != 70 || !strings.Contains(string(actual.stderr), probe.reason) {
						t.Fatalf("want exit 70 and %q: exit %d stdout %q stderr %q", probe.reason, actual.exitCode, actual.stdout, actual.stderr)
					}
				})
			}
		})
	}
}

func jsonEncoderGapNUL(t *testing.T, backend int) {
	// The review runner still marks this witness pending; exercise it directly
	// against Node on each backend, including its length and escaped-result checks.
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/review/agree/fxspptb_bdb950b_gap_nul.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNodeWith(t, inputRun{}, path)
	options := native.Options{Sanitize: backend == 0}
	if backend == 2 {
		if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
			t.Skip("set ADAMIC_ORACLE_WASI=1")
		}
		options.Target = "wasm32-wasi"
	}
	binary := filepath.Join(t.TempDir(), "program")
	if err := native.Build(native.C(program), binary, options); err != nil {
		t.Fatal(err)
	}
	var actual run
	if backend == 2 {
		actual = execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
	} else {
		actual = executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	}
	if difference := disagreement(want, actual); difference != "" {
		t.Fatal(difference)
	}
}

func TestJSONEncoderMainFixture_000(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 0, 0) }

func TestJSONEncoderMainFixture_001(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 0, 1) }

func TestJSONEncoderMainFixture_002(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 0, 2) }

func TestJSONEncoderMainFixture_003(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 0, 3) }

func TestJSONEncoderMainFixture_004(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 0, 4) }

func TestJSONEncoderMainFixture_005(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 1, 0) }

func TestJSONEncoderMainFixture_006(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 1, 1) }

func TestJSONEncoderMainFixture_007(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 1, 2) }

func TestJSONEncoderMainFixture_008(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 1, 3) }

func TestJSONEncoderMainFixture_009(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 1, 4) }

func TestJSONEncoderMainFixture_010(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 2, 0) }

func TestJSONEncoderMainFixture_011(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 2, 1) }

func TestJSONEncoderMainFixture_012(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 2, 2) }

func TestJSONEncoderMainFixture_013(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 2, 3) }

func TestJSONEncoderMainFixture_014(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 2, 4) }

func TestJSONEncoderMainFixture_015(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 3, 0) }

func TestJSONEncoderMainFixture_016(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 3, 1) }

func TestJSONEncoderMainFixture_017(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 3, 2) }

func TestJSONEncoderMainFixture_018(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 3, 3) }

func TestJSONEncoderMainFixture_019(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 3, 4) }

func TestJSONEncoderMainFixture_020(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 4, 0) }

func TestJSONEncoderMainFixture_021(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 4, 1) }

func TestJSONEncoderMainFixture_022(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 4, 2) }

func TestJSONEncoderMainFixture_023(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 4, 3) }

func TestJSONEncoderMainFixture_024(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 4, 4) }

func TestJSONEncoderMainFixture_025(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 5, 0) }

func TestJSONEncoderMainFixture_026(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 5, 1) }

func TestJSONEncoderMainFixture_027(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 5, 2) }

func TestJSONEncoderMainFixture_028(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 5, 3) }

func TestJSONEncoderMainFixture_029(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 5, 4) }

func TestJSONEncoderMainFixture_030(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 6, 0) }

func TestJSONEncoderMainFixture_031(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 6, 1) }

func TestJSONEncoderMainFixture_032(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 6, 2) }

func TestJSONEncoderMainFixture_033(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 6, 3) }

func TestJSONEncoderMainFixture_034(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 6, 4) }

func TestJSONEncoderMainFixture_035(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 7, 0) }

func TestJSONEncoderMainFixture_036(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 7, 1) }

func TestJSONEncoderMainFixture_037(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 7, 2) }

func TestJSONEncoderMainFixture_038(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 7, 3) }

func TestJSONEncoderMainFixture_039(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 7, 4) }

func TestJSONEncoderMainFixture_040(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 8, 0) }

func TestJSONEncoderMainFixture_041(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 8, 1) }

func TestJSONEncoderMainFixture_042(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 8, 2) }

func TestJSONEncoderMainFixture_043(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 8, 3) }

func TestJSONEncoderMainFixture_044(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 8, 4) }

func TestJSONEncoderMainFixture_045(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 9, 0) }

func TestJSONEncoderMainFixture_046(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 9, 1) }

func TestJSONEncoderMainFixture_047(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 9, 2) }

func TestJSONEncoderMainFixture_048(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 9, 3) }

func TestJSONEncoderMainFixture_049(t *testing.T) { t.Parallel(); jsonEncoderMainFixture(t, 9, 4) }

func TestJSONEncoderMissingProviders_000(t *testing.T) {
	t.Parallel()
	jsonEncoderMissingProviders(t, 0)
}

func TestJSONEncoderMissingProviders_001(t *testing.T) {
	t.Parallel()
	jsonEncoderMissingProviders(t, 1)
}

func TestJSONEncoderGapNUL_000(t *testing.T) { t.Parallel(); jsonEncoderGapNUL(t, 0) }

func TestJSONEncoderGapNUL_001(t *testing.T) { t.Parallel(); jsonEncoderGapNUL(t, 1) }

func TestJSONEncoderGapNUL_002(t *testing.T) { t.Parallel(); jsonEncoderGapNUL(t, 2) }
