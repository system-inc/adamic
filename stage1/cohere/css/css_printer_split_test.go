package css

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

const testCSSPrinterAgreesWithGoShards = 512

// ADAMIC_TEST_SHARD=i/n runs only shard i of n; unset runs every shard.
// Each mode/check group has 64 deterministic byte-balanced case ranges. The
// agreement group retains all three backends, sanitizer, leak and library
// comparisons. Each of the three mutant groups retains native and Node checks
// over every case. Cases 0, 2 and 12 retain their original kill witnesses in
// the corresponding range of each mode's mutant group. No check is sampled.
func TestCSSPrinterAgreesWithGo(t *testing.T) {
	setup := cssMeasurement(t)
	rawOracle := cssOracleProduct(t, false)
	printerOracle := cssOracleProduct(t, true)
	cases := cssPrinterCases(t, rawOracle)
	directory := portDirectory(t, nil)
	product := cssPrinterProduct(t, filepath.Join(directory, "print_main.ts"), "sanitized printer and lowered backends", true)
	leakBinary := product.binary
	if runtime.GOOS == "darwin" {
		leakBinary = cssPrinterProduct(t, product.source, "leak printer", false).binary
	}
	mutated := make([]cssExecutable, len(printerMutants))
	for index, mutation := range printerMutants {
		dir := portDirectory(t, &mutation)
		mutated[index] = cssPrinterProduct(t, filepath.Join(dir, "print_main.ts"), "mutant "+mutation.name, true)
	}
	shards := cssPrinterModeShards(t, cases, testCSSPrinterAgreesWithGoShards)
	if len(shards) < 13 || len(shards[12].cases) == 0 {
		t.Fatal("shards must retain the three original mutant witnesses")
	}
	start, end := cssSelected(t, testCSSPrinterAgreesWithGoShards)
	repo, _ := filepath.Abs(repository)
	setup()
	for shard := start; shard < end; shard++ {
		t.Run(fmt.Sprintf("shard-%03d", shard), func(t *testing.T) {
			t.Parallel()
			unit := shards[shard]
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
					t.Skip("set ADAMIC_CSS_PRINTER_LIBRARY to scratch Prettier 3.9.6")
				}
				t.Logf("%s: %d cases; three backends, sanitizers, leaks and libraries checked", unit.mode, len(unit.cases))
			} else {
				mutation := printerMutants[unit.mutant]
				for _, side := range []struct {
					name   string
					result run
				}{
					{"native ASan/UBSan", execute(t, environment, mutated[unit.mutant].binary, arguments...)},
					{"Node", onNode(t, mutated[unit.mutant].source, arguments...)},
				} {
					if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
						t.Fatalf("%s %s mutant %s must run: %d %s", unit.mode, side.name, mutation.name, side.result.exitCode, side.result.stderr)
					}
					difference := firstDifference(string(side.result.stdout), expected)
					if shard%(testCSSPrinterAgreesWithGoShards/8) == []int{0, 2, 12}[unit.mutant] {
						if difference == "" {
							t.Fatalf("%s %s printer mutant %s survived", unit.mode, side.name, mutation.name)
						}
						t.Logf("%s %s caught %s: %s", unit.mode, side.name, mutation.name, difference)
					}
				}
				t.Logf("%s mutant %s: %d cases; sanitized native and Node checked", unit.mode, mutation.name, len(unit.cases))
			}
		})
	}
}
