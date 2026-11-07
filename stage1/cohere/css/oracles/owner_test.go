package oracles

import (
	"os"
	"testing"
)

// Not parallel: establishes the shared helpers' css-relative working directory,
// grants this compiler-free binary ownership, and cleans shared build inputs.
func TestMain(suite *testing.M) {
	if chdirError := os.Chdir(".."); chdirError != nil {
		panic(chdirError)
	}
	EnableProductProduction()
	code := suite.Run()
	if cleanupTemporaryDirectoriesError := CleanupTemporaryDirectories(); cleanupTemporaryDirectoriesError != nil &&
		code == 0 {
		code = 1
	}
	os.Exit(code)
}
