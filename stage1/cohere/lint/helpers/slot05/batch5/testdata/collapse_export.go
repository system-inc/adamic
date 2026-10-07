// Oracle-only access to actual private methods and their observable store state.
package tailwind

import "sort"

type AdamicThemeEntry struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Options int    `json:"options"`
}
type AdamicThemeState struct {
	Prefix  string             `json:"prefix"`
	Dead    int                `json:"dead"`
	Order   []string           `json:"order"`
	Entries []AdamicThemeEntry `json:"entries"`
}

func AdamicThemeSnapshot(theme *Theme) AdamicThemeState {
	state := AdamicThemeState{Prefix: theme.Prefix, Dead: theme.deadKeys, Order: append([]string{}, theme.keyOrder...), Entries: []AdamicThemeEntry{}}
	keys := []string{}
	for key := range theme.values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		v := theme.values[key]
		state.Entries = append(state.Entries, AdamicThemeEntry{key, v.value, int(v.options)})
	}
	return state
}
func AdamicThemeRestore(state AdamicThemeState) *Theme {
	theme := NewTheme()
	theme.Prefix = state.Prefix
	theme.deadKeys = state.Dead
	theme.keyOrder = append([]string{}, state.Order...)
	for _, v := range state.Entries {
		theme.values[v.Key] = themeValue{v.Value, ThemeOptions(v.Options)}
	}
	return theme
}
func AdamicMutationTrace(theme *Theme, mode, key string) []AdamicThemeState {
	oldOrder, oldValues := theme.keyOrder, theme.values
	mutate := func() {
		switch mode {
		case "clear":
			theme.clearAll()
		case "compact":
			theme.compactKeyOrder()
		case "delete":
			theme.delete(key)
		default:
			panic("bad mode")
		}
	}
	mutate()
	trace := []AdamicThemeState{AdamicThemeSnapshot(theme), AdamicThemeSnapshot(&Theme{keyOrder: oldOrder, values: oldValues})}
	if len(oldOrder) > 0 {
		oldOrder[0] = "--alias-probe"
	}
	oldValues["--alias-probe"] = themeValue{"alias mutation", 31}
	trace = append(trace, AdamicThemeSnapshot(theme))
	mutate()
	return append(trace, AdamicThemeSnapshot(theme))
}
