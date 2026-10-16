package lower

import (
	"os"
	"testing"
)

func fx6Source(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile("testdata/fx6/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func TestFX6P35(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, fx6Source(t, "p35"))
}

func TestFX6P36(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, fx6Source(t, "p36"))
}

func TestFX6P37(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, fx6Source(t, "p37"))
}

func TestFX6P38(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, fx6Source(t, "p38"))
}

func TestFX6P51(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, fx6Source(t, "p51"))
}

func TestFX6P53(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, fx6Source(t, "p53"))
}

func TestFX6P38Number(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, fx6Source(t, "p38_number"))
}
