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
	path, absolutePathError := filepath.Abs(filepath.Join(repository, fixture))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := onNodeWith(test, inputRun{}, path)
	code := native.C(program)
	var actual run
	if backend == 2 {
		script := filepath.Join(test.TempDir(), "program.mjs")
		if writeError := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0600); writeError != nil {
			test.Fatal(writeError)
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
		if buildError := native.Build(code, binary, options); buildError != nil {
			test.Fatal(buildError)
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
			recorded, recordedCountsError := recordedCounts()
			if recordedCountsError != nil {
				test.Fatal(recordedCountsError)
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
			if buildError := native.Build(code, binary, options); buildError != nil {
				test.Fatal(buildError)
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

func jsonEncoderGapNul(test *testing.T, backend int) {
	// The review runner still marks this witness pending; exercise it directly
	// against Node on each backend, including its length and escaped-result checks.
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/review/agree/fxspptb_bdb950b_gap_nul.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
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
	if buildError := native.Build(native.C(program), binary, options); buildError != nil {
		test.Fatal(buildError)
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

func TestJsonEncoderMainFixture_000(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 0, 0)
}

func TestJsonEncoderMainFixture_001(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 0, 1)
}

func TestJsonEncoderMainFixture_002(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 0, 2)
}

func TestJsonEncoderMainFixture_003(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 0, 3)
}

func TestJsonEncoderMainFixture_004(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 0, 4)
}

func TestJsonEncoderMainFixture_005(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 1, 0)
}

func TestJsonEncoderMainFixture_006(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 1, 1)
}

func TestJsonEncoderMainFixture_007(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 1, 2)
}

func TestJsonEncoderMainFixture_008(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 1, 3)
}

func TestJsonEncoderMainFixture_009(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 1, 4)
}

func TestJsonEncoderMainFixture_010(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 2, 0)
}

func TestJsonEncoderMainFixture_011(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 2, 1)
}

func TestJsonEncoderMainFixture_012(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 2, 2)
}

func TestJsonEncoderMainFixture_013(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 2, 3)
}

func TestJsonEncoderMainFixture_014(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 2, 4)
}

func TestJsonEncoderMainFixture_015(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 3, 0)
}

func TestJsonEncoderMainFixture_016(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 3, 1)
}

func TestJsonEncoderMainFixture_017(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 3, 2)
}

func TestJsonEncoderMainFixture_018(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 3, 3)
}

func TestJsonEncoderMainFixture_019(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 3, 4)
}

func TestJsonEncoderMainFixture_020(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 4, 0)
}

func TestJsonEncoderMainFixture_021(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 4, 1)
}

func TestJsonEncoderMainFixture_022(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 4, 2)
}

func TestJsonEncoderMainFixture_023(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 4, 3)
}

func TestJsonEncoderMainFixture_024(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 4, 4)
}

func TestJsonEncoderMainFixture_025(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 5, 0)
}

func TestJsonEncoderMainFixture_026(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 5, 1)
}

func TestJsonEncoderMainFixture_027(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 5, 2)
}

func TestJsonEncoderMainFixture_028(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 5, 3)
}

func TestJsonEncoderMainFixture_029(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 5, 4)
}

func TestJsonEncoderMainFixture_030(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 6, 0)
}

func TestJsonEncoderMainFixture_031(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 6, 1)
}

func TestJsonEncoderMainFixture_032(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 6, 2)
}

func TestJsonEncoderMainFixture_033(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 6, 3)
}

func TestJsonEncoderMainFixture_034(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 6, 4)
}

func TestJsonEncoderMainFixture_035(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 7, 0)
}

func TestJsonEncoderMainFixture_036(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 7, 1)
}

func TestJsonEncoderMainFixture_037(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 7, 2)
}

func TestJsonEncoderMainFixture_038(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 7, 3)
}

func TestJsonEncoderMainFixture_039(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 7, 4)
}

func TestJsonEncoderMainFixture_040(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 8, 0)
}

func TestJsonEncoderMainFixture_041(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 8, 1)
}

func TestJsonEncoderMainFixture_042(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 8, 2)
}

func TestJsonEncoderMainFixture_043(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 8, 3)
}

func TestJsonEncoderMainFixture_044(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 8, 4)
}

func TestJsonEncoderMainFixture_045(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 9, 0)
}

func TestJsonEncoderMainFixture_046(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 9, 1)
}

func TestJsonEncoderMainFixture_047(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 9, 2)
}

func TestJsonEncoderMainFixture_048(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 9, 3)
}

func TestJsonEncoderMainFixture_049(test *testing.T) {
	test.Parallel()
	jsonEncoderMainFixture(test, 9, 4)
}

func TestJsonEncoderMissingProviders_000(test *testing.T) {
	test.Parallel()
	jsonEncoderMissingProviders(test, 0)
}

func TestJsonEncoderMissingProviders_001(test *testing.T) {
	test.Parallel()
	jsonEncoderMissingProviders(test, 1)
}

func TestJsonEncoderGapNul_000(test *testing.T) { test.Parallel(); jsonEncoderGapNul(test, 0) }

func TestJsonEncoderGapNul_001(test *testing.T) { test.Parallel(); jsonEncoderGapNul(test, 1) }

func TestJsonEncoderGapNul_002(test *testing.T) { test.Parallel(); jsonEncoderGapNul(test, 2) }
