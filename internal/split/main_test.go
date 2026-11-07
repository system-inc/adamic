package split_test

import (
	"os"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	_ "github.com/system-inc/adamic/internal/ccjobserver"
)

// Each package declares the same input-keyed tool product; no TestMain build.
func jsonTypes(test testing.TB) string {
	test.Helper()
	return buildcache.GoBuild(test, "json-types", "./oracle/json_types.go", nil)
}
func TestProduct_JSONTypes(test *testing.T) { test.Parallel(); jsonTypes(test) }

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
