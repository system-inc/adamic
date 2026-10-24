package oracle

import (
	"os"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// jsonTypes is oracle/json_types.go built as a product, keyed by everything it compiles: the tool node.mjs runs to
// give a program that decodes or encodes JSON its type descriptors. A raw go build in TestMain was keyed by nothing,
// so Loom's builder and its runners named it differently and no tree with it could publish.
func jsonTypes(t testing.TB) string {
	t.Helper()
	return buildcache.GoBuild(t, "json-types", "./oracle/json_types.go", nil)
}

// TestProduct_JSONTypes builds jsonTypes ahead, so no fixture's run pays for it.
func TestProduct_JSONTypes(t *testing.T) {
	t.Parallel()
	jsonTypes(t)
}

// jsonTypesEnvironment is what Node needs to run the program at path: ADAMIC_JSON_TYPES naming jsonTypes when the
// program decodes or encodes JSON, the same test node.mjs makes, and nothing otherwise, so a program without JSON never
// waits on the build. Without it node.mjs would go run the tool itself, unkeyed.
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
