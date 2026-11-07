package regexp

import (
	"fmt"
	"github.com/system-inc/adamic/internal/unicodeproperties"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"
)

// SyntaxError is an ECMAScript early error. Offset is a byte offset in Pattern.
type SyntaxError struct {
	Offset  int
	Message string
}

func (syntaxError *SyntaxError) Error() string {
	return fmt.Sprintf("regexp: %s at byte %d", syntaxError.Message, syntaxError.Offset)
}

// Parse parses pattern and flags according to ECMAScript 2025.
func Parse(pattern, flags string) (*Pattern, error) {
	return parsePattern(pattern, flags, false)
}

// Native compilation uses Node 24.19.0 V8 interval bounds; Parse remains the
// specification reference for syntax-oracle diagnostics.
func parsePattern(pattern, flags string, nativeBounds bool) (*Pattern, error) {
	f, parseFlagsError := parseFlags(flags)
	if parseFlagsError != nil {
		return nil, parseFlagsError
	}
	p := &parser{source: pattern, flags: f, nativeBounds: nativeBounds, names: make(map[string]bool)}
	// Decimal escapes depend on the total capture count, including captures to
	// their right. Count them without interpreting pattern contents first.
	p.captureCount, p.hasNamedCapture = countCaptures(pattern)
	body, disjunctionError := p.disjunction(0)
	if disjunctionError != nil {
		return nil, disjunctionError
	}
	if !p.done() {
		return nil, p.fail("unexpected character")
	}
	for name := range p.namedReferences {
		if !p.names[name] {
			return nil, p.failAt(p.namedReferences[name], "unknown capture name")
		}
	}
	if validateDuplicateNamesError := validateDuplicateNames(body); validateDuplicateNamesError != nil {
		return nil, validateDuplicateNamesError
	}
	return &Pattern{Flags: f, Body: body}, nil
}

// ParseUTF16 preserves JavaScript pattern strings, including lone surrogates.
func ParseUTF16(pattern []uint16, flags string) (*Pattern, error) {
	return Parse(patternFromUTF16(pattern), flags)
}
func patternFromUTF16(pattern []uint16) string {
	var encoded []byte
	for i := 0; i < len(pattern); i++ {
		c := rune(pattern[i])
		if c >= 0xd800 && c <= 0xdbff && i+1 < len(pattern) && pattern[i+1] >= 0xdc00 && pattern[i+1] <= 0xdfff {
			c = 0x10000 + (c-0xd800)*0x400 + rune(pattern[i+1]) - 0xdc00
			i++
		}
		if c >= 0xd800 && c <= 0xdfff {
			encoded = append(encoded, byte(0xe0|c>>12), byte(0x80|(c>>6)&0x3f), byte(0x80|c&0x3f))
		} else {
			encoded = utf8.AppendRune(encoded, c)
		}
	}
	return string(encoded)
}

// Adamic strings use WTF-8 so that lone UTF-16 surrogates remain observable.
func patternRune(source string) (rune, int) {
	r, width := utf8.DecodeRuneInString(source)
	if width == 1 &&
		r == utf8.RuneError &&
		len(source) >= 3 &&
		source[0] == 0xed &&
		source[1] >= 0xa0 &&
		source[1] <= 0xbf &&
		source[2] >= 0x80 &&
		source[2] <= 0xbf {
		return rune(source[0]&15)<<12 | rune(source[1]&63)<<6 | rune(source[2]&63), 3
	}
	return r, width
}

func parseFlags(s string) (Flags, error) {
	var f Flags
	seen := [128]bool{}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 128 || seen[c] {
			return f, &SyntaxError{i, "invalid or duplicate flag"}
		}
		seen[c] = true
		switch c {
		case 'd':
			f.HasIndices = true
		case 'g':
			f.Global = true
		case 'i':
			f.IgnoreCase = true
		case 'm':
			f.Multiline = true
		case 's':
			f.DotAll = true
		case 'u':
			f.Unicode = true
		case 'v':
			f.UnicodeSets = true
		case 'y':
			f.Sticky = true
		default:
			return f, &SyntaxError{i, "invalid flag"}
		}
	}
	if f.Unicode && f.UnicodeSets {
		return f, &SyntaxError{0, "u and v flags are mutually exclusive"}
	}
	return f, nil
}

