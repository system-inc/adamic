package regex

import "testing"

func TestProduct_RegexOracle(t *testing.T) {
	t.Parallel()
	cohereProgram(t, "testdata/oracle.go", "regex-oracle")
}

func TestProduct_RegexOptionDialects(t *testing.T) {
	t.Parallel()
	cohereProgram(t, "testdata/option_dialects.go", "regex-option-dialects")
}

func TestProduct_RegexInventory(t *testing.T) {
	t.Parallel()
	cohereProgram(t, "testdata/inventory.go", "regex-inventory")
}

func TestProduct_RegexShapesGate(t *testing.T) {
	t.Parallel()
	shapesGate(t)
}
