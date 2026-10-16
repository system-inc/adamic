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
func TestProduct_CSSStringsGoOracle(t *testing.T) { t.Parallel(); Printer(t) }
func TestProduct_CSSStringsGoAnswers(t *testing.T) {
	t.Parallel()
	c := Enumerate(t)
	p := Printer(t)
	FullAnswers(t, p, nil, Input(c.Texts), Identity(t, p))
}
func TestProduct_CSSStringsGoBatchAnswers(t *testing.T) {
	t.Parallel()
	c := Enumerate(t)
	p := Printer(t)
	BatchAnswers(t, p, Identity(t, p), BatchCases(c))
}
func TestProduct_CSSStringsGoRawAnswers(t *testing.T) {
	t.Parallel()
	c := Enumerate(t)
	p := Printer(t)
	RawAnswers(t, p, Identity(t, p), RawCases(c))
}
