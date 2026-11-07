package oracles

import (
	"os"
	"testing"
)

// Not parallel: establish the estree-relative working directory and production ownership before tests run.
func TestMain(suite *testing.M) {
	if chdirError := os.Chdir(".."); chdirError != nil {
		panic(chdirError)
	}
	EnableProductProduction()
	os.Exit(suite.Run())
}
