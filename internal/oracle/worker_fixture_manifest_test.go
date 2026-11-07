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
func TestProduct_OracleCompilingFixtures(test *testing.T) {
	test.Parallel()
	inputs, goTestInputsError := buildcache.GoTestInputs("", "oracle.test", "./internal/oracle", nil, nil)
	if goTestInputsError != nil {
		test.Fatal(goTestInputsError)
	}
	inputs.Name = "oracle compiling fixtures"
	buildcache.Product(test, inputs, func(directory string) error {
		paths := make([]string, 0, len(fixtures))
		for _, fixture := range fixtures {
			if fixture.lowers {
				paths = append(paths, fixture.path)
			}
		}
		slices.Sort(paths)
		paths = slices.Compact(paths)
		data, encodeError := json.Marshal(paths)
		if encodeError != nil {
			return encodeError
		}
		return os.WriteFile(filepath.Join(directory, "fixtures.json"), data, 0644)
	})
}
