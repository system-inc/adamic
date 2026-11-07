package oracles

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	producer = true
	os.Exit(m.Run())
}
func TestProduct_RegexOracle(test *testing.T) {
	test.Parallel()
	Program(test, "testdata/oracle.go", "regex-oracle")
}
func TestProduct_RegexOptionDialects(test *testing.T) {
	test.Parallel()
	Program(test, "testdata/option_dialects.go", "regex-option-dialects")
}
func TestProduct_RegexInventory(test *testing.T) {
	test.Parallel()
	Program(test, "testdata/inventory.go", "regex-inventory")
}
