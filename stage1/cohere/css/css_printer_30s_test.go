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
)

// The unmutated printer runs every case in both modes, in sixty-four content-hash partitions per mode. Keep ownership
// keyed by case bytes; doubling splits each former bucket in two without changing the checks or their own-work
// deadline. At sixteen partitions, alone on one core, _000 and _016 passed the 60 s own-work deadline, _027 took 54.6 s
// and _011 49.9 s (Workshop, Oct 11), and Server panicked two leaves past it (landable-18 run 3). A unit runs one leaf
// per CPU (testguard.Share), so a leaf's wall time is its one-core work at any share: sixty-four keeps it well under
// half. No leaf prepares another leaf's executable.
const cssPrinterPartitions = 64

// Each printer mutant runs, in each mode, only over its witnesses: the cases where a full-corpus run found it differs
// from Go (cssPrinterWitnesses), the smallest cssPrinterWitnessesRun of them. Run over the rest it could fail only by crashing, which says something about the
// mutant, never about the printer: on Oct 12 those runs took 1,780 of the family's 2,928 active seconds (#testaudit).
// 128 unmutated shards, both modes by sixty-four partitions, then one witness shard per mutant and mode.
const testCSSPrinterAgreesWithGoShards = 2*cssPrinterPartitions + 2*3

var cssPrinterCorpusOnce sync.Once
var cssPrinterCorpus []string

// cssPrinterCorpusLines reads the printer's corpus from the parser's corpus product (TestProduct_CSSParserCorpus):
// the same Go oracle request over the same roots, so the same cases, which TestCSSPrinterCorpusIsTheParserCorpus
// holds byte for byte. Enumerating it live cost every printer shard's process about 3.7 s of its setup (Home).
func cssPrinterCorpusLines(t *testing.T) []string {
	t.Helper()
	cssPrinterCorpusOnce.Do(func() {
		path, _ := cachedCSSParserOracleOutputs(t, cssParserOracle(t))
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		cssPrinterCorpus = strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	})
	return cssPrinterCorpus
}

// The printer reads the parser's corpus product only because a live enumeration through the printer's own Go
// oracle gives the same bytes. Mutant, run by hand with ADAMIC_CSS_FIXTURES set (without it there's no fixture root to
// drop): drop the fixture root from the live enumeration; the corpora then differ.
func TestCSSPrinterCorpusIsTheParserCorpus(t *testing.T) {
	t.Parallel()
	live, err := os.ReadFile(cssPrinterCases(t, cssOracleProduct(t, false)))
	if err != nil {
		t.Fatal(err)
	}
	product, _ := cachedCSSParserOracleOutputs(t, cssParserOracle(t))
	declared, err := os.ReadFile(product)
	if err != nil {
		t.Fatal(err)
	}
	if difference := firstDifference(string(live), string(declared)); difference != "" {
		t.Fatalf("the printer's live corpus differs from the parser's corpus product: %s", difference)
	}
	t.Logf("one corpus, %d cases, %d bytes", strings.Count(string(live), "\n"), len(live))
}

func cssPrinterOwnDeadline(t *testing.T) func() {
	t.Helper()
	started := time.Now()
	timer := time.AfterFunc(60*time.Second, func() { panic(fmt.Sprintf("%s: own work exceeded 60s", t.Name())) })
	return func() { timer.Stop(); t.Logf("own work: %s", time.Since(started)) }
}

