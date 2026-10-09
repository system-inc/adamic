// Package vendored is the installed tailwindcss the Tailwind rules' tests build a design system against: a
// snapshot of tailwindcss 4.3.3 in testdata, so those tests run on any machine instead of only where ahra's
// node_modules happens to be (#sycrdr6).
//
// Only the files the design-system loader reads are kept: index.css and the stylesheets it imports
// (theme.css, preflight.css, utilities.css), with package.json and the package's MIT LICENSE beside them.
// See testdata/README.md for the source and how to refresh it.
package vendored

import (
	"path/filepath"
	"runtime"
)

// TailwindVersion is the version of tailwindcss the snapshot holds.
const TailwindVersion = "4.3.3"

// TailwindPackageRoot is the snapshot's directory, the root a node_modules/tailwindcss would be.
func TailwindPackageRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", "tailwindcss")
}
