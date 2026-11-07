package oracles

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	EnableProductProduction()
	os.Exit(m.Run())
}
func TestProduct_CSSNumbersGoOracle(test *testing.T) { test.Parallel(); Printer(test) }
func TestProduct_CSSNumbersGoAnswers(test *testing.T) {
	test.Parallel()
	c := Enumerate(test)
	p := Printer(test)
	FullAnswers(test, p, nil, Input(c.Texts), Identity(test, p))
}
func TestProduct_CSSNumbersGoBatchAnswers(test *testing.T) {
	test.Parallel()
	c := Enumerate(test)
	p := Printer(test)
	BatchAnswers(test, p, Identity(test, p), BatchCases(c))
}
func TestProduct_CSSNumbersGoRawAnswers(test *testing.T) {
	test.Parallel()
	c := Enumerate(test)
	p := Printer(test)
	RawAnswers(test, p, Identity(test, p), RawCases(c))
}
