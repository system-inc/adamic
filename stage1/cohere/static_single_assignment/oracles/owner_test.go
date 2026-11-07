package oracles

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if chdirError := os.Chdir(".."); chdirError != nil {
		panic(chdirError)
	}
	producer = true
	os.Exit(m.Run())
}
func TestProduct_StaticSingleAssignmentGoOracle(test *testing.T) {
	test.Parallel()
	StaticSingleAssignment(test)
}
