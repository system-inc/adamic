package sourcename_test

import (
	"testing"

	"github.com/system-inc/cohere/internal/types/sourcename"
)

// An Adamic name reads as the same name ending in .ts, and every other name as itself: a directory, a name
// that is only the extension, and names that merely contain ".a" are not Adamic.
func TestAnAdamicNameIsTreatedAsTypeScript(t *testing.T) {
	t.Parallel()
	for name, want := range map[string]string{
		"/repo/Dog.a":         "/repo/Dog.ts",
		"Dog.test.a":          "Dog.test.ts",
		"/repo/geometry.a":    "/repo/geometry.ts",
		"/repo/Dog.ts":        "/repo/Dog.ts",
		"/repo/Dog.tsx":       "/repo/Dog.tsx",
		"/repo/data.archive":  "/repo/data.archive",
		"/repo/a":             "/repo/a",
		".a":                  ".a",
		"/repo/.a":            "/repo/.a",
		"/repo/Dog.d.ts":      "/repo/Dog.d.ts",
		"/repo/lib/libfoo.ab": "/repo/lib/libfoo.ab",
	} {
		if got := sourcename.TreatedAs(name); got != want {
			t.Errorf("TreatedAs(%q) = %q, want %q", name, got, want)
		}
		if got := sourcename.IsAdamic(name); got != (want != name) {
			t.Errorf("IsAdamic(%q) = %t", name, got)
		}
	}
}
