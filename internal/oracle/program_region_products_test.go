package oracle

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native/runtimeproducts"
)

var programRuntimeOptions = []native.Options{{ProgramRegion: true, Sanitize: true}, {ProgramRegion: true}, {ProgramRegion: true, Count: true}}

func programRegionRuntime(t *testing.T, options native.Options, features []string) string {
	t.Helper()
	request, err := native.RuntimeRequestFor(options, features...)
	if err != nil {
		t.Fatal(err)
	}
	jobserver, err := ccjobserver.Process()
	if err != nil {
		t.Fatal(err)
	}
	return buildcache.Product(t, runtimeproducts.Inputs(request), func(directory string) error { return request.Build(directory, 1, jobserver) })
}
func programRegionRuntimeShard(t *testing.T, variant, group int) {
	t.Helper()
	for i, set := range runtimeproducts.Sets {
		if i%3 == group {
			programRegionRuntime(t, programRuntimeOptions[variant], set.Features)
		}
	}
}
func programRegionBuild(t *testing.T, code, output string, options native.Options) error {
	t.Helper()
	if !options.ProgramRegion {
		return native.Build(code, output, options)
	}
	features := native.Features(code)
	if !slices.ContainsFunc(runtimeproducts.Sets, func(s runtimeproducts.Set) bool { return slices.Equal(features, s.Features) }) {
		t.Fatalf("undeclared Program runtime feature set: %v", features)
	}
	if !slices.Contains(programRuntimeOptions, options) {
		t.Fatalf("undeclared Program runtime options: %+v", options)
	}
	library := filepath.Join(programRegionRuntime(t, options, features), "runtime.a")
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
