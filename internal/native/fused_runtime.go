package native

// buildFusedRuntimeComparison reproduces contraction in a comparison Node runtime.
// Program arithmetic stays unfused. The archive key holds the complete source snapshot and flags.
func buildFusedRuntimeComparison(source, output string) error {
	options := Options{}
	files, flags, compiler, version, err := runtimeInputs("", options, featureFlags(source))
	if err != nil {
		return err
	}
	flags = append(flags, "-ffp-contract=on", "-DADAMIC_FUSED_RUNTIME")
	cache, err := cacheDirectory("native-runtime")
	if err != nil {
		return err
	}
	library, err := cachedRuntime(files, flags, compiler, version, cache)
	if err != nil {
		return err
	}
	return buildOneWithRuntime(source, output, options, sourceFlags(source, options), library)
}
