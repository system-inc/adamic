package worker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/worker/testtools"
)

// Each package declares the same input-keyed tool product; no TestMain build.
func jsonTypes(test testing.TB) string {
	test.Helper()
	inputs, err := buildcache.GoBuildInInputs("", "json-types", "./oracle/json_types.go", nil, nil)
	if err != nil {
		test.Fatal(err)
	}
	return filepath.Join(findWorkerProduct(test, inputs), "json-types")
}
func TestProduct_WorkerJSONTypes(test *testing.T) {
	test.Parallel()
	buildcache.GoBuild(test, "json-types", "./oracle/json_types.go", nil)
}

func jsonTypesEnvironment(test testing.TB, path string) []string {
	test.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		test.Fatal(err)
	}
	if !strings.Contains(string(source), "decodeJson") && !strings.Contains(string(source), "encodeJson") {
		return nil
	}
	return []string{"ADAMIC_JSON_TYPES=" + jsonTypes(test)}
}

// TestMain selects the installed SDK; all build inputs remain TestProduct_ units.
func TestMain(m *testing.M) { testtools.ConfigureWASI(); os.Exit(m.Run()) }
