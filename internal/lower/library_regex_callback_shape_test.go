package lower

import (
	"testing"
)

func TestLibraryRegexOffsetRequiresNoCaptures(t *testing.T) {
	for _, source := range []string{
		"console.log('a'.replace(/(a)/, (match: string, offset: number) => `${match}:${offset}`));",
		"console.log('a'.replace(/(?=(a))a/, (match: string, offset: number) => `${match}:${offset}`));",
		"function replace(pattern: RegExp): string { return 'a'.replace(pattern, (match: string, offset: number) => `${match}:${offset}`); }",
		"let pattern = /a/; console.log('a'.replace(pattern, (match: string, offset: number) => `${match}:${offset}`));",
	} {
		_, err := lowerSource(t, source)
		if err == nil {
			t.Fatalf("want capture-free producer stop, got %v", err)
		}
	}
}
