package regexp

import (
	"errors"
	"fmt"
	"math/big"
	"slices"
	"unicode/utf16"
)

// ErrStepLimit is an interrupted computation, never a failed match.
var ErrStepLimit = errors.New("regexp: instruction step limit exceeded")

// Capture holds UTF-16 offsets. {-1,-1} means undefined, distinct from empty.
type Capture struct{ Start, End int }

// Match contains capture zero followed by the parenthesis captures.
type Match struct {
	Captures []Capture
	Groups   map[string]Capture
}

// Program is immutable bytecode. No syntax tree is retained by the executor.
// Integer targets, range tables, registers and assertion subprograms can be
// serialized for a C executor without Go callbacks or interface dispatch.
type Program struct {
	code        []instruction
	flags       Flags
	captures    int
	names       map[string][]int
	repeats     int
	firstASCII  [2]uint64
	filterFirst bool
	prefix      []rune
	anchored    bool
}

type opcode uint8

const (
	opAccept opcode = iota
	opSet
	opSplit
	opJump
	opSave
	opAssert
	opLook
	opReference
	opRepeatInit
	opRepeat
	opRepeatEnd
)

type instruction struct {
	op         opcode
	x, y       int
	direction  int
	flags      Flags
	set        characterSet
	assertion  AssertionKind
	look       *Program
	negative   bool
	references []int
	min, max   *big.Int
	greedy     bool
	clear      []int
}

// Compile parses and compiles with the complete Unicode Character Database provider.
func Compile(pattern, flags string) (*Program, error) {
	return CompileWithProperties(pattern, flags, UnicodeProperties{})
}

// CompileWithProperties supplies the Unicode data seam used by UCD tables.
func CompileWithProperties(pattern, flags string, properties PropertyProvider) (*Program, error) {
	tree, err := Parse(pattern, flags)
	if err != nil {
		return nil, err
	}
	return compilePattern(tree, properties)
}

// CompileUTF16 accepts an exact JavaScript pattern string, including surrogates.
func CompileUTF16(pattern []uint16, flags string) (*Program, error) {
	return CompileUTF16WithProperties(pattern, flags, UnicodeProperties{})
}

// CompileUTF16WithProperties combines exact pattern strings and the UCD seam.
func CompileUTF16WithProperties(pattern []uint16, flags string, properties PropertyProvider) (*Program, error) {
	tree, err := ParseUTF16(pattern, flags)
	if err != nil {
		return nil, err
	}
	return compilePattern(tree, properties)
}
func compilePattern(tree *Pattern, properties PropertyProvider) (*Program, error) {
	p := &Program{flags: tree.Flags, names: map[string][]int{}}
	groups := map[*Group]int{}
	var number func(Node)
	number = func(n Node) {
		switch n := n.(type) {
		case *Disjunction:
			for _, a := range n.Alternatives {
				for _, n := range a.Terms {
					number(n)
				}
			}
		case *Group:
			if n.Kind == Capturing {
				p.captures++
				groups[n] = p.captures
				if n.Name != "" {
					p.names[n.Name] = append(p.names[n.Name], p.captures)
				}
			}
			number(n.Body)
		case *Quantifier:
			number(n.Atom)
		}
	}
	number(tree.Body)
	c := compiler{p: p, groups: groups, properties: properties}
	if err := c.disjunction(tree.Body, tree.Flags, 1); err != nil {
		return nil, err
	}
	p.code = append(p.code, instruction{op: opAccept})
	p.firstASCII, p.filterFirst = p.nativeFirstASCII()
	p.prefix = p.nativePrefix()
	p.anchored = p.nativeAnchored()
	return p, nil
}

type compiler struct {
	p          *Program
	groups     map[*Group]int
	properties PropertyProvider
}

