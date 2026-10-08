package estree

import (
	"bytes"
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPostfixValueGap(t *testing.T) {
	path, err := filepath.Abs("gaps/postfixValue.ts")
	if err != nil {
		t.Fatal(err)
	}
	got := onNode(t, path)
	if string(got) != "7\n1\n" {
		t.Fatalf("Node %q", got)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	if err == nil || !strings.Contains(err.Error(), "a PostfixUnaryExpression") {
		t.Fatalf("postfix value gap changed: %v", err)
	}
	t.Logf("Node 7, 1; lowering: %v", err)
}

func TestMethodReplacementGap(t *testing.T) {
	path, err := filepath.Abs("gaps/methodReplacement.ts")
	if err != nil {
		t.Fatal(err)
	}
	if got := string(onNode(t, path)); got != "replacement\noriginal\n" {
		t.Fatal(got)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	if err == nil || !strings.Contains(err.Error(), "unbound-method") {
		t.Fatalf("method replacement gap changed: %v", err)
	}
	t.Logf("Node replacement/original; Adamic: %v", err)
}

func TestRawInputGap(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "malformed.ts")
	second := filepath.Join(dir, "replacement.ts")
	a := append([]byte("//"), 0xf0, 0x90, 0x80)
	b := append([]byte("//"), 0xef, 0xbf, 0xbd)
	a = append(a, []byte("\nx;")...)
	b = append(b, []byte("\nx;")...)
	for path, data := range map[string][]byte{first: a, second: b} {
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	oracle := goOracle(t)
	left, right := execute(t, "", oracle, first), execute(t, "", oracle, second)
	if bytes.Equal(left, right) {
		t.Fatal("Go must distinguish these equal-size inputs")
	}
	path, err := filepath.Abs("gaps/rawInput.ts")
	if err != nil {
		t.Fatal(err)
	}
	binary, script := build(t, path, true)
	want := onNode(t, path, first)
	for _, input := range []string{first, second} {
		for _, got := range [][]byte{onNode(t, path, input), execute(t, "", binary, input), onNode(t, script, input)} {
			if !bytes.Equal(want, got) {
				t.Fatalf("reader changed: %q != %q", got, want)
			}
		}
	}
	t.Logf("Equal size, readTextFile text and utf8Length on Node/native/emitted JS: %q; Go distinguishes: %s", want, firstDifference(left, right))
}

// The shared parser once looped forever on an unterminated type literal. The
// parser's recovery port (88f4a83d) closed that gap, so the program now has to
// finish and print the same bytes from source Node, sanitized native and
// emitted JavaScript. The guard is a hang guard sized to a program that prints
// one line, never a timing assertion.
func TestParserRecoveryTerminates(t *testing.T) {
	path, err := filepath.Abs("gaps/parserRecovery.ts")
	if err != nil {
		t.Fatal(err)
	}
	binary, script := build(t, path, true)
	node := func(file string) []string {
		return []string{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), file}
	}
	var want []byte
	for _, argv := range [][]string{node(path), {binary}, node(script)} {
		command := exec.Command(argv[0], argv[1:]...)
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err := childguard.Run(command, childguard.Options{FirstOutput: 2 * time.Minute, Ceiling: 5 * time.Minute})
		if err != nil || stderr.Len() != 0 {
			t.Fatalf("%s: %v\n%s", argv[0], err, &stderr)
		}
		if stdout.Len() == 0 {
			t.Fatalf("%s printed nothing", argv[0])
		}
		if want == nil {
			want = stdout.Bytes()
		} else if !bytes.Equal(stdout.Bytes(), want) {
			t.Fatalf("%v printed %q, source Node printed %q", argv, stdout.Bytes(), want)
		}
	}
	t.Logf("source Node, native and emitted JavaScript all terminate and print %q", want)
}

func TestInterfaceDefaultGap(t *testing.T) {
	path, err := filepath.Abs("gaps/interfaceDefault.ts")
	if err != nil {
		t.Fatal(err)
	}
	if got := string(onNode(t, path)); got != "5\n" {
		t.Fatalf("source Node: %q", got)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), loaded)
	var diagnostic *lower.NotYet
	if !errors.As(err, &diagnostic) || diagnostic.Where != path+":10:12" || diagnostic.What != "a class method through a view that erases its prototype origin" {
		t.Fatalf("recorded lowering gap changed: %v", err)
	}
	t.Logf("Node prints 5; lowering refuses before native emission: %s", err)
}
