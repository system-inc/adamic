package load

import "strings"

// The stock declaration loses the null returned for an invalid Date.
func dateLibrary(text string) string {
	return strings.ReplaceAll(text, "toJSON(key?: any): string;", "toJSON(key?: any): string | null;")
}
