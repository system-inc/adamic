package lower

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestOverloadStructuralResults(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"block", "evaluator", "fields", "factory"} {
		source, err := os.ReadFile("../oracle/testdata/overload_structural_" + name + ".a")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := lowerSource(t, string(source)); err != nil {
			t.Fatal(err)
		}
	}
}
func checkOverloadStructuralRefuses(t *testing.T, selected string) {
	t.Helper()
	for _, probe := range []struct{ name, file, path, nodeOutput string }{
		{"Block factory covariance", "block-factory-liar", "readonly covariance requires string to fit", "Expression\n"},
		{"factory literal covariance", "factory-literal-liar", "readonly covariance requires string | number | undefined to fit", "bad\n"},
		{"factory writable invariance", "factory-writable", "writable invariance (reverse direction) requires string | undefined to fit string", "undefined\n"},
		{"factory field storage", "factory-storage", "result.value", "7\n"},
		{"factory result covariance", "factory-liar", "readonly covariance requires string | number | undefined to fit string | undefined", "1\n"},
		{"visitor domain liar", "visitor-domain-liar", "callback contravariance at callback.parameter1 requires Node to fit TIn", "undefined\n1\n"},
		{"visitor input contravariance", "visitor-domain", "callback contravariance at callback.parameter1 requires Node to fit TIn", "modifier\n1\n"},
		{"mixed field storage", "fields-mixed", "result.value without a single storage representation", "7\n"},
		{"missing Block field", "block-liar", "result.statements", "undefined\n"},
		{"shared wider result", "evaluator-shared-liar", "result.value", "1\n"},
		{"reverse readonly variance", "evaluator-reverse-liar", "result.value", "1\n"},
		{"writable variance", "evaluator-writable", "parameter input", "1\n"},
	} {
		func() {
			if probe.name != selected {
				return
			}
			file, err := filepath.Abs("testdata/overload-structural/" + probe.file + ".a")
			if err != nil {
				t.Fatal(err)
			}
			runner, _ := filepath.Abs("../../oracle/node.mjs")
			output, err := exec.Command("node", "--disable-warning=ExperimentalWarning", runner, file).CombinedOutput()
			if err != nil || string(output) != probe.nodeOutput {
				t.Fatalf("Node witness: %v %q", err, output)
			}
			source, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSource(t, string(source))
			if err == nil || !strings.Contains(err.Error(), probe.path) {
				t.Fatalf("wanted refusal naming %s, got %v", probe.path, err)
			}
		}()
	}
}

func TestOverloadStructuralRefusesBlockFactoryCovariance(t *testing.T) {
	t.Parallel()
	checkOverloadStructuralRefuses(t, "Block factory covariance")
}

func TestOverloadStructuralRefusesFactoryLiteralCovariance(t *testing.T) {
	t.Parallel()
	checkOverloadStructuralRefuses(t, "factory literal covariance")
}

func TestOverloadStructuralRefusesFactoryWritableInvariance(t *testing.T) {
	t.Parallel()
	checkOverloadStructuralRefuses(t, "factory writable invariance")
}

func TestOverloadStructuralRefusesFactoryFieldStorage(t *testing.T) {
	t.Parallel()
	checkOverloadStructuralRefuses(t, "factory field storage")
}

func TestOverloadStructuralRefusesFactoryResultCovariance(t *testing.T) {
	t.Parallel()
	checkOverloadStructuralRefuses(t, "factory result covariance")
}

func TestOverloadStructuralRefusesVisitorDomainLiar(t *testing.T) {
	t.Parallel()
	checkOverloadStructuralRefuses(t, "visitor domain liar")
}

func TestOverloadStructuralRefusesVisitorInputContravariance(t *testing.T) {
	t.Parallel()
	checkOverloadStructuralRefuses(t, "visitor input contravariance")
}

func TestOverloadStructuralRefusesMixedFieldStorage(t *testing.T) {
	t.Parallel()
	checkOverloadStructuralRefuses(t, "mixed field storage")
}

func TestOverloadStructuralRefusesMissingBlockField(t *testing.T) {
	t.Parallel()
	checkOverloadStructuralRefuses(t, "missing Block field")
}

func TestOverloadStructuralRefusesSharedWiderResult(t *testing.T) {
	t.Parallel()
	checkOverloadStructuralRefuses(t, "shared wider result")
}

func TestOverloadStructuralRefusesReverseReadonlyVariance(t *testing.T) {
	t.Parallel()
	checkOverloadStructuralRefuses(t, "reverse readonly variance")
}

func TestOverloadStructuralRefusesWritableVariance(t *testing.T) {
	t.Parallel()
	checkOverloadStructuralRefuses(t, "writable variance")
}
