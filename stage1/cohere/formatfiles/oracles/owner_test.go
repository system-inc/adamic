package oracles

import (
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"path/filepath"
	"testing"
)

// Not parallel: TestMain sets the parent working directory before any test runs.
func TestMain(suite *testing.M) {
	if chdirError := os.Chdir(".."); chdirError != nil {
		panic(chdirError)
	}
	producer = true
	os.Exit(suite.Run())
}
func TestProduct_FormatfilesGoOracle(test *testing.T) {
	test.Parallel()
	binary, _ := GoOracle(test)
	buildcache.RequireArtifacts(
		test,
		filepath.Dir(binary),
		buildcache.Artifact{Name: filepath.Base(binary), Executable: true, Arguments: []string{"-test.list=^$"}},
	)
}