type parser struct {
	nativeBounds           bool
	source                 string
	pos                    int
	flags                  Flags
	captureCount, captures int
	hasNamedCapture        bool
	pendingCharacter       *Character
	names                  map[string]bool
	namedReferences        map[string]int
}

func (parser *parser) disjunction(stop byte) (*Disjunction, error) {
	d := &Disjunction{}
	for {
		a, alternativeError := parser.alternative(stop)
		if alternativeError != nil {
			return nil, alternativeError
		}
		d.Alternatives = append(d.Alternatives, a)
		if parser.done() || parser.peek() != '|' {
			break
		}
		parser.pos++
	}
	return d, nil
}

func (parser *parser) alternative(stop byte) (*Alternative, error) {
	a := &Alternative{}
	for !parser.done() && parser.peek() != '|' && (stop == 0 || parser.peek() != stop) {
		if parser.peek() == ')' && stop == 0 {
			return nil, parser.fail("unmatched closing parenthesis")
		}
		n, quantifiable, termError := parser.term()
		if termError != nil {
			return nil, termError
		}
		if parser.quantifierAhead() {
			if !quantifiable {
				return nil, parser.fail("nothing to repeat")
			}
			var quantifierError error
			n, quantifierError = parser.quantifier(n)
			if quantifierError != nil {
				return nil, quantifierError
			}
			if parser.quantifierAhead() {
				return nil, parser.fail("nothing to repeat")
			}
		}
		a.Terms = append(a.Terms, n)
	}
	return a, nil
}

func (parser *parser) term() (Node, bool, error) {
	switch parser.peek() {
	case '^':
		parser.pos++
		return &Assertion{Kind: Start}, false, nil
	case '$':
		parser.pos++
		return &Assertion{Kind: End}, false, nil
	case '.':
		parser.pos++
		return &Dot{}, true, nil
	case '[':
		n, characterClassError := parser.characterClass()
		return n, true, characterClassError
	case '(':
		n, q, groupError := parser.group()
		return n, q, groupError
	case '\\':
		return parser.escape(false)
	case '*', '+', '?':
		return nil, false, parser.fail("nothing to repeat")
	case '{':
		if parser.quantifierAhead() {
			return nil, false, parser.fail("nothing to repeat")
		}
		if parser.flags.Unicode || parser.flags.UnicodeSets {
			return nil, false, parser.fail("incomplete quantifier")
		}
	case ']', '}':
		if parser.flags.Unicode || parser.flags.UnicodeSets {
			return nil, false, parser.fail("lone punctuation")
		}
	}
	return parser.literal()
}

func (parser *parser) group() (Node, bool, error) {
	parser.pos++
	kind, name := Capturing, ""
	var enable, disable Flags
	if parser.take('?') {
		switch {
		case parser.take(':'):
			kind = NonCapturing
		case parser.take('='):
			kind = PositiveLookahead
		case parser.take('!'):
			kind = NegativeLookahead
		case parser.take('<'):
			if parser.take('=') {
				kind = PositiveLookbehind
			} else if parser.take('!') {
				kind = NegativeLookbehind
			} else {
				kind = Capturing
				var groupNameError error
				name, groupNameError = parser.groupName()
				if groupNameError != nil {
					return nil, false, groupNameError
				}
				parser.names[name] = true
			}
		case strings.ContainsRune("ims-", rune(parser.peek())):
			var modifiersError error
			enable, disable, modifiersError = parser.modifiers()
			if modifiersError != nil {
				return nil, false, modifiersError
			}
			kind = NonCapturing
		default:
			return nil, false, parser.fail("invalid group")
		}
	}
	if kind == Capturing {
		parser.captures++
	}
	body, disjunctionError := parser.disjunction(')')
	if disjunctionError != nil {
		return nil, false, disjunctionError
	}
	if !parser.take(')') {
		return nil, false, parser.fail("unterminated group")
	}
	q := kind == Capturing || kind == NonCapturing
	if !parser.flags.Unicode && !parser.flags.UnicodeSets && (kind == PositiveLookahead || kind == NegativeLookahead) {
		q = true
	}
	return &Group{Kind: kind, Name: name, Body: body, Enable: enable, Disable: disable}, q, nil
}

