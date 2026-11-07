package regex

import "testing"

func TestProduct_RegexFindsOracle(test *testing.T) {
	test.Parallel()
	cohereProgram(test, "testdata/oracle.go", "regex-oracle")
}

func TestProduct_RegexFindsOptionDialects(test *testing.T) {
	test.Parallel()
	cohereProgram(test, "testdata/option_dialects.go", "regex-option-dialects")
}

func TestProduct_RegexFindsInventory(test *testing.T) {
	test.Parallel()
	cohereProgram(test, "testdata/inventory.go", "regex-inventory")
}

func TestProduct_RegexShapesGate(test *testing.T) {
	test.Parallel()
	shapesGate(test)
}

func TestProduct_RegexRunner(test *testing.T) {
	test.Parallel()
	runnerProduct(test, false)
}

func TestProduct_RegexRunnerMutant(test *testing.T) {
	test.Parallel()
	runnerProduct(test, true)
}
