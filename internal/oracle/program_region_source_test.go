package oracle

import (
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/programsource"
)

func programRegionSource(t *testing.T) string {
	t.Helper()
	inputs, build := programsource.Request(t, "../..")
	return buildcache.Product(t, inputs, build)
}
func TestProduct_ProgramTscSource(t *testing.T) { t.Parallel(); programRegionSource(t) }
