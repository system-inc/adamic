// Oracle-only access; actual Go constructor and map writes remain unchanged.
package tailwind

import "fmt"

func AdamicNewThemeTrace(key string) []string {
	first, second := NewTheme(), NewTheme()
	snapshot := func(t *Theme) string {
		if t.values == nil {
			panic("constructor values map is nil")
		}
		return fmt.Sprintf("%d %d %d %d", len(t.Prefix), len(t.values), len(t.keyOrder), t.deadKeys)
	}
	trace := []string{snapshot(first)}
	first.values[key] = themeValue{key, 31}
	first.keyOrder = append(first.keyOrder, key)
	first.Prefix = "probe"
	first.deadKeys = 7
	trace = append(trace, snapshot(second), snapshot(NewTheme()))
	return trace
}