func (c *compiler) emit(i instruction) int {
	at := len(c.p.code)
	c.p.code = append(c.p.code, i)
	return at
}
func (c *compiler) disjunction(d *Disjunction, f Flags, dir int) error {
	var jumps []int
	for i, a := range d.Alternatives {
		split := -1
		if i+1 < len(d.Alternatives) {
			split = c.emit(instruction{op: opSplit, x: len(c.p.code) + 1})
		}
		terms := slices.Clone(a.Terms)
		// The parser preserves raw astral literals. In legacy mode they are two
		// atoms; a suffix quantifier applies only to the trailing code unit.
		var normalized []Node
		for _, n := range terms {
			if ch, ok := n.(*Character); ok && ch.Value > 0xffff && !unicodeMode(f) {
				h, l := utf16.EncodeRune(ch.Value)
				normalized = append(normalized, &Character{Value: h}, &Character{Value: l})
				continue
			}
			if q, ok := n.(*Quantifier); ok && !unicodeMode(f) {
				if ch, ok := q.Atom.(*Character); ok && ch.Value > 0xffff {
					h, l := utf16.EncodeRune(ch.Value)
					copyQ := *q
					copyQ.Atom = &Character{Value: l}
					normalized = append(normalized, &Character{Value: h}, &copyQ)
					continue
				}
			}
			normalized = append(normalized, n)
		}
		terms = normalized
		if dir < 0 {
			slices.Reverse(terms)
		}
		for _, n := range terms {
			if err := c.node(n, f, dir); err != nil {
				return err
			}
		}
		if split >= 0 {
			jumps = append(jumps, c.emit(instruction{op: opJump}))
			c.p.code[split].y = len(c.p.code)
		}
	}
	for _, j := range jumps {
		c.p.code[j].x = len(c.p.code)
	}
	return nil
}
func (c *compiler) node(n Node, f Flags, dir int) error {
	base := instruction{direction: dir, flags: f}
	switch n := n.(type) {
	case *Disjunction:
		return c.disjunction(n, f, dir)
	case *Character:
		set, err := compileCharacter(n, f, c.properties)
		if err != nil {
			return err
		}
		base.op = opSet
		base.set = set
		c.emit(base)
	case *CharacterClass:
		set, err := compileClass(n.Expr, f, c.properties)
		if err != nil {
			return err
		}
		if n.Negated {
			set = negateSet(set, f)
		}
		base.op = opSet
		base.set = set
		c.emit(base)
	case *Dot:
		base.op = opSet
		base.set = characterSet{ranges: []RuneRange{{0, maxCharacter(f)}}}
		if !f.DotAll {
			base.set.ranges = subtractRanges(base.set.ranges, []RuneRange{{10, 10}, {13, 13}, {0x2028, 0x2029}})
		}
		c.emit(base)
	case *Assertion:
		base.op = opAssert
		base.assertion = n.Kind
		c.emit(base)
	case *Backreference:
		base.op = opReference
		if n.Name != "" {
			base.references = c.p.names[n.Name]
		} else {
			base.references = []int{n.Index}
		}
		c.emit(base)
	case *Group:
		f.IgnoreCase = (f.IgnoreCase || n.Enable.IgnoreCase) && !n.Disable.IgnoreCase
		f.Multiline = (f.Multiline || n.Enable.Multiline) && !n.Disable.Multiline
		f.DotAll = (f.DotAll || n.Enable.DotAll) && !n.Disable.DotAll
		if n.Kind >= PositiveLookahead {
			direction := 1
			if n.Kind == PositiveLookbehind || n.Kind == NegativeLookbehind {
				direction = -1
			}
			sub := &Program{flags: f, captures: c.p.captures, names: c.p.names}
			child := compiler{p: sub, groups: c.groups, properties: c.properties}
			if err := child.disjunction(n.Body, f, direction); err != nil {
				return err
			}
			child.emit(instruction{op: opAccept})
			base.op = opLook
			base.look = sub
			base.negative = n.Kind == NegativeLookahead || n.Kind == NegativeLookbehind
			c.emit(base)
		} else {
			id := c.groups[n]
			if id != 0 {
				c.emit(instruction{op: opSave, x: 2*id + boolInt(dir < 0)})
			}
			if err := c.disjunction(n.Body, f, dir); err != nil {
				return err
			}
			if id != 0 {
				c.emit(instruction{op: opSave, x: 2*id + boolInt(dir > 0)})
			}
		}
	case *Quantifier:
		id := c.p.repeats
		c.p.repeats++
		c.emit(instruction{op: opRepeatInit, x: id})
		choice := c.emit(instruction{op: opRepeat, x: id, min: n.Min, max: n.Max, greedy: n.Greedy})
		var collect func(Node)
		collect = func(n Node) {
			switch n := n.(type) {
			case *Group:
				if g := c.groups[n]; g != 0 {
					c.p.code[choice].clear = append(c.p.code[choice].clear, g)
				}
				collect(n.Body)
			case *Quantifier:
				collect(n.Atom)
			case *Disjunction:
				for _, a := range n.Alternatives {
					for _, t := range a.Terms {
						collect(t)
					}
				}
			}
		}
		collect(n.Atom)
		if err := c.node(n.Atom, f, dir); err != nil {
			return err
		}
		c.emit(instruction{op: opRepeatEnd, x: id, y: choice, min: n.Min})
		c.p.code[choice].y = len(c.p.code)
	default:
		return fmt.Errorf("regexp: unsupported syntax %T", n)
	}
	return nil
}
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
func unicodeMode(f Flags) bool { return f.Unicode || f.UnicodeSets }
func maxCharacter(f Flags) rune {
	if unicodeMode(f) {
		return 0x10ffff
	}
	return 0xffff
}

