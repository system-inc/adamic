package oracles

import (
	"os"
	"testing"
)

// Not parallel: establish the estree-relative working directory and production ownership before tests run.
func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	EnableProductProduction()
	os.Exit(m.Run())
}
