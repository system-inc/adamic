package lower

import (
	"strings"
	"testing"
)

func TestNullableStringAdmission(t *testing.T) {
	for _, source := range []string{
		`function read(x: string | null): string {return x === null ? 'none' : x;} console.log(read(null));`,
		`function read(x: string | null | undefined): string {return x ?? 'none';} console.log(read(undefined));`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
	for _, source := range []string{
		`function read(x: number | null): number {return x === null ? 0 : x;} console.log(String(read(null)));`,
		`function read(x: boolean | null): boolean {return x === null ? false : x;} console.log(read(null) ? 'yes':'no');`,
	} {
		if _, err := lowerSource(t, source); err == nil || !strings.Contains(err.Error(), "a value of type") {
			t.Fatalf("unimplemented nullable scalar must remain stopped, got %v", err)
		}
	}
}
