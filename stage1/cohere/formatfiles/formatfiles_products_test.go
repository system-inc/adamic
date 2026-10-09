package formatfiles

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Each product is a separate build-phase unit. The recipes and cache keys are
// shared with shard preparation; no unit relies on another test having run.
func TestProduct_FormatfilesGoOracle(t *testing.T) {
	t.Parallel()
	binary, _ := formatfilesOracle(t)
	buildcache.RequireArtifacts(t, filepath.Dir(binary), buildcache.Artifact{Name: filepath.Base(binary), Executable: true, Arguments: []string{"-test.list=^$"}})
}

func TestProduct_FormatfilesNative(t *testing.T) {
	t.Parallel()
	product := formatfilesPreparedPort(t)
	requireFormatfilesBinary(t, product.binary)
	buildcache.RequireArtifacts(t, filepath.Dir(product.javascript), buildcache.Artifact{Name: filepath.Base(product.javascript)})
}

// Used by the macOS leak check; building it also works on Linux.
func TestProduct_FormatfilesNativeUnsanitized(t *testing.T) {
	t.Parallel()
	requireFormatfilesBinary(t, formatfilesUnsanitized(t))
}

func TestProduct_FormatfilesNativeAdamicFilesDeclined(t *testing.T) {
	t.Parallel()
	requireFormatfilesBinary(t, formatfilesPreparedMutant(t, 0).binary)
}

func TestProduct_FormatfilesNativeBareAdamicName(t *testing.T) {
	t.Parallel()
	requireFormatfilesBinary(t, formatfilesPreparedMutant(t, 1).binary)
}

func TestProduct_FormatfilesNativeUppercaseAdamicExtension(t *testing.T) {
	t.Parallel()
	requireFormatfilesBinary(t, formatfilesPreparedMutant(t, 2).binary)
}

func TestProduct_FormatfilesNativeJavaScriptLowercase(t *testing.T) {
	t.Parallel()
	requireFormatfilesBinary(t, formatfilesPreparedMutant(t, 3).binary)
}

func TestProduct_FormatfilesNativeUTF16Sort(t *testing.T) {
	t.Parallel()
	requireFormatfilesBinary(t, formatfilesPreparedMutant(t, 4).binary)
}

func TestProduct_FormatfilesNativeLinkLoop(t *testing.T) {
	t.Parallel()
	requireFormatfilesBinary(t, formatfilesPreparedMutant(t, 5).binary)
}

func TestProduct_FormatfilesNativeQuoteDEL(t *testing.T) {
	t.Parallel()
	requireFormatfilesBinary(t, formatfilesPreparedMutant(t, 6).binary)
}

func TestProduct_FormatfilesNativeQuoteMark(t *testing.T) {
	t.Parallel()
	requireFormatfilesBinary(t, formatfilesPreparedMutant(t, 7).binary)
}

func TestProduct_FormatfilesNativeDanglingLink(t *testing.T) {
	t.Parallel()
	requireFormatfilesBinary(t, formatfilesPreparedMutant(t, 8).binary)
}

func TestProduct_FormatfilesNativeTrailingDot(t *testing.T) {
	t.Parallel()
	requireFormatfilesBinary(t, formatfilesPreparedMutant(t, 9).binary)
}

func TestProduct_FormatfilesNativeDanglingGit(t *testing.T) {
	t.Parallel()
	requireFormatfilesBinary(t, formatfilesPreparedMutant(t, 10).binary)
}

func TestProduct_FormatfilesNativeDanglingPrettierignore(t *testing.T) {
	t.Parallel()
	requireFormatfilesBinary(t, formatfilesPreparedMutant(t, 11).binary)
}

func requireFormatfilesBinary(t *testing.T, binary string) {
	t.Helper()
	manifest := filepath.Join(t.TempDir(), "empty-manifest")
	if err := os.WriteFile(manifest, nil, 0600); err != nil {
		t.Fatal(err)
	}
	buildcache.RequireArtifacts(t, filepath.Dir(binary), buildcache.Artifact{Name: filepath.Base(binary), Executable: true, Arguments: []string{manifest}})
}
