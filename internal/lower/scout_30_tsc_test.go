package lower

import "testing"

func TestScout30GlobalNumericPredicates(t *testing.T) {
	for _, source := range []string{
		`console.log(String(isFinite(1))); console.log(String(isNaN(NaN)));`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}
