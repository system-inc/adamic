package printer

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

// Every entry point prepares its own prerequisites before its work timer. The
// once values hold buildcache products, never paths owned by another test's cleanup.
type printerMutantProductOnce struct {
	once sync.Once
	path string
}

func (p *printerMutantProductOnce) get(test *testing.T, build func() string) string {
	test.Helper()
	p.once.Do(func() { p.path = build() })
	if p.path == "" {
		test.Fatal("shared GraphQL product preparation failed")
	}
	return p.path
}

var mutantSources, mutantLowered, mutantSanitized [3]printerMutantProductOnce

func printerMutantSource(test *testing.T, number int) string {
	test.Helper()
	return mutantSources[number].get(test, func() string {
		mutation := printerMutations[number]
		files := []string{
			"../token.ts",
			"../characterClasses.ts",
			"../blockString.ts",
			"../lexer.ts",
			"../parser.ts",
			"../../json/width.ts",
			"../../json/widthTables.ts",
			"doc.ts",
			"printer.ts",
			"main.ts",
			"printer_test.go",
			"grain_mutant_test.go",
		}
		return printerBuild(test, printerBuildInputs{
			Name:      fmt.Sprintf("GraphQL mutant %03d source", number),
			Files:     printerInputFiles(test, files...),
			Flags:     []string{mutation.file, mutation.from, mutation.to},
			Toolchain: runtime.Version(),
		}, func(dir string) error {
			printerDirectoryAt(test, dir, mutation.file, mutation.from, mutation.to)
			return nil
		}) + "/graphql/printer/main.ts"
	})
}

func printerMutantLowered(test *testing.T, number int) string {
	test.Helper()
	return mutantLowered[number].get(test, func() string {
		return printerLoweredProduct(test, printerMutantSource(test, number))
	})
}

func printerMutantSanitized(test *testing.T, number int) string {
	test.Helper()
	return mutantSanitized[number].get(test, func() string {
		return printerCompiledProduct(test, printerMutantLowered(test, number), native.Options{Sanitize: true})
	})
}

func printerMutantProducts(test *testing.T, number int) printerProducts {
	test.Helper()
	return printerProducts{source: printerMutantSource(test, number), sanitized: printerMutantSanitized(test, number)}
}

func TestPrinterMutants_000(test *testing.T) { test.Parallel(); printerMutantUnit(test, 0) }
func TestPrinterMutants_001(test *testing.T) { test.Parallel(); printerMutantUnit(test, 1) }
func TestPrinterMutants_002(test *testing.T) { test.Parallel(); printerMutantUnit(test, 2) }
func TestPrinterMutantUnion(test *testing.T) { test.Parallel(); printerMutantUnit(test, -1) }

// The mutants read the defaults corpus every other test reads (printerCases), which TestProduct_GraphQLPrinterDefaults
// declares, from the one Go oracle (printerOracle).
func TestProduct_GraphQLMutantSource000(test *testing.T) {
	test.Parallel()
	printerMutantSource(test, 0)
}
func TestProduct_GraphQLMutantSource001(test *testing.T) {
	test.Parallel()
	printerMutantSource(test, 1)
}
func TestProduct_GraphQLMutantSource002(test *testing.T) {
	test.Parallel()
	printerMutantSource(test, 2)
}
func TestProduct_GraphQLMutantLowered000(test *testing.T) {
	test.Parallel()
	printerMutantLowered(test, 0)
}
func TestProduct_GraphQLMutantLowered001(test *testing.T) {
	test.Parallel()
	printerMutantLowered(test, 1)
}
func TestProduct_GraphQLMutantLowered002(test *testing.T) {
	test.Parallel()
	printerMutantLowered(test, 2)
}
func TestProduct_GraphQLMutantSanitized000(test *testing.T) {
	test.Parallel()
	printerMutantSanitized(test, 0)
}
func TestProduct_GraphQLMutantSanitized001(test *testing.T) {
	test.Parallel()
	printerMutantSanitized(test, 1)
}
func TestProduct_GraphQLMutantSanitized002(test *testing.T) {
	test.Parallel()
	printerMutantSanitized(test, 2)
}

const testPrinterMutantsShards = 3

var printerMutations = [...]struct{ name, file, from, to string }{
	{"line width ignored", "doc.ts", "this.settings.printWidth - column", "100000 - column"},
	{
		"end of line comment leads next node",
		"printer.ts",
		"if(own && around.following >= 0)",
		"if((own || end) && around.following >= 0)",
	},
	{
		"block string indentation discarded",
		"printer.ts",
		"for(const line of lines) parts.push(documents.text(line));",
		"for(const line of lines) parts.push(documents.text(line.trim()));",
	},
}

