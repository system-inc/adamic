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
func TestProduct_RegexOracle(t *testing.T) {
	t.Parallel()
	Program(t, "testdata/oracle.go", "regex-oracle")
}
func TestProduct_RegexOptionDialects(t *testing.T) {
	t.Parallel()
	Program(t, "testdata/option_dialects.go", "regex-option-dialects")
}
func TestProduct_RegexInventory(t *testing.T) {
	t.Parallel()
	Program(t, "testdata/inventory.go", "regex-inventory")
}
