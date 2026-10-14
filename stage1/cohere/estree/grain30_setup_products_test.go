package estree

import (
	"bytes"
	"encoding/json"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var estreeGoOracleReady struct {
	once   sync.Once
	binary string
}

func goOracle(t *testing.T) string {
	t.Helper()
	estreeGoOracleReady.once.Do(func() {
		inputs := syntaxMutantSetupInputs(t)
		inputs.Name = "estree-go-oracle-v1"
		inputs.Flags = append(inputs.Flags, "go build "+strings.Join(buildcache.ReproducibleGoFlags(), " "))
		dir := buildcache.Product(t, inputs, func(dir string) error {
			estreeGoOracleBuild(t, dir)
			return nil
		})
		estreeGoOracleReady.binary = filepath.Join(dir, "oracle")
	})
	if estreeGoOracleReady.binary == "" {
		t.Fatal("Go oracle preparation failed")
	}
	return estreeGoOracleReady.binary
}

func estreeGoOracleBuild(t *testing.T, dir string) string {
	t.Helper()
	repo := root(t)
	source, err := filepath.Abs("testdata/oracle.go")
	if err != nil {
		t.Fatal(err)
	}
	virtual := filepath.Join(repo, "cohere/adamic_estree_oracle.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
	defer os.Remove(path)
	binary := filepath.Join(dir, "oracle")
	command := toolchain.CommandIn(filepath.Join(repo, "cohere"), "go", append(append([]string{"build"}, buildcache.ReproducibleGoFlags()...), "-overlay="+path, "-o", binary, virtual)...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil || stderr.Len() != 0 {
		t.Fatalf("go build %s: %v\n%s", virtual, err, &stderr)
	}
	return binary
}

func TestProduct_EstreeGoOracle(t *testing.T) { t.Parallel(); goOracle(t) }

func estreeMainPath(t *testing.T) string {
	t.Helper()
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	return main
}
func estreeStallMutantPath(t *testing.T) string {
	t.Helper()
	return miscMutant(t, "sourceStatements.ts", "if(this.parser.scanner.fullStart === start)", "if(false)")
}
func estreeExportMutantPath(t *testing.T) string {
	t.Helper()
	return miscMutant(t, "convert.ts", "this.arena.node(wrapper).start = this.start(exported);", "this.arena.node(wrapper).start = this.start(id);")
}
func estreeDiagnosticControlPath(t *testing.T) string {
	t.Helper()
	return miscMutant(t, "pipeline.ts", "if(syntax !== '')", "if(false)")
}
func estreeCookedSurrogateMutantPath(t *testing.T) string {
	t.Helper()
	return miscMutant(t, "protocol.ts", `result += '\\ufffd\\ufffd\\ufffd';`, `result += '\\ufffd';`)
}

// estreeInterfaceControlPath is TestInterfaceTypeMethodGap's default-free control: gaps/interfaceTypeMethod.ts with
// its defaulted method renamed and called through an explicit callback, written into a product so its build is one too.
func estreeInterfaceControlPath(t *testing.T) string {
	t.Helper()
	inputs := buildcache.Inputs{Name: "estree-interface-control-source", Files: []string{"stage1/cohere/estree/gaps/interfaceTypeMethod.ts", "stage1/cohere/estree/grain30_setup_products_test.go"}, Toolchain: []string{runtime.Version()}}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		source, err := os.ReadFile("gaps/interfaceTypeMethod.ts")
		if err != nil {
			return err
		}
		control := strings.Replace(strings.Replace(string(source), "type(minimum = 0, conditional = true)", "typeValue(minimum: number, conditional: boolean)", 1), "console.log(parse(new ParserLike()).toString());", "const parser = new ParserLike();\nconsole.log(parse({type: (minimum: number, conditional: boolean) => parser.typeValue(minimum, conditional)}).toString());", 1)
		return os.WriteFile(filepath.Join(directory, "control.ts"), []byte(control), 0644)
	})
	return filepath.Join(directory, "control.ts")
}
func estreeGapPath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs("gaps/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return path
}
func estreeBoundedAudits(t *testing.T) []boundedPortCase {
	t.Helper()
	var audited []boundedPortCase
	for _, item := range boundedPortCases(t) {
		if item.audit {
			audited = append(audited, item)
		}
	}
	return audited
}
func TestProduct_EstreeBoundedPort(t *testing.T) { t.Parallel(); miscBuild(t, estreeMainPath(t)) }
func TestProduct_EstreeBoundedAnswers(t *testing.T) {
	t.Parallel()
	miscAnswers(t, goOracle(t), estreeBoundedAudits(t))
}
func TestProduct_EstreeStallSource(t *testing.T)  { t.Parallel(); estreeStallMutantPath(t) }
func TestProduct_EstreeStallPort(t *testing.T)    { t.Parallel(); miscBuild(t, estreeStallMutantPath(t)) }
func TestProduct_EstreeExportSource(t *testing.T) { t.Parallel(); estreeExportMutantPath(t) }
func TestProduct_EstreeExportPort(t *testing.T) {
	t.Parallel()
	miscBuild(t, estreeExportMutantPath(t))
}
func TestProduct_EstreeExportAnswers(t *testing.T) {
	t.Parallel()
	miscTextAnswers(t, goOracle(t), decoratedExports(), ".ts")
}
func TestProduct_EstreeDiagnosticControlSource(t *testing.T) {
	t.Parallel()
	estreeDiagnosticControlPath(t)
}
func TestProduct_EstreeDiagnosticControlPort(t *testing.T) {
	t.Parallel()
	miscBuild(t, estreeDiagnosticControlPath(t))
}
func TestProduct_EstreeCookedSurrogateSource(t *testing.T) {
	t.Parallel()
	estreeCookedSurrogateMutantPath(t)
}
func TestProduct_EstreeCookedSurrogatePort(t *testing.T) {
	t.Parallel()
	miscBuild(t, estreeCookedSurrogateMutantPath(t))
}
func TestProduct_EstreeInterfaceControlSource(t *testing.T) {
	t.Parallel()
	estreeInterfaceControlPath(t)
}
func TestProduct_EstreeInterfaceControlPort(t *testing.T) {
	t.Parallel()
	miscBuild(t, estreeInterfaceControlPath(t))
}
func TestProduct_EstreeRawInputGapPort(t *testing.T) {
	t.Parallel()
	miscBuild(t, estreeGapPath(t, "rawInput.ts"))
}
func TestProduct_EstreeParserRecoveryGapPort(t *testing.T) {
	t.Parallel()
	miscBuild(t, estreeGapPath(t, "parserRecovery.ts"))
}
