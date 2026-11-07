package native_test

import (
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// jsonTypes is oracle/json_types.go built as a product, keyed by everything it compiles: the tool node.mjs runs to
// give a program that decodes or encodes JSON its type descriptors. Without ADAMIC_JSON_TYPES naming it, node.mjs
// would go run the tool itself, and Loom's runners never build. Each package declares the same input-keyed product
// (internal/oracle, internal/split), so Workshop builds it once.
func jsonTypes(test testing.TB) string {
	test.Helper()
	return buildcache.GoBuild(test, "json-types", "./oracle/json_types.go", nil)
}

// TestProduct_JSONTypes builds jsonTypes ahead, so no witness's run pays for it.
func TestProduct_JSONTypes(test *testing.T) {
	test.Parallel()
	jsonTypes(test)
}
