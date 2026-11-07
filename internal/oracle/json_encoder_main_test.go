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
var jsonEncoderFixtures = []string{
	"scalars",
	"values",
	"options",
	"escapes",
	"numbers",
	"undefined",
	"indent",
	"keys",
	"replacer",
	"census_append_overload",
}

// One fixture and one runtime configuration per unit, including cold archives.
func jsonEncoderMainFixture(test *testing.T, index, backend int) {
	test.Helper()
	name := jsonEncoderFixtures[index]
	fixture := "internal/oracle/testdata/json_stringify_" + name + ".a"
	if name == "census_append_overload" {
		fixture = "internal/oracle/testdata/" + name + ".a"
	}
	path, err := filepath.Abs(filepath.Join(repository, fixture))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	want := onNodeWith(test, inputRun{}, path)
	code := native.C(program)
	var actual run
	if backend == 2 {
		script := filepath.Join(test.TempDir(), "program.mjs")
		if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0600); err != nil {
			test.Fatal(err)
		}
		actual = onNodeWith(test, inputRun{}, script)
	} else {
		options := native.Options{Sanitize: backend == 0, Count: backend == 4}
		if backend == 3 {
			if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
				test.Skip("set ADAMIC_ORACLE_WASI=1")
			}
			options.Target = "wasm32-wasi"
		}
		binary := filepath.Join(test.TempDir(), "program")
		if err := native.Build(code, binary, options); err != nil {
			test.Fatal(err)
		}
		if backend == 3 {
			actual = execute(
				test,
				"node",
				"--disable-warning=ExperimentalWarning",
				filepath.Join(repository, "oracle/wasi.mjs"),
				binary,
			)
		} else {
			actual = executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
		}
		if backend == 4 {
			match := countsLine.FindSubmatch(actual.stderr)
			if match == nil {
				test.Fatalf("missing counts: %s", actual.stderr)
			}
			numbers := []string{}
			for _, value := range match[1:] {
				numbers = append(numbers, string(value))
			}
			row := "| " + fixture + " | " + strings.Join(numbers, " | ") + " |"
			recorded, err := recordedCounts()
			if err != nil {
				test.Fatal(err)
			}
			if recorded[fixture] != row {
				test.Fatalf("counts differ: got %s; recorded %s", row, recorded[fixture])
			}
			actual.stderr = countsLine.ReplaceAll(actual.stderr, nil)
		}
	}
	if difference := disagreement(want, actual); difference != "" {
		test.Fatalf("backend %d: %s", backend, difference)
	}
}

// Production callers supply no providers; missing metadata must never be approximated.
func jsonEncoderMissingProviders(test *testing.T, backend int) {
	code := jsonEncoderSource(test)
	for _, target := range []string{"native", "wasm32-wasi"}[backend : backend+1] {
		test.Run(target, func(test *testing.T) {
			test.Parallel()
			options := native.Options{}
			if target == "wasm32-wasi" {
				if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
					test.Skip("set ADAMIC_ORACLE_WASI=1")
				}
				options.Target = target
			}
			binary := filepath.Join(test.TempDir(), "probe")
			if err := native.Build(code, binary, options); err != nil {
				test.Fatal(err)
			}
			for _, probe := range []struct{ arg, reason string }{
				{"empty_array", "JSON.stringify runtime array without complete element descriptors"},
				{"array", "JSON.stringify runtime array without complete element descriptors"},
				{"union", "JSON union lacks proven container metadata"},
				{"missing_hook", "JSON toJSON lacks a typed runtime result provider"},
			} {
				test.Run(probe.arg, func(test *testing.T) {
					var actual run
					if target == "native" {
						actual = execute(test, binary, probe.arg)
					} else {
						actual = execute(
							test,
							"node",
							"--disable-warning=ExperimentalWarning",
							filepath.Join(repository, "oracle", "wasi.mjs"),
							binary,
							probe.arg,
						)
					}
					if actual.exitCode != 70 || !strings.Contains(string(actual.stderr), probe.reason) {
						test.Fatalf(
							"want exit 70 and %q: exit %d stdout %q stderr %q",
							probe.reason,
							actual.exitCode,
							actual.stdout,
							actual.stderr,
						)
					}
				})
			}
		})
	}
}

