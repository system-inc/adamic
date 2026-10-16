package oracles

import (
	"os"
	"testing"
)

// Not parallel: establishes the shared helpers' tsprinter-relative working directory and grants this compiler-free
// binary ownership of the oracle products.
func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	EnableProductProduction()
	os.Exit(m.Run())
}
