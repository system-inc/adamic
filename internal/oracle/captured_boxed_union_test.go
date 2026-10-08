package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/captured_boxed_union.a", true, false})
}
func TestCapturedUnionWriteMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/captured_boxed_union.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	for index := range program.Functions {
		f := &program.Functions[index]
		for slot, statement := range f.Body {
			if write, ok := statement.(ir.Assign); ok && program.Locals[write.Local].Captured && program.Locals[write.Local].Type == ir.Union {
				f.Body[slot] = ir.Evaluate{Value: write.Value}
				changed++
			}
		}
	}
	if changed != 1 {
		t.Fatalf("want one removed union cell write, got %d", changed)
	}
	native, _ := natively(t, program)
	for backend, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 {
			t.Fatalf("%s mutant failed to run: %+v", backend, got)
		}
		if diff := disagreement(truth, got); diff != "stdout differs" {
			t.Fatalf("%s mutant not caught by Node stdout: %s", backend, diff)
		}
		t.Logf("%s: write mutant exits normally, stdout %q differs from Node %q", backend, got.stdout, truth.stdout)
	}
}
