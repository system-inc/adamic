// Oracle-only access to unchanged private Go modifier and dependency helpers.
package tailwind

type AdamicModifierCase struct {
	Name      string            `json:"name"`
	Source    string            `json:"source"`
	Decodes   map[string]string `json:"decodes"`
	Arbitrary map[string]bool   `json:"validArbitrary"`
	Blank     map[string]bool   `json:"blank"`
	Named     map[string]bool   `json:"validNamed"`
}

func AdamicModifierInput(name, input string) AdamicModifierCase {
	inner := input
	if len(input) >= 2 {
		inner = input[1 : len(input)-1]
	}
	row := AdamicModifierCase{name, input, map[string]string{}, map[string]bool{}, map[string]bool{}, map[string]bool{}}
	for _, text := range []string{input, inner, "var(" + inner + ")"} {
		decoded := decodeArbitraryValue(text)
		row.Decodes[text] = decoded
		for _, check := range []string{text, decoded} {
			row.Arbitrary[check] = isValidArbitrary(check)
			row.Blank[check] = isBlank(check)
			row.Named[check] = isValidNamedValue(check)
		}
	}
	return row
}
func AdamicParseModifier(input string) *ParsedModifier { return parseModifier(input) }
