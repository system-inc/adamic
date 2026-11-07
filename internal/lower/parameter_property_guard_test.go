package lower

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

const parameterPropertyValueSource = "class C {constructor(public value:number){}} const c=new C(7); console.log(`${c.value}`);"

func TestParameterPropertyValueMatchesNode(test *testing.T) {
	test.Parallel()
	program := lowersAndAgreesWithNode(test, parameterPropertyValueSource)
	directory := test.TempDir()
	run := func(name, executable string, arguments ...string) []byte {
		test.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, executable, arguments...)
		var stderr bytes.Buffer
		command.Stderr = &stderr
		stdout, err := command.Output()
		if err != nil || stderr.Len() != 0 {
			test.Fatalf("%s: error %v, stdout %q, stderr %q", name, err, stdout, stderr.Bytes())
		}
		return stdout
	}
	want := []byte("7\n")
	binary := filepath.Join(directory, "native")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	if got := run("native backend with ASan/UBSan", binary); !bytes.Equal(got, want) {
		test.Errorf("native backend stdout = %q, source Node = %q", got, want)
	}
}
