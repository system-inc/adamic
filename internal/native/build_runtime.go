package native

import "github.com/system-inc/adamic/internal/native_runtime"

// BuildWithRuntime compiles source with a supplied archive and its matching headers.
// Alternate checkout and allocator fixtures keep the same relative paths and prefix maps as Build.
func BuildWithRuntime(source string, output string, options native_runtime.Options, library string) error {
	if err := native_runtime.ValidateOptions(options); err != nil {
		return err
	}
	return buildOneWithRuntime(source, output, options, native_runtime.SourceFlags(source, options), library)
}
