package tailwind

func AdamicSplitThemeKey(key string) []string     { return splitThemeKey(key) }
func AdamicJoinSegments(segments []string) string { return joinSegments(segments) }
func AdamicBreakpointGroupOrder(registrations []FrameworkVariantRegistration) (int, bool) {
	original := FrameworkVariantRegistrations
	FrameworkVariantRegistrations = registrations
	defer func() { FrameworkVariantRegistrations = original }()
	return breakpointGroupOrder()
}
