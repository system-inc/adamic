package lower

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
)

// Local Node comparison: compiler/lower-agree is not pushed.
const parameterPropertyValueSource = "class C {constructor(public value:number){}} const c=new C(7); console.log(`${c.value}`);"

func TestParameterPropertyValueMatchesNode(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	source := filepath.Join(directory, "main.a")
	if err := os.WriteFile(source, []byte(parameterPropertyValueSource), 0o644); err != nil {
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
	if string(want) != "7\n" {
		t.Fatalf("source Node stdout = %q, want 7", want)
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
