package tsprinter

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The .txt suffix keeps this separately executed native gap out of the formatter corpus.
func TestClassInterfaceMethodGap(t *testing.T) {
	data, err := os.ReadFile("gaps/classInterfaceMethod.ts.txt")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	port := filepath.Join(directory, "gap.ts")
	if err = os.WriteFile(port, data, 0644); err != nil {
		t.Fatal(err)
	}
	node := onNode(t, port)
	if node.exitCode != 0 || len(node.stderr) != 0 || string(node.stdout) != "17\n" {
		t.Fatalf("Node truth: %+v", node)
	}
	loaded, err := load.Load([]string{port})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), loaded)
	isGap := func(err error) bool {
		var diagnostic *lower.NotYet
		return errors.As(err, &diagnostic) && diagnostic.Where == port+":6:28" && diagnostic.What == "a class method through a view that erases its prototype origin"
	}
	if !isGap(err) {
		t.Fatalf("recorded lowering gap changed: %v", err)
	}
	t.Logf("Node prints 17; lowering refuses before native emission: %s", err)
	// A safe explicit callback is also the mutant proving this gap check can reject a normal run.
	fixed := strings.Replace(strings.Replace(string(data), "read(): number", "readValue(): number", 1), "console.log(`${new Consumer(new Box()).run()}`);", "const box = new Box();\nconsole.log(`${new Consumer({read: () => box.readValue()}).run()}`);", 1)
	workaround := filepath.Join(directory, "workaround.ts")
	if err = os.WriteFile(workaround, []byte(fixed), 0644); err != nil {
		t.Fatal(err)
	}
	loaded, err = load.Load([]string{workaround})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil || isGap(err) {
		t.Fatalf("explicit callback mutant lowering: %v", err)
	}
	safe, binary := natively(t, program)
	for _, result := range []run{onNode(t, workaround), safe, onJavaScriptBackend(t, program)} {
		if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != "17\n" {
			t.Fatalf("explicit callback: %+v", result)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Log("explicit callback prints 17 on Node, native and backend; leak-free; gap check rejects this normal-run mutant")
}

func TestOptionalBooleanFunctionGap(t *testing.T) {
	data, err := os.ReadFile("gaps/optionalBooleanFunction.ts.txt")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "gap.ts")
	if err = os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || len(node.stderr) != 0 || string(node.stdout) != "absent\n" {
		t.Fatalf("Node truth: %+v", node)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), loaded)
	isGap := func(err error) bool {
		return err != nil && strings.Contains(err.Error(), "stage 0 can't lower a function value taking boolean | undefined yet")
	}
	if !isGap(err) {
		t.Fatalf("optional boolean gap changed: %v", err)
	}
	t.Logf("Node prints absent; stage 0 %s", err)
	// All layout callback callers supply known arguments, so a required parameter suffices.
	workaround := filepath.Join(t.TempDir(), "required.ts")
	if err = os.WriteFile(workaround, []byte("const show = (flag: boolean): string => flag ? 'present' : 'absent';\nconsole.log(show(false));\n"), 0644); err != nil {
		t.Fatal(err)
	}
	loaded, err = load.Load([]string{workaround})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil || isGap(err) {
		t.Fatalf("required parameter lowering: %v", err)
	}
	safe, binary := natively(t, program)
	for _, result := range []run{onNode(t, workaround), safe, onJavaScriptBackend(t, program)} {
		if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != "absent\n" {
			t.Fatalf("required parameter: %+v", result)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Log("required callback parameter prints absent on all three, leak-free; gap check rejects successful-lowering mutant")
}
