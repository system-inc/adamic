package regexp

import (
	"fmt"
	"slices"
	"strings"
)

// V8DivergenceError is a backend compatibility refusal, not an ECMAScript
// SyntaxError. Keep parsing and reference bytecode available for diagnosis.
type V8DivergenceError struct {
	Behavior string
	Section  string
}

func (e *V8DivergenceError) Error() string {
	return fmt.Sprintf("RegExp refused: V8 %s departs from ECMA-262 %s; native and JavaScript must agree", e.Behavior, e.Section)
}

// NativeCompatibility also gates stage 0's JavaScript backend, which lowers
// through the same compiled program. A refused program emits neither backend.
func (p *Program) NativeCompatibility() error { return p.v8Refusal }

func modifiedFlags(f Flags, g *Group) Flags {
	f.IgnoreCase = (f.IgnoreCase || g.Enable.IgnoreCase) && !g.Disable.IgnoreCase
	f.Multiline = (f.Multiline || g.Enable.Multiline) && !g.Disable.Multiline
	f.DotAll = (f.DotAll || g.Enable.DotAll) && !g.Disable.DotAll
	return f
}

func v8Divergence(tree *Pattern, properties PropertyProvider) error {
	// RegExpParserImpl::flags_ changes on '(' but not ')', unlike each
	// builder's scoped flags. Track source order, including alternatives and
	// lookbehind, not the VM's execution order. An ordinary '(' resets it.
	parserFlags := tree.Flags
	legacyAlternative := false
	var visit func(Node, Flags) error
	visit = func(node Node, flags Flags) error {
		switch node := node.(type) {
		case *Disjunction:
			outerAlternative := legacyAlternative
			defer func() { legacyAlternative = outerAlternative }()
			for index, alternative := range node.Alternatives {
				legacyAlternative = outerAlternative || index != 0
				for _, term := range alternative.Terms {
					if err := visit(term, flags); err != nil {
						return err
					}
				}
			}
		case *Quantifier:
			return visit(node.Atom, flags)
		case *Group:
			flags = modifiedFlags(flags, node)
			parserFlags = flags
			return visit(node.Body, flags)
		case *CharacterClass:
			set, err := compileClass(node.Expr, flags, properties)
			if err != nil {
				return err
			}
			// In legacy mode V8 loses an enabled scoped i when compiling
			// a negated class on a later alternative. The first branch and
			// case-closed classes are controls, not blanket refusals.
			if !unicodeMode(flags) && flags.IgnoreCase && !tree.Flags.IgnoreCase && legacyAlternative && node.Negated {
				actualFlags := flags
				actualFlags.IgnoreCase = false
				actual, err := compileClass(node.Expr, actualFlags, properties)
				if err != nil {
					return err
				}
				if !sameCharacterSets(matchedCharacters(set, flags), actual) {
					return &V8DivergenceError{"drops scoped i on a later alternative's negated legacy class", "22.2.2.7 CompileAtom and 22.2.2.7.4 UpdateModifiers"}
				}
			}
			if flags.UnicodeSets {
				// The reported ab|a| shape also hangs V8's replacement slow
				// path when an empty match is advanced into a surrogate pair.
				// Direct exec ordering agrees with 22.2.2.7; do not call it a
				// demonstrated ordering bug. Refuse the ruled-on mixed shape.
				if len(set.ranges) > 0 && slices.ContainsFunc(set.strings, func(s []rune) bool { return len(s) == 0 }) && slices.ContainsFunc(set.strings, func(s []rune) bool { return len(s) > 1 }) {
					return &V8DivergenceError{"can repeat an earlier empty match and hang replacement for mixed-length \\q alternatives", "22.2.2.7 CompileAtom and 22.2.6.11 RegExp.prototype [ Symbol.replace ] / AdvanceStringIndex"}
				}
				actual, err := v8Class(node.Expr, parserFlags, properties)
				if err != nil {
					return err
				}
				// V8 parses class strings with its stale flags, but the
				// string matcher uses the builder's correctly scoped flags.
				// Canonicalize strings again for this comparison only; do
				// not expand singleton ranges, which carry the other bug.
				actual.strings = foldSet(characterSet{strings: actual.strings}, flags).strings
				if node.Negated {
					set = negateSet(set, flags)
					actual.ranges = subtractRanges([]RuneRange{{0, maxCharacter(flags)}}, actual.ranges)
				}
				if !sameCharacterSets(matchedCharacters(set, flags), actual) {
					if parserFlags.IgnoreCase != flags.IgnoreCase {
						return modifierDivergence()
					}
					return &V8DivergenceError{"does not case-fold singleton \\q class strings under iv", "22.2.2.9 CompileToCharSet and 22.2.2.10 CompileClassSetString"}
				}
			} else if flags.Unicode && parserFlags.IgnoreCase != flags.IgnoreCase && classWordEscape(node.Expr, !parserFlags.IgnoreCase) {
				actual, err := v8UnicodeClass(node.Expr, flags, parserFlags, properties)
				if err != nil {
					return err
				}
				actual = foldSet(actual, flags)
				if node.Negated {
					set = negateSet(set, flags)
					actual = negateSet(actual, flags)
				}
				if !sameCharacterSets(matchedCharacters(set, flags), matchedCharacters(actual, flags)) {
					return modifierDivergence()
				}
			}
		case *Character:
			if unicodeMode(flags) && parserFlags.IgnoreCase != flags.IgnoreCase {
				if flags.UnicodeSets && node.Kind == PropertyEscape {
					expected, err := compileCharacter(node, flags, properties)
					if err != nil {
						return err
					}
					actual, err := compileCharacter(node, parserFlags, properties)
					if err != nil {
						return err
					}
					if !sameCharacterSets(matchedCharacters(expected, flags), matchedCharacters(actual, parserFlags)) {
						return modifierDivergence()
					}
				} else if node.Kind == ClassEscape && (node.Raw == `\W` || parserFlags.IgnoreCase && strings.EqualFold(node.Raw, `\w`)) {
					return modifierDivergence()
				}
			}
		}
		return nil
	}
	return visit(tree.Body, tree.Flags)
}

