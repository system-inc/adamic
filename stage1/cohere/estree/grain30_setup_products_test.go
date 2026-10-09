package estree

import (
	"encoding/json"
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"path/filepath"
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
	binary := filepath.Join(dir, "oracle")
	execute(t, filepath.Join(repo, "cohere"), "go", "build", "-overlay="+path, "-o", binary, virtual)
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
