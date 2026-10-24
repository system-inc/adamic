package oracle

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native/runtimeproducts"
)

// The JSON families share JSONTypes and these input-keyed archives. identity asks
// for three base archives even for Node-only witnesses; a fixture can also turn
// on closure features. Fetch every actual request in a product unit, before any
// family runs, rather than building an archive in the first shard that needs it.
// The existing group-0 runtime products remain the sole builders under the
// trace. Archive refuses a new feature set from another group until its owner
// is explicitly added to this product's order dependencies.
func TestProduct_JSONOracleRuntimes(t *testing.T) {
	t.Parallel()
	sets := map[string][]string{"": nil}
	for _, name := range append(slices.Clone(jsonDecodeFixtures), jsonEncodeFixtures...) {
		if jsonDecodeEnumerationRefused(name) {
			continue
		}
		path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		features := native.Features(native.C(program))
		sets[strings.Join(features, ",")] = features
	}
	keys := make([]string, 0, len(sets))
	for key := range sets {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		features := sets[key]
		for _, variant := range []struct{ name, owner string }{{"sanitized", "Sanitized"}, {"release", "Release"}, {"counted", "Counted"}} {
			// identity's counted archive has no features.
			if key != "" && variant.name == "counted" {
				continue
			}
			runtimeproducts.Archive(t, variant.name, features, "internal/oracle TestProduct_NativeRuntimes"+variant.owner, 0)
		}
	}
}
