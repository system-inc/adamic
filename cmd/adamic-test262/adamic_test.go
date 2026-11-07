package main

import (
	"os"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// prepare is a subprocess engine over this checkout, running the shared stage 0 compiler product.
func prepare(test testing.TB, test262 string, work string) (*engine, error) {
	test.Helper()
	// Reads: internal/lower/*.go
	return prepareMode("../..", test262, work, productAdamic(test), nil, false)
}

// productAdamic links the shared stage 0 compiler (buildcache.Adamic) where an engine wants it, so no test's engine
// runs go build.
func productAdamic(test testing.TB) adamicBuilder {
	test.Helper()
	adamic := buildcache.Adamic(test)
	return func(output string) error { return os.Symlink(adamic, output) }
}

func TestProduct_Adamic(test *testing.T) {
	test.Parallel()
	buildcache.AdamicBuiltElsewhere(test)
}
