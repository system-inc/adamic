package native

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestAsyncUnexpectedExceptionStops(test *testing.T) {
	test.Parallel()
	e := emitter{function: &ir.Function{Name: "proved_safe"}, asyncSlots: map[int]int{}}
	e.checkThrown()
	source := "#include \"adamic.h\"\nint main(void) { adamic_thrown = (adamic_object *)1;\n" +
		e.out.String() +
		"return 0; }\n"
	binary := filepath.Join(test.TempDir(), "unexpected-exception")
	if buildError := Build(source, binary, Options{}); buildError != nil {
		test.Fatal(buildError)
	}
	output, combinedOutputError := exec.Command(binary).CombinedOutput()
	exit, ok := combinedOutputError.(*exec.ExitError)
	if !ok ||
		exit.ExitCode() != 70 ||
		!strings.Contains(string(output), "proved_safe: exception where the flow graph proved none") {
		test.Fatalf("want named exit 70, got %v: %s", combinedOutputError, output)
	}
}
