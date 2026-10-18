package tsprinter

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

// ADAMIC_TEST_SHARD=i/n runs the shards whose number modulo n is i;
// unset runs all shards. The mixed live repository/pinned TypeScript corpus uses
// a fixed 16 shards and SHA-256(relative path, file/mode-local index) modulo 16.
// Build products are prepared once and shared by leaves. statementReadyShared
// calls this once per process, after the partition and the ADAMIC_TS_PRETTIER skip.
func statementPrepareCommon(t *testing.T, shards *oracles.StatementGateShards) statementSharedProducts {
	t.Helper()
	setup := time.Now()
	cpu := oracles.StatementCPU()
	t.Cleanup(func() { t.Logf("shared setup CPU including builds: %.3fs", oracles.StatementCPU()-cpu) })
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	port, _ := filepath.Abs("statementsMain.ts")
	loweredDir := printerGateLowered(t, "statements")
	binary, release := printerGateNative(t, "statements", true)+"/port", printerGateNative(t, "statements", false)+"/port"
	if keep := os.Getenv("ADAMIC_TS_STATEMENT_ARTIFACTS"); keep != "" {
		if err := os.MkdirAll(keep, 0755); err != nil {
			t.Fatal(err)
		}
		for name, source := range map[string]string{"port.c": filepath.Join(loweredDir, "port.c"), "port": release} {
			data, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			mode := os.FileMode(0644)
			if name == "port" {
				mode = 0755
			}
			if err := os.WriteFile(filepath.Join(keep, name), data, mode); err != nil {
				t.Fatal(err)
			}
		}
	}
	libraryHash := oracles.StatementLibraryHashOnce(t)
	elapsed := time.Since(setup).Seconds()
	t.Logf("shared setup: %.3fs, %d cases in %d shards", elapsed, strings.Count(shards.Want, "\n"), len(shards.Shards))
	return statementSharedProducts{Cases: shards.Cases, Want: shards.Want, Specs: shards.Specs, Port: port, LoweredDir: loweredDir, Binary: binary, Release: release, Library: library, LibraryHash: libraryHash, Elapsed: elapsed, Shards: shards}
}

func statementAgainstGoAndPrettierShard(t *testing.T, shardNumber int) {
	t.Helper()
	if !statementShardSelection(t)(shardNumber) {
		return
	}
	products := statementReadyShared(t)
	setup := time.Now()
	cpu := oracles.StatementCPU()
	t.Cleanup(func() { t.Logf("shard CPU: %.3fs", oracles.StatementCPU()-cpu) })
	// The partition is the process's; only this shard's own files are written here.
	shard := products.Shards.Shard(t, shardNumber)
	port := products.Port
	binary, release, library := products.Binary, products.Release, products.Library
	loweredDir := products.LoweredDir

	gapInput := ">const x:number=1;\n>const {x}=value;\n>export const x=1;\n>if(x)f();\n>#!/usr/bin/env node\\nf();\n"
	gapWant := "notyet\tvariable-types\nnotyet\tvariable-pattern\nnotyet\tvariable-modifiers\nnotyet\tIfStatement\nnotyet\tcomment-attachment\n"
	gapPath := filepath.Join(t.TempDir(), "statement-gaps.txt")
	if err := os.WriteFile(gapPath, []byte(gapInput), 0644); err != nil {
		t.Fatal(err)
	}
	backend := filepath.Join(loweredDir, "program.mjs")
	libraryHash := products.LibraryHash
	var oracleAnswers [2][]byte
	for side := 0; side < 2; side++ {
		directory := oracles.StatementLibraryOracle(t, shardNumber, side, shard, libraryHash, library)
		answers, err := os.ReadFile(filepath.Join(directory, "answers.txt"))
		if err != nil {
			t.Fatal(err)
		}
		oracleAnswers[side] = answers
	}
	t.Logf("shard-local input preparation: %.3fs; union: %d cases in %d shards", time.Since(setup).Seconds(), strings.Count(products.Want, "\n"), len(products.Shards.Shards))
	caseStart := time.Now()
	caseContext, cancelCases := context.WithTimeout(context.Background(), 60*time.Second)
	oracles.StatementCaseContexts.Store(t, caseContext)
	t.Cleanup(func() {
		cancelCases()
		oracles.StatementCaseContexts.Delete(t)
		t.Logf("case time after shared setup: %.3fs", time.Since(caseStart).Seconds())
	})
	compare := func(name string, result oracles.Run) {
		if err := statementDisagreement(name, result, shard.Want, shard.Labels); err != nil {
			t.Fatal(err)
		}
	}
	compare("Node", statementOnNode(t, port, "--cases", shard.Text, "80"))
	compare("native", oracles.StatementExecute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, "--cases", shard.Text, "80"))
	compare("backend", statementOnNode(t, backend, "--cases", shard.Text, "80"))
	for _, side := range []struct {
		name   string
		result oracles.Run
	}{
		{"Node gaps", statementOnNode(t, port, "--cases", gapPath, "80")},
		{"native gaps", oracles.StatementExecute(t, nil, binary, "--cases", gapPath, "80")},
		{"backend gaps", statementOnNode(t, backend, "--cases", gapPath, "80")},
	} {
		if err := statementDisagreement(side.name, side.result, gapWant, nil); err != nil {
			t.Fatal(err)
		}
	}
	if report := statementLeaks(t, binary, release, "--cases", shard.Text, "80"); report != "" {
		t.Fatal(report)
	}
	statementPrinterLibrary(t, "npm Prettier", oracles.Run{Stdout: oracleAnswers[0]}, shard.Specs, false)
	statementPrinterLibrary(t, "embedded Prettier", oracles.Run{Stdout: oracleAnswers[1]}, shard.Specs, true)
	compare("release", oracles.StatementExecute(t, nil, release, "--cases", shard.Text, "80"))
	t.Logf("%d statement/program fragments byte-identical", len(shard.Indices))
}