func (parser *parser) modifiers() (Flags, Flags, error) {
	var enable, disable Flags
	disabling, any := false, false
	seen := make(map[byte]bool)
	for !parser.done() && parser.peek() != ':' {
		c := parser.peek()
		if c == '-' && !disabling {
			disabling = true
			parser.pos++
			continue
		}
		if c != 'i' && c != 'm' && c != 's' || seen[c] {
			return enable, disable, parser.fail("invalid modifiers")
		}
		seen[c], any = true, true
		target := &enable
		if disabling {
			target = &disable
		}
		switch c {
		case 'i':
			target.IgnoreCase = true
		case 'm':
			target.Multiline = true
		case 's':
			target.DotAll = true
		}
		parser.pos++
	}
	if !any || !parser.take(':') {
		return enable, disable, parser.fail("invalid modifiers")
	}
	return enable, disable, nil
}

func (parser *parser) quantifier(atom Node) (Node, error) {
	minimum, maximum := big.NewInt(0), (*big.Int)(nil)
	switch parser.peek() {
	case '*':
		parser.pos++
	case '+':
		parser.pos++
		minimum = big.NewInt(1)
	case '?':
		parser.pos++
		maximum = big.NewInt(1)
	case '{':
		start := parser.pos
		parser.pos++
		m, ok := parser.decimal()
		if !ok {
			parser.pos = start
			return nil, parser.fail("incomplete quantifier")
		}
		minimum, maximum = m, new(big.Int).Set(m)
		if parser.take(',') {
			maximum = nil
			if n, ok := parser.decimal(); ok {
				maximum = n
			}
		}
		if !parser.take('}') {
			parser.pos = start
			return nil, parser.fail("incomplete quantifier")
		}
		if maximum != nil && minimum.Cmp(maximum) > 0 {
			// Exact decimal MVs determine the spec's early error. Node's
			// clamped comparison instead accepts this ruled-on divergence.
			if maximum.Cmp(big.NewInt(2147483647)) >= 0 {
				return nil, &V8DivergenceError{
					Behavior: "clamps quantifier bounds above 2^31-1 before the min > max check",
					Section:  "22.2.1.1 Static Semantics: Early Errors, QuantifierPrefix",
				}
			}
			return nil, parser.failAt(start, "quantifier range out of order")
		}
		if parser.nativeBounds {
			// Port of V8 RegExpParserImpl::ParseIntervalQuantifier after the
			// exact-order refusal. kInfinity is the unbounded maximum sentinel.
			limit := big.NewInt(2147483647)
			if minimum.Cmp(limit) > 0 {
				minimum.Set(limit)
			}
			if maximum != nil && maximum.Cmp(limit) >= 0 {
				maximum = nil
			}
		}
	}
	greedy := !parser.take('?')
	return &Quantifier{Atom: atom, Min: minimum, Max: maximum, Greedy: greedy}, nil
}

