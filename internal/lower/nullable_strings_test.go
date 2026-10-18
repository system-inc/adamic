package lower

import (
	"testing"
)

func TestNullableStringAdmission(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function read(x: string | null): string {return x === null ? 'none' : x;} console.log(read(null));`,
		`function read(x: string | null | undefined): string {return x ?? 'none';} console.log(read(undefined));`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}
