package oracles

import (
	"os"
	"testing"
)

func TestMain(suite *testing.M) {
	if chdirError := os.Chdir(".."); chdirError != nil {
		panic(chdirError)
	}
	EnableProductProduction()
	os.Exit(suite.Run())
}
func TestProduct_CSSStringsGoOracle(test *testing.T) { test.Parallel(); Printer(test) }
func TestProduct_CSSStringsGoAnswers(test *testing.T) {
	test.Parallel()
	corpus := Enumerate(test)
	p := Printer(test)
	FullAnswers(test, p, nil, Input(corpus.Texts), Identity(test, p))
}
func TestProduct_CSSStringsGoBatchAnswers(test *testing.T) {
	test.Parallel()
	corpus := Enumerate(test)
	p := Printer(test)
	BatchAnswers(test, p, Identity(test, p), BatchCases(corpus))
}
func TestProduct_CSSStringsGoRawAnswers(test *testing.T) {
	test.Parallel()
	corpus := Enumerate(test)
	p := Printer(test)
	RawAnswers(test, p, Identity(test, p), RawCases(corpus))
}
