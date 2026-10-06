package regexp

import (
	"fmt"
	"strconv"
	"strings"
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
	return &Pattern{Flags: f, Body: body}, nil
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
				if p.names[name] {
					return nil, false, p.fail("duplicate capture name")
				}
				p.names[name] = true
			}
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
	return &Group{Kind: kind, Name: name, Body: body}, q, nil
}

func (p *parser) quantifier(atom Node) (Node, error) {
	min, max := 0, -1
	switch p.peek() {
	case '*':
		p.pos++
	case '+':
		p.pos++
		min = 1
	case '?':
		p.pos++
		max = 1
	case '{':
		start := p.pos
		p.pos++
		m, ok := p.decimal()
		if !ok {
			p.pos = start
			return nil, p.fail("incomplete quantifier")
		}
		min, max = m, m
		if p.take(',') {
			max = -1
			if n, ok := p.decimal(); ok {
				max = n
			}
		}
		if !p.take('}') {
			p.pos = start
			return nil, p.fail("incomplete quantifier")
		}
		if max >= 0 && min > max {
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
		if n <= p.captureCount {
			return &Backreference{Index: n}, true, nil
		}
		if p.flags.Unicode || p.flags.UnicodeSets {
			return nil, false, p.failAt(start, "invalid decimal escape")
		}
		p.pos = start + 1
		return p.legacyOctal(start)
	}
	if !inClass && c == 'k' {
		if !p.hasNamedCapture {
			return p.char(start, Escape)
		}
		if !p.take('<') {
			if p.flags.Unicode || p.flags.UnicodeSets {
				return nil, false, p.fail("invalid named reference")
			}
			return p.char(start, Escape)
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
		if !(p.flags.Unicode || p.flags.UnicodeSets) || !p.take('{') {
			return nil, false, p.failAt(start, "invalid property escape")
		}
		end := strings.IndexByte(p.source[p.pos:], '}')
		if end < 0 {
			return nil, false, p.fail("unterminated property escape")
		}
		property := p.source[p.pos : p.pos+end]
		p.pos += end + 1
		if !validProperty(property, p.flags.UnicodeSets) {
			return nil, false, p.failAt(start, "invalid Unicode property")
		}
		return &Character{Raw: p.source[start:p.pos], Kind: PropertyEscape}, true, nil
	}
	if strings.ContainsRune("dDsSwW", rune(c)) {
		return &Character{Raw: p.source[start:p.pos], Kind: ClassEscape}, true, nil
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
		if !p.done() && (p.peek() >= 'a' && p.peek() <= 'z' || p.peek() >= 'A' && p.peek() <= 'Z') {
			value := p.peek() % 32
			p.pos++
			return &Character{Value: rune(value), Raw: p.source[start:p.pos], Kind: Escape}, true, nil
		}
		if p.flags.Unicode || p.flags.UnicodeSets {
			return nil, false, p.failAt(start, "invalid control escape")
		}
		return &Character{Value: 'c', Raw: p.source[start:p.pos], Kind: Escape}, true, nil
	}
	if c == 'u' {
		ch, err := p.unicodeEscape(start)
		return ch, true, err
	}
	if c == 'x' {
		ch, err := p.hexEscape(start, 2)
		return ch, true, err
	}
	if strings.ContainsRune("fnrtv", rune(c)) {
		return &Character{Raw: p.source[start:p.pos], Kind: Escape}, true, nil
	}
	if (p.flags.Unicode || p.flags.UnicodeSets) && isIdentifierPart(c) {
		return nil, false, p.failAt(start, "invalid identity escape")
	}
	return &Character{Value: rune(c), Raw: p.source[start:p.pos], Kind: Escape}, true, nil
}

func (p *parser) characterClass() (Node, error) {
	p.pos++
	neg := p.take('^')
	union := &ClassUnion{}
	for !p.done() && p.peek() != ']' {
		if p.flags.UnicodeSets && p.peek() == '[' {
			nested, err := p.characterClass()
			if err != nil {
				return nil, err
			}
			union.Operands = append(union.Operands, nested.(*CharacterClass).Expr)
			continue
		}
		if p.flags.UnicodeSets && p.match("\\q{") {
			p.pos += 3
			s := &ClassString{Alternatives: [][]*Character{{}}}
			for !p.done() && p.peek() != '}' {
				if p.take('|') {
					s.Alternatives = append(s.Alternatives, []*Character{})
					continue
				}
				ch, err := p.classCharacter()
				if err != nil {
					return nil, err
				}
				last := len(s.Alternatives) - 1
				s.Alternatives[last] = append(s.Alternatives[last], ch)
			}
			if !p.take('}') {
				return nil, p.fail("unterminated class string")
			}
			union.Operands = append(union.Operands, s)
			continue
		}
		left, err := p.classCharacter()
		if err != nil {
			return nil, err
		}
		var expr ClassExpression = &ClassCharacter{Character: left}
		if p.peek() == '-' && p.pos+1 < len(p.source) && p.source[p.pos+1] != ']' && !(p.flags.UnicodeSets && p.source[p.pos:p.pos+2] == "--") {
			p.pos++
			right, e := p.classCharacter()
			if e != nil {
				return nil, e
			}
			if left.Kind == ClassEscape || left.Kind == PropertyEscape || right.Kind == ClassEscape || right.Kind == PropertyEscape || left.Value > right.Value {
				return nil, p.fail("invalid character class range")
			}
			expr = &ClassRange{From: left, To: right}
		}
		union.Operands = append(union.Operands, expr)
		if p.flags.UnicodeSets && (p.match("&&") || p.match("--")) {
			op := p.source[p.pos : p.pos+2]
			p.pos += 2
			right, e := p.classSetOperand()
			if e != nil {
				return nil, e
			}
			var leftExpr ClassExpression = union
			if op == "&&" {
				return p.finishClass(neg, &ClassIntersection{Left: leftExpr, Right: right})
			}
			return p.finishClass(neg, &ClassSubtraction{Left: leftExpr, Right: right})
		}
	}
	return p.finishClass(neg, union)
}

func (p *parser) finishClass(neg bool, expr ClassExpression) (Node, error) {
	if !p.take(']') {
		return nil, p.fail("unterminated character class")
	}
	return &CharacterClass{Negated: neg, Expr: expr}, nil
}
func (p *parser) classSetOperand() (ClassExpression, error) {
	if p.peek() == '[' {
		n, e := p.characterClass()
		if e != nil {
			return nil, e
		}
		return n.(*CharacterClass).Expr, nil
	}
	c, e := p.classCharacter()
	if e != nil {
		return nil, e
	}
	return &ClassCharacter{Character: c}, nil
}
func (p *parser) classCharacter() (*Character, error) {
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
	if p.flags.UnicodeSets && strings.ContainsRune("(){}|/", rune(p.peek())) {
		return nil, p.fail("reserved character in Unicode set")
	}
	n, _, e := p.literal()
	return n.(*Character), e
}

func (p *parser) literal() (Node, bool, error) {
	start := p.pos
	r, n := utf8.DecodeRuneInString(p.source[p.pos:])
	if r == utf8.RuneError && n == 1 {
		return nil, false, p.fail("invalid UTF-8")
	}
	p.pos += n
	return &Character{Value: r, Raw: p.source[start:p.pos], Kind: Literal}, true, nil
}
func (p *parser) char(start int, kind CharacterKind) (Node, bool, error) {
	return &Character{Raw: p.source[start:p.pos], Kind: kind}, true, nil
}
func (p *parser) unicodeEscape(start int) (*Character, error) {
	if p.take('{') {
		if !(p.flags.Unicode || p.flags.UnicodeSets) {
			return nil, p.failAt(start, "invalid Unicode escape")
		}
		n, ok := p.hexNumber('}')
		if !ok || n > 0x10ffff {
			return nil, p.failAt(start, "invalid Unicode escape")
		}
		return &Character{Value: rune(n), Raw: p.source[start:p.pos], Kind: Escape}, nil
	}
	return p.hexEscape(start, 4)
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
	n := 0
	for i := 0; i < 3 && !p.done() && p.peek() >= '0' && p.peek() <= '7'; i++ {
		n = n*8 + int(p.peek()-'0')
		p.pos++
	}
	return &Character{Value: rune(n), Raw: p.source[start:p.pos], Kind: Escape}, true, nil
}

func (p *parser) groupName() (string, error) {
	start := p.pos
	if p.done() || !isNameStart(p.peek()) {
		return "", p.fail("invalid capture name")
	}
	for !p.done() && p.peek() != '>' {
		c := p.peek()
		if !isNamePart(c) {
			return "", p.fail("invalid capture name")
		}
		p.pos++
	}
	if p.pos == start || !p.take('>') {
		return "", p.fail("invalid capture name")
	}
	return p.source[start : p.pos-1], nil
}
func (p *parser) decimal() (int, bool) {
	start := p.pos
	n := 0
	for !p.done() && p.peek() >= '0' && p.peek() <= '9' {
		if n > 1_000_000_000 {
			return 0, false
		}
		n = n*10 + int(p.peek()-'0')
		p.pos++
	}
	return n, p.pos > start
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
func (p *parser) done() bool { return p.pos >= len(p.source) }
func (p *parser) peek() byte {
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
func isNamePart(c byte) bool {
	return c == '_' || c == '$' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}
func isNameStart(c byte) bool {
	return c == '_' || c == '$' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
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
