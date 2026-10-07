package tailwind

import "sort"

func AdamicEscapeTerminator(character byte) bool { return isEscapeTerminator(character) }
func AdamicFollowedByWhitespace(input string, index int, peek func(int) byte) bool {
	return isFollowedByWhitespace(input, index, peek)
}
func AdamicIgnoredThemeKey(key, namespace string) bool { return isIgnoredThemeKey(key, namespace) }
func AdamicIgnoredNamespaces() []string {
	result := []string{}
	for key := range ignoredThemeKeys {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
func AdamicIgnoredKeys(namespace string) []string { return ignoredThemeKeys[namespace] }