func (parser *parser) escape(inClass bool) (Node, bool, error) {
	start := parser.pos
	parser.pos++
	if parser.done() {
		return nil, false, parser.fail("trailing escape")
	}
	if parser.peek() >= utf8.RuneSelf {
		if parser.flags.Unicode || parser.flags.UnicodeSets {
			return nil, false, parser.failAt(start, "invalid identity escape")
		}
		node, _, literalError := parser.literal()
		if literalError != nil {
			return nil, false, literalError
		}
		character := node.(*Character)
		character.Kind = Escape
		character.Raw = parser.source[start:parser.pos]
		return character, true, nil
	}
	c := parser.peek()
	parser.pos++
	if !inClass && (c == 'b' || c == 'B') {
		kind := WordBoundary
		if c == 'B' {
			kind = NotWordBoundary
		}
		return &Assertion{Kind: kind}, false, nil
	}
	if c >= '1' && c <= '9' {
		parser.pos--
		n, _ := parser.decimal()
		if !inClass && n.IsInt64() && n.Int64() <= int64(parser.captureCount) {
			return &Backreference{Index: int(n.Int64())}, true, nil
		}
		if parser.flags.Unicode || parser.flags.UnicodeSets {
			return nil, false, parser.failAt(start, "invalid decimal escape")
		}
		// Annex B.1.2 ClassEscape / LegacyOctalEscapeSequence also
		// applies inside legacy classes, independently of capture count.
		parser.pos = start + 1
		return parser.legacyOctal(start)
	}
	if !inClass && c == 'k' {
		if !parser.hasNamedCapture {
			if parser.flags.Unicode || parser.flags.UnicodeSets {
				return nil, false, parser.failAt(start, "invalid named reference")
			}
			return &Character{Value: 'k', Raw: parser.source[start:parser.pos], Kind: Escape}, true, nil
		}
		if !parser.take('<') {
			return nil, false, parser.fail("invalid named reference")
		}
		name, groupNameError := parser.groupName()
		if groupNameError != nil {
			return nil, false, groupNameError
		}
		if parser.namedReferences == nil {
			parser.namedReferences = make(map[string]int)
		}
		parser.namedReferences[name] = start
		return &Backreference{Name: name}, true, nil
	}
	if c == 'p' || c == 'P' {
		if !(parser.flags.Unicode || parser.flags.UnicodeSets) {
			return &Character{Value: rune(c), Raw: parser.source[start:parser.pos], Kind: Escape}, true, nil
		}
		if !parser.take('{') {
			return nil, false, parser.failAt(start, "invalid property escape")
		}
		end := strings.IndexByte(parser.source[parser.pos:], '}')
		if end < 0 {
			return nil, false, parser.fail("unterminated property escape")
		}
		property := parser.source[parser.pos : parser.pos+end]
		parser.pos += end + 1
		if !validProperty(property, parser.flags.UnicodeSets) || c == 'P' && stringProperties[property] {
			return nil, false, parser.failAt(start, "invalid Unicode property")
		}
		return &Character{Raw: parser.source[start:parser.pos], Kind: PropertyEscape}, true, nil
	}
	if strings.ContainsRune("dDsSwW", rune(c)) {
		return &Character{Raw: parser.source[start:parser.pos], Kind: ClassEscape}, true, nil
	}
	if inClass && c == 'b' {
		return &Character{Value: '\b', Raw: parser.source[start:parser.pos], Kind: Escape}, true, nil
	}
	if c == '0' {
		if !parser.done() && parser.peek() >= '0' && parser.peek() <= '9' {
			if parser.flags.Unicode || parser.flags.UnicodeSets {
				return nil, false, parser.failAt(start, "invalid decimal escape")
			}
			parser.pos = start + 1
			return parser.legacyOctal(start)
		}
		return &Character{Value: 0, Raw: parser.source[start:parser.pos], Kind: Escape}, true, nil
	}
	if c == 'c' {
		if !parser.done() &&
			(parser.peek() >= 'a' && parser.peek() <= 'z' || parser.peek() >= 'A' && parser.peek() <= 'Z' || inClass &&
				!parser.flags.Unicode &&
				!parser.flags.UnicodeSets &&
				(parser.peek() >= '0' && parser.peek() <= '9' || parser.peek() == '_')) {
			value := parser.peek() % 32
			parser.pos++
			return &Character{Value: rune(value), Raw: parser.source[start:parser.pos], Kind: Escape}, true, nil
		}
		if parser.flags.Unicode || parser.flags.UnicodeSets {
			return nil, false, parser.failAt(start, "invalid control escape")
		}
		parser.pos = start + 1
		return &Character{Value: '\\', Raw: parser.source[start:parser.pos], Kind: Literal}, true, nil
	}
	if c == 'u' {
		ch, unicodeEscapeError := parser.unicodeEscape(start)
		return ch, true, unicodeEscapeError
	}
	if c == 'x' {
		ch, hexEscapeError := parser.hexEscape(start, 2)
		if hexEscapeError != nil && !parser.flags.Unicode && !parser.flags.UnicodeSets {
			parser.pos = start + 2
			return &Character{Value: 'x', Raw: parser.source[start:parser.pos], Kind: Escape}, true, nil
		}
		return ch, true, hexEscapeError
	}
	if strings.ContainsRune("fnrtv", rune(c)) {
		return &Character{
			Value: map[byte]rune{'f': 12, 'n': 10, 'r': 13, 't': 9, 'v': 11}[c],
			Raw:   parser.source[start:parser.pos],
			Kind:  Escape,
		}, true, nil
	}
	if (parser.flags.Unicode || parser.flags.UnicodeSets) &&
		!strings.ContainsRune("^$\\.*+?()[]{}|/", rune(c)) &&
		!(inClass && (c == '-' || parser.flags.UnicodeSets && strings.ContainsRune("!#%&,:;<=>@`~", rune(c)))) {
		return nil, false, parser.failAt(start, "invalid identity escape")
	}
	return &Character{Value: rune(c), Raw: parser.source[start:parser.pos], Kind: Escape}, true, nil
}

