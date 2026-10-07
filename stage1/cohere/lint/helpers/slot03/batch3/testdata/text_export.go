package text

import "sort"

func AdamicHexValue(character rune) int { return hexValue(character) }

var adamicEntityBodies []string

func AdamicDecodeEntity(item string) (string, bool) { return decodeEntity(item) }
func AdamicEntityNames() []string {
	names := []string{}
	for name := range xhtmlEntities {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
func AdamicUnescapeObservation(value string) (string, []string) {
	adamicEntityBodies = []string{}
	result := UnescapeStringLiteralText(value)
	return result, append([]string{}, adamicEntityBodies...)
}
