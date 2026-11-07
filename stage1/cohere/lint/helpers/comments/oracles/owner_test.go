package oracles

import (
	"os"
	"testing"
)

func TestMain(suite *testing.M) {
	if chdirError := os.Chdir(".."); chdirError != nil {
		panic(chdirError)
	}
	producer = true
	os.Exit(suite.Run())
}
func TestProduct_CommentsOracle(test *testing.T) { test.Parallel(); Comments(test) }
