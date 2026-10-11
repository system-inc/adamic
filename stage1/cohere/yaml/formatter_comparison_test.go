package yaml

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// formatterComparison is cohere's command/formatter_comparison with testdata/<adapter> laid over its main by an
// overlay, and mutations planted in cohere's sources: every Go oracle of this package, a product built ahead under
// output and keyed through the overlay's map, the adapter's content and each mutation (#5qykzj5).
func formatterComparison(t *testing.T, output, adapter string, mutations ...buildcache.Mutation) string {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	source, err := filepath.Abs(filepath.Join("testdata", adapter))
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "cohere/command/formatter_comparison/main.go"): source}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	arguments := []string{"-overlay=" + path}
	if len(mutations) == 0 {
		return buildcache.GoBuildIn(t, "cohere", output, "./command/formatter_comparison", arguments)
	}
	return buildcache.GoBuildInMutated(t, "cohere", output, "./command/formatter_comparison", arguments, mutations)
}

// numbersExports are the two plantings that open cohere's compose package to an adapter: imports added at the top of
// numbers.go's import block, and testdata/<exports> appended at its end, after the file's last top-level function.
func numbersExports(t *testing.T, imports, exports string) []buildcache.Mutation {
	t.Helper()
	const numbers = "cohere/internal/format/yaml/compose/numbers.go"
	original, err := os.ReadFile(filepath.Join(repository, numbers))
	if err != nil {
		t.Fatal(err)
	}
	appended, err := os.ReadFile(filepath.Join("testdata", exports))
	if err != nil {
		t.Fatal(err)
	}
	last := strings.LastIndex(string(original), "\nfunc ")
	if last < 0 {
		t.Fatalf("%s has no top-level function to append %s after", numbers, exports)
	}
	tail := string(original[last:])
	return []buildcache.Mutation{
		{File: numbers, Before: "import (", After: "import (" + imports},
		{File: numbers, Before: tail, After: tail + string(appended)},
	}
}
