package css

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/stage1/cohere/css/oracles"
)

// The unmutated printer runs every case in both modes, in sixty-four content-hash partitions per mode. Keep ownership
// keyed by case bytes; doubling splits each former bucket in two without changing the checks or their own-work
// deadline. At sixteen partitions, alone on one core, _000 and _016 passed the 60 s own-work deadline, _027 took 54.6 s
// and _011 49.9 s (Workshop, Oct 11), and Server panicked two leaves past it (landable-18 run 3). A unit runs one leaf
// per CPU (testguard.Share), so a leaf's wall time is its one-core work at any share: sixty-four keeps it well under
// half. No leaf prepares another leaf's executable.
const cssPrinterPartitions = 64

// Each printer mutant runs, in each mode, only over its witnesses: the cases where a full-corpus run found it differs
// from Go (cssPrinterWitnesses), the smallest cssPrinterWitnessesRun of them. Run over the rest it could fail only by
// crashing, which says something about the mutant, never about the printer: on Oct 12 those runs took 1,780 of the
// family's 2,928 active seconds (#testaudit). 128 unmutated shards, both modes by sixty-four partitions, then one
// witness shard per mutant and mode.
const testCssPrinterAgreesWithGoShards = 2*cssPrinterPartitions + 2*3

var cssPrinterCorpusOnce sync.Once
var cssPrinterCorpus []string

// cssPrinterCorpusLines reads the printer's corpus from the parser's corpus product (TestProduct_CSSParserCorpus):
// the same Go oracle request over the same roots, so the same cases, which TestCSSPrinterCorpusIsTheParserCorpus
// holds byte for byte. Enumerating it live cost every printer shard's process about 3.7 s of its setup (Home).
func cssPrinterCorpusLines(test *testing.T) []string {
	test.Helper()
	cssPrinterCorpusOnce.Do(func() {
		path, _ := oracles.CachedParserOutputs(test)
		data, readError := os.ReadFile(path)
		if readError != nil {
			test.Fatal(readError)
		}
		cssPrinterCorpus = strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	})
	return cssPrinterCorpus
}

// The printer reads the parser's corpus product only because a live enumeration through the printer's own Go
// oracle gives the same bytes. Mutant, run by hand with ADAMIC_CSS_FIXTURES set (without it there's no fixture root to
// drop): drop the fixture root from the live enumeration; the corpora then differ.
func TestCssPrinterCorpusIsTheParserCorpus(test *testing.T) {
	test.Parallel()
	live, readError := os.ReadFile(cssPrinterCases(test, oracles.ComparisonOracle(test, false)))
	if readError != nil {
		test.Fatal(readError)
	}
	product, _ := oracles.CachedParserOutputs(test)
	declared, readError := os.ReadFile(product)
	if readError != nil {
		test.Fatal(readError)
	}
	if difference := firstDifference(string(live), string(declared)); difference != "" {
		test.Fatalf("the printer's live corpus differs from the parser's corpus product: %s", difference)
	}
	test.Logf("one corpus, %d cases, %d bytes", strings.Count(string(live), "\n"), len(live))
}

func cssPrinterOwnDeadline(test *testing.T) func() {
	test.Helper()
	started := time.Now()
	timer := time.AfterFunc(60*time.Second, func() { panic(fmt.Sprintf("%s: own work exceeded 60s", test.Name())) })
	return func() { timer.Stop(); test.Logf("own work: %s", time.Since(started)) }
}

// cssPrinterSanitizedAgreement runs the sanitized printer once and holds it to Go's answers. On Linux leak detection is
// on in that run, so one run checks the output, ASan/UBSan and LeakSanitizer, where a second run checked leaks before:
// the slowest corpus case (a 30 KB Tailwind preflight, about 7 s sanitized, #z3p4j5q) ran twice in its shards.
// macOS checks leaks with its own run (printerLeaks). The error says which check failed: a leak, LeakSanitizer
// unable to run (a red, never a pass), a sanitizer's error, or an output that isn't Go's.
func cssPrinterSanitizedAgreement(test *testing.T, binary, expected string, arguments ...string) error {
	test.Helper()
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=1"}
	}
	return cssPrinterSanitizedVerdict(execute(test, environment, binary, arguments...), expected)
}

