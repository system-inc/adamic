package regexp

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// SyntaxError is an ECMAScript early error. Offset is a byte offset in Pattern.
type SyntaxError struct {
	Offset  int
	Message string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("regexp: %s at byte %d", e.Message, e.Offset)
}

// Parse parses pattern and flags according to ECMAScript 2025.
func Parse(pattern, flags string) (*Pattern, error) {
	f, err := parseFlags(flags)
	if err != nil {
		return nil, err
	}
	p := &parser{source: pattern, flags: f, names: make(map[string]bool)}
	// Decimal escapes depend on the total capture count, including captures to
	// their right. Count them without interpreting pattern contents first.
	p.captureCount, p.hasNamedCapture = countCaptures(pattern)
	body, err := p.disjunction(0)
	if err != nil {
		return nil, err
	}
	if !p.done() {
		return nil, p.fail("unexpected character")
	}
	for name := range p.namedReferences {
		if !p.names[name] {
			return nil, p.failAt(p.namedReferences[name], "unknown capture name")
		}
	}
	if err := validateDuplicateNames(body); err != nil {
		return nil, err
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
	if width == 1 && r == utf8.RuneError && len(source) >= 3 && source[0] == 0xed && source[1] >= 0xa0 && source[1] <= 0xbf && source[2] >= 0x80 && source[2] <= 0xbf {
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
	source                 string
	pos                    int
	flags                  Flags
	captureCount, captures int
	hasNamedCapture        bool
	pendingCharacter       *Character
	names                  map[string]bool
	namedReferences        map[string]int
}

func (p *parser) disjunction(stop byte) (*Disjunction, error) {
	d := &Disjunction{}
	for {
		a, err := p.alternative(stop)
		if err != nil {
			return nil, err
		}
		d.Alternatives = append(d.Alternatives, a)
		if p.done() || p.peek() != '|' {
			break
		}
		p.pos++
	}
	return d, nil
}

func (p *parser) alternative(stop byte) (*Alternative, error) {
	a := &Alternative{}
	for !p.done() && p.peek() != '|' && p.peek() != stop {
		if p.peek() == ')' && stop == 0 {
			return nil, p.fail("unmatched closing parenthesis")
		}
		n, quantifiable, err := p.term()
		if err != nil {
			return nil, err
		}
		if p.quantifierAhead() {
			if !quantifiable {
				return nil, p.fail("nothing to repeat")
			}
			n, err = p.quantifier(n)
			if err != nil {
				return nil, err
			}
			if p.quantifierAhead() {
				return nil, p.fail("nothing to repeat")
			}
		}
		a.Terms = append(a.Terms, n)
	}
	return a, nil
}

func (p *parser) term() (Node, bool, error) {
	switch p.peek() {
	case '^':
		p.pos++
		return &Assertion{Kind: Start}, false, nil
	case '$':
		p.pos++
		return &Assertion{Kind: End}, false, nil
	case '.':
		p.pos++
		return &Dot{}, true, nil
	case '[':
		n, e := p.characterClass()
		return n, true, e
	case '(':
		n, q, e := p.group()
		return n, q, e
	case '\\':
		return p.escape(false)
	case '*', '+', '?':
		return nil, false, p.fail("nothing to repeat")
	case '{':
		if p.quantifierAhead() {
			return nil, false, p.fail("nothing to repeat")
		}
		if p.flags.Unicode || p.flags.UnicodeSets {
			return nil, false, p.fail("incomplete quantifier")
		}
	case ']', '}':
		if p.flags.Unicode || p.flags.UnicodeSets {
			return nil, false, p.fail("lone punctuation")
		}
	}
	return p.literal()
}

func (p *parser) group() (Node, bool, error) {
	p.pos++
	kind, name := Capturing, ""
	var enable, disable Flags
	if p.take('?') {
		switch {
		case p.take(':'):
			kind = NonCapturing
		case p.take('='):
			kind = PositiveLookahead
		case p.take('!'):
			kind = NegativeLookahead
		case p.take('<'):
			if p.take('=') {
				kind = PositiveLookbehind
			} else if p.take('!') {
				kind = NegativeLookbehind
			} else {
				kind = Capturing
				var err error
				name, err = p.groupName()
				if err != nil {
					return nil, false, err
				}
				p.names[name] = true
			}
		case strings.ContainsRune("ims-", rune(p.peek())):
			var err error
			enable, disable, err = p.modifiers()
			if err != nil {
				return nil, false, err
			}
			kind = NonCapturing
		default:
			return nil, false, p.fail("invalid group")
		}
	}
	if kind == Capturing {
		p.captures++
	}
	body, err := p.disjunction(')')
	if err != nil {
		return nil, false, err
	}
	if !p.take(')') {
		return nil, false, p.fail("unterminated group")
	}
	q := kind == Capturing || kind == NonCapturing
	if !p.flags.Unicode && !p.flags.UnicodeSets && (kind == PositiveLookahead || kind == NegativeLookahead) {
		q = true
	}
	return &Group{Kind: kind, Name: name, Body: body, Enable: enable, Disable: disable}, q, nil
}

func (p *parser) modifiers() (Flags, Flags, error) {
	var enable, disable Flags
	disabling, any := false, false
	seen := make(map[byte]bool)
	for !p.done() && p.peek() != ':' {
		c := p.peek()
		if c == '-' && !disabling {
			disabling = true
			p.pos++
			continue
		}
		if c != 'i' && c != 'm' && c != 's' || seen[c] {
			return enable, disable, p.fail("invalid modifiers")
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
		p.pos++
	}
	if !any || !p.take(':') {
		return enable, disable, p.fail("invalid modifiers")
	}
	return enable, disable, nil
}

func (p *parser) quantifier(atom Node) (Node, error) {
	min, max := big.NewInt(0), (*big.Int)(nil)
	switch p.peek() {
	case '*':
		p.pos++
	case '+':
		p.pos++
		min = big.NewInt(1)
	case '?':
		p.pos++
		max = big.NewInt(1)
	case '{':
		start := p.pos
		p.pos++
		m, ok := p.decimal()
		if !ok {
			p.pos = start
			return nil, p.fail("incomplete quantifier")
		}
		min, max = m, new(big.Int).Set(m)
		if p.take(',') {
			max = nil
			if n, ok := p.decimal(); ok {
				max = n
			}
		}
		if !p.take('}') {
			p.pos = start
			return nil, p.fail("incomplete quantifier")
		}
		if max != nil && min.Cmp(max) > 0 {
			return nil, p.failAt(start, "quantifier range out of order")
		}
	}
	greedy := !p.take('?')
	return &Quantifier{Atom: atom, Min: min, Max: max, Greedy: greedy}, nil
}

func (p *parser) escape(inClass bool) (Node, bool, error) {
	start := p.pos
	p.pos++
	if p.done() {
		return nil, false, p.fail("trailing escape")
	}
	if p.peek() >= utf8.RuneSelf {
		if p.flags.Unicode || p.flags.UnicodeSets {
			return nil, false, p.failAt(start, "invalid identity escape")
		}
		node, _, err := p.literal()
		if err != nil {
			return nil, false, err
		}
		character := node.(*Character)
		character.Kind = Escape
		character.Raw = p.source[start:p.pos]
		return character, true, nil
	}
	c := p.peek()
	p.pos++
	if !inClass && (c == 'b' || c == 'B') {
		kind := WordBoundary
		if c == 'B' {
			kind = NotWordBoundary
		}
		return &Assertion{Kind: kind}, false, nil
	}
	if !inClass && c >= '1' && c <= '9' {
		p.pos--
		n, _ := p.decimal()
		if n.IsInt64() && n.Int64() <= int64(p.captureCount) {
			return &Backreference{Index: int(n.Int64())}, true, nil
		}
		if p.flags.Unicode || p.flags.UnicodeSets {
			return nil, false, p.failAt(start, "invalid decimal escape")
		}
		p.pos = start + 1
		return p.legacyOctal(start)
	}
	if !inClass && c == 'k' {
		if !p.hasNamedCapture {
			if p.flags.Unicode || p.flags.UnicodeSets {
				return nil, false, p.failAt(start, "invalid named reference")
			}
			return &Character{Value: 'k', Raw: p.source[start:p.pos], Kind: Escape}, true, nil
		}
		if !p.take('<') {
			return nil, false, p.fail("invalid named reference")
		}
		name, err := p.groupName()
		if err != nil {
			return nil, false, err
		}
		if p.namedReferences == nil {
			p.namedReferences = make(map[string]int)
		}
		p.namedReferences[name] = start
		return &Backreference{Name: name}, true, nil
	}
	if c == 'p' || c == 'P' {
		if !(p.flags.Unicode || p.flags.UnicodeSets) {
			return &Character{Value: rune(c), Raw: p.source[start:p.pos], Kind: Escape}, true, nil
		}
		if !p.take('{') {
			return nil, false, p.failAt(start, "invalid property escape")
		}
		end := strings.IndexByte(p.source[p.pos:], '}')
		if end < 0 {
			return nil, false, p.fail("unterminated property escape")
		}
		property := p.source[p.pos : p.pos+end]
		p.pos += end + 1
		if !validProperty(property, p.flags.UnicodeSets) || c == 'P' && stringProperties[property] {
			return nil, false, p.failAt(start, "invalid Unicode property")
		}
		return &Character{Raw: p.source[start:p.pos], Kind: PropertyEscape}, true, nil
	}
	if strings.ContainsRune("dDsSwW", rune(c)) {
		return &Character{Raw: p.source[start:p.pos], Kind: ClassEscape}, true, nil
	}
	if inClass && c == 'b' {
		return &Character{Value: '\b', Raw: p.source[start:p.pos], Kind: Escape}, true, nil
	}
	if c == '0' {
		if !p.done() && p.peek() >= '0' && p.peek() <= '9' {
			if p.flags.Unicode || p.flags.UnicodeSets {
				return nil, false, p.failAt(start, "invalid decimal escape")
			}
			p.pos = start + 1
			return p.legacyOctal(start)
		}
		return &Character{Value: 0, Raw: p.source[start:p.pos], Kind: Escape}, true, nil
	}
	if c == 'c' {
		if !p.done() && (p.peek() >= 'a' && p.peek() <= 'z' || p.peek() >= 'A' && p.peek() <= 'Z' || inClass && !p.flags.Unicode && !p.flags.UnicodeSets && (p.peek() >= '0' && p.peek() <= '9' || p.peek() == '_')) {
			value := p.peek() % 32
			p.pos++
			return &Character{Value: rune(value), Raw: p.source[start:p.pos], Kind: Escape}, true, nil
		}
		if p.flags.Unicode || p.flags.UnicodeSets {
			return nil, false, p.failAt(start, "invalid control escape")
		}
		p.pos = start + 1
		return &Character{Value: '\\', Raw: p.source[start:p.pos], Kind: Literal}, true, nil
	}
	if c == 'u' {
		ch, err := p.unicodeEscape(start)
		return ch, true, err
	}
	if c == 'x' {
		ch, err := p.hexEscape(start, 2)
		if err != nil && !p.flags.Unicode && !p.flags.UnicodeSets {
			p.pos = start + 2
			return &Character{Value: 'x', Raw: p.source[start:p.pos], Kind: Escape}, true, nil
		}
		return ch, true, err
	}
	if strings.ContainsRune("fnrtv", rune(c)) {
		return &Character{Value: map[byte]rune{'f': 12, 'n': 10, 'r': 13, 't': 9, 'v': 11}[c], Raw: p.source[start:p.pos], Kind: Escape}, true, nil
	}
	if (p.flags.Unicode || p.flags.UnicodeSets) && !strings.ContainsRune("^$\\.*+?()[]{}|/", rune(c)) && !(inClass && (c == '-' || p.flags.UnicodeSets && strings.ContainsRune("!#%&,:;<=>@`~", rune(c)))) {
		return nil, false, p.failAt(start, "invalid identity escape")
	}
	return &Character{Value: rune(c), Raw: p.source[start:p.pos], Kind: Escape}, true, nil
}

func (p *parser) characterClass() (Node, error) {
	p.pos++
	neg := p.take('^')
	union := &ClassUnion{}
	for !p.done() && p.peek() != ']' {
		operand, err := p.classSetOperand()
		if err != nil {
			return nil, err
		}
		union.Operands = append(union.Operands, operand)
		if p.flags.UnicodeSets && (p.match("&&") || p.match("--")) {
			if len(union.Operands) != 1 {
				return nil, p.fail("invalid Unicode set operation")
			}
			if _, rangeOperand := operand.(*ClassRange); rangeOperand {
				return nil, p.fail("range must be nested in Unicode set operation")
			}
			op := p.source[p.pos : p.pos+2]
			p.pos += 2
			left := ClassExpression(union)
			for {
				right, operandErr := p.classSetOperand()
				if operandErr != nil {
					return nil, operandErr
				}
				if _, rangeOperand := right.(*ClassRange); rangeOperand {
					return nil, p.fail("range must be nested in Unicode set operation")
				}
				if op == "&&" {
					left = &ClassIntersection{Left: left, Right: right}
				} else {
					left = &ClassSubtraction{Left: left, Right: right}
				}
				if !p.match(op) {
					break
				}
				p.pos += 2
			}
			return p.finishClass(neg, left)
		}
	}
	return p.finishClass(neg, union)
}

func (p *parser) finishClass(neg bool, expr ClassExpression) (Node, error) {
	if !p.take(']') {
		return nil, p.fail("unterminated character class")
	}
	if neg && classMayContainStrings(expr) {
		return nil, p.fail("cannot negate a class containing strings")
	}
	return &CharacterClass{Negated: neg, Expr: expr}, nil
}
func (p *parser) classSetOperand() (ClassExpression, error) {
	if p.flags.UnicodeSets && p.peek() == '[' {
		n, e := p.characterClass()
		if e != nil {
			return nil, e
		}
		class := n.(*CharacterClass)
		if class.Negated {
			return &ClassNegation{Operand: class.Expr}, nil
		}
		return class.Expr, nil
	}
	if p.flags.UnicodeSets && p.match("\\q{") {
		p.pos += 3
		result := &ClassString{Alternatives: [][]*Character{{}}}
		for !p.done() && p.peek() != '}' {
			if p.take('|') {
				result.Alternatives = append(result.Alternatives, []*Character{})
				continue
			}
			character, err := p.classCharacter()
			if err != nil {
				return nil, err
			}
			last := len(result.Alternatives) - 1
			result.Alternatives[last] = append(result.Alternatives[last], character)
		}
		if !p.take('}') {
			return nil, p.fail("unterminated class string")
		}
		return result, nil
	}
	c, e := p.classCharacter()
	if e != nil {
		return nil, e
	}
	if p.peek() == '-' && p.pos+1 < len(p.source) && p.source[p.pos+1] != ']' && !(p.flags.UnicodeSets && p.match("--")) {
		p.pos++
		right, err := p.classCharacter()
		if err != nil {
			return nil, err
		}
		if c.Kind == ClassEscape || c.Kind == PropertyEscape || right.Kind == ClassEscape || right.Kind == PropertyEscape {
			if !p.flags.Unicode && !p.flags.UnicodeSets {
				return &ClassUnion{Operands: []ClassExpression{&ClassCharacter{Character: c}, &ClassCharacter{Character: &Character{Value: '-', Raw: "-", Kind: Literal}}, &ClassCharacter{Character: right}}}, nil
			}
			return nil, p.fail("invalid character class range")
		}
		if c.Value > right.Value {
			return nil, p.fail("invalid character class range")
		}
		return &ClassRange{From: c, To: right}, nil
	}
	return &ClassCharacter{Character: c}, nil
}
func (p *parser) classCharacter() (*Character, error) {
	if p.pendingCharacter != nil {
		character := p.pendingCharacter
		p.pendingCharacter = nil
		return character, nil
	}
	if p.done() {
		return nil, p.fail("unterminated character class")
	}
	if p.peek() == '\\' {
		n, _, e := p.escape(true)
		if e != nil {
			return nil, e
		}
		c, ok := n.(*Character)
		if !ok {
			return nil, p.fail("invalid class escape")
		}
		return c, nil
	}
	if p.flags.UnicodeSets && (strings.ContainsRune("(){}|/-", rune(p.peek())) || p.pos+1 < len(p.source) && p.source[p.pos] == p.source[p.pos+1] && strings.ContainsRune("!#$%&*+,.:;<=>?@^`~", rune(p.peek()))) {
		return nil, p.fail("reserved character in Unicode set")
	}
	n, _, e := p.literal()
	return n.(*Character), e
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

func (p *parser) literal() (Node, bool, error) {
	if p.pendingCharacter != nil {
		character := p.pendingCharacter
		p.pendingCharacter = nil
		return character, true, nil
	}
	start := p.pos
	r, n := patternRune(p.source[p.pos:])
	if r == utf8.RuneError && n == 1 {
		return nil, false, p.fail("invalid UTF-8")
	}
	p.pos += n
	if r > 0xffff && !p.flags.Unicode && !p.flags.UnicodeSets {
		value := r - 0x10000
		p.pendingCharacter = &Character{Value: 0xdc00 + value%0x400, Raw: p.source[start:p.pos], Kind: Literal}
		r = 0xd800 + value/0x400
	}
	return &Character{Value: r, Raw: p.source[start:p.pos], Kind: Literal}, true, nil
}
func (p *parser) char(start int, kind CharacterKind) (Node, bool, error) {
	return &Character{Raw: p.source[start:p.pos], Kind: kind}, true, nil
}
func (p *parser) unicodeEscape(start int) (*Character, error) {
	if p.take('{') {
		if !(p.flags.Unicode || p.flags.UnicodeSets) {
			p.pos = start + 2
			return &Character{Value: 'u', Raw: p.source[start:p.pos], Kind: Escape}, nil
		}
		n, ok := p.hexNumber('}')
		if !ok || n > 0x10ffff {
			return nil, p.failAt(start, "invalid Unicode escape")
		}
		return &Character{Value: rune(n), Raw: p.source[start:p.pos], Kind: Escape}, nil
	}
	first, err := p.hexEscape(start, 4)
	if err != nil && !p.flags.Unicode && !p.flags.UnicodeSets {
		p.pos = start + 2
		return &Character{Value: 'u', Raw: p.source[start:p.pos], Kind: Escape}, nil
	}
	if err != nil {
		return nil, err
	}
	// In Unicode mode a paired Unicode escape is one atom, before a
	// quantifier or class range is parsed.
	if (p.flags.Unicode || p.flags.UnicodeSets) && first.Value >= 0xd800 && first.Value <= 0xdbff && p.match("\\u") {
		secondStart := p.pos
		p.pos += 2
		second, secondErr := p.hexEscape(secondStart, 4)
		if secondErr == nil && second.Value >= 0xdc00 && second.Value <= 0xdfff {
			first.Value = 0x10000 + (first.Value-0xd800)*0x400 + second.Value - 0xdc00
			first.Raw = p.source[start:p.pos]
		} else {
			p.pos = secondStart
		}
	}
	return first, nil
}
func (p *parser) hexEscape(start, n int) (*Character, error) {
	if p.pos+n > len(p.source) {
		return nil, p.failAt(start, "invalid hex escape")
	}
	v, e := strconv.ParseUint(p.source[p.pos:p.pos+n], 16, 32)
	if e != nil {
		return nil, p.failAt(start, "invalid hex escape")
	}
	p.pos += n
	return &Character{Value: rune(v), Raw: p.source[start:p.pos], Kind: Escape}, nil
}
func (p *parser) legacyOctal(start int) (Node, bool, error) {
	if p.peek() == '8' || p.peek() == '9' {
		value := rune(p.peek())
		p.pos++
		return &Character{Value: value, Raw: p.source[start:p.pos], Kind: Escape}, true, nil
	}
	limit := 3
	if p.peek() >= '4' {
		limit = 2
	}
	n := 0
	for i := 0; i < limit && !p.done() && p.peek() >= '0' && p.peek() <= '7'; i++ {
		n = n*8 + int(p.peek()-'0')
		p.pos++
	}
	return &Character{Value: rune(n), Raw: p.source[start:p.pos], Kind: Escape}, true, nil
}

func (p *parser) groupName() (string, error) {
	start := p.pos
	var name []rune
	for !p.done() && p.peek() != '>' {
		r, err := p.groupNameRune()
		if err != nil || len(name) == 0 && !isRegExpIdentifierStart(r) || len(name) != 0 && !isRegExpIdentifierPart(r) {
			return "", p.failAt(start, "invalid capture name")
		}
		name = append(name, r)
	}
	if len(name) == 0 || !p.take('>') {
		return "", p.fail("invalid capture name")
	}
	return string(name), nil
}

func (p *parser) groupNameRune() (rune, error) {
	if p.peek() != '\\' {
		r, width := patternRune(p.source[p.pos:])
		if r == utf8.RuneError && width == 1 {
			return 0, p.fail("invalid UTF-8")
		}
		p.pos += width
		return r, nil
	}
	start := p.pos
	p.pos++
	if !p.take('u') {
		return 0, p.failAt(start, "invalid capture name escape")
	}
	if p.take('{') {
		value, ok := p.hexNumber('}')
		if !ok || value > utf8.MaxRune {
			return 0, p.failAt(start, "invalid capture name escape")
		}
		return rune(value), nil
	}
	first, err := p.hexEscape(start, 4)
	if err != nil {
		return 0, err
	}
	if first.Value >= 0xD800 && first.Value <= 0xDBFF && p.match("\\u") {
		secondStart := p.pos
		p.pos += 2
		second, secondErr := p.hexEscape(secondStart, 4)
		if secondErr == nil && second.Value >= 0xDC00 && second.Value <= 0xDFFF {
			return utf8.RuneSelf + (first.Value-0xD800)*0x400 + second.Value - 0xDC00 - utf8.RuneSelf + 0x10000, nil
		}
		p.pos = secondStart
	}
	return first.Value, nil
}

func isRegExpIdentifierStart(r rune) bool {
	return r == '$' || r == '_' || unicode.IsLetter(r) || unicode.In(r, unicode.Nl)
}

func isRegExpIdentifierPart(r rune) bool {
	return isRegExpIdentifierStart(r) || r == 0x200C || r == 0x200D || unicode.In(r, unicode.Mn, unicode.Mc, unicode.Nd, unicode.Pc)
}
func (p *parser) decimal() (*big.Int, bool) {
	start := p.pos
	for !p.done() && p.peek() >= '0' && p.peek() <= '9' {
		p.pos++
	}
	if p.pos == start {
		return nil, false
	}
	n, ok := new(big.Int).SetString(p.source[start:p.pos], 10)
	return n, ok
}
func (p *parser) hexNumber(end byte) (int, bool) {
	start := p.pos
	n := 0
	for !p.done() && p.peek() != end {
		c := p.peek()
		d := -1
		if c >= '0' && c <= '9' {
			d = int(c - '0')
		} else if c >= 'a' && c <= 'f' {
			d = int(c-'a') + 10
		} else if c >= 'A' && c <= 'F' {
			d = int(c-'A') + 10
		}
		if d < 0 {
			return 0, false
		}
		n = n*16 + d
		p.pos++
	}
	if p.pos == start || !p.take(end) {
		return 0, false
	}
	return n, true
}
func (p *parser) done() bool { return p.pendingCharacter == nil && p.pos >= len(p.source) }
func (p *parser) peek() byte {
	if p.pendingCharacter != nil {
		// A pending low surrogate cannot begin ASCII pattern syntax.
		return 0xff
	}
	if p.done() {
		return 0
	}
	return p.source[p.pos]
}
func (p *parser) take(c byte) bool {
	if p.peek() == c {
		p.pos++
		return true
	}
	return false
}
func (p *parser) match(s string) bool          { return strings.HasPrefix(p.source[p.pos:], s) }
func (p *parser) fail(s string) error          { return p.failAt(p.pos, s) }
func (p *parser) failAt(n int, s string) error { return &SyntaxError{Offset: n, Message: s} }
func (p *parser) quantifierAhead() bool {
	c := p.peek()
	if c == '*' || c == '+' || c == '?' {
		return true
	}
	if c != '{' {
		return false
	}
	i := p.pos + 1
	if i >= len(p.source) || p.source[i] < '0' || p.source[i] > '9' {
		return false
	}
	for i < len(p.source) && p.source[i] >= '0' && p.source[i] <= '9' {
		i++
	}
	if i < len(p.source) && p.source[i] == ',' {
		i++
		for i < len(p.source) && p.source[i] >= '0' && p.source[i] <= '9' {
			i++
		}
	}
	return i < len(p.source) && p.source[i] == '}'
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
		if !inClass && c == '(' && (i+1 >= len(s) || s[i+1] != '?' || (i+2 < len(s) && s[i+1:i+3] == "?<" && (i+3 >= len(s) || s[i+3] != '=' && s[i+3] != '!'))) {
			n++
			if i+2 < len(s) && s[i+1:i+3] == "?<" {
				named = true
			}
		}
	}
	return n, named
}
