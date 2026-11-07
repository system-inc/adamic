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
const testCSSPrinterAgreesWithGoShards = 2*cssPrinterPartitions + 2*3

var cssPrinterCorpusOnce sync.Once
var cssPrinterCorpus []string

// cssPrinterCorpusLines reads the printer's corpus from the parser's corpus product (TestProduct_CSSParserCorpus):
// the same Go oracle request over the same roots, so the same cases, which TestCSSPrinterCorpusIsTheParserCorpus
// holds byte for byte. Enumerating it live cost every printer shard's process about 3.7 s of its setup (Home).
func cssPrinterCorpusLines(test *testing.T) []string {
	test.Helper()
	cssPrinterCorpusOnce.Do(func() {
		path, _ := oracles.CachedParserOutputs(test)
		data, err := os.ReadFile(path)
		if err != nil {
			test.Fatal(err)
		}
		cssPrinterCorpus = strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	})
	return cssPrinterCorpus
}

// The printer reads the parser's corpus product only because a live enumeration through the printer's own Go
// oracle gives the same bytes. Mutant, run by hand with ADAMIC_CSS_FIXTURES set (without it there's no fixture root to
// drop): drop the fixture root from the live enumeration; the corpora then differ.
func TestCSSPrinterCorpusIsTheParserCorpus(test *testing.T) {
	test.Parallel()
	live, err := os.ReadFile(cssPrinterCases(test, oracles.ComparisonOracle(test, false)))
	if err != nil {
		test.Fatal(err)
	}
	product, _ := oracles.CachedParserOutputs(test)
	declared, err := os.ReadFile(product)
	if err != nil {
		test.Fatal(err)
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
func TestCSSPrinterSanitizedRunNamesEachFailure(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "linux" {
		test.Skip("Linux runs leak detection inside the sanitized run")
	}
	cases := filepath.Join(test.TempDir(), "cases.txt")
	if err := os.WriteFile(
		cases,
		[]byte(">Ca{color:red;margin:0.50px}\n>Sa{--x:{b:c};color:rgb(1,2,3)}\n"),
		0644,
	); err != nil {
		test.Fatal(err)
	}
	expected := cssPrinterAnswers(test, oracles.ComparisonOracle(test, true), cases, "default")
	arguments := []string{cases, "output", "once", "default"}
	if err := cssPrinterSanitizedAgreement(
		test,
		cssPrinterProduct(test, -1).binary,
		expected,
		arguments...,
	); err != nil {
		test.Fatalf("the unmutated printer: %v", err)
	}
	err := cssPrinterSanitizedAgreement(test, composedMemoryProduct(test, 2).sanitized, expected, arguments...)
	if err == nil || !strings.HasPrefix(err.Error(), "leak ") {
		test.Fatalf("a printer that drops its releases read as %v, want a leak", err)
	}
	test.Logf("planted leak: %.120s", err)
	err = cssPrinterSanitizedAgreement(test, composedMemoryProduct(test, 0).sanitized, expected, arguments...)
	if err == nil || !strings.HasPrefix(err.Error(), "exit ") || !strings.Contains(err.Error(), "AddressSanitizer") {
		test.Fatalf("a printer that reads freed memory read as %v, want a sanitizer's error", err)
	}
	test.Logf("planted use after free: %.120s", err)
	changed := strings.Replace(expected, "red", "blue", 1)
	if changed == expected {
		test.Fatal("no answer to change")
	}
	err = cssPrinterSanitizedAgreement(test, cssPrinterProduct(test, -1).binary, changed, arguments...)
	if err == nil || !strings.HasPrefix(err.Error(), "output differs from Go") {
		test.Fatalf("a changed answer read as %v, want an output that isn't Go's", err)
	}
	err = cssPrinterSanitizedVerdict(run{
		stdout:   []byte(expected),
		stderr:   []byte("==1==LeakSanitizer has encountered a fatal error.\n"),
		exitCode: 1,
	}, expected)
	if err == nil || !strings.HasPrefix(err.Error(), "LeakSanitizer could not run") {
		test.Fatalf("LeakSanitizer unable to run read as %v, want a red", err)
	}
}

func runCSSPrinterShard(test *testing.T, shard int) {
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
		if err := cssPrinterSanitizedAgreement(test, product.binary, expected, arguments...); err != nil {
			test.Fatalf("%s native ASan/UBSan: %v", unit.mode, err)
		}
		for _, side := range []struct {
			name   string
			result run
		}{
			{"Node", onNode(test, product.source, arguments...)},
			{"JavaScript backend", onNode(test, product.javascript, arguments...)},
		} {
			if err := cssAgreement(side.result, expected); err != nil {
				test.Fatalf("%s %s: %v", unit.mode, side.name, err)
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
			differing, err := cssPrinterDifferingCases(string(side.result.stdout), expected)
			if err != nil {
				test.Fatalf("%s %s printer mutant %s: %v", unit.mode, side.name, mutation.name, err)
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

func TestCSSPrinterAgreesWithGoUnion(test *testing.T) {
	test.Parallel()
	lines := cssPrinterCorpusLines(test)
	witnesses := cssPrinterAllWitnesses(test)
	defer cssPrinterOwnDeadline(test)()
	plan := cssModePlan(lines, witnesses)
	if err := cssModeUnion(lines, plan, witnesses); err != nil {
		test.Fatal(err)
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

func TestCSSPrinterAgreesWithGo_000(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 0) }
func TestCSSPrinterAgreesWithGo_001(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 1) }
func TestCSSPrinterAgreesWithGo_002(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 2) }
func TestCSSPrinterAgreesWithGo_003(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 3) }
func TestCSSPrinterAgreesWithGo_004(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 4) }
func TestCSSPrinterAgreesWithGo_005(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 5) }
func TestCSSPrinterAgreesWithGo_006(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 6) }
func TestCSSPrinterAgreesWithGo_007(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 7) }
func TestCSSPrinterAgreesWithGo_008(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 8) }
func TestCSSPrinterAgreesWithGo_009(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 9) }
func TestCSSPrinterAgreesWithGo_010(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 10) }
func TestCSSPrinterAgreesWithGo_011(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 11) }
func TestCSSPrinterAgreesWithGo_012(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 12) }
func TestCSSPrinterAgreesWithGo_013(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 13) }
func TestCSSPrinterAgreesWithGo_014(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 14) }
func TestCSSPrinterAgreesWithGo_015(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 15) }
func TestCSSPrinterAgreesWithGo_016(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 16) }
func TestCSSPrinterAgreesWithGo_017(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 17) }
func TestCSSPrinterAgreesWithGo_018(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 18) }
func TestCSSPrinterAgreesWithGo_019(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 19) }
func TestCSSPrinterAgreesWithGo_020(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 20) }
func TestCSSPrinterAgreesWithGo_021(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 21) }
func TestCSSPrinterAgreesWithGo_022(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 22) }
func TestCSSPrinterAgreesWithGo_023(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 23) }
func TestCSSPrinterAgreesWithGo_024(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 24) }
func TestCSSPrinterAgreesWithGo_025(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 25) }
func TestCSSPrinterAgreesWithGo_026(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 26) }
func TestCSSPrinterAgreesWithGo_027(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 27) }
func TestCSSPrinterAgreesWithGo_028(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 28) }
func TestCSSPrinterAgreesWithGo_029(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 29) }
func TestCSSPrinterAgreesWithGo_030(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 30) }
func TestCSSPrinterAgreesWithGo_031(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 31) }
func TestCSSPrinterAgreesWithGo_032(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 32) }
func TestCSSPrinterAgreesWithGo_033(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 33) }
func TestCSSPrinterAgreesWithGo_034(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 34) }
func TestCSSPrinterAgreesWithGo_035(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 35) }
func TestCSSPrinterAgreesWithGo_036(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 36) }
func TestCSSPrinterAgreesWithGo_037(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 37) }
func TestCSSPrinterAgreesWithGo_038(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 38) }
func TestCSSPrinterAgreesWithGo_039(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 39) }
func TestCSSPrinterAgreesWithGo_040(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 40) }
func TestCSSPrinterAgreesWithGo_041(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 41) }
func TestCSSPrinterAgreesWithGo_042(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 42) }
func TestCSSPrinterAgreesWithGo_043(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 43) }
func TestCSSPrinterAgreesWithGo_044(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 44) }
func TestCSSPrinterAgreesWithGo_045(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 45) }
func TestCSSPrinterAgreesWithGo_046(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 46) }
func TestCSSPrinterAgreesWithGo_047(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 47) }
func TestCSSPrinterAgreesWithGo_048(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 48) }
func TestCSSPrinterAgreesWithGo_049(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 49) }
func TestCSSPrinterAgreesWithGo_050(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 50) }
func TestCSSPrinterAgreesWithGo_051(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 51) }
func TestCSSPrinterAgreesWithGo_052(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 52) }
func TestCSSPrinterAgreesWithGo_053(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 53) }
func TestCSSPrinterAgreesWithGo_054(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 54) }
func TestCSSPrinterAgreesWithGo_055(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 55) }
func TestCSSPrinterAgreesWithGo_056(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 56) }
func TestCSSPrinterAgreesWithGo_057(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 57) }
func TestCSSPrinterAgreesWithGo_058(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 58) }
func TestCSSPrinterAgreesWithGo_059(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 59) }
func TestCSSPrinterAgreesWithGo_060(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 60) }
func TestCSSPrinterAgreesWithGo_061(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 61) }
func TestCSSPrinterAgreesWithGo_062(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 62) }
func TestCSSPrinterAgreesWithGo_063(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 63) }
func TestCSSPrinterAgreesWithGo_064(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 64) }
func TestCSSPrinterAgreesWithGo_065(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 65) }
func TestCSSPrinterAgreesWithGo_066(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 66) }
func TestCSSPrinterAgreesWithGo_067(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 67) }
func TestCSSPrinterAgreesWithGo_068(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 68) }
func TestCSSPrinterAgreesWithGo_069(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 69) }
func TestCSSPrinterAgreesWithGo_070(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 70) }
func TestCSSPrinterAgreesWithGo_071(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 71) }
func TestCSSPrinterAgreesWithGo_072(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 72) }
func TestCSSPrinterAgreesWithGo_073(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 73) }
func TestCSSPrinterAgreesWithGo_074(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 74) }
func TestCSSPrinterAgreesWithGo_075(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 75) }
func TestCSSPrinterAgreesWithGo_076(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 76) }
func TestCSSPrinterAgreesWithGo_077(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 77) }
func TestCSSPrinterAgreesWithGo_078(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 78) }
func TestCSSPrinterAgreesWithGo_079(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 79) }
func TestCSSPrinterAgreesWithGo_080(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 80) }
func TestCSSPrinterAgreesWithGo_081(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 81) }
func TestCSSPrinterAgreesWithGo_082(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 82) }
func TestCSSPrinterAgreesWithGo_083(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 83) }
func TestCSSPrinterAgreesWithGo_084(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 84) }
func TestCSSPrinterAgreesWithGo_085(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 85) }
func TestCSSPrinterAgreesWithGo_086(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 86) }
func TestCSSPrinterAgreesWithGo_087(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 87) }
func TestCSSPrinterAgreesWithGo_088(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 88) }
func TestCSSPrinterAgreesWithGo_089(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 89) }
func TestCSSPrinterAgreesWithGo_090(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 90) }
func TestCSSPrinterAgreesWithGo_091(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 91) }
func TestCSSPrinterAgreesWithGo_092(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 92) }
func TestCSSPrinterAgreesWithGo_093(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 93) }
func TestCSSPrinterAgreesWithGo_094(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 94) }
func TestCSSPrinterAgreesWithGo_095(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 95) }
func TestCSSPrinterAgreesWithGo_096(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 96) }
func TestCSSPrinterAgreesWithGo_097(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 97) }
func TestCSSPrinterAgreesWithGo_098(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 98) }
func TestCSSPrinterAgreesWithGo_099(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 99) }
func TestCSSPrinterAgreesWithGo_100(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 100) }
func TestCSSPrinterAgreesWithGo_101(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 101) }
func TestCSSPrinterAgreesWithGo_102(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 102) }
func TestCSSPrinterAgreesWithGo_103(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 103) }
func TestCSSPrinterAgreesWithGo_104(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 104) }
func TestCSSPrinterAgreesWithGo_105(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 105) }
func TestCSSPrinterAgreesWithGo_106(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 106) }
func TestCSSPrinterAgreesWithGo_107(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 107) }
func TestCSSPrinterAgreesWithGo_108(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 108) }
func TestCSSPrinterAgreesWithGo_109(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 109) }
func TestCSSPrinterAgreesWithGo_110(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 110) }
func TestCSSPrinterAgreesWithGo_111(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 111) }
func TestCSSPrinterAgreesWithGo_112(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 112) }
func TestCSSPrinterAgreesWithGo_113(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 113) }
func TestCSSPrinterAgreesWithGo_114(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 114) }
func TestCSSPrinterAgreesWithGo_115(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 115) }
func TestCSSPrinterAgreesWithGo_116(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 116) }
func TestCSSPrinterAgreesWithGo_117(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 117) }
func TestCSSPrinterAgreesWithGo_118(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 118) }
func TestCSSPrinterAgreesWithGo_119(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 119) }
func TestCSSPrinterAgreesWithGo_120(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 120) }
func TestCSSPrinterAgreesWithGo_121(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 121) }
func TestCSSPrinterAgreesWithGo_122(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 122) }
func TestCSSPrinterAgreesWithGo_123(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 123) }
func TestCSSPrinterAgreesWithGo_124(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 124) }
func TestCSSPrinterAgreesWithGo_125(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 125) }
func TestCSSPrinterAgreesWithGo_126(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 126) }
func TestCSSPrinterAgreesWithGo_127(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 127) }
func TestCSSPrinterAgreesWithGo_128(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 128) }
func TestCSSPrinterAgreesWithGo_129(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 129) }
func TestCSSPrinterAgreesWithGo_130(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 130) }
func TestCSSPrinterAgreesWithGo_131(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 131) }
func TestCSSPrinterAgreesWithGo_132(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 132) }
func TestCSSPrinterAgreesWithGo_133(test *testing.T) { test.Parallel(); runCSSPrinterShard(test, 133) }
