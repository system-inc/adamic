package formatfiles

import "testing"

// Each product is a separate build-phase unit. The recipes and cache keys are
// shared with shard preparation; no unit relies on another test having run.
func TestProduct_FormatfilesGoOracle(t *testing.T) {
	t.Parallel()
	formatfilesOracle(t)
}

func TestProduct_FormatfilesNative(t *testing.T) {
	t.Parallel()
	formatfilesPreparedPort(t)
}

// Used by the macOS leak check; building it also works on Linux.
func TestProduct_FormatfilesNativeUnsanitized(t *testing.T) {
	t.Parallel()
	formatfilesUnsanitized(t)
}

func TestProduct_FormatfilesNativeAdamicFilesDeclined(t *testing.T) {
	t.Parallel()
	formatfilesPreparedMutant(t, 0)
}

func TestProduct_FormatfilesNativeBareAdamicName(t *testing.T) {
	t.Parallel()
	formatfilesPreparedMutant(t, 1)
}

func TestProduct_FormatfilesNativeUppercaseAdamicExtension(t *testing.T) {
	t.Parallel()
	formatfilesPreparedMutant(t, 2)
}

func TestProduct_FormatfilesNativeJavaScriptLowercase(t *testing.T) {
	t.Parallel()
	formatfilesPreparedMutant(t, 3)
}

func TestProduct_FormatfilesNativeUTF16Sort(t *testing.T) {
	t.Parallel()
	formatfilesPreparedMutant(t, 4)
}

func TestProduct_FormatfilesNativeLinkLoop(t *testing.T) {
	t.Parallel()
	formatfilesPreparedMutant(t, 5)
}

func TestProduct_FormatfilesNativeQuoteDEL(t *testing.T) {
	t.Parallel()
	formatfilesPreparedMutant(t, 6)
}

func TestProduct_FormatfilesNativeQuoteMark(t *testing.T) {
	t.Parallel()
	formatfilesPreparedMutant(t, 7)
}

func TestProduct_FormatfilesNativeDanglingLink(t *testing.T) {
	t.Parallel()
	formatfilesPreparedMutant(t, 8)
}

func TestProduct_FormatfilesNativeTrailingDot(t *testing.T) {
	t.Parallel()
	formatfilesPreparedMutant(t, 9)
}

func TestProduct_FormatfilesNativeDanglingGit(t *testing.T) {
	t.Parallel()
	formatfilesPreparedMutant(t, 10)
}

func TestProduct_FormatfilesNativeDanglingPrettierignore(t *testing.T) {
	t.Parallel()
	formatfilesPreparedMutant(t, 11)
}
