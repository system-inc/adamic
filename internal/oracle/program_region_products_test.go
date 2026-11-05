package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native/runtimeproducts"
)

func programRegionRuntimeShard(t *testing.T, variant, group int) {
	t.Helper()
	for i, set := range runtimeproducts.Sets {
		if i%3 == group {
			runtimeproducts.ProgramArchive(t, runtimeproducts.ProgramVariants[variant].Options, set.Features)
		}
	}
}
func programRegionBuild(t *testing.T, code, output string, options native.Options) error {
	t.Helper()
	if !options.ProgramRegion {
		return native.Build(code, output, options)
	}
	library := filepath.Join(runtimeproducts.ProgramArchive(t, options, native.Features(code)), "runtime.a")
	return native.BuildWithRuntime(code, output, options, library)
}
func TestProduct_ProgramRuntimesSanitized0(t *testing.T) {
	t.Parallel()
	programRegionRuntimeShard(t, 0, 0)
}
func TestProduct_ProgramRuntimesSanitized1(t *testing.T) {
	t.Parallel()
	programRegionRuntimeShard(t, 0, 1)
}
func TestProduct_ProgramRuntimesSanitized2(t *testing.T) {
	t.Parallel()
	programRegionRuntimeShard(t, 0, 2)
}
func TestProduct_ProgramRuntimesRelease0(t *testing.T) {
	t.Parallel()
	programRegionRuntimeShard(t, 1, 0)
}
func TestProduct_ProgramRuntimesRelease1(t *testing.T) {
	t.Parallel()
	programRegionRuntimeShard(t, 1, 1)
}
func TestProduct_ProgramRuntimesRelease2(t *testing.T) {
	t.Parallel()
	programRegionRuntimeShard(t, 1, 2)
}
func TestProduct_ProgramRuntimesCounted0(t *testing.T) {
	t.Parallel()
	programRegionRuntimeShard(t, 2, 0)
}
func TestProduct_ProgramRuntimesCounted1(t *testing.T) {
	t.Parallel()
	programRegionRuntimeShard(t, 2, 1)
}
func TestProduct_ProgramRuntimesCounted2(t *testing.T) {
	t.Parallel()
	programRegionRuntimeShard(t, 2, 2)
}