func (parser *parser) characterClass() (Node, error) {
	parser.pos++
	neg := parser.take('^')
	union := &ClassUnion{}
	for !parser.done() && parser.peek() != ']' {
		operand, classSetOperandError := parser.classSetOperand()
		if classSetOperandError != nil {
			return nil, classSetOperandError
		}
		union.Operands = append(union.Operands, operand)
		if parser.flags.UnicodeSets && (parser.match("&&") || parser.match("--")) {
			if len(union.Operands) != 1 {
				return nil, parser.fail("invalid Unicode set operation")
			}
			if _, rangeOperand := operand.(*ClassRange); rangeOperand {
				return nil, parser.fail("range must be nested in Unicode set operation")
			}
			op := parser.source[parser.pos : parser.pos+2]
			parser.pos += 2
			left := ClassExpression(union)
			for {
				right, operandError := parser.classSetOperand()
				if operandError != nil {
					return nil, operandError
				}
				if _, rangeOperand := right.(*ClassRange); rangeOperand {
					return nil, parser.fail("range must be nested in Unicode set operation")
				}
				if op == "&&" {
					left = &ClassIntersection{Left: left, Right: right}
				} else {
					left = &ClassSubtraction{Left: left, Right: right}
				}
				if !parser.match(op) {
					break
				}
				parser.pos += 2
			}
			return parser.finishClass(neg, left)
		}
	}
	return parser.finishClass(neg, union)
}

func (parser *parser) finishClass(neg bool, expr ClassExpression) (Node, error) {
	if !parser.take(']') {
		return nil, parser.fail("unterminated character class")
	}
	if neg && classMayContainStrings(expr) {
		return nil, parser.fail("cannot negate a class containing strings")
	}
	return &CharacterClass{Negated: neg, Expr: expr}, nil
}
func (parser *parser) classSetOperand() (ClassExpression, error) {
	if parser.flags.UnicodeSets && parser.peek() == '[' {
		n, characterClassError := parser.characterClass()
		if characterClassError != nil {
			return nil, characterClassError
		}
		class := n.(*CharacterClass)
		if class.Negated {
			return &ClassNegation{Operand: class.Expr}, nil
		}
		return class.Expr, nil
	}
	if parser.flags.UnicodeSets && parser.match("\\q{") {
		parser.pos += 3
		result := &ClassString{Alternatives: [][]*Character{{}}}
		for !parser.done() && parser.peek() != '}' {
			if parser.take('|') {
				result.Alternatives = append(result.Alternatives, []*Character{})
				continue
			}
			character, classCharacterError := parser.classCharacter()
			if classCharacterError != nil {
				return nil, classCharacterError
			}
			last := len(result.Alternatives) - 1
			result.Alternatives[last] = append(result.Alternatives[last], character)
		}
		if !parser.take('}') {
			return nil, parser.fail("unterminated class string")
		}
		return result, nil
	}
	c, classCharacterError := parser.classCharacter()
	if classCharacterError != nil {
		return nil, classCharacterError
	}
	if parser.peek() == '-' &&
		parser.pos+1 < len(parser.source) &&
		parser.source[parser.pos+1] != ']' &&
		!(parser.flags.UnicodeSets && parser.match("--")) {
		parser.pos++
		right, classCharacterError := parser.classCharacter()
		if classCharacterError != nil {
			return nil, classCharacterError
		}
		if c.Kind == ClassEscape ||
			c.Kind == PropertyEscape ||
			right.Kind == ClassEscape ||
			right.Kind == PropertyEscape {
			if !parser.flags.Unicode && !parser.flags.UnicodeSets {
				return &ClassUnion{Operands: []ClassExpression{
					&ClassCharacter{Character: c},
					&ClassCharacter{Character: &Character{Value: '-', Raw: "-", Kind: Literal}},
					&ClassCharacter{Character: right},
				}}, nil
			}
			return nil, parser.fail("invalid character class range")
		}
		if c.Value > right.Value {
			return nil, parser.fail("invalid character class range")
		}
		return &ClassRange{From: c, To: right}, nil
	}
	return &ClassCharacter{Character: c}, nil
}
func (parser *parser) classCharacter() (*Character, error) {
	if parser.pendingCharacter != nil {
		character := parser.pendingCharacter
		parser.pendingCharacter = nil
		return character, nil
	}
	if parser.done() {
		return nil, parser.fail("unterminated character class")
	}
	if parser.peek() == '\\' {
		n, _, escapeError := parser.escape(true)
		if escapeError != nil {
			return nil, escapeError
		}
		c, ok := n.(*Character)
		if !ok {
			return nil, parser.fail("invalid class escape")
		}
		return c, nil
	}
	if parser.flags.UnicodeSets &&
		(strings.ContainsRune("(){}|/-", rune(parser.peek())) || parser.pos+1 < len(parser.source) &&
			parser.source[parser.pos] == parser.source[parser.pos+1] &&
			strings.ContainsRune("!#$%&*+,.:;<=>?@^`~", rune(parser.peek()))) {
		return nil, parser.fail("reserved character in Unicode set")
	}
	n, _, literalError := parser.literal()
	return n.(*Character), literalError
}

