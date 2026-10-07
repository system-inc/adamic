package tailwind

func AdamicBorder() *FunctionalUtilityDescription { return borderSideDescription() }
func AdamicMask() *FunctionalUtilityDescription   { return maskStopDescription() }
func AdamicResolve(value string, guard bool, keys []string, theme *Theme) (string, bool) {
	candidate := &ParsedCandidate{Value: &ParsedValue{Value: value}, Modifier: &ParsedModifier{Value: "50"}}
	arm := &FunctionalUtilityArm{ThemeKeys: keys, IsColor: guard, RefusesModifier: guard, InferTypes: []DataType{DataTypePercentage}, PercentagePassesThrough: true}
	return resolveArmColor(candidate, arm, theme)
}
