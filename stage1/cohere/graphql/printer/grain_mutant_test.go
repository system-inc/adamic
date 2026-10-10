package printer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
)

// Every entry point prepares its own prerequisites before its work timer. The
// once values hold buildcache products, never paths owned by another test's cleanup.
type printerMutantProductOnce struct {
	once sync.Once
	path string
}

func (p *printerMutantProductOnce) get(t *testing.T, build func() string) string {
	t.Helper()
	p.once.Do(func() { p.path = build() })
	if p.path == "" {
		t.Fatal("shared GraphQL product preparation failed")
	}
	return p.path
}

var mutantSources, mutantLowered, mutantSanitized [3]printerMutantProductOnce
var mutantOracle printerMutantProductOnce
var mutantCorpus struct {
	once       sync.Once
	path, want string
}

func printerMutantSource(t *testing.T, number int) string {
	t.Helper()
	return mutantSources[number].get(t, func() string {
		mutation := printerMutations[number]
		files := []string{"../token.ts", "../characterClasses.ts", "../blockString.ts", "../lexer.ts", "../parser.ts", "../../json/width.ts", "../../json/widthTables.ts", "doc.ts", "printer.ts", "main.ts", "printer_test.go", "grain_mutant_test.go"}
		return printerBuild(t, printerBuildInputs{
			Name:      fmt.Sprintf("GraphQL mutant %03d source", number),
			Files:     printerInputFiles(t, files...),
			Flags:     []string{mutation.file, mutation.from, mutation.to},
			Toolchain: runtime.Version(),
		}, func(dir string) error {
			printerDirectoryAt(t, dir, mutation.file, mutation.from, mutation.to)
			return nil
		}) + "/graphql/printer/main.ts"
	})
}

func printerMutantLowered(t *testing.T, number int) string {
	t.Helper()
	return mutantLowered[number].get(t, func() string {
		return printerLoweredProduct(t, printerMutantSource(t, number))
	})
}

func printerMutantSanitized(t *testing.T, number int) string {
	t.Helper()
	return mutantSanitized[number].get(t, func() string {
		return printerCompiledProduct(t, printerMutantLowered(t, number), native.Options{Sanitize: true})
	})
}

func printerMutantProducts(t *testing.T, number int) printerProducts {
	t.Helper()
	return printerProducts{source: printerMutantSource(t, number), sanitized: printerMutantSanitized(t, number)}
}

func printerMutantOracle(t *testing.T) string {
	t.Helper()
	return mutantOracle.get(t, func() string {
		// GoBuild is not available yet. Hash the entire pinned Cohere tree rather
		// than a hand-selected dependency list, plus the overlay and workspace.
		files := printerInputFiles(t, filepath.Join(repository, "cohere"), "testdata/cohere_side_test.go", "../testdata/cohere_side_test.go", "shards_test.go", "grain_mutant_test.go", filepath.Join(repository, "go.mod"), filepath.Join(repository, "go.work"))
		for _, name := range []string{"go.sum", "go.work.sum"} {
			path := filepath.Join(repository, name)
			if _, err := os.Stat(path); err == nil {
				files = append(files, printerInputFiles(t, path)...)
			} else if !os.IsNotExist(err) {
				t.Fatal(err)
			}
		}

		return printerBuild(t, printerBuildInputs{
			Name: "GraphQL mutant Go oracle", Files: files,
			Flags:     []string{"go test -c -trimpath -ldflags=-buildid= -overlay ./internal/format/graphql"},
			Toolchain: buildcache.Tool("go", "env", "-json", "GOVERSION", "GOOS", "GOARCH", "GOEXPERIMENT", "GOFLAGS", "CGO_ENABLED", "GOTOOLCHAIN", "GOAMD64", "GOARM64", "GOARM", "GO386", "GOMIPS", "GOMIPS64", "GOPPC64", "GORISCV64", "GOWASM"),
		}, func(dir string) error {
			binary := printerOracle(t)
			data, err := os.ReadFile(binary)
			if err != nil {
				return err
			}
			return os.WriteFile(dir+"/oracle", data, 0755)
		}) + "/oracle"
	})
}

func printerMutantCorpus(t *testing.T) (string, string) {
	t.Helper()
	mutantCorpus.once.Do(func() { mutantCorpus.path, mutantCorpus.want = printerCases(t, "defaults", printerMutantOracle(t)) })
	if mutantCorpus.path == "" {
		t.Fatal("shared GraphQL corpus preparation failed")
	}
	return mutantCorpus.path, mutantCorpus.want
}

func TestPrinterMutants_000(t *testing.T) { t.Parallel(); printerMutantUnit(t, 0) }
func TestPrinterMutants_001(t *testing.T) { t.Parallel(); printerMutantUnit(t, 1) }
func TestPrinterMutants_002(t *testing.T) { t.Parallel(); printerMutantUnit(t, 2) }
func TestPrinterMutantUnion(t *testing.T) { t.Parallel(); printerMutantUnit(t, -1) }

