package regex

import "testing"

func TestProduct_RegexFindsOracle(t *testing.T) {
	t.Parallel()
	cohereProgram(t, "testdata/oracle.go", "regex-oracle")
}

func TestProduct_RegexFindsOptionDialects(t *testing.T) {
	t.Parallel()
	cohereProgram(t, "testdata/option_dialects.go", "regex-option-dialects")
}

func TestProduct_RegexFindsInventory(t *testing.T) {
	t.Parallel()
	cohereProgram(t, "testdata/inventory.go", "regex-inventory")
}

func TestProduct_RegexShapesGate(t *testing.T) {
	t.Parallel()
	shapesGate(t)
}

func TestProduct_RegexRunner(t *testing.T) {
	t.Parallel()
	runnerProduct(t, false)
}

func TestProduct_RegexRunnerMutant(t *testing.T) {
	t.Parallel()
	runnerProduct(t, true)
}
