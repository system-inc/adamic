package printer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/corpusfiles"
)

func printerCases(t *testing.T, mode string, oracle ...string) (string, string) {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	// Mode and mutant shards use this fixed corpus: Cohere commit
	// 7945d102a6c18dd36adf9114a758ce646e8b2359 plus generated controls
	// from cohere_side_test.go, whose random seed is 20261006. Compare
	// live case totals in the union; guard the pinned file count separately.
	files := corpusfiles.Upstream(t, filepath.Join(root, "cohere"), corpusfiles.CohereCommit,
		[]string{"internal/format/graphql"}, []string{"*_test.go"})
	if len(files) != 3 {
		t.Fatalf("pinned GraphQL corpus: %d files, want 3", len(files))
	}
	t.Log("GraphQL documents: no tracked .graphql corpus; pinned upstream test constants and generated controls")
	if len(oracle) == 0 {
		oracle = []string{printerOracle(t)}
	}
	directory := printerBuild(t, printerBuildInputs{
		Name:  "GraphQL printer oracle outputs",
		Files: []string{oracle[0], root + "/stage1/cohere/graphql/printer/printer_test.go", root + "/stage1/cohere/graphql/printer/shards_test.go"},
		Flags: []string{"mode=" + mode, "test-run=TestAdamicPrinter"}, Toolchain: runtime.Version(),
	}, func(directory string) error {
		request, err := json.Marshal(map[string]any{"Sources": []string(nil), "Cases": directory + "/cases.txt", "Answers": directory + "/answers.txt", "Mode": mode, "Coverage": directory + "/coverage.json"})
		if err != nil {
			return err
		}
		path := filepath.Join(t.TempDir(), "request.json")
		if err := os.WriteFile(path, request, 0644); err != nil {
			return err
		}
		command := bounded(t, oracle[0], "-test.run=^TestAdamicPrinter$", "-test.count=1", "-test.timeout=0")
		command.Dir = root + "/cohere"
		command.Env = append(os.Environ(), "ADAMIC_PRINTER_REQUEST="+path)
		if output, err := combinedOutput(command); err != nil {
			return fmt.Errorf("Go oracle: %w\n%s", err, output)
		}
		return nil
	})
	cases, answers := directory+"/cases.txt", directory+"/answers.txt"
	data, err := os.ReadFile(answers)
	if err != nil {
		t.Fatal(err)
	}
	if keep := os.Getenv("ADAMIC_GRAPHQL_PRINTER_KEEP"); keep != "" {
		for _, name := range []string{"cases.txt", "answers.txt", "coverage.json"} {
			data, err := os.ReadFile(filepath.Join(directory, name))
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(keep, name), data, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return cases, string(data)
}

const testPrinterUpstreamPreflightShards = 4

// ADAMIC_TEST_SHARD=i/n selects indices modulo n equal to i; unset runs all.
// The four fixed option modes run as shard-NNN and share their Go oracle build.
func TestPrinterUpstreamPreflight(t *testing.T) { printerUpstreamShards(t) }

func printerDirectory(t *testing.T, file, from, to string) string {
	t.Helper()
	directory := t.TempDir()
	for _, group := range []struct {
		source, target string
		files          []string
	}{
		{"..", "graphql", []string{"token.ts", "characterClasses.ts", "blockString.ts", "lexer.ts", "parser.ts"}},
		{"../../json", "json", []string{"width.ts", "widthTables.ts"}},
		{".", "graphql/printer", []string{"doc.ts", "printer.ts", "main.ts"}},
	} {
		target := filepath.Join(directory, group.target)
		if err := os.MkdirAll(target, 0755); err != nil {
			t.Fatal(err)
		}
		for _, name := range group.files {
			data, err := os.ReadFile(filepath.Join(group.source, name))
			if err != nil {
				t.Fatal(err)
			}
			if group.source == "." && file == name {
				if strings.Count(string(data), from) != 1 {
					t.Fatalf("mutant must match once: %s", file)
				}
				data = []byte(strings.Replace(string(data), from, to, 1))
			}
			if err = os.WriteFile(filepath.Join(target, name), data, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return filepath.Join(directory, "graphql/printer/main.ts")
}

const testPrinterMutantsShards = 3

var printerMutations = [...]struct{ name, file, from, to string }{
	{"line width ignored", "doc.ts", "this.settings.printWidth - column", "100000 - column"},
	{"end of line comment leads next node", "printer.ts", "if(own && around.following >= 0)", "if((own || end) && around.following >= 0)"},
	{"block string indentation discarded", "printer.ts", "for(const line of lines) parts.push(documents.text(line));", "for(const line of lines) parts.push(documents.text(line.trim()));"},
}

// Each fixed mutant owns one shard-NNN and checks the entire corpus on both
// original sides. ADAMIC_TEST_SHARD=i/n selects indices modulo n equal to i;
// unset runs all. Builds are shared inputs prepared before t.Parallel.
func TestPrinterMutants(t *testing.T) {
	if len(printerMutations) != testPrinterMutantsShards {
		t.Fatalf("enumerated %d shards, declared %d", len(printerMutations), testPrinterMutantsShards)
	}
	cases, want := printerCases(t, "defaults", printerOracle(t))
	enumeration := enumeratePrinter(t, "defaults", cases, want)
	var whole []printerCase
	for number := range printerMutations {
		whole = append(whole, printerMutantCases(number, enumeration)...)
	}
	shards := make([]printerShard, len(printerMutations))
	products := make([]printerProducts, len(printerMutations))
	for number, mutation := range printerMutations {
		shards[number] = printerShard{mode: "defaults", path: cases, cases: printerMutantCases(number, enumeration)}
		path := printerDirectory(t, mutation.file, mutation.from, mutation.to)
		products[number] = preparePrinterProducts(t, path)
	}
	if len(shards) != testPrinterMutantsShards {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), testPrinterMutantsShards)
	}
	if err := printerShardUnion(whole, shards); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d unique mutant/case ids across %d shards (%d cases per mutant)", len(whole), len(shards), len(enumeration))
	selected, err := printerShardSelection(os.Getenv("ADAMIC_TEST_SHARD"), len(shards))
	if err != nil {
		t.Fatal(err)
	}
	for number, mutation := range printerMutations {
		if !selected[number] {
			continue
		}
		product := products[number]
		t.Run(fmt.Sprintf("shard-%03d", number), func(t *testing.T) {
			t.Parallel()
			start := time.Now()
			t.Cleanup(func() {
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
		})
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

// The same binary's ordinary file driver is held to the batch oracle too.
func TestPrinterFileDriver(t *testing.T) {
	t.Parallel()
	source := filepath.Join(t.TempDir(), "probe.graphql")
	if err := os.WriteFile(source, []byte("query{hello(a:1,b:2)}"), 0644); err != nil {
		t.Fatal(err)
	}
	path, _ := filepath.Abs("main.ts")
	products := preparePrinterProducts(t, path)
	want := "query {\n    hello(a: 1, b: 2)\n}\n"
	for _, side := range []run{onNode(t, path, source), execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized, source)} {
		if side.exitCode != 0 || string(side.stdout) != want {
			t.Fatalf("driver: exit %d, stdout %q, stderr %q", side.exitCode, side.stdout, side.stderr)
		}
	}

}
