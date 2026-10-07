package tailwind

import "fmt"

func wave06Snapshot(theme *Theme) string {
	return fmt.Sprintf("%t %d %t %d %d", theme.Prefix == "", len(theme.values), theme.keyOrder == nil, len(theme.keyOrder), theme.deadKeys)
}
func AdamicWave06NewTheme(key string) string {
	first, second := NewTheme(), NewTheme()
	result := wave06Snapshot(first) + "\n" + wave06Snapshot(second) + fmt.Sprintf("\n%t\n", first == second)
	first.Prefix = key
	first.values[key] = themeValue{value: key, options: 31}
	first.keyOrder = []string{key}
	first.deadKeys = 7
	result += wave06Snapshot(second) + "\n"
	third := NewTheme()
	result += wave06Snapshot(third) + "\n"
	_, secondHas := second.values[key]
	_, thirdHas := third.values[key]
	return result + fmt.Sprintf("%d %t %t\n", len(first.values), secondHas, thirdHas)
}
