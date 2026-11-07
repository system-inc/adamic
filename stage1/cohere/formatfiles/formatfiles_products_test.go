package formatfiles

import (
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"path/filepath"
	"testing"
)

// Each product is a separate build-phase unit. The recipes and cache keys are
// shared with shard preparation; no unit relies on another test having run.
func TestProduct_FormatfilesFindsGoOracle(test *testing.T) {
	test.Parallel()
	binary, _ := formatfilesOracle(test)
	buildcache.RequireArtifacts(
		test,
		filepath.Dir(binary),
		buildcache.Artifact{Name: filepath.Base(binary), Executable: true, Arguments: []string{"-test.list=^$"}},
	)
}

func TestProduct_FormatfilesNative(test *testing.T) {
	test.Parallel()
	product := formatfilesPreparedPort(test)
	requireFormatfilesBinary(test, product.binary)
	buildcache.RequireArtifacts(
		test,
		filepath.Dir(product.javascript),
		buildcache.Artifact{Name: filepath.Base(product.javascript)},
	)
}

// Used by the macOS leak check; building it also works on Linux.
func TestProduct_FormatfilesNativeUnsanitized(test *testing.T) {
	test.Parallel()
	requireFormatfilesBinary(test, formatfilesUnsanitized(test))
}

func TestProduct_FormatfilesNativeAdamicFilesDeclined(test *testing.T) {
	test.Parallel()
	requireFormatfilesBinary(test, formatfilesPreparedMutant(test, 0).binary)
}

func TestProduct_FormatfilesNativeBareAdamicName(test *testing.T) {
	test.Parallel()
	requireFormatfilesBinary(test, formatfilesPreparedMutant(test, 1).binary)
}

func TestProduct_FormatfilesNativeUppercaseAdamicExtension(test *testing.T) {
	test.Parallel()
	requireFormatfilesBinary(test, formatfilesPreparedMutant(test, 2).binary)
}

func TestProduct_FormatfilesNativeJavaScriptLowercase(test *testing.T) {
	test.Parallel()
	requireFormatfilesBinary(test, formatfilesPreparedMutant(test, 3).binary)
}

func TestProduct_FormatfilesNativeUTF16Sort(test *testing.T) {
	test.Parallel()
	requireFormatfilesBinary(test, formatfilesPreparedMutant(test, 4).binary)
}

func TestProduct_FormatfilesNativeLinkLoop(test *testing.T) {
	test.Parallel()
	requireFormatfilesBinary(test, formatfilesPreparedMutant(test, 5).binary)
}

func TestProduct_FormatfilesNativeQuoteDEL(test *testing.T) {
	test.Parallel()
	requireFormatfilesBinary(test, formatfilesPreparedMutant(test, 6).binary)
}

func TestProduct_FormatfilesNativeQuoteMark(test *testing.T) {
	test.Parallel()
	requireFormatfilesBinary(test, formatfilesPreparedMutant(test, 7).binary)
}

func TestProduct_FormatfilesNativeDanglingLink(test *testing.T) {
	test.Parallel()
	requireFormatfilesBinary(test, formatfilesPreparedMutant(test, 8).binary)
}

func TestProduct_FormatfilesNativeTrailingDot(test *testing.T) {
	test.Parallel()
	requireFormatfilesBinary(test, formatfilesPreparedMutant(test, 9).binary)
}

func TestProduct_FormatfilesNativeDanglingGit(test *testing.T) {
	test.Parallel()
	requireFormatfilesBinary(test, formatfilesPreparedMutant(test, 10).binary)
}

func TestProduct_FormatfilesNativeDanglingPrettierignore(test *testing.T) {
	test.Parallel()
	requireFormatfilesBinary(test, formatfilesPreparedMutant(test, 11).binary)
}

func requireFormatfilesBinary(test *testing.T, binary string) {
	test.Helper()
	manifest := filepath.Join(test.TempDir(), "empty-manifest")
	if writeError := os.WriteFile(manifest, nil, 0600); writeError != nil {
		test.Fatal(writeError)
	}
	buildcache.RequireArtifacts(
		test,
		filepath.Dir(binary),
		buildcache.Artifact{Name: filepath.Base(binary), Executable: true, Arguments: []string{manifest}},
	)
}
