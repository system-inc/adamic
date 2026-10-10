package regexp

import "unicode/utf8"

// NativeLiteralTestDeclarations permits only a fresh literal's test on this exact
// literal input. The pattern-wide refusal still applies to every other use.
func NativeLiteralTestDeclarations(pattern, flags, input, name string) (string, error) {
	tree, err := parsePattern(pattern, flags, true)
	if err != nil {
		return "", err
	}
	properties := UnicodeProperties{}
	program, err := compilePattern(tree, properties)
	if err != nil {
		return "", err
	}
	if program.v8Refusal == nil {
		return program.NativeDeclarations(name)
	}
	if !literalClassTestAgrees(tree, program.v8Refusal, input, properties) {
		return "", program.v8Refusal
	}
	program.v8Refusal = nil
	return program.NativeDeclarations(name)
}

// For one unquantified class with identical string alternatives, excluding every
// differing code point from the input proves all its possible match starts agree.
func literalClassTestAgrees(tree *Pattern, refusal error, input string, properties PropertyProvider) bool {
	divergence, ok := refusal.(*V8DivergenceError)
	if !ok || divergence.Behavior != "does not case-fold singleton \\q class strings under iv" ||
		tree.Flags.Global || tree.Flags.Sticky || !utf8.ValidString(input) ||
		len(tree.Body.Alternatives) != 1 || len(tree.Body.Alternatives[0].Terms) != 1 {
		return false
	}
	class, ok := tree.Body.Alternatives[0].Terms[0].(*CharacterClass)
	if !ok || class.Negated {
		return false
	}
	expected, err := compileClass(class.Expr, tree.Flags, properties)
	if err != nil {
		return false
	}
	actual, err := v8Class(class.Expr, tree.Flags, properties)
	if err != nil {
		return false
	}
	actual.strings = foldSet(characterSet{strings: actual.strings}, tree.Flags).strings
	if !sameCharacterSets(characterSet{strings: expected.strings}, characterSet{strings: actual.strings}) {
		return false
	}
	expected = matchedCharacters(expected, tree.Flags)
	difference := characterSet{ranges: normalizeRanges(append(subtractRanges(expected.ranges, actual.ranges), subtractRanges(actual.ranges, expected.ranges)...))}
	for _, point := range input {
		if difference.contains(point) {
			return false
		}
	}
	return true
}
