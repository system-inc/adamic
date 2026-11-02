package native

// BuildWithRuntimeForTest is Build's one-unit path against the runtime archive library, as a mutant runtime's probe
// needs it: compiled and linked through clangIn with the build directory and the runtime mapped out, so the probe's
// bytes are a function of its inputs, as every product's are (timers_test.go's timerRuntime.build).
func BuildWithRuntimeForTest(source string, output string, options Options, library string) error {
	if err := validateBuild(source, options); err != nil {
		return err
	}
	return buildOneWithRuntime(source, output, options, Flags(options), library)
}