// RegExp owns JavaScript's mutable lastIndex. Programs can be shared by instances.
// LastIndex is already converted by ToLength by the embedding runtime.
type RegExp struct {
	Program   *Program
	LastIndex uint64
	StepLimit uint64
}

func (p *Program) New() *RegExp { return &RegExp{Program: p} }

// Exec accepts exact UTF-16, including lone surrogates. A zero StepLimit is unlimited.
func (r *RegExp) Exec(input []uint16) (*Match, error) {
	p := r.Program
	start := uint64(0)
	stateful := p.flags.Global || p.flags.Sticky
	if stateful {
		start = r.LastIndex
	}
	budget := executionBudget{limit: r.StepLimit}
	for start <= uint64(len(input)) {
		position := int(start)
		if unicodeMode(p.flags) && position > 0 && position < len(input) && high(input[position-1]) && low(input[position]) {
			position--
		}
		if p.anchored && position != 0 {
			break
		}
		if !p.anchored && !p.flags.Sticky {
			for position < len(input) {
				c := input[position]
				possible := !p.filterFirst || c >= 128 || p.firstASCII[c/64]&(uint64(1)<<uint(c%64)) != 0
				prefix := len(p.prefix) == 0
				if len(p.prefix) > 0 && position+len(p.prefix) <= len(input) {
					prefix = true
					for j, c := range p.prefix {
						if input[position+j] != uint16(c) {
							prefix = false
							break
						}
					}
				}
				if possible && prefix {
					break
				}
				_, position, _ = readCharacter(input, position, 1, unicodeMode(p.flags))
			}
		}
		caps := make([]int, 2*(p.captures+1))
		for i := range caps {
			caps[i] = -1
		}
		caps[0] = position
		state, ok, err := p.run(input, position, caps, &budget)
		if err != nil {
			return nil, err
		}
		if ok {
			state.caps[1] = state.pos
			match := &Match{Captures: make([]Capture, p.captures+1), Groups: map[string]Capture{}}
			for i := range match.Captures {
				match.Captures[i] = Capture{state.caps[2*i], state.caps[2*i+1]}
			}
			for name, ids := range p.names {
				capture := Capture{-1, -1}
				for _, id := range ids {
					if match.Captures[id].Start >= 0 {
						capture = match.Captures[id]
					}
				}
				match.Groups[name] = capture
			}
			if stateful {
				r.LastIndex = uint64(state.pos)
			}
			return match, nil
		}
		if p.anchored || p.flags.Sticky {
			break
		}
		_, next, ok := readCharacter(input, position, 1, unicodeMode(p.flags))
		if !ok {
			break
		}
		start = uint64(next)
	}
	if stateful {
		r.LastIndex = 0
	}
	return nil, nil
}

// ExecString is a convenience for well-formed UTF-8; use Exec for lone surrogates.
func (r *RegExp) ExecString(input string) (*Match, error) { return r.Exec(utf16.Encode([]rune(input))) }

type repeatRegister struct {
	count *big.Int
	start int
}
type machineState struct {
	pc, pos int
	caps    []int
	repeats []repeatRegister
}

func (s machineState) clone() machineState {
	s.caps = slices.Clone(s.caps)
	s.repeats = slices.Clone(s.repeats)
	return s
}

type executionBudget struct{ limit, steps uint64 }

