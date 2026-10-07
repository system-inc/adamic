package tailwind

func AdamicGeneric(value string) bool { return isGenericName(value) }
func AdamicMath(value string) bool    { return hasMathFunction(value) }
func AdamicUtilityHandle(handle int) (int, bool, bool, int, bool) {
	arena := []*UtilityEvaluator{{}, {}}
	var utility *UtilityEvaluator
	if handle >= 0 {
		utility = arena[handle]
	}
	system := &LoadedDesignSystem{utility: utility}
	got := system.Utilities()
	result := -1
	for i, v := range arena {
		if v == got {
			result = i
		}
	}
	alias := got == utility
	unchanged := system.utility == utility
	if got != nil {
		got.Theme = NewTheme()
		got.Theme.Prefix = "mutated"
		alias = alias && utility.Theme.Prefix == "mutated"
	}
	repeated := system.Utilities()
	second := -1
	for i, v := range arena {
		if v == repeated {
			second = i
		}
	}
	return result, alias, unchanged, second, system.utility == utility
}