func classMayContainStrings(expression ClassExpression) bool {
	switch expression := expression.(type) {
	case *ClassString:
		for _, alternative := range expression.Alternatives {
			if len(alternative) != 1 {
				return true
			}
		}
	case *ClassCharacter:
		if expression.Character.Kind == PropertyEscape {
			raw := expression.Character.Raw
			if open := strings.IndexByte(raw, '{'); open >= 0 && len(raw) > open+2 {
				return stringProperties[raw[open+1:len(raw)-1]]
			}
		}
	case *ClassUnion:
		for _, operand := range expression.Operands {
			if classMayContainStrings(operand) {
				return true
			}
		}
	case *ClassIntersection:
		return classMayContainStrings(expression.Left) && classMayContainStrings(expression.Right)
	case *ClassSubtraction:
		return classMayContainStrings(expression.Left)
	}
	return false
}

type namedCaptureLocation struct {
	path map[*Disjunction]int
}

func validateDuplicateNames(body *Disjunction) error {
	locations := make(map[string][]namedCaptureLocation)
	var visitDisjunction func(*Disjunction, map[*Disjunction]int)
	var visitNode func(Node, map[*Disjunction]int)
	visitNode = func(node Node, path map[*Disjunction]int) {
		switch node := node.(type) {
		case *Group:
			if node.Kind == Capturing && node.Name != "" {
				copyPath := make(map[*Disjunction]int, len(path))
				for d, a := range path {
					copyPath[d] = a
				}
				locations[node.Name] = append(locations[node.Name], namedCaptureLocation{path: copyPath})
			}
			visitDisjunction(node.Body, path)
		case *Quantifier:
			visitNode(node.Atom, path)
		}
	}
	visitDisjunction = func(disjunction *Disjunction, path map[*Disjunction]int) {
		for alternativeIndex, alternative := range disjunction.Alternatives {
			path[disjunction] = alternativeIndex
			for _, term := range alternative.Terms {
				visitNode(term, path)
			}
		}
		delete(path, disjunction)
	}
	visitDisjunction(body, make(map[*Disjunction]int))
	for name, captures := range locations {
		for i := range captures {
			for j := 0; j < i; j++ {
				disjoint := false
				for d, a := range captures[i].path {
					if b, ok := captures[j].path[d]; ok && a != b {
						disjoint = true
						break
					}
				}
				if !disjoint {
					return &SyntaxError{Message: "duplicate capture name " + name}
				}
			}
		}
	}
	return nil
}

