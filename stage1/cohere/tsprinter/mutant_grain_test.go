package tsprinter

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testMutant(t *testing.T, i int) {
	t.Helper()
	started := time.Now()
	change := mutations[i]
	// Setup belongs to this leaf so its measured cost includes the build.
	// Selecting one shard never prepares any other mutant or corpus.
	var cases, want string
	switch change.entry {
	case "docMain.ts":
		cases, want, _ = documentCorpus(t)
	default:
		cases, want = mutantOracle(t, change)
	}
	t.Logf("grain oracle setup: %.3fs", time.Since(started).Seconds())
	path := mutatedPort(t, change)
	program := lowered(t, path)
	t.Logf("grain oracle and lower setup: %.3fs", time.Since(started).Seconds())
	binary := nativeBinary(t, program, t.TempDir())
	t.Logf("grain total build setup: %.3fs", time.Since(started).Seconds())
	ownStarted := time.Now()
	defer func() { t.Logf("grain own work: %.3fs", time.Since(ownStarted).Seconds()) }()
	arguments := []string{cases}
	if change.entry == "main.ts" {
		arguments = []string{"--cases", cases, "80"}
	}
	if change.entry == "statementsMain.ts" {
		arguments = []string{"--cases", cases, "80"}
	}
	if change.name == "hashbang loses its refusal" {
		cases = filepath.Join(filepath.Dir(cases), "gaps.txt")
		gapWant, err := os.ReadFile(filepath.Join(filepath.Dir(cases), "gap-answers.txt"))
		if err != nil {
			t.Fatal(err)
		}
		want = string(gapWant)
		arguments = []string{"--cases", cases, "80"}
	}
	if change.entry != "docMain.ts" && change.name != "hashbang loses its refusal" {
		cases, want = mutantCorpus(t, change, cases, want)
		arguments = []string{"--cases", cases, "80"}
	}
	t.Logf("mutant %d/%d: %s", i+1, testMutantsShards, change.name)
	node := onNode(t, path, arguments...)
	native := execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, arguments...)
	for _, side := range []struct {
		name   string
		result run
	}{{"Node", node}, {"native", native}} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
			t.Fatalf("%s mutant must finish normally: exit %d stderr %s", side.name, side.result.exitCode, side.result.stderr)
		}
		if string(side.result.stdout) == want {
			t.Fatalf("%s mutant escaped the byte comparison", side.name)
		}
		difference := firstDifference(string(side.result.stdout), want)
		if filepath.Base(cases) != "gaps.txt" && filepath.Base(cases) != "witnesses.txt" {
			difference = corpusDifference(t, cases, string(side.result.stdout), want)
		}
		t.Logf("%s caught by successful-run output mismatch: %s", side.name, difference)
	}
}