func jsonEncoderGapNUL(test *testing.T, backend int) {
	// The review runner still marks this witness pending; exercise it directly
	// against Node on each backend, including its length and escaped-result checks.
	path, err := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/review/agree/fxspptb_bdb950b_gap_nul.a"),
	)
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	want := onNodeWith(test, inputRun{}, path)
	options := native.Options{Sanitize: backend == 0}
	if backend == 2 {
		if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
			test.Skip("set ADAMIC_ORACLE_WASI=1")
		}
		options.Target = "wasm32-wasi"
	}
	binary := filepath.Join(test.TempDir(), "program")
	if err := native.Build(native.C(program), binary, options); err != nil {
		test.Fatal(err)
	}
	var actual run
	if backend == 2 {
		actual = execute(
			test,
			"node",
			"--disable-warning=ExperimentalWarning",
			filepath.Join(repository, "oracle/wasi.mjs"),
			binary,
		)
	} else {
		actual = executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	}
	if difference := disagreement(want, actual); difference != "" {
		test.Fatal(difference)
	}
}

func TestJSONEncoderMainFixture_000(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 0, 0)
}

func TestJSONEncoderMainFixture_001(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 0, 1)
}

func TestJSONEncoderMainFixture_002(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 0, 2)
}

func TestJSONEncoderMainFixture_003(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 0, 3)
}

func TestJSONEncoderMainFixture_004(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 0, 4)
}

func TestJSONEncoderMainFixture_005(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 1, 0)
}

func TestJSONEncoderMainFixture_006(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 1, 1)
}

func TestJSONEncoderMainFixture_007(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 1, 2)
}

func TestJSONEncoderMainFixture_008(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 1, 3)
}

func TestJSONEncoderMainFixture_009(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 1, 4)
}

func TestJSONEncoderMainFixture_010(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 2, 0)
}

func TestJSONEncoderMainFixture_011(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 2, 1)
}

func TestJSONEncoderMainFixture_012(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 2, 2)
}

func TestJSONEncoderMainFixture_013(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 2, 3)
}

func TestJSONEncoderMainFixture_014(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 2, 4)
}

func TestJSONEncoderMainFixture_015(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 3, 0)
}

func TestJSONEncoderMainFixture_016(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 3, 1)
}

func TestJSONEncoderMainFixture_017(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 3, 2)
}

func TestJSONEncoderMainFixture_018(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 3, 3)
}

func TestJSONEncoderMainFixture_019(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 3, 4)
}

func TestJSONEncoderMainFixture_020(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 4, 0)
}

func TestJSONEncoderMainFixture_021(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 4, 1)
}

func TestJSONEncoderMainFixture_022(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 4, 2)
}

func TestJSONEncoderMainFixture_023(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 4, 3)
}

func TestJSONEncoderMainFixture_024(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 4, 4)
}

func TestJSONEncoderMainFixture_025(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 5, 0)
}

func TestJSONEncoderMainFixture_026(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 5, 1)
}

func TestJSONEncoderMainFixture_027(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 5, 2)
}

func TestJSONEncoderMainFixture_028(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 5, 3)
}

func TestJSONEncoderMainFixture_029(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 5, 4)
}

func TestJSONEncoderMainFixture_030(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 6, 0)
}

func TestJSONEncoderMainFixture_031(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 6, 1)
}

func TestJSONEncoderMainFixture_032(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 6, 2)
}

func TestJSONEncoderMainFixture_033(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 6, 3)
}

func TestJSONEncoderMainFixture_034(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 6, 4)
}

func TestJSONEncoderMainFixture_035(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 7, 0)
}

func TestJSONEncoderMainFixture_036(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 7, 1)
}

func TestJSONEncoderMainFixture_037(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 7, 2)
}

func TestJSONEncoderMainFixture_038(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 7, 3)
}

func TestJSONEncoderMainFixture_039(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 7, 4)
}

func TestJSONEncoderMainFixture_040(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 8, 0)
}

func TestJSONEncoderMainFixture_041(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 8, 1)
}

func TestJSONEncoderMainFixture_042(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 8, 2)
}

func TestJSONEncoderMainFixture_043(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 8, 3)
}

func TestJSONEncoderMainFixture_044(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 8, 4)
}

func TestJSONEncoderMainFixture_045(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 9, 0)
}

func TestJSONEncoderMainFixture_046(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 9, 1)
}

func TestJSONEncoderMainFixture_047(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 9, 2)
}

func TestJSONEncoderMainFixture_048(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 9, 3)
}

func TestJSONEncoderMainFixture_049(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 9, 4)
}

func TestJSONEncoderMissingProviders_000(test *testing.T) {
	test.Parallel()
	jsonEncoderMissingProviders(test, 0)
}

func TestJSONEncoderMissingProviders_001(test *testing.T) {
	test.Parallel()
	jsonEncoderMissingProviders(test, 1)
}

func TestJSONEncoderGapNUL_000(test *testing.T) { test.Parallel(); jsonEncoderGapNUL(test, 0) }

func TestJSONEncoderGapNUL_001(test *testing.T) { test.Parallel(); jsonEncoderGapNUL(test, 1) }

func TestJSONEncoderGapNUL_002(test *testing.T) { test.Parallel(); jsonEncoderGapNUL(test, 2) }
