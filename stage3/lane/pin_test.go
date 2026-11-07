package lane

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// typescriptPin is the commit stage3/source.json pins. The adapted tree's build keeps its own
// (stage3/lane/adaptedtree), and lane keeps this one rather than import that package, since a lane test binary linking
// it would key lane's units on the adapted tree's build code.
func typescriptPin(test testing.TB, repository string) string {
	test.Helper()
	data, err := os.ReadFile(filepath.Join(repository, "stage3/source.json"))
	if err != nil {
		test.Fatal(err)
	}
	var source struct {
		Commit string `json:"commit"`
	}
	if err := json.Unmarshal(data, &source); err != nil {
		test.Fatal(err)
	}
	return source.Commit
}
