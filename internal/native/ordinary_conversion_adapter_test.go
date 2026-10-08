package native_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestOrdinaryConversionAdapterMutant(t *testing.T) {
	t.Parallel()
	path := filepath.Join("..", "oracle", "testdata", "object_semantics", "conversion_runtime.a")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := prototypeNode(t, string(source))
	if strings.Count(want, "TypeError: Cannot convert object to primitive value\n") != 2 {
		t.Fatalf("Node must throw for both hints: %q", want)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	if strings.Count(code, "= adamic_ordinary_to_primitive(") < 2 {
		t.Fatal("template and addition must enter the runtime protocol")
	}
	// The generated program is unchanged. Only the runtime's exhausted-protocol
	// TypeError is removed, so the observation must catch a zero-exit wrong value.
	binary := primitiveMutant(t, code, "adamic_thrown = error;", "adamic_release(error);")
	run := primitiveRun(binary, 0, "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
	if run.ExitCode != 0 {
		t.Fatalf("mutant must finish: exit %d, stdout %q, stderr %s", run.ExitCode, run.Stdout, run.Stderr)
	}
	if bytes.Equal(run.Stdout, []byte(want)) {
		t.Fatal("missing conversion TypeError survived the source Node comparison")
	}
	if report := leakcheck.Unbalanced(run); report != "" {
		t.Fatal(report)
	}
	if report := leakcheck.Report(t, code, binary, "0"); report != "" {
		t.Fatal(report)
	}
	t.Logf("Node caught missing runtime TypeError; mutant exited zero without leaks: %q", run.Stdout)
}
