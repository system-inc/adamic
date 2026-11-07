package estree

import (
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/stage1/cohere/estree/oracles"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func goOracle(test *testing.T) string { test.Helper(); return oracles.ParserOracle(test) }

func TestProduct_EstreeFindsGoOracle(test *testing.T) { test.Parallel(); goOracle(test) }

func estreeMainPath(test *testing.T) string {
	test.Helper()
	main, err := filepath.Abs("main.ts")
	if err != nil {
		test.Fatal(err)
	}
	return main
}
func estreeStallMutantPath(test *testing.T) string {
	test.Helper()
	return miscMutant(test, "sourceStatements.ts", "if(this.parser.scanner.fullStart === start)", "if(false)")
}
func estreeExportMutantPath(test *testing.T) string {
	test.Helper()
	return miscMutant(
		test,
		"convert.ts",
		"this.arena.node(wrapper).start = this.start(exported);",
		"this.arena.node(wrapper).start = this.start(id);",
	)
}
func estreeDiagnosticControlPath(test *testing.T) string {
	test.Helper()
	return miscMutant(test, "pipeline.ts", "if(syntax !== '')", "if(false)")
}
func estreeCookedSurrogateMutantPath(test *testing.T) string {
	test.Helper()
	return miscMutant(test, "protocol.ts", `result += '\\ufffd\\ufffd\\ufffd';`, `result += '\\ufffd';`)
}

// estreeInterfaceControlPath is TestInterfaceTypeMethodGap's default-free control: gaps/interfaceTypeMethod.ts with
// its defaulted method renamed and called through an explicit callback, written into a product so its build is one too.
func estreeInterfaceControlPath(test *testing.T) string {
	test.Helper()
	inputs := buildcache.Inputs{Name: "estree-interface-control-source", Files: []string{
		"stage1/cohere/estree/gaps/interfaceTypeMethod.ts",
		"stage1/cohere/estree/grain30_setup_products_test.go",
	}, Toolchain: []string{runtime.Version()}}
	directory := buildcache.Product(test, inputs, func(directory string) error {
		source, err := os.ReadFile("gaps/interfaceTypeMethod.ts")
		if err != nil {
			return err
		}
		control := strings.Replace(
			strings.Replace(
				string(source),
				"type(minimum = 0, conditional = true)",
				"typeValue(minimum: number, conditional: boolean)",
				1,
			),
			"console.log(parse(new ParserLike()).toString());",
			"const parser = new ParserLike();\nconsole.log(parse({type: (minimum: number, conditional: boolean) => parser.typeValue(minimum, conditional)}).toString());",
			1,
		)
		return os.WriteFile(filepath.Join(directory, "control.ts"), []byte(control), 0644)
	})
	return filepath.Join(directory, "control.ts")
}
func estreeGapPath(test *testing.T, name string) string {
	test.Helper()
	path, err := filepath.Abs("gaps/" + name)
	if err != nil {
		test.Fatal(err)
	}
	return path
}
func estreeBoundedAudits(test *testing.T) []boundedPortCase {
	test.Helper()
	var audited []boundedPortCase
	for _, item := range boundedPortCases(test) {
		if item.audit {
			audited = append(audited, item)
		}
	}
	return audited
}
func TestProduct_EstreeBoundedPort(test *testing.T) {
	test.Parallel()
	miscBuild(test, estreeMainPath(test))
}
func TestProduct_EstreeFindsBoundedAnswers(test *testing.T) {
	test.Parallel()
	miscAnswers(test, goOracle(test), estreeBoundedAudits(test))
}
func TestProduct_EstreeStallSource(test *testing.T) { test.Parallel(); estreeStallMutantPath(test) }
func TestProduct_EstreeStallPort(test *testing.T) {
	test.Parallel()
	miscBuild(test, estreeStallMutantPath(test))
}
func TestProduct_EstreeExportSource(test *testing.T) { test.Parallel(); estreeExportMutantPath(test) }
func TestProduct_EstreeExportPort(test *testing.T) {
	test.Parallel()
	miscBuild(test, estreeExportMutantPath(test))
}
func TestProduct_EstreeFindsExportAnswers(test *testing.T) {
	test.Parallel()
	miscTextAnswers(test, goOracle(test), decoratedExports(), ".ts")
}
func TestProduct_EstreeDiagnosticControlSource(test *testing.T) {
	test.Parallel()
	estreeDiagnosticControlPath(test)
}
func TestProduct_EstreeDiagnosticControlPort(test *testing.T) {
	test.Parallel()
	miscBuild(test, estreeDiagnosticControlPath(test))
}
func TestProduct_EstreeCookedSurrogateSource(test *testing.T) {
	test.Parallel()
	estreeCookedSurrogateMutantPath(test)
}
func TestProduct_EstreeCookedSurrogatePort(test *testing.T) {
	test.Parallel()
	miscBuild(test, estreeCookedSurrogateMutantPath(test))
}
func TestProduct_EstreeInterfaceControlSource(test *testing.T) {
	test.Parallel()
	estreeInterfaceControlPath(test)
}
func TestProduct_EstreeInterfaceControlPort(test *testing.T) {
	test.Parallel()
	miscBuild(test, estreeInterfaceControlPath(test))
}
func TestProduct_EstreeRawInputGapPort(test *testing.T) {
	test.Parallel()
	miscBuild(test, estreeGapPath(test, "rawInput.ts"))
}
func TestProduct_EstreeParserRecoveryGapPort(test *testing.T) {
	test.Parallel()
	miscBuild(test, estreeGapPath(test, "parserRecovery.ts"))
}
