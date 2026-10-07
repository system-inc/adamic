package tailwind

func AdamicDecode(input string) ([]ValueNode, []ValueNode, string, string, string) {
	original := ParseValue(input)
	decoded := ParseValue(input)
	recursivelyDecodeArbitraryValues(decoded)
	css := ValueToCss(decoded)
	return original, decoded, css, addWhitespaceAroundMathOperators(css), decodeArbitraryValue(input)
}
func AdamicRecursive(nodes []ValueNode) { recursivelyDecodeArbitraryValues(nodes) }
func AdamicBreakpoint(registry *VariantRegistry, theme *Theme, order int, present bool) {
	original := FrameworkVariantRegistrations
	defer func() { FrameworkVariantRegistrations = original }()
	FrameworkVariantRegistrations = nil
	if present {
		FrameworkVariantRegistrations = []FrameworkVariantRegistration{{Name: "sm", Order: order, Kind: ParsedVariantKindStatic}}
	}
	registerThemeBreakpointVariants(registry, theme)
}