func (b *executionBudget) step() error {
	if b.limit != 0 && b.steps >= b.limit {
		return ErrStepLimit
	}
	b.steps++
	return nil
}
func (p *Program) run(input []uint16, pos int, caps []int, budget *executionBudget) (machineState, bool, error) {
	s := machineState{pos: pos, caps: slices.Clone(caps), repeats: make([]repeatRegister, p.repeats)}
	var stack []machineState
	one := big.NewInt(1)
	for {
		if err := budget.step(); err != nil {
			return s, false, err
		}
		i := p.code[s.pc]
		failed := false
		switch i.op {
		case opAccept:
			return s, true, nil
		case opJump:
			s.pc = i.x
			continue
		case opSplit:
			alternative := s.clone()
			alternative.pc = i.y
			stack = append(stack, alternative)
			s.pc = i.x
			continue
		case opSave:
			s.caps[i.x] = s.pos
		case opSet:
			if len(i.set.strings) == 0 {
				c, next, exists := readCharacter(input, s.pos, i.direction, unicodeMode(i.flags))
				if exists && i.set.contains(canonicalize(c, i.flags)) {
					s.pos = next
				} else {
					failed = true
				}
				break
			}
			positions := i.set.match(input, s.pos, i.direction, i.flags)
			if len(positions) == 0 {
				failed = true
			} else {
				for k := len(positions) - 1; k > 0; k-- {
					alt := s.clone()
					alt.pc++
					alt.pos = positions[k]
					stack = append(stack, alt)
				}
				s.pos = positions[0]
			}
		case opAssert:
			before := s.pos > 0 && lineTerminator(rune(input[s.pos-1]))
			after := s.pos < len(input) && lineTerminator(rune(input[s.pos]))
			switch i.assertion {
			case Start:
				failed = !(s.pos == 0 || i.flags.Multiline && before)
			case End:
				failed = !(s.pos == len(input) || i.flags.Multiline && after)
			default:
				left, _, l := readCharacter(input, s.pos, -1, unicodeMode(i.flags))
				right, _, r := readCharacter(input, s.pos, 1, unicodeMode(i.flags))
				boundary := (l && wordCharacter(left, i.flags)) != (r && wordCharacter(right, i.flags))
				failed = boundary == (i.assertion == NotWordBoundary)
			}
		case opLook:
			sub, ok, err := i.look.run(input, s.pos, s.caps, budget)
			if err != nil {
				return s, false, err
			}
			failed = ok == i.negative
			if ok && !i.negative {
				s.caps = sub.caps
			}
		case opReference:
			capture := Capture{-1, -1}
			for _, id := range i.references {
				if s.caps[2*id] >= 0 && s.caps[2*id+1] >= 0 {
					capture = Capture{s.caps[2*id], s.caps[2*id+1]}
				}
			}
			if capture.Start >= 0 {
				at, end := capture.Start, capture.End
				if i.direction < 0 {
					at, end = end, at
				}
				for at != end {
					a, next, ok := readCharacter(input, at, i.direction, unicodeMode(i.flags))
					b, target, exists := readCharacter(input, s.pos, i.direction, unicodeMode(i.flags))
					if !ok || !exists || canonicalize(a, i.flags) != canonicalize(b, i.flags) {
						failed = true
						break
					}
					at = next
					s.pos = target
				}
			}
		case opRepeatInit:
			s.repeats[i.x] = repeatRegister{count: new(big.Int)}
		case opRepeat:
			reg := s.repeats[i.x]
			canExit := reg.count.Cmp(i.min) >= 0
			canBody := i.max == nil || reg.count.Cmp(i.max) < 0
			if !canBody && !canExit {
				failed = true
				break
			}
			enter := func(state *machineState) {
				state.pc++
				state.repeats[i.x].start = state.pos
				for _, id := range i.clear {
					state.caps[2*id] = -1
					state.caps[2*id+1] = -1
				}
			}
			if canBody && canExit {
				alternative := s.clone()
				if i.greedy {
					alternative.pc = i.y
				} else {
					enter(&alternative)
				}
				stack = append(stack, alternative)
			}
			if canBody && (!canExit || i.greedy) {
				enter(&s)
			} else {
				s.pc = i.y
			}
			continue
		case opRepeatEnd:
			reg := s.repeats[i.x]
			if reg.count.Cmp(i.min) >= 0 && reg.start == s.pos {
				failed = true
			} else {
				s.repeats[i.x].count = new(big.Int).Add(reg.count, one)
				s.pc = i.y
				continue
			}
		}
		if failed {
			if len(stack) == 0 {
				return s, false, nil
			}
			s = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
		} else {
			s.pc++
		}
	}
}
func high(c uint16) bool { return c >= 0xd800 && c <= 0xdbff }
func low(c uint16) bool  { return c >= 0xdc00 && c <= 0xdfff }
func readCharacter(input []uint16, pos, dir int, unicode bool) (rune, int, bool) {
	if dir > 0 {
		if pos >= len(input) {
			return 0, pos, false
		}
		c := input[pos]
		if unicode && high(c) && pos+1 < len(input) && low(input[pos+1]) {
			return utf16.DecodeRune(rune(c), rune(input[pos+1])), pos + 2, true
		}
		return rune(c), pos + 1, true
	}
	if pos <= 0 {
		return 0, pos, false
	}
	c := input[pos-1]
	if unicode && low(c) && pos > 1 && high(input[pos-2]) {
		return utf16.DecodeRune(rune(input[pos-2]), rune(c)), pos - 2, true
	}
	return rune(c), pos - 1, true
}
func lineTerminator(c rune) bool { return c == 10 || c == 13 || c == 0x2028 || c == 0x2029 }
