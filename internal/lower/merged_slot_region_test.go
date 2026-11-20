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
function make(n: number): Merged {
    const m = new Merged();
    m.a = n;
    return m;
}
function weigh(n: number): number {
    const m = make(n);
    return m.a + (m.f === undefined ? 10 : 20);
}
function fill(n: number): string {
    const m = new Merged();
    m.a = n;
    m.f = "x" + String(n);
    return String(m.f) + ":" + Object.keys(m).join(",");
}
console.log(String(weigh(3)));
console.log(fill(4));
console.log(Object.keys(make(5)).join(","));
`
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	code := native.C(program)
	version := regexp.MustCompile(`adamic_\w*Merged\w*_in\b|adamic_region `)
	for _, line := range strings.Split(code, "\n") {
		if version.MatchString(line) || strings.Contains(line, "adamic_object_construct") {
			t.Logf("C: %s", strings.TrimSpace(line))
		}
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
	const want = "13\nx4:a,f\na\n"
	if out.String() != want || errs.Len() != 0 {
		t.Fatalf("native under the sanitizers: stdout %q stderr %q; want Node's %q", out.String(), errs.String(), want)
	}
	if report := leakcheck.Report(t, code, binary); report != "" {
		t.Fatal(report)
	}
}