func TestProduct_GraphQLMutantOracle(t *testing.T)       { t.Parallel(); printerMutantOracle(t) }
func TestProduct_GraphQLMutantDefaults(t *testing.T)     { t.Parallel(); printerMutantCorpus(t) }
func TestProduct_GraphQLMutantSource000(t *testing.T)    { t.Parallel(); printerMutantSource(t, 0) }
func TestProduct_GraphQLMutantSource001(t *testing.T)    { t.Parallel(); printerMutantSource(t, 1) }
func TestProduct_GraphQLMutantSource002(t *testing.T)    { t.Parallel(); printerMutantSource(t, 2) }
func TestProduct_GraphQLMutantLowered000(t *testing.T)   { t.Parallel(); printerMutantLowered(t, 0) }
func TestProduct_GraphQLMutantLowered001(t *testing.T)   { t.Parallel(); printerMutantLowered(t, 1) }
func TestProduct_GraphQLMutantLowered002(t *testing.T)   { t.Parallel(); printerMutantLowered(t, 2) }
func TestProduct_GraphQLMutantSanitized000(t *testing.T) { t.Parallel(); printerMutantSanitized(t, 0) }
func TestProduct_GraphQLMutantSanitized001(t *testing.T) { t.Parallel(); printerMutantSanitized(t, 1) }
func TestProduct_GraphQLMutantSanitized002(t *testing.T) { t.Parallel(); printerMutantSanitized(t, 2) }

const testPrinterMutantsShards = 3

var printerMutations = [...]struct{ name, file, from, to string }{
	{"line width ignored", "doc.ts", "this.settings.printWidth - column", "100000 - column"},
	{"end of line comment leads next node", "printer.ts", "if(own && around.following >= 0)", "if((own || end) && around.following >= 0)"},
	{"block string indentation discarded", "printer.ts", "for(const line of lines) parts.push(documents.text(line));", "for(const line of lines) parts.push(documents.text(line.trim()));"},
}

// Each fixed mutant owns one top-level TestPrinterMutants_NNN and checks the corpus on both
// original sides. ADAMIC_TEST_SHARD=i/n selects indices modulo n equal to i;
// unset runs all. Build inputs are prepared before case timing.
func printerMutantUnit(t *testing.T, unit int) {
	setupStart := time.Now()
	if len(printerMutations) != testPrinterMutantsShards {
		t.Fatalf("enumerated %d shards, declared %d", len(printerMutations), testPrinterMutantsShards)
	}
	cases, want := printerMutantCorpus(t)
	enumeration := enumeratePrinter(t, "defaults", cases, want)
	var whole []printerCase
	for number := range printerMutations {
		whole = append(whole, printerMutantCases(number, enumeration)...)
	}
	shards := make([]printerShard, len(printerMutations))
	products := make([]printerProducts, len(printerMutations))
	for number := range printerMutations {
		shards[number] = printerShard{mode: "defaults", path: cases, cases: printerMutantCases(number, enumeration)}
		if number == unit {
			products[number] = printerMutantProducts(t, number)
		}
	}
	if len(shards) != testPrinterMutantsShards {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), testPrinterMutantsShards)
	}
	if err := printerShardUnion(whole, shards); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d unique mutant/case ids across %d shards (%d cases per mutant)", len(whole), len(shards), len(enumeration))
	if unit < 0 {
		return
	}
	selected, err := printerShardSelection(os.Getenv("ADAMIC_TEST_SHARD"), len(shards))
	if err != nil {
		t.Fatal(err)
	}
	for number, mutation := range printerMutations {
		if number != unit || !selected[number] {
			continue
		}
		product := products[number]
		{
			t.Logf("setup wall %.3fs", time.Since(setupStart).Seconds())
			start := time.Now()
			t.Cleanup(func() {
				t.Logf("own work wall %.3fs", time.Since(start).Seconds())
				if elapsed := time.Since(start); elapsed > 30*time.Second {
					t.Errorf("invalid test unit: %.3fs exceeds 30s", elapsed.Seconds())
				}
			})
			t.Logf("mutant %s, cases 0..%d", mutation.name, len(shards[number].cases)-1)
			for _, side := range []struct {
				name   string
				result run
			}{
				{"native", execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, product.sanitized, "--cases", cases)},
				{"Node", onNode(t, product.source, "--cases", cases)},
			} {
				difference, err := printerMutantDisagreement(number, side.name, side.result, want)
				if err != nil {
					t.Error(err)
				} else {
					t.Logf("%s caught: %s", side.name, difference)
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
		return "", fmt.Errorf("shard-%03d %s mutant must run: exit %d, %s", number, side, result.exitCode, result.stderr)
	}
	difference := firstDifference(string(result.stdout), want)
	if difference == "" {
		return "", fmt.Errorf("shard-%03d %s mutant escaped oracle", number, side)
	}
	return difference, nil
}

// A real process emits an unchanged answer for one planted surviving mutant;
// exactly its owning shard must reject it through the production comparison.
func TestPrinterMutantPlantedSurvivor(t *testing.T) {
	t.Parallel()
	caught := 0
	for number := range printerMutations {
		answer := "ok\tmutated\n"
		if number == 1 {
			answer = "ok\toriginal\n"
		}
		encoded, _ := json.Marshal(answer)
		result := execute(t, nil, "node", "-e", "process.stdout.write("+string(encoded)+")")
		_, err := printerMutantDisagreement(number, "Node", result, "ok\toriginal\n")
		if err == nil {
			if number == 1 {
				t.Fatal("planted survivor escaped shard-001")
			}
			continue
		}
		if number != 1 || !strings.Contains(err.Error(), "shard-001 Node mutant escaped oracle") {
			t.Fatalf("wrong owner: %v", err)
		}
		t.Log(err)
		caught++
	}
	if caught != 1 {
		t.Fatalf("%d shards caught planted survivor, want 1", caught)
	}
}