func (parser *parser) literal() (Node, bool, error) {
	if parser.pendingCharacter != nil {
		character := parser.pendingCharacter
		parser.pendingCharacter = nil
		return character, true, nil
	}
	start := parser.pos
	r, n := patternRune(parser.source[parser.pos:])
	if r == utf8.RuneError && n == 1 {
		return nil, false, parser.fail("invalid UTF-8")
	}
	parser.pos += n
	if r > 0xffff && !parser.flags.Unicode && !parser.flags.UnicodeSets {
		value := r - 0x10000
		parser.pendingCharacter = &Character{
			Value: 0xdc00 + value%0x400,
			Raw:   parser.source[start:parser.pos],
			Kind:  Literal,
		}
		r = 0xd800 + value/0x400
	}
	return &Character{Value: r, Raw: parser.source[start:parser.pos], Kind: Literal}, true, nil
}
func (parser *parser) char(start int, kind CharacterKind) (Node, bool, error) {
	return &Character{Raw: parser.source[start:parser.pos], Kind: kind}, true, nil
}
func (parser *parser) unicodeEscape(start int) (*Character, error) {
	if parser.take('{') {
		if !(parser.flags.Unicode || parser.flags.UnicodeSets) {
			parser.pos = start + 2
			return &Character{Value: 'u', Raw: parser.source[start:parser.pos], Kind: Escape}, nil
		}
		n, ok := parser.hexNumber('}')
		if !ok || n > 0x10ffff {
			return nil, parser.failAt(start, "invalid Unicode escape")
		}
		return &Character{Value: rune(n), Raw: parser.source[start:parser.pos], Kind: Escape}, nil
	}
	first, hexEscapeError := parser.hexEscape(start, 4)
	if hexEscapeError != nil && !parser.flags.Unicode && !parser.flags.UnicodeSets {
		parser.pos = start + 2
		return &Character{Value: 'u', Raw: parser.source[start:parser.pos], Kind: Escape}, nil
	}
	if hexEscapeError != nil {
		return nil, hexEscapeError
	}
	// In Unicode mode a paired Unicode escape is one atom, before a
	// quantifier or class range is parsed.
	if (parser.flags.Unicode || parser.flags.UnicodeSets) &&
		first.Value >= 0xd800 &&
		first.Value <= 0xdbff &&
		parser.match("\\u") {
		secondStart := parser.pos
		parser.pos += 2
		second, secondError := parser.hexEscape(secondStart, 4)
		if secondError == nil && second.Value >= 0xdc00 && second.Value <= 0xdfff {
			first.Value = 0x10000 + (first.Value-0xd800)*0x400 + second.Value - 0xdc00
			first.Raw = parser.source[start:parser.pos]
		} else {
			parser.pos = secondStart
		}
	}
	return first, nil
}
func (parser *parser) hexEscape(start, n int) (*Character, error) {
	if parser.pos+n > len(parser.source) {
		return nil, parser.failAt(start, "invalid hex escape")
	}
	v, parseError := strconv.ParseUint(parser.source[parser.pos:parser.pos+n], 16, 32)
	if parseError != nil {
		return nil, parser.failAt(start, "invalid hex escape")
	}
	parser.pos += n
	return &Character{Value: rune(v), Raw: parser.source[start:parser.pos], Kind: Escape}, nil
}
func (parser *parser) legacyOctal(start int) (Node, bool, error) {
	if parser.peek() == '8' || parser.peek() == '9' {
		value := rune(parser.peek())
		parser.pos++
		return &Character{Value: value, Raw: parser.source[start:parser.pos], Kind: Escape}, true, nil
	}
	limit := 3
	if parser.peek() >= '4' {
		limit = 2
	}
	n := 0
	for i := 0; i < limit && !parser.done() && parser.peek() >= '0' && parser.peek() <= '7'; i++ {
		n = n*8 + int(parser.peek()-'0')
		parser.pos++
	}
	return &Character{Value: rune(n), Raw: parser.source[start:parser.pos], Kind: Escape}, true, nil
}

func (parser *parser) groupName() (string, error) {
	start := parser.pos
	var name []rune
	for !parser.done() && parser.peek() != '>' {
		r, groupNameRuneError := parser.groupNameRune()
		if groupNameRuneError != nil ||
			len(name) == 0 && !isRegExpIdentifierStart(r) ||
			len(name) != 0 && !isRegExpIdentifierPart(r) {
			return "", parser.failAt(start, "invalid capture name")
		}
		name = append(name, r)
	}
	if len(name) == 0 || !parser.take('>') {
		return "", parser.fail("invalid capture name")
	}
	return string(name), nil
}

func (parser *parser) groupNameRune() (rune, error) {
	if parser.peek() != '\\' {
		r, width := patternRune(parser.source[parser.pos:])
		if r == utf8.RuneError && width == 1 {
			return 0, parser.fail("invalid UTF-8")
		}
		parser.pos += width
		return r, nil
	}
	start := parser.pos
	parser.pos++
	if !parser.take('u') {
		return 0, parser.failAt(start, "invalid capture name escape")
	}
	if parser.take('{') {
		value, ok := parser.hexNumber('}')
		if !ok || value > utf8.MaxRune {
			return 0, parser.failAt(start, "invalid capture name escape")
		}
		return rune(value), nil
	}
	first, hexEscapeError := parser.hexEscape(start, 4)
	if hexEscapeError != nil {
		return 0, hexEscapeError
	}
	if first.Value >= 0xD800 && first.Value <= 0xDBFF && parser.match("\\u") {
		secondStart := parser.pos
		parser.pos += 2
		second, secondError := parser.hexEscape(secondStart, 4)
		if secondError == nil && second.Value >= 0xDC00 && second.Value <= 0xDFFF {
			return utf8.RuneSelf + (first.Value-0xD800)*0x400 + second.Value - 0xDC00 - utf8.RuneSelf + 0x10000, nil
		}
		parser.pos = secondStart
	}
	return first.Value, nil
}

