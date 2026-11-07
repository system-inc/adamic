package lower

import (
	"testing"
)

func TestNullableStringAdmission(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`function read(x: string | null): string {return x === null ? 'none' : x;} console.log(read(null));`,
		`function read(x: string | null | undefined): string {return x ?? 'none';} console.log(read(undefined));`,
	} {
		if _, err := lowerSource(test, source); err != nil {
			test.Fatal(err)
		}
	}
}