// cssPrinterSanitizedVerdict is cssPrinterSanitizedAgreement's reading of one run.
func cssPrinterSanitizedVerdict(result run, expected string) error {
	stderr := string(result.stderr)
	switch {
	case strings.Contains(stderr, "LeakSanitizer has encountered a fatal error"):
		return fmt.Errorf("LeakSanitizer could not run, so leaks went unchecked (exit %d): %s", result.exitCode, stderr)
	case strings.Contains(stderr, "ERROR: LeakSanitizer: detected memory leaks"):
		return fmt.Errorf("leak (exit %d): %s", result.exitCode, stderr)
	case result.exitCode != 0 || len(stderr) != 0:
		return fmt.Errorf("exit %d stderr %s", result.exitCode, stderr)
	}
	if difference := firstDifference(string(result.stdout), expected); difference != "" {
		return fmt.Errorf("output differs from Go: %s", difference)
	}
	return nil
}

// The one sanitized run still names each failure. A printer whose releases are dropped (the composed memory leak
// mutant) reads as a leak, a changed answer as an output that isn't Go's, and LeakSanitizer unable to run as a red;
// the unmutated printer passes.
func TestCssPrinterSanitizedRunNamesEachFailure(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "linux" {
		test.Skip("Linux runs leak detection inside the sanitized run")
	}
	cases := filepath.Join(test.TempDir(), "cases.txt")
	if writeError := os.WriteFile(
		cases,
		[]byte(">Ca{color:red;margin:0.50px}\n>Sa{--x:{b:c};color:rgb(1,2,3)}\n"),
		0644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	expected := cssPrinterAnswers(test, oracles.ComparisonOracle(test, true), cases, "default")
	arguments := []string{cases, "output", "once", "default"}
	if cssPrinterSanitizedAgreementError := cssPrinterSanitizedAgreement(
		test,
		cssPrinterProduct(test, -1).binary,
		expected,
		arguments...,
	); cssPrinterSanitizedAgreementError != nil {
		test.Fatalf("the unmutated printer: %v", cssPrinterSanitizedAgreementError)
	}
	cssPrinterSanitizedAgreementError := cssPrinterSanitizedAgreement(
		test,
		composedMemoryProduct(test, 2).sanitized,
		expected,
		arguments...,
	)
	if cssPrinterSanitizedAgreementError == nil ||
		!strings.HasPrefix(cssPrinterSanitizedAgreementError.Error(), "leak ") {
		test.Fatalf("a printer that drops its releases read as %v, want a leak", cssPrinterSanitizedAgreementError)
	}
	test.Logf("planted leak: %.120s", cssPrinterSanitizedAgreementError)
	cssPrinterSanitizedAgreementError = cssPrinterSanitizedAgreement(
		test,
		composedMemoryProduct(test, 0).sanitized,
		expected,
		arguments...,
	)
	if cssPrinterSanitizedAgreementError == nil ||
		!strings.HasPrefix(cssPrinterSanitizedAgreementError.Error(), "exit ") ||
		!strings.Contains(cssPrinterSanitizedAgreementError.Error(), "AddressSanitizer") {
		test.Fatalf(
			"a printer that reads freed memory read as %v, want a sanitizer's error",
			cssPrinterSanitizedAgreementError,
		)
	}
	test.Logf("planted use after free: %.120s", cssPrinterSanitizedAgreementError)
	changed := strings.Replace(expected, "red", "blue", 1)
	if changed == expected {
		test.Fatal("no answer to change")
	}
	cssPrinterSanitizedAgreementError = cssPrinterSanitizedAgreement(
		test,
		cssPrinterProduct(test, -1).binary,
		changed,
		arguments...,
	)
	if cssPrinterSanitizedAgreementError == nil ||
		!strings.HasPrefix(cssPrinterSanitizedAgreementError.Error(), "output differs from Go") {
		test.Fatalf("a changed answer read as %v, want an output that isn't Go's", cssPrinterSanitizedAgreementError)
	}
	cssPrinterSanitizedVerdictError := cssPrinterSanitizedVerdict(run{
		stdout:   []byte(expected),
		stderr:   []byte("==1==LeakSanitizer has encountered a fatal error.\n"),
		exitCode: 1,
	}, expected)
	if cssPrinterSanitizedVerdictError == nil ||
		!strings.HasPrefix(cssPrinterSanitizedVerdictError.Error(), "LeakSanitizer could not run") {
		test.Fatalf("LeakSanitizer unable to run read as %v, want a red", cssPrinterSanitizedVerdictError)
	}
}

func runCssPrinterShard(test *testing.T, shard int) {
	test.Helper()
	setup := time.Now()
	lines := cssPrinterCorpusLines(test)
	var unit cssModeShard
	if shard < 2*cssPrinterPartitions {
		unit = cssPrinterUnmutatedPlan(lines)[shard]
	} else {
		variant, modeIndex := (shard-2*cssPrinterPartitions)/2, (shard-2*cssPrinterPartitions)%2
		ids := cssPrinterWitnesses(test, variant)[modeIndex]
		if err := cssPrinterWitnessError(printerMutants[variant].name, cssPrinterModes[modeIndex], ids); err != nil {
			test.Fatal(err)
		}
		unit = cssPrinterMutantShard(lines, variant, modeIndex, ids)
	}
	printerOracle := oracles.ComparisonOracle(test, true)
	product := cssPrinterProduct(test, unit.mutant)
	leakBinary := product.binary
	if unit.mutant < 0 && runtime.GOOS == "darwin" {
		leakBinary = cssPrinterProduct(test, 3).binary
	}
	repo, _ := filepath.Abs(repository)
	test.Logf("setup including products: %s", time.Since(setup))
	defer cssPrinterOwnDeadline(test)()
	path := cssShardFile(test, unit.cases)
	expected := cssPrinterAnswers(test, printerOracle, path, unit.mode)
	arguments := []string{path, "output", "once", unit.mode}
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	if unit.mutant < 0 {
		if cssPrinterSanitizedAgreementError := cssPrinterSanitizedAgreement(
			test,
			product.binary,
			expected,
			arguments...,
		); cssPrinterSanitizedAgreementError != nil {
			test.Fatalf("%s native ASan/UBSan: %v", unit.mode, cssPrinterSanitizedAgreementError)
		}
		for _, side := range []struct {
			name   string
			result run
		}{
			{"Node", onNode(test, product.source, arguments...)},
			{"JavaScript backend", onNode(test, product.javascript, arguments...)},
		} {
			if cssAgreementError := cssAgreement(side.result, expected); cssAgreementError != nil {
				test.Fatalf("%s %s: %v", unit.mode, side.name, cssAgreementError)
			}
		}
		if runtime.GOOS != "linux" {
			if report := printerLeaks(test, leakBinary, arguments...); report != "" {
				test.Fatal(report)
			}
		}
		if library := os.Getenv("ADAMIC_CSS_PRINTER_LIBRARY"); library != "" {
			comparePrinterLibrary(test, path, expected, library, unit.mode, "npm")
			comparePrinterLibrary(
				test,
				path,
				expected,
				filepath.Join(repo, "cohere", "internal", "format", "prettier", "bundles"),
				unit.mode,
				"fork",
			)
		} else {
			test.Log(
				"optional library comparisons unavailable: set ADAMIC_CSS_PRINTER_LIBRARY to scratch Prettier 3.9.6",
			)
		}
		test.Logf("%s: %d cases; three backends, sanitizers and leaks checked", unit.mode, len(unit.cases))
	} else {
		mutation := printerMutants[unit.mutant]
		for _, side := range []struct {
			name   string
			result run
		}{
			{"native ASan/UBSan", execute(test, environment, product.binary, arguments...)},
			{"Node", onNode(test, product.source, arguments...)},
		} {
			if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
				test.Fatalf(
					"%s %s mutant %s must run: %d %s",
					unit.mode,
					side.name,
					mutation.name,
					side.result.exitCode,
					side.result.stderr,
				)
			}
			differing, cssPrinterDifferingCasesError := cssPrinterDifferingCases(string(side.result.stdout), expected)
			if cssPrinterDifferingCasesError != nil {
				test.Fatalf(
					"%s %s printer mutant %s: %v",
					unit.mode,
					side.name,
					mutation.name,
					cssPrinterDifferingCasesError,
				)
			}
			if len(differing) != len(unit.cases) {
				test.Fatalf(
					"%s %s printer mutant %s differs from Go on %d of its %d chosen witnesses",
					unit.mode,
					side.name,
					mutation.name,
					len(differing),
					len(unit.cases),
				)
			}
			test.Logf(
				"%s %s caught %s on all %d chosen witnesses: %s",
				unit.mode,
				side.name,
				mutation.name,
				len(unit.cases),
				firstDifference(string(side.result.stdout), expected),
			)
		}
		test.Logf(
			"%s mutant %s: %d chosen witnesses; sanitized native and Node checked",
			unit.mode,
			mutation.name,
			len(unit.cases),
		)
	}
}