// cssPrinterSanitizedAgreement runs the sanitized printer once and holds it to Go's answers. On Linux leak detection is
// on in that run, so one run checks the output, ASan/UBSan and LeakSanitizer, where a second run checked leaks before:
// the slowest corpus case (a 30 KB Tailwind preflight, about 7 s sanitized, #z3p4j5q) ran twice in its shards.
// macOS checks leaks with its own run (printerLeaks). The error says which check failed: a leak, LeakSanitizer
// unable to run (a red, never a pass), a sanitizer's error, or an output that isn't Go's.
func cssPrinterSanitizedAgreement(t *testing.T, binary, expected string, arguments ...string) error {
	t.Helper()
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=1"}
	}
	return cssPrinterSanitizedVerdict(execute(t, environment, binary, arguments...), expected)
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
func TestCSSPrinterSanitizedRunNamesEachFailure(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux runs leak detection inside the sanitized run")
	}
	cases := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(cases, []byte(">Ca{color:red;margin:0.50px}\n>Sa{--x:{b:c};color:rgb(1,2,3)}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	expected := cssPrinterAnswers(t, cssOracleProduct(t, true), cases, "default")
	arguments := []string{cases, "output", "once", "default"}
	if err := cssPrinterSanitizedAgreement(t, cssPrinterProduct(t, -1).binary, expected, arguments...); err != nil {
		t.Fatalf("the unmutated printer: %v", err)
	}
	err := cssPrinterSanitizedAgreement(t, composedMemoryProduct(t, 2).sanitized, expected, arguments...)
	if err == nil || !strings.HasPrefix(err.Error(), "leak ") {
		t.Fatalf("a printer that drops its releases read as %v, want a leak", err)
	}
	t.Logf("planted leak: %.120s", err)
	err = cssPrinterSanitizedAgreement(t, composedMemoryProduct(t, 0).sanitized, expected, arguments...)
	if err == nil || !strings.HasPrefix(err.Error(), "exit ") || !strings.Contains(err.Error(), "AddressSanitizer") {
		t.Fatalf("a printer that reads freed memory read as %v, want a sanitizer's error", err)
	}
	t.Logf("planted use after free: %.120s", err)
	changed := strings.Replace(expected, "red", "blue", 1)
	if changed == expected {
		t.Fatal("no answer to change")
	}
	err = cssPrinterSanitizedAgreement(t, cssPrinterProduct(t, -1).binary, changed, arguments...)
	if err == nil || !strings.HasPrefix(err.Error(), "output differs from Go") {
		t.Fatalf("a changed answer read as %v, want an output that isn't Go's", err)
	}
	err = cssPrinterSanitizedVerdict(run{stdout: []byte(expected), stderr: []byte("==1==LeakSanitizer has encountered a fatal error.\n"), exitCode: 1}, expected)
	if err == nil || !strings.HasPrefix(err.Error(), "LeakSanitizer could not run") {
		t.Fatalf("LeakSanitizer unable to run read as %v, want a red", err)
	}
}

func runCSSPrinterShard(t *testing.T, shard int) {
	t.Helper()
	setup := time.Now()
	lines := cssPrinterCorpusLines(t)
	var unit cssModeShard
	if shard < 2*cssPrinterPartitions {
		unit = cssPrinterUnmutatedPlan(lines)[shard]
	} else {
		variant, modeIndex := (shard-2*cssPrinterPartitions)/2, (shard-2*cssPrinterPartitions)%2
		ids := cssPrinterWitnesses(t, variant)[modeIndex]
		if err := cssPrinterWitnessError(printerMutants[variant].name, cssPrinterModes[modeIndex], ids); err != nil {
			t.Fatal(err)
		}
		unit = cssPrinterMutantShard(lines, variant, modeIndex, ids)
	}
	printerOracle := cssOracleProduct(t, true)
	product := cssPrinterProduct(t, unit.mutant)
	leakBinary := product.binary
	if unit.mutant < 0 && runtime.GOOS == "darwin" {
		leakBinary = cssPrinterProduct(t, 3).binary
	}
	repo, _ := filepath.Abs(repository)
	t.Logf("setup including products: %s", time.Since(setup))
	defer cssPrinterOwnDeadline(t)()
	path := cssShardFile(t, unit.cases)
	expected := cssPrinterAnswers(t, printerOracle, path, unit.mode)
	arguments := []string{path, "output", "once", unit.mode}
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	if unit.mutant < 0 {
		if err := cssPrinterSanitizedAgreement(t, product.binary, expected, arguments...); err != nil {
			t.Fatalf("%s native ASan/UBSan: %v", unit.mode, err)
		}
		for _, side := range []struct {
			name   string
			result run
		}{
			{"Node", onNode(t, product.source, arguments...)},
			{"JavaScript backend", onNode(t, product.javascript, arguments...)},
		} {
			if err := cssAgreement(side.result, expected); err != nil {
				t.Fatalf("%s %s: %v", unit.mode, side.name, err)
			}
		}
		if runtime.GOOS != "linux" {
			if report := printerLeaks(t, leakBinary, arguments...); report != "" {
				t.Fatal(report)
			}
		}
		if library := os.Getenv("ADAMIC_CSS_PRINTER_LIBRARY"); library != "" {
			comparePrinterLibrary(t, path, expected, library, unit.mode, "npm")
			comparePrinterLibrary(t, path, expected, filepath.Join(repo, "cohere", "internal", "format", "prettier", "bundles"), unit.mode, "fork")
		} else {
			t.Log("optional library comparisons unavailable: set ADAMIC_CSS_PRINTER_LIBRARY to scratch Prettier 3.9.6")
		}
		t.Logf("%s: %d cases; three backends, sanitizers and leaks checked", unit.mode, len(unit.cases))
	} else {
		mutation := printerMutants[unit.mutant]
		for _, side := range []struct {
			name   string
			result run
		}{
			{"native ASan/UBSan", execute(t, environment, product.binary, arguments...)},
			{"Node", onNode(t, product.source, arguments...)},
		} {
			if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
				t.Fatalf("%s %s mutant %s must run: %d %s", unit.mode, side.name, mutation.name, side.result.exitCode, side.result.stderr)
			}
			differing, err := cssPrinterDifferingCases(string(side.result.stdout), expected)
			if err != nil {
				t.Fatalf("%s %s printer mutant %s: %v", unit.mode, side.name, mutation.name, err)
			}
			if len(differing) != len(unit.cases) {
				t.Fatalf("%s %s printer mutant %s differs from Go on %d of its %d chosen witnesses", unit.mode, side.name, mutation.name, len(differing), len(unit.cases))
			}
			t.Logf("%s %s caught %s on all %d chosen witnesses: %s", unit.mode, side.name, mutation.name, len(unit.cases), firstDifference(string(side.result.stdout), expected))
		}
		t.Logf("%s mutant %s: %d chosen witnesses; sanitized native and Node checked", unit.mode, mutation.name, len(unit.cases))
	}
}

