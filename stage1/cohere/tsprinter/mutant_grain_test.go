package tsprinter

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"

	"github.com/system-inc/adamic/internal/childguard"
)

func testMutant(test *testing.T, index int) {
	test.Helper()
	started := time.Now()
	change := oracles.Mutations[index]
	cases, want := oracles.MutantOracleProduct(test, index)
	test.Logf("grain oracle setup: %.3fs", time.Since(started).Seconds())
	binary := mutantNativeProduct(test, index)
	path := mutatedPort(test, change)
	test.Logf("grain total build setup: %.3fs", time.Since(started).Seconds())
	// Only this leaf's witness selection, execution and comparisons use the
	// case deadline. A cold miss is still covered by go test's 90 s timeout.
	timeoutContext, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	ownStarted := time.Now()
	defer func() { test.Logf("grain own work: %.3fs", time.Since(ownStarted).Seconds()) }()
	arguments := []string{cases}
	if change.Entry == "main.ts" {
		arguments = []string{"--cases", cases, "80"}
	}
	if change.Entry == "statementsMain.ts" {
		arguments = []string{"--cases", cases, "80"}
	}
	if change.Name == "hashbang loses its refusal" {
		cases = filepath.Join(filepath.Dir(cases), "gaps.txt")
		gapWant, readError := os.ReadFile(filepath.Join(filepath.Dir(cases), "gap-answers.txt"))
		if readError != nil {
			test.Fatal(readError)
		}
		want = string(gapWant)
		arguments = []string{"--cases", cases, "80"}
	}
	if change.Entry != "docMain.ts" && change.Name != "hashbang loses its refusal" {
		cases, want = mutantCorpus(test, change, cases, want)
		arguments = []string{"--cases", cases, "80"}
	}
	test.Logf("mutant %d/%d: %s", index+1, oracles.MutantsShards, change.Name)
	runner, absolutePathError := filepath.Abs(filepath.Join(oracles.Repository, "oracle", "node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	node := mutantExecute(
		test,
		timeoutContext,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...,
	)
	native := mutantExecute(test, timeoutContext, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, arguments...)
	for _, side := range []struct {
		name   string
		result oracles.Run
	}{{"Node", node}, {"native", native}} {
		if side.result.ExitCode != 0 || len(side.result.Stderr) != 0 {
			test.Fatalf(
				"%s mutant must finish normally: exit %d stderr %s",
				side.name,
				side.result.ExitCode,
				side.result.Stderr,
			)
		}
		if string(side.result.Stdout) == want {
			test.Fatalf("%s mutant escaped the byte comparison", side.name)
		}
		difference := firstDifference(string(side.result.Stdout), want)
		if filepath.Base(cases) != "gaps.txt" && filepath.Base(cases) != "witnesses.txt" {
			difference = corpusDifference(test, cases, string(side.result.Stdout), want)
		}
		test.Logf("%s caught by successful-run output mismatch: %s", side.name, difference)
	}
	if timeoutContext.Err() != nil {
		test.Fatalf("mutant own-work deadline: %v", timeoutContext.Err())
	}
}

// The case-only context covers both output oracles. A killed or crashed child
// is a test failure, never evidence that a mutant was caught. Output is read
// through childguard.Capture, to its end with no timer deciding when: exec's own
// copying under a one-second WaitDelay reported exec.ErrWaitDelay whenever the
// reads hadn't reached EOF a second after the child exited, so a Node or native
// mutant run that had finished could fail its shard under load (#n3gk22h's
// shape). Capture has no stall guard, and the runs lose nothing by it:
// childguard.Run's windows (two minutes of silence, thirty before the first
// byte) are longer than the case's 60 s deadline, which ends a hung run sooner.
func mutantExecute(
	test *testing.T,
	callerContext context.Context,
	environment []string,
	name string,
	arguments ...string,
) oracles.Run {
	test.Helper()
	command := exec.CommandContext(callerContext, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	stdout, stderr, captureError := childguard.Capture(callerContext, command)
	if callerContext.Err() != nil {
		test.Fatalf("mutant own-work deadline: %v", callerContext.Err())
	}
	var exitError *exec.ExitError
	if captureError != nil && !errors.As(captureError, &exitError) {
		test.Fatalf("running %s: %v", name, captureError)
	}
	return oracles.Run{Stdout: stdout, Stderr: stderr, ExitCode: command.ProcessState.ExitCode()}
}