func isRegExpIdentifierStart(r rune) bool {
	property, _ := unicodeproperties.Lookup("ID_Start", false)
	return r == '$' || r == '_' || property.Contains(r)
}

func isRegExpIdentifierPart(r rune) bool {
	property, _ := unicodeproperties.Lookup("ID_Continue", false)
	return isRegExpIdentifierStart(r) || r == 0x200C || r == 0x200D || property.Contains(r)
}
func (parser *parser) decimal() (*big.Int, bool) {
	start := parser.pos
	for !parser.done() && parser.peek() >= '0' && parser.peek() <= '9' {
		parser.pos++
	}
	if parser.pos == start {
		return nil, false
	}
	n, ok := new(big.Int).SetString(parser.source[start:parser.pos], 10)
	return n, ok
}
func (parser *parser) hexNumber(end byte) (int, bool) {
	start := parser.pos
	n := 0
	for !parser.done() && parser.peek() != end {
		c := parser.peek()
		d := -1
		if c >= '0' && c <= '9' {
			d = int(c - '0')
		} else if c >= 'a' && c <= 'f' {
			d = int(c-'a') + 10
		} else if c >= 'A' && c <= 'F' {
			d = int(c-'A') + 10
		}
		if d < 0 || n > (utf8.MaxRune-d)/16 {
			return 0, false
		}
		n = n*16 + d
		parser.pos++
	}
	if parser.pos == start || !parser.take(end) {
		return 0, false
	}
	return n, true
}
func (parser *parser) done() bool {
	return parser.pendingCharacter == nil && parser.pos >= len(parser.source)
}
func (parser *parser) peek() byte {
	if parser.pendingCharacter != nil {
		// A pending low surrogate cannot begin ASCII pattern syntax.
		return 0xff
	}
	if parser.done() {
		return 0
	}
	return parser.source[parser.pos]
}
func (parser *parser) take(c byte) bool {
	if parser.peek() == c {
		parser.pos++
		return true
	}
	return false
}
func (parser *parser) match(s string) bool          { return strings.HasPrefix(parser.source[parser.pos:], s) }
func (parser *parser) fail(s string) error          { return parser.failAt(parser.pos, s) }
func (parser *parser) failAt(n int, s string) error { return &SyntaxError{Offset: n, Message: s} }
func (parser *parser) quantifierAhead() bool {
	c := parser.peek()
	if c == '*' || c == '+' || c == '?' {
		return true
	}
	if c != '{' {
		return false
	}
	i := parser.pos + 1
	if i >= len(parser.source) || parser.source[i] < '0' || parser.source[i] > '9' {
		return false
	}
	for i < len(parser.source) && parser.source[i] >= '0' && parser.source[i] <= '9' {
		i++
	}
	if i < len(parser.source) && parser.source[i] == ',' {
		i++
		for i < len(parser.source) && parser.source[i] >= '0' && parser.source[i] <= '9' {
			i++
		}
	}
	return i < len(parser.source) && parser.source[i] == '}'
}
func isIdentifierPart(c byte) bool {
	return c == '_' || c == '$' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}
func countCaptures(s string) (int, bool) {
	n := 0
	named := false
	escaped, inClass := false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' {
			escaped = true
			continue
		}
		if c == '[' {
			inClass = true
			continue
		}
		if c == ']' {
			inClass = false
			continue
		}
		if !inClass && c == '(' && (i+1 >= len(s) ||
			s[i+1] != '?' ||
			(i+2 < len(s) && s[i+1:i+3] == "?<" && (i+3 >= len(s) || s[i+3] != '=' && s[i+3] != '!'))) {
			n++
			if i+2 < len(s) && s[i+1:i+3] == "?<" {
				named = true
			}
		}
	}
	return n, named
}
