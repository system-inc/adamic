package tsprinter

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

func testMutant(t *testing.T, i int) {
	t.Helper()
	started := time.Now()
	change := mutations[i]
	cases, want := mutantOracleProduct(t, i)
	t.Logf("grain oracle setup: %.3fs", time.Since(started).Seconds())
	binary := mutantNativeProduct(t, i)
	path := mutatedPort(t, change)
	t.Logf("grain total build setup: %.3fs", time.Since(started).Seconds())
	// Only this leaf's witness selection, execution and comparisons use the
	// case deadline. A cold miss is still covered by go test's 90 s timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
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
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	node := mutantExecute(t, ctx, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...)
	native := mutantExecute(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, arguments...)
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
	if ctx.Err() != nil {
		t.Fatalf("mutant own-work deadline: %v", ctx.Err())
	}
}

// The case-only context covers both output oracles. A killed or crashed child
// is a test failure, never evidence that a mutant was caught.
func mutantExecute(t *testing.T, ctx context.Context, environment []string, name string, arguments ...string) run {
	t.Helper()
	command := exec.CommandContext(ctx, name, arguments...)
	command.WaitDelay = time.Second
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := childguard.Run(command, childguard.Options{})
	if ctx.Err() != nil {
		t.Fatalf("mutant own-work deadline: %v", ctx.Err())
	}
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}
