package main

import (
	"os"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// prepare is a subprocess engine over this checkout, running the shared stage 0 compiler product.
func prepare(t testing.TB, test262 string, work string) (*engine, error) {
	t.Helper()
	return prepareMode("../..", test262, work, productAdamic(t), nil, false)
}

// productAdamic links the shared stage 0 compiler (buildcache.Adamic) where an engine wants it, so no test's engine
// runs go build.
func productAdamic(t testing.TB) adamicBuilder {
	t.Helper()
	adamic := buildcache.Adamic(t)
	return func(output string) error { return os.Symlink(adamic, output) }
}
