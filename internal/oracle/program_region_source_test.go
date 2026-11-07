package oracle

import (
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/programsource"
)

func programRegionSource(test *testing.T) string {
	test.Helper()
	inputs, build := programsource.Request(test, "../..")
	return buildcache.Product(test, inputs, build)
}
func TestProduct_ProgramTscSource(test *testing.T) { test.Parallel(); programRegionSource(test) }
