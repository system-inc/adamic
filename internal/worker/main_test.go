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
	inputs, goBuildInInputsError := buildcache.GoBuildInInputs("", "json-types", "./oracle/json_types.go", nil, nil)
	if goBuildInInputsError != nil {
		test.Fatal(goBuildInInputsError)
	}
	return filepath.Join(findWorkerProduct(test, inputs), "json-types")
}
func TestProduct_WorkerJsonTypes(test *testing.T) {
	test.Parallel()
	buildcache.GoBuild(test, "json-types", "./oracle/json_types.go", nil)
}

func jsonTypesEnvironment(test testing.TB, path string) []string {
	test.Helper()
	source, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	if !strings.Contains(string(source), "decodeJson") && !strings.Contains(string(source), "encodeJson") {
		return nil
	}
	return []string{"ADAMIC_JSON_TYPES=" + jsonTypes(test)}
}

// TestMain selects the installed SDK; all build inputs remain TestProduct_ units.
func TestMain(suite *testing.M) { testtools.ConfigureWasi(); os.Exit(suite.Run()) }
