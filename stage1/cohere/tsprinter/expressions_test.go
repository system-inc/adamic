package tsprinter

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

func expressionCorpus(t *testing.T, processPlan ...bool) (string, string, string) {
	t.Helper()
	root, _ := filepath.Abs(repository)
	files := printerCorpusFiles(t)
	gaps, _ := filepath.Abs("testdata/notyet.json")
	directory := tsPrinterOracleOutputs(t, root, files, gaps, "expressions", printerGateGoOracle(t)+"/oracle")
	answers, err := os.ReadFile(directory + "/answers.txt")
	if err != nil {
		t.Fatal(err)
	}
	if keep := os.Getenv("ADAMIC_TS_PRINTER_KEEP"); keep != "" {
		if err := os.MkdirAll(keep, 0755); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"coverage.json", "cases.json", "cases.txt", "answers.txt", "gaps.txt", "gap-answers.txt", "gaps.json"} {
			data, err := os.ReadFile(directory + "/" + name)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(keep+"/"+name, data, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if len(processPlan) == 0 || processPlan[0] {
		planCorpus(t, directory+"/cases.txt")
	}
	return directory + "/cases.txt", string(answers), directory + "/cases.json"
}

const testExpressionsAgainstGoAndPrettierShards = 65

func expressionGateShard(t *testing.T, number int) {
	t.Helper()
	if expressionGateProofEnabled() {
		expressionGateProofShard(t, number)
		return
	}
	if !expressionSelection(t, testExpressionsAgainstGoAndPrettierShards)[number] {
		t.Skip("assigned to another box")
	}
	setup := time.Now()
	units := expressionGateUnitsOnce(t)
	plan, answers := units.plan, units.answers
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		t.Skip("set ADAMIC_TS_PRETTIER to prettier@3.9.6")
	}
	products := printerGateFamilyProducts(t, "expressions")
	port := products.source
	script, _ := filepath.Abs("testdata/expressions.mjs")
	var shard corpusShard
	var directory string
	var oracleAnswers [2][]byte
	if number < len(plan.shards) {
		shard = units.unit(t, number)
		// npm and embedded Prettier never read Adamic's output: a shard's answers are theirs over its cases, so they are
		// products built ahead (TestProduct_TSPrinterExpressionsNPM_ and _Embedded_), as statements' are, and the shard
		// compares their bytes as it compared the live runs. The two live passes were at least 55% of this family's
		// ~1,292 CPU seconds a run (test audit).
		libraryHash := statementDirectoryHash(t, library)
		for side := 0; side < 2; side++ {
			answers, err := os.ReadFile(filepath.Join(expressionLibraryOracle(t, number, side, shard, libraryHash, library), "answers.txt"))
			if err != nil {
				t.Fatal(err)
			}
			oracleAnswers[side] = answers
		}
	} else {
		directory = units.gaps(t)
	}
	t.Logf("setup: %.3fs", time.Since(setup).Seconds())
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	started := time.Now()
	defer func() { t.Logf("own work: %.3fs", time.Since(started).Seconds()) }()
	runCommand := func(environment []string, name string, args ...string) run {
		return tscAgreementExecute(t, ctx, environment, name, args...)
	}
	node := func(path string, args ...string) run {
		runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
		if err != nil {
			t.Fatal(err)
		}
		return runCommand(nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, args...)...)
	}
	if number < len(plan.shards) {
		check := func(name string, result run) {
			if err := expressionDisagreement(plan, number, answers[number], result); err != nil {
				t.Fatalf("%s: %v", name, err)
			}
		}
		check("Node", node(port, "--cases", shard.text, "80"))
		check("native", runCommand([]string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized, "--cases", shard.text, "80"))
		check("backend", node(products.backend, "--cases", shard.text, "80"))
		if runtime.GOOS == "darwin" {
			report := runCommand(nil, "leaks", "--atExit", "--", products.release, "--cases", shard.text, "80")
			if report.exitCode != 0 {
				t.Fatalf("leaks: exit %d stdout %s stderr %s", report.exitCode, report.stdout, report.stderr)
			}
		} else if runtime.GOOS == "linux" {
			check("leaks", runCommand([]string{"ASAN_OPTIONS=detect_leaks=1"}, products.sanitized, "--cases", shard.text, "80"))
		} else {
			t.Fatalf("no leak check for %s", runtime.GOOS)
		}
		check("native release", runCommand(nil, products.release, "--cases", shard.text, "80"))
		comparePrinterLibraryCases(t, "Prettier", run{stdout: oracleAnswers[0]}, shard.specs, "expressions", false, false)
		comparePrinterLibraryCases(t, "embedded Prettier", run{stdout: oracleAnswers[1]}, shard.specs, "expressions", true, false)
		return
	}
	for _, side := range []struct {
		name   string
		result run
	}{
		{"Node", node(port, "--cases", directory+"/gaps.txt", "80")},
		{"native", runCommand([]string{"ASAN_OPTIONS=detect_leaks=1"}, products.sanitized, "--cases", directory+"/gaps.txt", "80")},
		{"backend", node(products.backend, "--cases", directory+"/gaps.txt", "80")},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != string(units.gapWant) {
			t.Fatalf("%s loud gaps: exit %d stderr %s diff %s", side.name, side.result.exitCode, side.result.stderr, firstDifference(string(side.result.stdout), string(units.gapWant)))
		}
	}
	result := runCommand(nil, "node", script, library, directory+"/gaps.json")
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("Go/Prettier gap proof: %s", result.stderr)
	}
}

