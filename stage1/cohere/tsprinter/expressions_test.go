package tsprinter

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

func expressionGateShard(test *testing.T, number int) {
	test.Helper()
	if expressionGateProofEnabled() {
		expressionGateProofShard(test, number)
		return
	}
	if !expressionSelection(test, oracles.ExpressionsAgainstGoAndPrettierShards)[number] {
		test.Skip("assigned to another box")
	}
	setup := time.Now()
	units := oracles.ExpressionGateUnitsOnce(test)
	plan, answers := units.Plan, units.Answers
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		test.Skip("set ADAMIC_TS_PRETTIER to prettier@3.9.6")
	}
	products := printerGateFamilyProducts(test, "expressions")
	port := products.source
	script, _ := filepath.Abs("testdata/expressions.mjs")
	var shard oracles.CorpusShard
	var directory string
	var oracleAnswers [2][]byte
	if number < len(plan.Shards) {
		shard = units.Unit(test, number)
		// npm and embedded Prettier never read Adamic's output: a shard's answers are theirs over its cases, so they
		// are products built ahead (TestProduct_TSPrinterExpressionsNPM_ and _Embedded_), as statements' are, and the
		// shard compares their bytes as it compared the live runs. The two live passes were at least 55% of this
		// family's ~1,292 CPU seconds a run (test audit).
		libraryHash := oracles.StatementDirectoryHash(test, library)
		for side := 0; side < 2; side++ {
			answers, err := os.ReadFile(filepath.Join(
				oracles.ExpressionLibraryOracle(test, number, side, shard, libraryHash, library),
				"answers.txt",
			))
			if err != nil {
				test.Fatal(err)
			}
			oracleAnswers[side] = answers
		}
	} else {
		directory = units.Gaps(test)
	}
	test.Logf("setup: %.3fs", time.Since(setup).Seconds())
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	started := time.Now()
	defer func() { test.Logf("own work: %.3fs", time.Since(started).Seconds()) }()
	runCommand := func(environment []string, name string, args ...string) oracles.Run {
		return tscAgreementExecute(test, ctx, environment, name, args...)
	}
	node := func(path string, args ...string) oracles.Run {
		runner, err := filepath.Abs(filepath.Join(oracles.Repository, "oracle/node.mjs"))
		if err != nil {
			test.Fatal(err)
		}
		return runCommand(
			nil,
			"node",
			append([]string{"--disable-warning=ExperimentalWarning", runner, path}, args...)...,
		)
	}
	if number < len(plan.Shards) {
		check := func(name string, result oracles.Run) {
			if err := expressionDisagreement(plan, number, answers[number], result); err != nil {
				test.Fatalf("%s: %v", name, err)
			}
		}
		check("Node", node(port, "--cases", shard.Text, "80"))
		check(
			"native",
			runCommand([]string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized, "--cases", shard.Text, "80"),
		)
		check("backend", node(products.backend, "--cases", shard.Text, "80"))
		if runtime.GOOS == "darwin" {
			report := runCommand(nil, "leaks", "--atExit", "--", products.release, "--cases", shard.Text, "80")
			if report.ExitCode != 0 {
				test.Fatalf("leaks: exit %d stdout %s stderr %s", report.ExitCode, report.Stdout, report.Stderr)
			}
		} else if runtime.GOOS == "linux" {
			check(
				"leaks",
				runCommand([]string{"ASAN_OPTIONS=detect_leaks=1"}, products.sanitized, "--cases", shard.Text, "80"),
			)
		} else {
			test.Fatalf("no leak check for %s", runtime.GOOS)
		}
		check("native release", runCommand(nil, products.release, "--cases", shard.Text, "80"))
		comparePrinterLibraryCases(
			test,
			"Prettier",
			oracles.Run{Stdout: oracleAnswers[0]},
			shard.Specs,
			"expressions",
			false,
			false,
		)
		comparePrinterLibraryCases(
			test,
			"embedded Prettier",
			oracles.Run{Stdout: oracleAnswers[1]},
			shard.Specs,
			"expressions",
			true,
			false,
		)
		return
	}
	for _, side := range []struct {
		name   string
		result oracles.Run
	}{
		{"Node", node(port, "--cases", directory+"/gaps.txt", "80")},
		{"native", runCommand(
			[]string{"ASAN_OPTIONS=detect_leaks=1"},
			products.sanitized,
			"--cases",
			directory+"/gaps.txt",
			"80",
		)},
		{"backend", node(products.backend, "--cases", directory+"/gaps.txt", "80")},
	} {
		if side.result.ExitCode != 0 ||
			len(side.result.Stderr) != 0 ||
			string(side.result.Stdout) != string(units.GapWant) {
			test.Fatalf(
				"%s loud gaps: exit %d stderr %s diff %s",
				side.name,
				side.result.ExitCode,
				side.result.Stderr,
				firstDifference(string(side.result.Stdout), string(units.GapWant)),
			)
		}
	}
	result := runCommand(nil, "node", script, library, directory+"/gaps.json")
	if result.ExitCode != 0 || len(result.Stderr) != 0 {
		test.Fatalf("Go/Prettier gap proof: %s", result.Stderr)
	}
}
