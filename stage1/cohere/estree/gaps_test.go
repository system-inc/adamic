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

func TestParserRecoveryGap(t *testing.T) {
	path, err := filepath.Abs("gaps/parserRecovery.ts")
	if err != nil {
		t.Fatal(err)
	}
	binary, _ := build(t, path, true)
	for _, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), path}, {binary}} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		command := exec.CommandContext(ctx, argv[0], argv[1:]...)
		output, err := os.CreateTemp(t.TempDir(), "recovery-log")
		if err != nil {
			t.Fatal(err)
		}
		command.Stdout, command.Stderr = output, output
		err = command.Run()
		output.Close()
		timedOut := ctx.Err() == context.DeadlineExceeded
		cancel()
		if !timedOut {
			t.Fatalf("expected bounded external timeout; got %v", err)
		}
		t.Logf("%s: unported parser recovery does not terminate within 1s", argv[0])
	}
}

func TestParserBoundaryRefusals(t *testing.T) {
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	binary, script := build(t, main, true)
	oracle := goOracle(t)
	cases := []struct{ name, text, diagnostic string }{
		{"modifier.ts", "class C { readonly!: number; }", "recovered an identifier"},
		{"surrogate.ts", "'\\ud800a\\udc00';", "unpaired surrogates"},
		{"jsx.tsx", "const node = <A/>;", "unported JSX"},
	}
	for _, sample := range cases {
		t.Run(sample.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), sample.name)
			if err := os.WriteFile(path, []byte(sample.text), 0644); err != nil {
				t.Fatal(err)
			}
			if len(execute(t, "", oracle, path)) == 0 {
				t.Fatal("Go should produce a nonempty tree")
			}
			for _, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, path}, {binary, path}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), script, path}} {
				command := exec.Command(argv[0], argv[1:]...)
				stdout, err := os.CreateTemp(t.TempDir(), "stdout")
				if err != nil {
					t.Fatal(err)
				}
				var stderr bytes.Buffer
				command.Stdout, command.Stderr = stdout, &stderr
				err = command.Run()
				stdout.Close()
				info, statErr := os.Stat(stdout.Name())
				if statErr != nil {
					t.Fatal(statErr)
				}
				if err == nil || info.Size() != 0 || !strings.Contains(stderr.String(), sample.diagnostic) {
					t.Fatalf("%v: exit=%v stdout=%d stderr=%s", argv, err, info.Size(), &stderr)
				}
			}
			t.Logf("Go accepts %q; source Node, sanitized native and emitted JS explicitly refuse: %s", sample.text, sample.diagnostic)
		})
	}
}

func TestInterfaceDefaultGap(t *testing.T) {
	path, err := filepath.Abs("gaps/interfaceDefault.ts")
	if err != nil {
		t.Fatal(err)
	}
	if got := string(onNode(t, path)); got != "5\n" {
		t.Fatalf("source Node: %q", got)
	}
	binary, script := build(t, path, true)
	command := exec.Command(binary)
	stdout, err := os.CreateTemp(t.TempDir(), "native-gap-stdout")
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	command.Stdout, command.Stderr = stdout, &stderr
	err = command.Run()
	stdout.Close()
	if err == nil || !strings.Contains(stderr.String(), "AddressSanitizer: stack-buffer-overflow") {
		t.Fatalf("native gap changed: %v\n%s", err, &stderr)
	}
	t.Logf("Sanitized native: %v\n%s", err, &stderr)
	if got := string(onNode(t, script)); got != "5\n" {
		t.Fatalf("emitted JS: %q", got)
	}
	t.Log("Node and emitted JS: 5; sanitized native reads past interface argument storage; release output 4 was separately observed and is undefined behavior")
}
