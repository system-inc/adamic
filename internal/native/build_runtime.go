package native

// BuildWithRuntime compiles source with a supplied archive and its matching headers.
// Alternate checkout and allocator fixtures keep the same relative paths and prefix maps as Build.
func BuildWithRuntime(source string, output string, options Options, library string) error {
	if err := ValidateOptions(options); err != nil {
		return err
	}
	return buildOneWithRuntime(source, output, options, SourceFlags(source, options), library)
}