func TestCssPrinterAgreesWithGoUnion(test *testing.T) {
	test.Parallel()
	lines := cssPrinterCorpusLines(test)
	witnesses := cssPrinterAllWitnesses(test)
	defer cssPrinterOwnDeadline(test)()
	plan := cssModePlan(lines, witnesses)
	if cssModeUnionError := cssModeUnion(lines, plan, witnesses); cssModeUnionError != nil {
		test.Fatal(cssModeUnionError)
	}
	for offset, unit := range plan[2*cssPrinterPartitions:] {
		test.Logf(
			"mutant %s, %s: %d witnesses, its shard runs the %d smallest",
			printerMutants[unit.mutant].name,
			unit.mode,
			len(witnesses[unit.mutant][offset%2]),
			len(unit.cases),
		)
	}
	test.Logf(
		"counted union: %d cases x 2 modes unmutated in %d shards, and %d mutant witness shards; %d shards",
		len(lines),
		2*cssPrinterPartitions,
		len(plan)-2*cssPrinterPartitions,
		len(plan),
	)
}

func TestCssPrinterAgreesWithGo_000(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 0) }
func TestCssPrinterAgreesWithGo_001(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 1) }
func TestCssPrinterAgreesWithGo_002(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 2) }
func TestCssPrinterAgreesWithGo_003(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 3) }
func TestCssPrinterAgreesWithGo_004(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 4) }
func TestCssPrinterAgreesWithGo_005(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 5) }
func TestCssPrinterAgreesWithGo_006(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 6) }
func TestCssPrinterAgreesWithGo_007(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 7) }
func TestCssPrinterAgreesWithGo_008(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 8) }
func TestCssPrinterAgreesWithGo_009(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 9) }
func TestCssPrinterAgreesWithGo_010(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 10) }
func TestCssPrinterAgreesWithGo_011(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 11) }
func TestCssPrinterAgreesWithGo_012(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 12) }
func TestCssPrinterAgreesWithGo_013(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 13) }
func TestCssPrinterAgreesWithGo_014(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 14) }
func TestCssPrinterAgreesWithGo_015(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 15) }
func TestCssPrinterAgreesWithGo_016(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 16) }
func TestCssPrinterAgreesWithGo_017(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 17) }
func TestCssPrinterAgreesWithGo_018(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 18) }
func TestCssPrinterAgreesWithGo_019(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 19) }
func TestCssPrinterAgreesWithGo_020(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 20) }
func TestCssPrinterAgreesWithGo_021(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 21) }
func TestCssPrinterAgreesWithGo_022(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 22) }
func TestCssPrinterAgreesWithGo_023(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 23) }
func TestCssPrinterAgreesWithGo_024(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 24) }
func TestCssPrinterAgreesWithGo_025(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 25) }
func TestCssPrinterAgreesWithGo_026(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 26) }
func TestCssPrinterAgreesWithGo_027(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 27) }
func TestCssPrinterAgreesWithGo_028(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 28) }
func TestCssPrinterAgreesWithGo_029(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 29) }
func TestCssPrinterAgreesWithGo_030(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 30) }
func TestCssPrinterAgreesWithGo_031(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 31) }
func TestCssPrinterAgreesWithGo_032(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 32) }
func TestCssPrinterAgreesWithGo_033(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 33) }
func TestCssPrinterAgreesWithGo_034(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 34) }
func TestCssPrinterAgreesWithGo_035(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 35) }
func TestCssPrinterAgreesWithGo_036(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 36) }
func TestCssPrinterAgreesWithGo_037(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 37) }
func TestCssPrinterAgreesWithGo_038(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 38) }
func TestCssPrinterAgreesWithGo_039(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 39) }
func TestCssPrinterAgreesWithGo_040(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 40) }
func TestCssPrinterAgreesWithGo_041(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 41) }
func TestCssPrinterAgreesWithGo_042(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 42) }
func TestCssPrinterAgreesWithGo_043(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 43) }
func TestCssPrinterAgreesWithGo_044(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 44) }
func TestCssPrinterAgreesWithGo_045(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 45) }
func TestCssPrinterAgreesWithGo_046(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 46) }
func TestCssPrinterAgreesWithGo_047(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 47) }
func TestCssPrinterAgreesWithGo_048(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 48) }
func TestCssPrinterAgreesWithGo_049(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 49) }
func TestCssPrinterAgreesWithGo_050(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 50) }
func TestCssPrinterAgreesWithGo_051(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 51) }
func TestCssPrinterAgreesWithGo_052(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 52) }
func TestCssPrinterAgreesWithGo_053(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 53) }
func TestCssPrinterAgreesWithGo_054(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 54) }
func TestCssPrinterAgreesWithGo_055(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 55) }
func TestCssPrinterAgreesWithGo_056(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 56) }
func TestCssPrinterAgreesWithGo_057(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 57) }
func TestCssPrinterAgreesWithGo_058(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 58) }
func TestCssPrinterAgreesWithGo_059(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 59) }
func TestCssPrinterAgreesWithGo_060(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 60) }
func TestCssPrinterAgreesWithGo_061(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 61) }
func TestCssPrinterAgreesWithGo_062(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 62) }
func TestCssPrinterAgreesWithGo_063(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 63) }
func TestCssPrinterAgreesWithGo_064(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 64) }
func TestCssPrinterAgreesWithGo_065(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 65) }
func TestCssPrinterAgreesWithGo_066(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 66) }
func TestCssPrinterAgreesWithGo_067(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 67) }
func TestCssPrinterAgreesWithGo_068(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 68) }
func TestCssPrinterAgreesWithGo_069(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 69) }
func TestCssPrinterAgreesWithGo_070(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 70) }
func TestCssPrinterAgreesWithGo_071(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 71) }
func TestCssPrinterAgreesWithGo_072(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 72) }
func TestCssPrinterAgreesWithGo_073(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 73) }
func TestCssPrinterAgreesWithGo_074(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 74) }
func TestCssPrinterAgreesWithGo_075(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 75) }
func TestCssPrinterAgreesWithGo_076(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 76) }
func TestCssPrinterAgreesWithGo_077(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 77) }
func TestCssPrinterAgreesWithGo_078(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 78) }
func TestCssPrinterAgreesWithGo_079(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 79) }
func TestCssPrinterAgreesWithGo_080(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 80) }
func TestCssPrinterAgreesWithGo_081(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 81) }
func TestCssPrinterAgreesWithGo_082(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 82) }
func TestCssPrinterAgreesWithGo_083(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 83) }
func TestCssPrinterAgreesWithGo_084(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 84) }
func TestCssPrinterAgreesWithGo_085(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 85) }
func TestCssPrinterAgreesWithGo_086(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 86) }
func TestCssPrinterAgreesWithGo_087(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 87) }
func TestCssPrinterAgreesWithGo_088(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 88) }
func TestCssPrinterAgreesWithGo_089(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 89) }
func TestCssPrinterAgreesWithGo_090(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 90) }
func TestCssPrinterAgreesWithGo_091(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 91) }
func TestCssPrinterAgreesWithGo_092(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 92) }
func TestCssPrinterAgreesWithGo_093(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 93) }
func TestCssPrinterAgreesWithGo_094(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 94) }
func TestCssPrinterAgreesWithGo_095(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 95) }
func TestCssPrinterAgreesWithGo_096(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 96) }
func TestCssPrinterAgreesWithGo_097(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 97) }
func TestCssPrinterAgreesWithGo_098(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 98) }
func TestCssPrinterAgreesWithGo_099(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 99) }
func TestCssPrinterAgreesWithGo_100(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 100) }
func TestCssPrinterAgreesWithGo_101(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 101) }
func TestCssPrinterAgreesWithGo_102(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 102) }
func TestCssPrinterAgreesWithGo_103(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 103) }
func TestCssPrinterAgreesWithGo_104(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 104) }
func TestCssPrinterAgreesWithGo_105(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 105) }
func TestCssPrinterAgreesWithGo_106(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 106) }
func TestCssPrinterAgreesWithGo_107(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 107) }
func TestCssPrinterAgreesWithGo_108(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 108) }
func TestCssPrinterAgreesWithGo_109(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 109) }
func TestCssPrinterAgreesWithGo_110(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 110) }
func TestCssPrinterAgreesWithGo_111(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 111) }
func TestCssPrinterAgreesWithGo_112(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 112) }
func TestCssPrinterAgreesWithGo_113(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 113) }
func TestCssPrinterAgreesWithGo_114(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 114) }
func TestCssPrinterAgreesWithGo_115(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 115) }
func TestCssPrinterAgreesWithGo_116(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 116) }
func TestCssPrinterAgreesWithGo_117(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 117) }
func TestCssPrinterAgreesWithGo_118(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 118) }
func TestCssPrinterAgreesWithGo_119(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 119) }
func TestCssPrinterAgreesWithGo_120(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 120) }
func TestCssPrinterAgreesWithGo_121(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 121) }
func TestCssPrinterAgreesWithGo_122(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 122) }
func TestCssPrinterAgreesWithGo_123(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 123) }
func TestCssPrinterAgreesWithGo_124(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 124) }
func TestCssPrinterAgreesWithGo_125(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 125) }
func TestCssPrinterAgreesWithGo_126(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 126) }
func TestCssPrinterAgreesWithGo_127(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 127) }
func TestCssPrinterAgreesWithGo_128(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 128) }
func TestCssPrinterAgreesWithGo_129(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 129) }
func TestCssPrinterAgreesWithGo_130(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 130) }
func TestCssPrinterAgreesWithGo_131(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 131) }
func TestCssPrinterAgreesWithGo_132(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 132) }
func TestCssPrinterAgreesWithGo_133(test *testing.T) { test.Parallel(); runCssPrinterShard(test, 133) }
