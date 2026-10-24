package split_test

import (
	"os"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	_ "github.com/system-inc/adamic/internal/ccjobserver"
)

// Each package declares the same input-keyed tool product; no TestMain build.
func jsonTypes(t testing.TB) string {
	t.Helper()
	return buildcache.GoBuild(t, "json-types", "./oracle/json_types.go", nil)
}
func TestProduct_JSONTypes(t *testing.T) { t.Parallel(); jsonTypes(t) }

func jsonTypesEnvironment(t testing.TB, path string) []string {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), "decodeJson") && !strings.Contains(string(source), "encodeJson") {
		return nil
	}
	return []string{"ADAMIC_JSON_TYPES=" + jsonTypes(t)}
}
