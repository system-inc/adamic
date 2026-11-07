package native

// BuildWithRuntime compiles source with a supplied archive and its matching headers.
// Alternate checkout and allocator fixtures keep the same relative paths and prefix maps as Build.
func BuildWithRuntime(source string, output string, options Options, library string) error {
	if validateOptionsError := ValidateOptions(options); validateOptionsError != nil {
		return validateOptionsError
	}
	return buildOneWithRuntime(source, output, options, SourceFlags(source, options), library)
}
