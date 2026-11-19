package load

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
)

// IsLibrary's prefix, read once, is the library path itself (#715zjph). Mutant: libraryPrefix returning "", which
// makes every file a library file, and this goes red.
func TestLibraryPrefixIsTheLibraryPath(t *testing.T) {
	t.Parallel()
	if got, want := libraryPrefix(), bundled.LibPath().AsString(); got != want || got == "" {
		t.Fatalf("libraryPrefix() = %q, want bundled.LibPath() %q", got, want)
	}
}
