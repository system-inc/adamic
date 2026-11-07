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

func TestClockGenericReturnsT01Mutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/clock_generic_returns_t_01.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := onNode(test, path)
	if want.exitCode != 0 || string(want.stdout) != "object undefined 2\n" {
		test.Fatalf("unexpected Node result: %+v", want)
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
		test.Fatalf("clock_generic_returns_t_01_present_as_undefined changed %d returns", changed)
	}
	native, binary := natively(test, program)
	for name, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(test, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 {
			test.Fatalf("%s mutant must execute cleanly: %+v", name, got)
		}
		if difference := disagreement(want, got); difference != "stdout differs" {
			test.Fatalf("%s mutant: want stdout disagreement, got %q", name, difference)
		}
		if string(got.stdout) != "undefined undefined 2\n" {
			test.Fatalf("%s mutant missed branches or calls: %q", name, got.stdout)
		}
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
}
