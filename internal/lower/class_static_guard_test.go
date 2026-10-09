package lower

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
)

const staticSideEffectSource = `class Box { static { console.log('static side effect'); } } console.log('done');`

func TestClassStaticInitializerCallIsEmitted(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, staticSideEffectSource)
	if err != nil {
		t.Fatal(err)
	}
	initializer := -1
	for index, function := range program.Functions {
		if function.Name == "Box_static_initialize" {
			if initializer != -1 {
				t.Fatal("duplicate static initializer")
			}
			initializer = index
		}
	}
	if initializer == -1 {
		t.Fatal("static initializer function is missing")
	}
	calls := 0
	for index, statement := range program.Main {
		evaluate, ok := statement.(ir.Evaluate)
		if !ok {
			continue
		}
		call, ok := evaluate.Value.(ir.Call)
		if !ok || call.Function != initializer {
			continue
		}
		calls++
		if index == 0 || len(call.Arguments) != 1 {
			t.Fatal("static initializer must receive the allocated class object")
		}
		receiver, ok := call.Arguments[0].(ir.Read)
		allocation, allocated := program.Main[index-1].(ir.Declare)
		if !ok || !allocated || allocation.Local != receiver.Local || receiver.Of != ir.Object {
			t.Fatal("static initializer must run immediately after its class object allocation")
		}
		if _, ok := allocation.Value.(ir.ObjectLiteral); !ok {
			t.Fatal("static initializer receiver is not a class object allocation")
		}
	}
	if calls != 1 {
		t.Fatalf("static initializer call count = %d, want 1", calls)
	}
}

func TestClassStaticSideEffectMatchesNode(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	source := filepath.Join(directory, "main.a")
	if err := os.WriteFile(source, []byte(staticSideEffectSource), 0o644); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	run := func(name, executable string, arguments ...string) []byte {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, executable, arguments...)
		var stderr bytes.Buffer
		command.Stderr = &stderr
		stdout, err := command.Output()
		if err != nil || stderr.Len() != 0 {
			t.Fatalf("%s: error %v, stdout %q, stderr %q", name, err, stdout, stderr.Bytes())
		}
		return stdout
	}
	want := run("source Node", "node", "--disable-warning=ExperimentalWarning", runner, source)
	if string(want) != "static side effect\ndone\n" {
		t.Fatalf("source Node stdout = %q, want static side effect then done", want)
	}
	checked, err := load.Load([]string{source})
	if err != nil {
		t.Fatal(err)
	}
	program, err := Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	generated := filepath.Join(directory, "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := run("JavaScript backend", "node", "--disable-warning=ExperimentalWarning", runner, generated); !bytes.Equal(got, want) {
		t.Errorf("JavaScript backend stdout = %q, source Node = %q", got, want)
	}
	binary := filepath.Join(directory, "native")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if got := run("native backend with ASan/UBSan", binary); !bytes.Equal(got, want) {
		t.Errorf("native backend stdout = %q, source Node = %q", got, want)
	}
}
