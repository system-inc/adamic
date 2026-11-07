package estree

import (
	"bytes"
	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPostfixValueGap(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("gaps/postfixValue.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	got := onNode(test, path)
	if string(got) != "7\n1\n" {
		test.Fatalf("Node %q", got)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, lowerError := lower.Lower(context.Background(), program)
	if lowerError == nil || !strings.Contains(lowerError.Error(), "a PostfixUnaryExpression") {
		test.Fatalf("postfix value gap changed: %v", lowerError)
	}
	test.Logf("Node 7, 1; lowering: %v", lowerError)
}

func TestMethodReplacementGap(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("gaps/methodReplacement.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	if got := string(onNode(test, path)); got != "replacement\noriginal\n" {
		test.Fatal(got)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, lowerError := lower.Lower(context.Background(), program)
	if lowerError == nil || !strings.Contains(lowerError.Error(), "unbound-method") {
		test.Fatalf("method replacement gap changed: %v", lowerError)
	}
	test.Logf("Node replacement/original; Adamic: %v", lowerError)
}

// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestRawInputGap(test *testing.T) {
	dir := test.TempDir()
	first := filepath.Join(dir, "malformed.ts")
	second := filepath.Join(dir, "replacement.ts")
	a := append([]byte("//"), 0xf0, 0x90, 0x80)
	b := append([]byte("//"), 0xef, 0xbf, 0xbd)
	a = append(a, []byte("\nx;")...)
	b = append(b, []byte("\nx;")...)
	for path, data := range map[string][]byte{first: a, second: b} {
		if writeError := os.WriteFile(path, data, 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	oracle := goOracle(test)
	left, right := execute(test, "", oracle, first), execute(test, "", oracle, second)
	if bytes.Equal(left, right) {
		test.Fatal("Go must distinguish these equal-size inputs")
	}
	path, absolutePathError := filepath.Abs("gaps/rawInput.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	binary, script := build(test, path, true)
	want := onNode(test, path, first)
	for _, input := range []string{first, second} {
		for _, got := range [][]byte{
			onNode(test, path, input),
			execute(test, "", binary, input),
			onNode(test, script, input),
		} {
			if !bytes.Equal(want, got) {
				test.Fatalf("reader changed: %q != %q", got, want)
			}
		}
	}
	test.Logf(
		"Equal size, readTextFile text and utf8Length on Node/native/emitted JS: %q; Go distinguishes: %s",
		want,
		firstDifference(left, right),
	)
}

// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestParserRecoveryGap(test *testing.T) {
	path, absolutePathError := filepath.Abs("gaps/parserRecovery.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	binary, _ := build(test, path, true)
	for _, argv := range [][]string{
		{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(test), "oracle/node.mjs"), path},
		{binary},
	} {
		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Second)
		command := exec.CommandContext(timeoutContext, argv[0], argv[1:]...)
		output, temporaryFileError := os.CreateTemp(test.TempDir(), "recovery-log")
		if temporaryFileError != nil {
			test.Fatal(temporaryFileError)
		}
		command.Stdout, command.Stderr = output, output
		runError := command.Run()
		output.Close()
		timedOut := timeoutContext.Err() == context.DeadlineExceeded
		cancel()
		if !timedOut {
			test.Fatalf("expected bounded external timeout; got %v", runError)
		}
		test.Logf("%s: unported parser recovery does not terminate within 1s", argv[0])
	}
}
