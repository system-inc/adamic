package regexp

import "unicode/utf8"

// NativeLiteralTestDeclarations permits only a fresh literal's test on this exact
// literal input. The pattern-wide refusal still applies to every other use.
func NativeLiteralTestDeclarations(pattern, flags, input, name string) (string, error) {
	tree, parsePatternError := parsePattern(pattern, flags, true)
	if parsePatternError != nil {
		return "", parsePatternError
	}
	properties := UnicodeProperties{}
	program, compilePatternError := compilePattern(tree, properties)
	if compilePatternError != nil {
		return "", compilePatternError
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
	expected, compileClassError := compileClass(class.Expr, tree.Flags, properties)
	if compileClassError != nil {
		return false
	}
	actual, v8ClassError := v8Class(class.Expr, tree.Flags, properties)
	if v8ClassError != nil {
		return false
	}
	actual.strings = foldSet(characterSet{strings: actual.strings}, tree.Flags).strings
	if !sameCharacterSets(characterSet{strings: expected.strings}, characterSet{strings: actual.strings}) {
		return false
	}
	expected = matchedCharacters(expected, tree.Flags)
	difference := characterSet{ranges: normalizeRanges(
		append(subtractRanges(expected.ranges, actual.ranges), subtractRanges(actual.ranges, expected.ranges)...),
	)}
	for _, point := range input {
		if difference.contains(point) {
			return false
		}
	}
	return true
}