// Each fixed mutant owns one top-level TestPrinterMutants_NNN and checks the corpus on both
// original sides. ADAMIC_TEST_SHARD=i/n selects indices modulo n equal to i;
// unset runs all. Build inputs are prepared before case timing.
func printerMutantUnit(test *testing.T, unit int) {
	setupStart := time.Now()
	if len(printerMutations) != testPrinterMutantsShards {
		test.Fatalf("enumerated %d shards, declared %d", len(printerMutations), testPrinterMutantsShards)
	}
	cases, want := printerCases(test, "defaults")
	enumeration := enumeratePrinter(test, "defaults", cases, want)
	var whole []printerCase
	for number := range printerMutations {
		whole = append(whole, printerMutantCases(number, enumeration)...)
	}
	shards := make([]printerShard, len(printerMutations))
	products := make([]printerProducts, len(printerMutations))
	for number := range printerMutations {
		shards[number] = printerShard{mode: "defaults", path: cases, cases: printerMutantCases(number, enumeration)}
		if number == unit {
			products[number] = printerMutantProducts(test, number)
		}
	}
	if len(shards) != testPrinterMutantsShards {
		test.Fatalf("enumerated %d shards, declared %d", len(shards), testPrinterMutantsShards)
	}
	if printerShardUnionError := printerShardUnion(whole, shards); printerShardUnionError != nil {
		test.Fatal(printerShardUnionError)
	}
	test.Logf(
		"union: %d unique mutant/case ids across %d shards (%d cases per mutant)",
		len(whole),
		len(shards),
		len(enumeration),
	)
	if unit < 0 {
		return
	}
	selected, printerShardSelectionError := printerShardSelection(os.Getenv("ADAMIC_TEST_SHARD"), len(shards))
	if printerShardSelectionError != nil {
		test.Fatal(printerShardSelectionError)
	}
	for number, mutation := range printerMutations {
		if number != unit || !selected[number] {
			continue
		}
		product := products[number]
		{
			test.Logf("setup wall %.3fs", time.Since(setupStart).Seconds())
			start := time.Now()
			test.Cleanup(func() {
				test.Logf("own work wall %.3fs", time.Since(start).Seconds())
				if elapsed := time.Since(start); elapsed > 30*time.Second {
					test.Errorf("invalid test unit: %.3fs exceeds 30s", elapsed.Seconds())
				}
			})
			test.Logf("mutant %s, cases 0..%d", mutation.name, len(shards[number].cases)-1)
			for _, side := range []struct {
				name   string
				result run
			}{
				{"native", execute(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, product.sanitized, "--cases", cases)},
				{"Node", onNode(test, product.source, "--cases", cases)},
			} {
				difference, printerMutantDisagreementError := printerMutantDisagreement(
					number,
					side.name,
					side.result,
					want,
				)
				if printerMutantDisagreementError != nil {
					test.Error(printerMutantDisagreementError)
				} else {
					test.Logf("%s caught: %s", side.name, difference)
				}
			}
		}
	}
}

func printerMutantCases(number int, enumeration []printerCase) []printerCase {
	owned := make([]printerCase, len(enumeration))
	for i, item := range enumeration {
		item.id = fmt.Sprintf("mutant-%03d/%s", number, item.id)
		owned[i] = item
	}
	return owned
}

func printerMutantDisagreement(number int, side string, result run, want string) (string, error) {
	if result.exitCode != 0 || len(result.stderr) != 0 {
		return "", fmt.Errorf(
			"shard-%03d %s mutant must run: exit %d, %s",
			number,
			side,
			result.exitCode,
			result.stderr,
		)
	}
	difference := firstDifference(string(result.stdout), want)
	if difference == "" {
		return "", fmt.Errorf("shard-%03d %s mutant escaped oracle", number, side)
	}
	return difference, nil
}

// A real process emits an unchanged answer for one planted surviving mutant;
// exactly its owning shard must reject it through the production comparison.
func TestPrinterMutantPlantedSurvivor(test *testing.T) {
	test.Parallel()
	caught := 0
	for number := range printerMutations {
		answer := "ok\tmutated\n"
		if number == 1 {
			answer = "ok\toriginal\n"
		}
		encoded, _ := json.Marshal(answer)
		result := execute(test, nil, "node", "-e", "process.stdout.write("+string(encoded)+")")
		_, printerMutantDisagreementError := printerMutantDisagreement(number, "Node", result, "ok\toriginal\n")
		if printerMutantDisagreementError == nil {
			if number == 1 {
				test.Fatal("planted survivor escaped shard-001")
			}
			continue
		}
		if number != 1 ||
			!strings.Contains(printerMutantDisagreementError.Error(), "shard-001 Node mutant escaped oracle") {
			test.Fatalf("wrong owner: %v", printerMutantDisagreementError)
		}
		test.Log(printerMutantDisagreementError)
		caught++
	}
	if caught != 1 {
		test.Fatalf("%d shards caught planted survivor, want 1", caught)
	}
}
