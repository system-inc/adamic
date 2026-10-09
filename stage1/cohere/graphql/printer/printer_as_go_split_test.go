package printer

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"
)

const testPrinterAsGoCohereShards = 4

// The four fixed option modes own every case in their mode. ADAMIC_TEST_SHARD=i/n
// selects shard indices modulo n equal to i; unset runs all. The gate can instead
// select TestPrinterAsGoCohere_NNN directly. Products are prepared before case timing.
func printerAsGoUnit(t *testing.T, unit int) {
	oracle := printerOracle(t)
	path := printerDirectory(t, "", "", "")
	products := preparePrinterProducts(t, path)
	var whole []printerCase
	var shards []printerShard
	for _, mode := range []string{"defaults", "narrow", "tight", "tabs"} {
		cases, want := printerCases(t, mode, oracle)
		enumeration := enumeratePrinter(t, mode, cases, want)
		whole = append(whole, enumeration...)
		shards = append(shards, printerShard{mode: mode, path: cases, cases: enumeration})
	}
	if len(shards) != testPrinterAsGoCohereShards {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), testPrinterAsGoCohereShards)
	}
	if err := printerShardUnion(whole, shards); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d unique mode/case ids across %d shards", len(whole), len(shards))
	if unit < 0 {
		return
	}
	selected, err := printerShardSelection(os.Getenv("ADAMIC_TEST_SHARD"), len(shards))
	if err != nil {
		t.Fatal(err)
	}
	for number, shard := range shards {
		if number != unit || !selected[number] {
			continue
		}
		{
			start := time.Now()
			t.Cleanup(func() {
				if elapsed := time.Since(start); elapsed > 30*time.Second {
					t.Errorf("invalid test unit: %.3fs exceeds 30s", elapsed.Seconds())
				}
			})
			t.Logf("mode %s, cases 0..%d", shard.mode, len(shard.cases)-1)
			check := func(name string, result run) {
				if err := printerShardDisagreement(number, shard, result); err != nil {
					t.Errorf("%s: %v", name, err)
				}
			}
			args := []string{"--cases", shard.path, shard.mode}
			check("Node", onNode(t, products.source, args...))
			check("native", execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized, args...))
			check("JS backend", onNode(t, products.backend, args...))
			switch runtime.GOOS {
			case "linux":
				check("leaks", execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, products.sanitized, args...))
			case "darwin":
				report := execute(t, nil, "leaks", append([]string{"--atExit", "--", products.release}, args...)...)
				if report.exitCode != 0 {
					t.Errorf("leaks: exit %d stdout %s stderr %s", report.exitCode, report.stdout, report.stderr)
				}
			default:
				t.Fatalf("no leak check for %s", runtime.GOOS)
			}
		}
	}
}

func TestPrinterAsGoCohere_Setup(t *testing.T) { t.Parallel(); printerAsGoUnit(t, -1) }
func TestPrinterAsGoCohere_000(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 0) }
func TestPrinterAsGoCohere_001(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 1) }
func TestPrinterAsGoCohere_002(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 2) }
func TestPrinterAsGoCohere_003(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 3) }
