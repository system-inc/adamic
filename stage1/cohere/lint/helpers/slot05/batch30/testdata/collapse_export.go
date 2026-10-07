package tailwind

func AdamicColor(keys []string) FunctionalUtilityArm { return colorArm(keys...) }
func AdamicTheme(key string) FunctionalUtilityArm    { return themeArm(key) }
func AdamicWidth(key, suffix string, guard bool) FunctionalUtilityArm {
	return widthArm(key, suffix, guard)
}