// A non-v Unicode class uses the builder's flags for literals and ranges,
// but V8 expands word escapes with its parser flags first. Combine the raw
// operands before the builder folds them, preserving masks such as [\w\W].
func v8UnicodeClass(expr ClassExpression, flags, parserFlags Flags, properties PropertyProvider) (characterSet, error) {
	switch expr := expr.(type) {
	case *ClassCharacter:
		operandFlags := flags
		if expr.Character.Kind == ClassEscape && strings.EqualFold(expr.Character.Raw, `\w`) {
			operandFlags = parserFlags
		}
		set, err := compileCharacter(expr.Character, operandFlags, properties)
		return matchedCharacters(set, operandFlags), err
	case *ClassUnion:
		var result characterSet
		for _, operand := range expr.Operands {
			set, err := v8UnicodeClass(operand, flags, parserFlags, properties)
			if err != nil {
				return result, err
			}
			result.ranges = append(result.ranges, set.ranges...)
		}
		result.ranges = normalizeRanges(result.ranges)
		return result, nil
	default:
		set, err := compileClass(expr, flags, properties)
		return matchedCharacters(set, flags), err
	}
}

func modifierDivergence() error {
	return &V8DivergenceError{"leaks or drops a scoped i modifier in a later Unicode class or word escape", "22.2.2.7 CompileAtom and 22.2.2.7.4 UpdateModifiers"}
}

// Convert canonical-domain ranges into the actual code points accepted by a
// matcher. This lets us retain classes whose other operands mask V8's bug.
func matchedCharacters(set characterSet, flags Flags) characterSet {
	if flags.IgnoreCase {
		ranges := slices.Clone(set.ranges)
		table := simpleCaseFold[:]
		if !unicodeMode(flags) {
			table = legacyUppercase[:]
		}
		for _, pair := range table {
			if set.contains(pair[1]) {
				ranges = append(ranges, RuneRange{pair[0], pair[0]})
			}
		}
		set.ranges = normalizeRanges(ranges)
	}
	return set
}

func sameCharacterSets(a, b characterSet) bool {
	if !slices.Equal(a.ranges, b.ranges) || len(a.strings) != len(b.strings) {
		return false
	}
	for _, text := range a.strings {
		if !slices.ContainsFunc(b.strings, func(other []rune) bool { return slices.Equal(text, other) }) {
			return false
		}
	}
	return true
}

// V8 expands ordinary v operands to their equivalence classes, but puts a
// one-code-point ClassString directly in the ranges as its folded spelling.
func v8Class(expr ClassExpression, flags Flags, properties PropertyProvider) (characterSet, error) {
	switch expr := expr.(type) {
	case *ClassString:
		return compileClass(expr, flags, properties)
	case *ClassUnion:
		var result characterSet
		for _, operand := range expr.Operands {
			set, err := v8Class(operand, flags, properties)
			if err != nil {
				return result, err
			}
			result.ranges = append(result.ranges, set.ranges...)
			result.strings = append(result.strings, set.strings...)
		}
		result.ranges = normalizeRanges(result.ranges)
		return result, nil
	case *ClassIntersection:
		a, err := v8Class(expr.Left, flags, properties)
		if err != nil {
			return a, err
		}
		b, err := v8Class(expr.Right, flags, properties)
		return combineSets(a, b, true), err
	case *ClassSubtraction:
		a, err := v8Class(expr.Left, flags, properties)
		if err != nil {
			return a, err
		}
		b, err := v8Class(expr.Right, flags, properties)
		return combineSets(a, b, false), err
	case *ClassNegation:
		set, err := v8Class(expr.Operand, flags, properties)
		set.ranges = subtractRanges([]RuneRange{{0, maxCharacter(flags)}}, set.ranges)
		return set, err
	default:
		set, err := compileClass(expr, flags, properties)
		return matchedCharacters(set, flags), err
	}
}

func classWordEscape(expr ClassExpression, uppercaseOnly bool) bool {
	switch expr := expr.(type) {
	case *ClassCharacter:
		return expr.Character.Kind == ClassEscape && (expr.Character.Raw == `\W` || !uppercaseOnly && expr.Character.Raw == `\w`)
	case *ClassUnion:
		return slices.ContainsFunc(expr.Operands, func(e ClassExpression) bool { return classWordEscape(e, uppercaseOnly) })
	case *ClassNegation:
		return classWordEscape(expr.Operand, uppercaseOnly)
	}
	return false
}
