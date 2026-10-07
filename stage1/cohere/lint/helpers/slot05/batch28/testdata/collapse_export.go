package tailwind

import "reflect"

func AdamicPredicate(kind, value string) (string, bool) {
	p := bareValuePredicate(BareValueKind(kind))
	if p == nil {
		return "none", false
	}
	for _, entry := range []struct {
		name string
		f    func(string) bool
	}{
		{"positive", isPositiveInteger}, {"opacity", isValidOpacityValue}, {"strict", isStrictPositiveInteger}, {"spacing", isValidSpacingMultiplier}, {"fontStretch", isFontStretchPercentage},
	} {
		if reflect.ValueOf(p).Pointer() == reflect.ValueOf(entry.f).Pointer() {
			return entry.name, p(value)
		}
	}
	panic("unknown predicate identity")
}
func AdamicTransform(kind, value string) string {
	return bareValueTransform(BareValueKind(kind), value)
}
func AdamicOpacity(value string) bool                   { return isValidOpacityValue(value) }
func AdamicMultiple(value string, divisor float64) bool { return isMultipleOf(value, divisor) }
func AdamicDependencies(value string) []bool {
	return []bool{isPositiveInteger(value), isValidOpacityValue(value), isStrictPositiveInteger(value), isValidSpacingMultiplier(value), isFontStretchPercentage(value)}
}
