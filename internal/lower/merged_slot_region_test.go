package lower

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// A merged member's slot is laid Absent, the shape that sends a constructor's region version down the
// construction path (emit_objects.go's constructionFields, #837zjkt slice 2), so an instance newed where the
// caller plans a region, then read and keyed, must stay clean under ASan and agree with Node (Lower's hold on
// #86cmqrp head (2)).
func TestMergedSlotRegionConstruction(t *testing.T) {
	t.Parallel()
	const source = `class Merged { a = 1; }
interface Merged { f?: string; }
function weigh(m: Merged): number {
    return m.a + (m.f === undefined ? 10 : 20);
}
function keys(m: Merged): string {
    return Object.keys(m).join(",");
}
function filled(m: Merged): string {
    m.f = "x";
    return String(m.f) + ":" + Object.keys(m).join(",");
}
console.log(String(weigh(new Merged())));
console.log(keys(new Merged()));
console.log(filled(new Merged()));
`
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	code := native.C(program)
	version := regexp.MustCompile(`adamic_\w*Merged_allocate_[0-9a-f]{32}_in\b`)
	called := false
	for _, line := range strings.Split(code, "\n") {
		if version.MatchString(line) || strings.Contains(line, "adamic_region") {
			t.Logf("C: %s", strings.TrimSpace(line))
		}
		if version.MatchString(line) && !strings.HasPrefix(strings.TrimSpace(line), "static ") {
			called = true
		}
	}
	if !called {
		t.Fatal("no caller reaches Merged's region version, so this probe can't see the construction hazard")
	}
	binary := filepath.Join(t.TempDir(), "native")
	if err := native.Build(code, binary, native_runtime.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
	var out, errs bytes.Buffer
	command.Stdout, command.Stderr = &out, &errs
	if err := command.Run(); err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatal(err)
		}
	}
	const want = "11\na\nx:a,f\n"
	if out.String() != want || errs.Len() != 0 {
		t.Fatalf("native under the sanitizers: stdout %q stderr %q; want Node's %q", out.String(), errs.String(), want)
	}
	if report := leakcheck.Report(t, code, binary); report != "" {
		t.Fatal(report)
	}
}
