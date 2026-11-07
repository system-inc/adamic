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
	path, err := filepath.Abs("gaps/postfixValue.ts")
	if err != nil {
		test.Fatal(err)
	}
	got := onNode(test, path)
	if string(got) != "7\n1\n" {
		test.Fatalf("Node %q", got)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		test.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	if err == nil || !strings.Contains(err.Error(), "a PostfixUnaryExpression") {
		test.Fatalf("postfix value gap changed: %v", err)
	}
	test.Logf("Node 7, 1; lowering: %v", err)
}

func TestMethodReplacementGap(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs("gaps/methodReplacement.ts")
	if err != nil {
		test.Fatal(err)
	}
	if got := string(onNode(test, path)); got != "replacement\noriginal\n" {
		test.Fatal(got)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		test.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	if err == nil || !strings.Contains(err.Error(), "unbound-method") {
		test.Fatalf("method replacement gap changed: %v", err)
	}
	test.Logf("Node replacement/original; Adamic: %v", err)
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
		if err := os.WriteFile(path, data, 0644); err != nil {
			test.Fatal(err)
		}
	}
	oracle := goOracle(test)
	left, right := execute(test, "", oracle, first), execute(test, "", oracle, second)
	if bytes.Equal(left, right) {
		test.Fatal("Go must distinguish these equal-size inputs")
	}
	path, err := filepath.Abs("gaps/rawInput.ts")
	if err != nil {
		test.Fatal(err)
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
	path, err := filepath.Abs("gaps/parserRecovery.ts")
	if err != nil {
		test.Fatal(err)
	}
	binary, _ := build(test, path, true)
	for _, argv := range [][]string{
		{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(test), "oracle/node.mjs"), path},
		{binary},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		command := exec.CommandContext(ctx, argv[0], argv[1:]...)
		output, err := os.CreateTemp(test.TempDir(), "recovery-log")
		if err != nil {
			test.Fatal(err)
		}
		command.Stdout, command.Stderr = output, output
		err = command.Run()
		output.Close()
		timedOut := ctx.Err() == context.DeadlineExceeded
		cancel()
		if !timedOut {
			test.Fatalf("expected bounded external timeout; got %v", err)
		}
		test.Logf("%s: unported parser recovery does not terminate within 1s", argv[0])
	}
}
