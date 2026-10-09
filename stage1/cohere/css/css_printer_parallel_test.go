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

// Eight independent mode/check groups, each with eight content-hash partitions.
// Every original case is checked in both modes against all original backends
// and all three mutants. No leaf prepares another leaf's executable.
const testCSSPrinterAgreesWithGoShards = 64

var cssPrinterCorpusOnce sync.Once
var cssPrinterCorpus []string

func cssPrinterCorpusLines(t *testing.T) []string {
	t.Helper()
	cssPrinterCorpusOnce.Do(func() {
		path := cssPrinterCases(t, cssOracleProduct(t, false))
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		cssPrinterCorpus = strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	})
	return cssPrinterCorpus
}

func cssPrinterUnitHasWitness(unit cssModeShard, count, witness int) bool {
	for _, c := range unit.cases {
		if c.id%count == witness {
			return true
		}
	}
	return false
}

func cssPrinterOwnDeadline(t *testing.T) func() {
	t.Helper()
	started := time.Now()
	timer := time.AfterFunc(60*time.Second, func() { panic(fmt.Sprintf("%s: own work exceeded 60s", t.Name())) })
	return func() { timer.Stop(); t.Logf("own work: %s", time.Since(started)) }
}

func runCSSPrinterShard(t *testing.T, shard int) {
	t.Helper()
	setup := time.Now()
	lines := cssPrinterCorpusLines(t)
	unit := cssModePlan(lines, testCSSPrinterAgreesWithGoShards)[shard]
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
		for _, side := range []struct {
			name   string
			result run
		}{
			{"native ASan/UBSan", execute(t, environment, product.binary, arguments...)},
			{"Node", onNode(t, product.source, arguments...)},
			{"JavaScript backend", onNode(t, product.javascript, arguments...)},
		} {
			if err := cssAgreement(side.result, expected); err != nil {
				t.Fatalf("%s %s: %v", unit.mode, side.name, err)
			}
		}
		if report := printerLeaks(t, leakBinary, arguments...); report != "" {
			t.Fatal(report)
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
			difference := firstDifference(string(side.result.stdout), expected)
			if cssPrinterUnitHasWitness(unit, len(lines), []int{0, 2, 12}[unit.mutant]) {
				if difference == "" {
					t.Fatalf("%s %s printer mutant %s survived", unit.mode, side.name, mutation.name)
				}
				t.Logf("%s %s caught %s: %s", unit.mode, side.name, mutation.name, difference)
			}
		}
		t.Logf("%s mutant %s: %d cases; sanitized native and Node checked", unit.mode, mutation.name, len(unit.cases))
	}
}

func TestCSSPrinterAgreesWithGoUnion(t *testing.T) {
	t.Parallel()
	lines := cssPrinterCorpusLines(t)
	defer cssPrinterOwnDeadline(t)()
	plan := cssModePlan(lines, testCSSPrinterAgreesWithGoShards)
	if err := cssModeUnion(lines, plan, testCSSPrinterAgreesWithGoShards); err != nil {
		t.Fatal(err)
	}
	if len(printerMutants) != 3 {
		t.Fatal("update mode enumeration when adding a mutant")
	}
	for variant := 0; variant < 3; variant++ {
		for mode := 0; mode < 2; mode++ {
			witnesses := 0
			for _, unit := range plan {
				if unit.mutant == variant && unit.mode == []string{"default", "narrow"}[mode] && cssPrinterUnitHasWitness(unit, len(lines), []int{0, 2, 12}[variant]) {
					witnesses++
				}
			}
			if witnesses != 1 {
				t.Fatalf("mutant %d mode %d: witness owned %d times", variant, mode, witnesses)
			}
		}
	}
	t.Logf("counted union: %d cases x 2 modes x 4 check groups = %d checks in %d shards", len(lines), len(lines)*8, len(plan))
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
