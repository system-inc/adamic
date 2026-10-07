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
func AdamicResolveKey(theme *Theme, candidate string, present bool, namespaces []string) (string, bool) {
	return theme.resolveKey(candidate, present, namespaces)
}
func AdamicIgnoredKey(key, namespace string) bool { return isIgnoredThemeKey(key, namespace) }
