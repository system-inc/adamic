package native

import (
	"bytes"
	"github.com/system-inc/adamic/internal/ir"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestViewCallableBoxingUnknownProducer(t *testing.T) {
	e := &emitter{program: &ir.Program{}}
	invoke := e.viewCallableBoxedInvokeTypes(nil, ir.Union, false)
	source := "#include \"adamic.h\"\n#include <stdio.h>\n" + strings.Join(e.declarations, "\n") + `
static adamic_value foreign(adamic_closure *self, adamic_value *arguments, size_t count) {
 (void)self; (void)arguments; (void)count;
 return (adamic_value){.number = 7};
}
int main(void) {
 adamic_closure *closure = adamic_closure_new(foreign, 0);
 adamic_value result = ` + invoke + `(closure, NULL, 0, false);
 printf("%.0f\n", result.number);
 adamic_release(closure);
 return 0;
}
`
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "unknown-producer")
		if err := Build(source, binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(binary)
		var out, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &stderr
		err := cmd.Run()
		failure, ok := err.(*exec.ExitError)
		if !ok || failure.ExitCode() != 70 || out.Len() != 0 || stderr.String() != "adamic: panic: callable ABI adapter: unknown producer signature\n" {
			t.Fatalf("unknown producer exit %v stdout %q stderr %q", err, out.String(), stderr.String())
		}
	}
}
