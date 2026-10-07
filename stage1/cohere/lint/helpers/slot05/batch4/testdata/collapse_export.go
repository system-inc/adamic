// Oracle-only access to the actual private Go helpers; production cohere is unchanged.
package tailwind

import "sort"

type AdamicThemeEntry struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Options int    `json:"options"`
}
type AdamicThemeState struct {
	Order   []string           `json:"order"`
	Entries []AdamicThemeEntry `json:"entries"`
}

func AdamicThemeSnapshot(theme *Theme) AdamicThemeState {
	state := AdamicThemeState{Order: append([]string{}, theme.keyOrder...), Entries: []AdamicThemeEntry{}}
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
	theme.keyOrder = append([]string{}, state.Order...)
	for _, v := range state.Entries {
		theme.values[v.Key] = themeValue{v.Value, ThemeOptions(v.Options)}
	}
	return theme
}
func AdamicLiveTrace(theme *Theme, stop int, mutation string) []AdamicThemeEntry {
	trace := []AdamicThemeEntry{}
	theme.liveKeys(func(key string, value themeValue) bool {
		trace = append(trace, AdamicThemeEntry{key, value.value, int(value.options)})
		if len(trace) == 1 {
			switch mutation {
			case "deleteNext":
				delete(theme.values, "--seed-c")
			case "overwriteNext":
				theme.values["--seed-c"] = themeValue{"changed", 17}
			case "append":
				theme.keyOrder = append(theme.keyOrder, "--late")
				theme.values["--late"] = themeValue{"late", 16}
			case "replaceOrder":
				theme.keyOrder = []string{"--late"}
				theme.values["--late"] = themeValue{"late", 16}
			case "replaceValues":
				theme.values = map[string]themeValue{"--seed-c": {"replacement", 31}}
			case "changeNextSlot":
				if len(theme.keyOrder) > 1 {
					theme.keyOrder[1] = "--seed-b"
				}
			}
		}
		return stop < 0 || len(trace) < stop
	})
	return trace
}

func AdamicIgnoredThemeKey(key, namespace string) bool { return isIgnoredThemeKey(key, namespace) }
func AdamicKeysRefuse(theme *Theme, namespaces []string) (refused bool) {
	defer func() {
		if recover() != nil {
			refused = true
		}
	}()
	theme.KeysInNamespaces(namespaces)
	return false
}