func TestCSSPrinterAgreesWithGoUnion(t *testing.T) {
	t.Parallel()
	lines := cssPrinterCorpusLines(t)
	witnesses := cssPrinterAllWitnesses(t)
	defer cssPrinterOwnDeadline(t)()
	plan := cssModePlan(lines, witnesses)
	if err := cssModeUnion(lines, plan, witnesses); err != nil {
		t.Fatal(err)
	}
	for offset, unit := range plan[2*cssPrinterPartitions:] {
		t.Logf("mutant %s, %s: %d witnesses, its shard runs the %d smallest", printerMutants[unit.mutant].name, unit.mode, len(witnesses[unit.mutant][offset%2]), len(unit.cases))
	}
	t.Logf("counted union: %d cases x 2 modes unmutated in %d shards, and %d mutant witness shards; %d shards", len(lines), 2*cssPrinterPartitions, len(plan)-2*cssPrinterPartitions, len(plan))
}

func TestCSSPrinterAgreesWithGo_000(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 0) }
func TestCSSPrinterAgreesWithGo_001(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 1) }
func TestCSSPrinterAgreesWithGo_002(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 2) }
func TestCSSPrinterAgreesWithGo_003(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 3) }
func TestCSSPrinterAgreesWithGo_004(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 4) }
func TestCSSPrinterAgreesWithGo_005(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 5) }
func TestCSSPrinterAgreesWithGo_006(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 6) }
func TestCSSPrinterAgreesWithGo_007(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 7) }
func TestCSSPrinterAgreesWithGo_008(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 8) }
func TestCSSPrinterAgreesWithGo_009(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 9) }
func TestCSSPrinterAgreesWithGo_010(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 10) }
func TestCSSPrinterAgreesWithGo_011(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 11) }
func TestCSSPrinterAgreesWithGo_012(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 12) }
func TestCSSPrinterAgreesWithGo_013(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 13) }
func TestCSSPrinterAgreesWithGo_014(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 14) }
func TestCSSPrinterAgreesWithGo_015(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 15) }
func TestCSSPrinterAgreesWithGo_016(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 16) }
func TestCSSPrinterAgreesWithGo_017(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 17) }
func TestCSSPrinterAgreesWithGo_018(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 18) }
func TestCSSPrinterAgreesWithGo_019(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 19) }
func TestCSSPrinterAgreesWithGo_020(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 20) }
func TestCSSPrinterAgreesWithGo_021(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 21) }
func TestCSSPrinterAgreesWithGo_022(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 22) }
func TestCSSPrinterAgreesWithGo_023(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 23) }
func TestCSSPrinterAgreesWithGo_024(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 24) }
func TestCSSPrinterAgreesWithGo_025(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 25) }
func TestCSSPrinterAgreesWithGo_026(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 26) }
func TestCSSPrinterAgreesWithGo_027(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 27) }
func TestCSSPrinterAgreesWithGo_028(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 28) }
func TestCSSPrinterAgreesWithGo_029(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 29) }
func TestCSSPrinterAgreesWithGo_030(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 30) }
func TestCSSPrinterAgreesWithGo_031(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 31) }
func TestCSSPrinterAgreesWithGo_032(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 32) }
func TestCSSPrinterAgreesWithGo_033(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 33) }
func TestCSSPrinterAgreesWithGo_034(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 34) }
func TestCSSPrinterAgreesWithGo_035(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 35) }
func TestCSSPrinterAgreesWithGo_036(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 36) }
func TestCSSPrinterAgreesWithGo_037(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 37) }
func TestCSSPrinterAgreesWithGo_038(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 38) }
func TestCSSPrinterAgreesWithGo_039(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 39) }
func TestCSSPrinterAgreesWithGo_040(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 40) }
func TestCSSPrinterAgreesWithGo_041(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 41) }
func TestCSSPrinterAgreesWithGo_042(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 42) }
func TestCSSPrinterAgreesWithGo_043(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 43) }
func TestCSSPrinterAgreesWithGo_044(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 44) }
func TestCSSPrinterAgreesWithGo_045(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 45) }
func TestCSSPrinterAgreesWithGo_046(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 46) }
func TestCSSPrinterAgreesWithGo_047(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 47) }
func TestCSSPrinterAgreesWithGo_048(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 48) }
func TestCSSPrinterAgreesWithGo_049(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 49) }
func TestCSSPrinterAgreesWithGo_050(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 50) }
func TestCSSPrinterAgreesWithGo_051(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 51) }
func TestCSSPrinterAgreesWithGo_052(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 52) }
func TestCSSPrinterAgreesWithGo_053(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 53) }
func TestCSSPrinterAgreesWithGo_054(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 54) }
func TestCSSPrinterAgreesWithGo_055(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 55) }
func TestCSSPrinterAgreesWithGo_056(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 56) }
func TestCSSPrinterAgreesWithGo_057(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 57) }
func TestCSSPrinterAgreesWithGo_058(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 58) }
func TestCSSPrinterAgreesWithGo_059(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 59) }
func TestCSSPrinterAgreesWithGo_060(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 60) }
func TestCSSPrinterAgreesWithGo_061(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 61) }
func TestCSSPrinterAgreesWithGo_062(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 62) }
func TestCSSPrinterAgreesWithGo_063(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 63) }
func TestCSSPrinterAgreesWithGo_064(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 64) }
func TestCSSPrinterAgreesWithGo_065(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 65) }
func TestCSSPrinterAgreesWithGo_066(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 66) }
func TestCSSPrinterAgreesWithGo_067(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 67) }
func TestCSSPrinterAgreesWithGo_068(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 68) }
func TestCSSPrinterAgreesWithGo_069(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 69) }
func TestCSSPrinterAgreesWithGo_070(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 70) }
func TestCSSPrinterAgreesWithGo_071(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 71) }
func TestCSSPrinterAgreesWithGo_072(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 72) }
func TestCSSPrinterAgreesWithGo_073(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 73) }
func TestCSSPrinterAgreesWithGo_074(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 74) }
func TestCSSPrinterAgreesWithGo_075(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 75) }
func TestCSSPrinterAgreesWithGo_076(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 76) }
func TestCSSPrinterAgreesWithGo_077(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 77) }
func TestCSSPrinterAgreesWithGo_078(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 78) }
func TestCSSPrinterAgreesWithGo_079(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 79) }
func TestCSSPrinterAgreesWithGo_080(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 80) }
func TestCSSPrinterAgreesWithGo_081(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 81) }
func TestCSSPrinterAgreesWithGo_082(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 82) }
func TestCSSPrinterAgreesWithGo_083(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 83) }
func TestCSSPrinterAgreesWithGo_084(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 84) }
func TestCSSPrinterAgreesWithGo_085(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 85) }
func TestCSSPrinterAgreesWithGo_086(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 86) }
func TestCSSPrinterAgreesWithGo_087(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 87) }
func TestCSSPrinterAgreesWithGo_088(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 88) }
func TestCSSPrinterAgreesWithGo_089(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 89) }
func TestCSSPrinterAgreesWithGo_090(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 90) }
func TestCSSPrinterAgreesWithGo_091(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 91) }
func TestCSSPrinterAgreesWithGo_092(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 92) }
func TestCSSPrinterAgreesWithGo_093(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 93) }
func TestCSSPrinterAgreesWithGo_094(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 94) }
func TestCSSPrinterAgreesWithGo_095(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 95) }
func TestCSSPrinterAgreesWithGo_096(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 96) }
func TestCSSPrinterAgreesWithGo_097(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 97) }
func TestCSSPrinterAgreesWithGo_098(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 98) }
func TestCSSPrinterAgreesWithGo_099(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 99) }
func TestCSSPrinterAgreesWithGo_100(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 100) }
func TestCSSPrinterAgreesWithGo_101(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 101) }
func TestCSSPrinterAgreesWithGo_102(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 102) }
func TestCSSPrinterAgreesWithGo_103(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 103) }
func TestCSSPrinterAgreesWithGo_104(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 104) }
func TestCSSPrinterAgreesWithGo_105(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 105) }
func TestCSSPrinterAgreesWithGo_106(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 106) }
func TestCSSPrinterAgreesWithGo_107(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 107) }
func TestCSSPrinterAgreesWithGo_108(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 108) }
func TestCSSPrinterAgreesWithGo_109(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 109) }
func TestCSSPrinterAgreesWithGo_110(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 110) }
func TestCSSPrinterAgreesWithGo_111(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 111) }
func TestCSSPrinterAgreesWithGo_112(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 112) }
func TestCSSPrinterAgreesWithGo_113(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 113) }
func TestCSSPrinterAgreesWithGo_114(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 114) }
func TestCSSPrinterAgreesWithGo_115(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 115) }
func TestCSSPrinterAgreesWithGo_116(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 116) }
func TestCSSPrinterAgreesWithGo_117(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 117) }
func TestCSSPrinterAgreesWithGo_118(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 118) }
func TestCSSPrinterAgreesWithGo_119(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 119) }
func TestCSSPrinterAgreesWithGo_120(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 120) }
func TestCSSPrinterAgreesWithGo_121(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 121) }
func TestCSSPrinterAgreesWithGo_122(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 122) }
func TestCSSPrinterAgreesWithGo_123(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 123) }
func TestCSSPrinterAgreesWithGo_124(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 124) }
func TestCSSPrinterAgreesWithGo_125(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 125) }
func TestCSSPrinterAgreesWithGo_126(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 126) }
func TestCSSPrinterAgreesWithGo_127(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 127) }
func TestCSSPrinterAgreesWithGo_128(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 128) }
func TestCSSPrinterAgreesWithGo_129(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 129) }
func TestCSSPrinterAgreesWithGo_130(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 130) }
func TestCSSPrinterAgreesWithGo_131(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 131) }
func TestCSSPrinterAgreesWithGo_132(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 132) }
func TestCSSPrinterAgreesWithGo_133(t *testing.T) { t.Parallel(); runCSSPrinterShard(t, 133) }