// expressionLibraryOracle is a case shard's npm (side 0) or embedded (side 1) Prettier answers, the product
// statementLibraryOracle makes for a statement shard and keyed the same way. The gap unit's proof still runs live.
func expressionLibraryOracle(t *testing.T, number, side int, shard corpusShard, libraryHash, library string) string {
	t.Helper()
	name, script, files, oracleFlags := "npm", "testdata/expressions.mjs", []string{"stage1/cohere/tsprinter/testdata/expressions.mjs", "stage1/cohere/tsprinter/testdata/tsc-upstream-differences.json", "stage1/cohere/tsprinter/testdata/prettier-differences.json"}, []string{"npm bytes sha256=" + libraryHash}
	if side == 1 {
		name, script = "embedded", "testdata/embedded.mjs"
		library, _ = filepath.Abs(filepath.Join(repository, "cohere/internal/format/prettier/bundles"))
		files = []string{"stage1/cohere/tsprinter/testdata/embedded.mjs", "stage1/cohere/tsprinter/testdata/tsc-upstream-differences.json", "cohere/internal/format/prettier/bundles"}
		oracleFlags = nil
	}
	script, _ = filepath.Abs(script)
	return printerGateOnce(t, fmt.Sprintf("expression-library-%d-%d", number, side), func() string {
		// The cases name their files by absolute path, so the hash reads them as the tree spells them; the npm library is
		// keyed by its bytes, the embedded one by Files, neither by where it is, and the test binary isn't an input, all as
		// statementLibraryOracle says why (#t37sw0f, landable-10).
		root, _ := filepath.Abs(repository)
		upstream, _ := filepath.Abs(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"))
		specs, err := os.ReadFile(shard.specs)
		if err != nil {
			t.Fatal(err)
		}
		cases := strings.ReplaceAll(strings.ReplaceAll(string(specs), upstream+"/", "typescript/"), root+"/", "repo/")
		flags := append(oracleFlags, "cases sha256="+statementBytesHash([]byte(cases)), "library="+name, "script="+script, "width=80", "NODE_OPTIONS="+os.Getenv("NODE_OPTIONS"))
		return statementProduct(t, buildcache.Inputs{
			Name:  fmt.Sprintf("tsprinter-expressions-%s-oracle-shard-%03d-v1", name, number),
			Files: files, Flags: flags,
			Toolchain: []string{runtime.GOOS, runtime.GOARCH, buildcache.Tool("node", "--version")},
		}, func(dir string) error {
			result := statementExecute(t, nil, "node", script, library, shard.specs)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				return fmt.Errorf("%s oracle exit %d: %s", name, result.exitCode, result.stderr)
			}
			return os.WriteFile(filepath.Join(dir, "answers.txt"), result.stdout, 0644)
		})
	})
}

// Inputs describe non-Go read-only shared products; Go builds stay separate.
type printerBuildInputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

func expressionBuild(t *testing.T, inputs printerBuildInputs, build func(dir string) error) string {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	key := expressionKey(t, inputs)
	digest, err := buildcache.Key(root, key)
	if err != nil {
		t.Fatal(err)
	}
	return printerGateOnce(t, "expression-product-"+digest, func() string {
		return buildcache.Product(t, key, build)
	})
}

// expressionKey is the key expressionBuild builds inputs under: each repository file by its path, and a generated input
// outside the repository by its bytes.
func expressionKey(t *testing.T, inputs printerBuildInputs) buildcache.Inputs {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	key := buildcache.Inputs{Name: inputs.Name, Flags: append([]string{}, inputs.Flags...), Toolchain: []string{inputs.Toolchain, runtime.Version(), runtime.GOOS, runtime.GOARCH}}
	for _, file := range inputs.Files {
		relative, err := filepath.Rel(root, file)
		if err != nil {
			t.Fatal(err)
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			// Generated inputs live in a read-only product outside the repository.
			// Its bytes, rather than the machine-specific cache path, identify it.
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			key.Flags = append(key.Flags, fmt.Sprintf("generated:%s:%x", filepath.Base(file), sha256.Sum256(data)))
		} else {
			key.Files = append(key.Files, filepath.ToSlash(relative))
		}
	}
	key.Flags = append(key.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	return key
}
