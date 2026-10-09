package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{
		"internal/oracle/testdata/clock_generic_returns_t_01.a", true, false,
	})
}

// Not parallel: oracle helpers write the shared os.UserCacheDir()/adamic/gate and adamic/runtime directories.
func TestClockGenericReturnsT01Mutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/clock_generic_returns_t_01.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 0 || string(want.stdout) != "object undefined 2\n" {
		t.Fatalf("unexpected Node result: %+v", want)
	}
	changed := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		for statement, body := range function.Body {
			returned, ok := body.(ir.Return)
			if !ok {
				continue
			}
			if _, present := returned.Value.(ir.ObjectLiteral); !present {
				continue
			}
			returned.Value = ir.Undefined{Of: function.Returns}
			function.Body[statement] = returned
			changed++
		}
	}
	if changed != 1 {
		t.Fatalf("clock_generic_returns_t_01_present_as_undefined changed %d returns", changed)
	}
	native, binary := natively(t, program)
	for name, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 {
			t.Fatalf("%s mutant must execute cleanly: %+v", name, got)
		}
		if difference := disagreement(want, got); difference != "stdout differs" {
			t.Fatalf("%s mutant: want stdout disagreement, got %q", name, difference)
		}
		if string(got.stdout) != "undefined undefined 2\n" {
			t.Fatalf("%s mutant missed branches or calls: %q", name, got.stdout)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
