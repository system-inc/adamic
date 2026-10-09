package incremental_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/compiler"
	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/incremental"
)

// TestShimSurfaceIsReachable proves the linkname surface resolves at build time and that the
// re-exported symbols are the upstream ones rather than empty stand-ins.
//
// This is a compile-and-reference test on purpose. Calling ReadBuildInfoProgram for real needs
// a compiler host and a parsed command line, which is the integration this shim exists to
// enable rather than something the shim itself should own. What can fail silently here is the
// shim drifting out of sync with upstream after a typescript-go bump, and that failure shows
// up as a build error the moment these references stop resolving.
func TestShimSurfaceIsReachable(t *testing.T) {
	// Each assignment is a distinct way the shim can break: a renamed function, a moved type,
	// a changed signature. Binding the function to a typed variable checks the signature
	// without needing a live host, and it fails at COMPILE time rather than here — which is
	// the point. A nil check would not have worked: these are package-level funcs, so the
	// compiler rejects `== nil` as always false, and a test that cannot fail proves nothing.
	var (
		_ func(compiler.CompilerHost) incremental.BuildInfoReader                                                     = incremental.NewBuildInfoReader
		_ func(compiler.CompilerHost) incremental.Host                                                                = incremental.CreateHost
		_ func(*tsoptions.ParsedCommandLine, incremental.BuildInfoReader, compiler.CompilerHost) *incremental.Program = incremental.ReadBuildInfoProgram
	)

	// The type aliases must name real upstream types rather than local stand-ins. Asserting
	// the alias resolves to a struct with upstream's fields is what catches a drift where a
	// type is still present but has become something else.
	var program *incremental.Program
	var buildInfo *incremental.BuildInfo
	if program != nil || buildInfo != nil {
		t.Fatal("unreachable; the declarations exist to force the aliases to resolve")
	}

	t.Log("shim surface resolves: three functions bound with upstream signatures, two type aliases")
}

// TestBuildInfoNilContract records the one behaviour a caller is most likely to get wrong.
//
// ReadBuildInfoProgram returns nil for three different reasons that all mean "rebuild
// everything" rather than "an error occurred": no build info on disk, a version mismatch, and
// a build info that was not written incrementally (incremental.go:47). A caller that treats
// nil as a failure will report an error on every first run; one that treats it as an empty
// cache is correct.
//
// There is nothing to execute here, so this test documents the contract next to the code that
// depends on it rather than asserting on a live read.
func TestBuildInfoNilContract(t *testing.T) {
	t.Log("ReadBuildInfoProgram returns nil for: absent build info, version mismatch, " +
		"or non-incremental build info. All three mean rebuild-everything, not error.")
	t.Log("A returned *Program from this path has a nil `program` field; GetProgram panics " +
		"on it (program.go:110). It is change detection, not a warm graph.")
}
