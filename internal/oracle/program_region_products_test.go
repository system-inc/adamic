package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native/runtimeproducts"
)

func programRegionRuntimeShard(test *testing.T, variant, group int) {
	test.Helper()
	for index, set := range runtimeproducts.Sets {
		if index%3 == group {
			runtimeproducts.ProgramArchive(test, runtimeproducts.ProgramVariants[variant].Options, set.Features)
		}
	}
}
func programRegionBuild(test *testing.T, code, output string, options native.Options) error {
	test.Helper()
	if !options.ProgramRegion {
		return native.Build(code, output, options)
	}
	library := filepath.Join(runtimeproducts.ProgramArchive(test, options, native.Features(code)), "runtime.a")
	return native.BuildWithRuntime(code, output, options, library)
}
func TestProduct_ProgramRuntimesSanitized0(test *testing.T) {
	test.Parallel()
	programRegionRuntimeShard(test, 0, 0)
}
func TestProduct_ProgramRuntimesSanitized1(test *testing.T) {
	test.Parallel()
	programRegionRuntimeShard(test, 0, 1)
}
func TestProduct_ProgramRuntimesSanitized2(test *testing.T) {
	test.Parallel()
	programRegionRuntimeShard(test, 0, 2)
}
func TestProduct_ProgramRuntimesRelease0(test *testing.T) {
	test.Parallel()
	programRegionRuntimeShard(test, 1, 0)
}
func TestProduct_ProgramRuntimesRelease1(test *testing.T) {
	test.Parallel()
	programRegionRuntimeShard(test, 1, 1)
}
func TestProduct_ProgramRuntimesRelease2(test *testing.T) {
	test.Parallel()
	programRegionRuntimeShard(test, 1, 2)
}
func TestProduct_ProgramRuntimesCounted0(test *testing.T) {
	test.Parallel()
	programRegionRuntimeShard(test, 2, 0)
}
func TestProduct_ProgramRuntimesCounted1(test *testing.T) {
	test.Parallel()
	programRegionRuntimeShard(test, 2, 1)
}
func TestProduct_ProgramRuntimesCounted2(test *testing.T) {
	test.Parallel()
	programRegionRuntimeShard(test, 2, 2)
}
