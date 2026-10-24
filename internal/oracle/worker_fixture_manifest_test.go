package oracle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// The registry is assembled by this package's init functions. Export its data,
// rather than duplicating the table or parsing just the literal in oracle_test.go.
func TestProduct_OracleCompilingFixtures(t *testing.T) {
	t.Parallel()
	inputs, err := buildcache.GoTestInputs("", "oracle.test", "./internal/oracle", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	inputs.Name = "oracle compiling fixtures"
	buildcache.Product(t, inputs, func(directory string) error {
		paths := make([]string, 0, len(fixtures))
		for _, fixture := range fixtures {
			if fixture.lowers {
				paths = append(paths, fixture.path)
			}
		}
		slices.Sort(paths)
		paths = slices.Compact(paths)
		data, err := json.Marshal(paths)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "fixtures.json"), data, 0644)
	})
}
