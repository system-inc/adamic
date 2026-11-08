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
	firstAscii  [2]uint64
	filterFirst bool
	prefix      []rune
	anchored    bool
	regular     *regularProgram
	v8Refusal   error
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
	tree, parsePatternError := parsePattern(pattern, flags, true)
	if parsePatternError != nil {
		return nil, parsePatternError
	}
	return compilePattern(tree, properties)
}

// CompileUtf16 accepts an exact JavaScript pattern string, including surrogates.
func CompileUtf16(pattern []uint16, flags string) (*Program, error) {
	return CompileUtf16WithProperties(pattern, flags, UnicodeProperties{})
}

// CompileUtf16WithProperties combines exact pattern strings and the UCD seam.
func CompileUtf16WithProperties(pattern []uint16, flags string, properties PropertyProvider) (*Program, error) {
	tree, parsePatternError := parsePattern(patternFromUtf16(pattern), flags, true)
	if parsePatternError != nil {
		return nil, parsePatternError
	}
	return compilePattern(tree, properties)
}
func compilePattern(tree *Pattern, properties PropertyProvider) (*Program, error) {
	program := &Program{flags: tree.Flags, names: map[string][]int{}}
	groups := map[*Group]int{}
	var number func(Node)
	number = func(node Node) {
		switch typed := node.(type) {
		case *Disjunction:
			for _, alternative := range typed.Alternatives {
				for _, term := range alternative.Terms {
					number(term)
				}
			}
		case *Group:
			if typed.Kind == Capturing {
				program.captures++
				groups[typed] = program.captures
				if typed.Name != "" {
					program.names[typed.Name] = append(program.names[typed.Name], program.captures)
				}
			}
			number(typed.Body)
		case *Quantifier:
			number(typed.Atom)
		}
	}
	number(tree.Body)
	compiler := compiler{p: program, groups: groups, properties: properties}
	if disjunctionError := compiler.disjunction(tree.Body, tree.Flags, 1); disjunctionError != nil {
		return nil, disjunctionError
	}
	program.code = append(program.code, instruction{op: opAccept})
	program.firstAscii, program.filterFirst = program.nativeFirstAscii()
	program.prefix = program.nativePrefix()
	program.anchored = program.nativeAnchored()
	program.regular = program.compileRegular()
	program.v8Refusal = v8Divergence(tree, properties)
	return program, nil
}

type compiler struct {
	p          *Program
	groups     map[*Group]int
	properties PropertyProvider
}

func (regexpCompiler *compiler) emit(instruction instruction) int {
	at := len(regexpCompiler.p.code)
	regexpCompiler.p.code = append(regexpCompiler.p.code, instruction)
	return at
}
func (regexpCompiler *compiler) disjunction(disjunction *Disjunction, flags Flags, dir int) error {
	var jumps []int
	for index, alternative := range disjunction.Alternatives {
		split := -1
		if index+1 < len(disjunction.Alternatives) {
			split = regexpCompiler.emit(instruction{op: opSplit, x: len(regexpCompiler.p.code) + 1})
		}
		terms := slices.Clone(alternative.Terms)
		// The parser preserves raw astral literals. In legacy mode they are two
		// atoms; a suffix quantifier applies only to the trailing code unit.
		var normalized []Node
		for _, term := range terms {
			if ch, ok := term.(*Character); ok && ch.Value > 0xffff && !unicodeMode(flags) {
				h, l := utf16.EncodeRune(ch.Value)
				normalized = append(normalized, &Character{Value: h}, &Character{Value: l})
				continue
			}
			if quantifier, ok := term.(*Quantifier); ok && !unicodeMode(flags) {
				if ch, ok := quantifier.Atom.(*Character); ok && ch.Value > 0xffff {
					h, l := utf16.EncodeRune(ch.Value)
					copyQ := *quantifier
					copyQ.Atom = &Character{Value: l}
					normalized = append(normalized, &Character{Value: h}, &copyQ)
					continue
				}
			}
			normalized = append(normalized, term)
		}
		terms = normalized
		if dir < 0 {
			slices.Reverse(terms)
		}
		for _, term := range terms {
			if nodeError := regexpCompiler.node(term, flags, dir); nodeError != nil {
				return nodeError
			}
		}
		if split >= 0 {
			jumps = append(jumps, regexpCompiler.emit(instruction{op: opJump}))
			regexpCompiler.p.code[split].y = len(regexpCompiler.p.code)
		}
	}
	for _, jump := range jumps {
		regexpCompiler.p.code[jump].x = len(regexpCompiler.p.code)
	}
	return nil
}
func (regexpCompiler *compiler) node(node Node, flags Flags, dir int) error {
	base := instruction{direction: dir, flags: flags}
	switch typed := node.(type) {
	case *Disjunction:
		return regexpCompiler.disjunction(typed, flags, dir)
	case *Character:
		set, compileCharacterError := compileCharacter(typed, flags, regexpCompiler.properties)
		if compileCharacterError != nil {
			return compileCharacterError
		}
		base.op = opSet
		base.set = set
		regexpCompiler.emit(base)
	case *CharacterClass:
		set, compileClassError := compileClass(typed.Expr, flags, regexpCompiler.properties)
		if compileClassError != nil {
			return compileClassError
		}
		if typed.Negated {
			set = negateSet(set, flags)
		}
		base.op = opSet
		base.set = set
		regexpCompiler.emit(base)
	case *Dot:
		base.op = opSet
		base.set = characterSet{ranges: []RuneRange{{0, maxCharacter(flags)}}}
		if !flags.DotAll {
			base.set.ranges = subtractRanges(base.set.ranges, []RuneRange{{10, 10}, {13, 13}, {0x2028, 0x2029}})
		}
		regexpCompiler.emit(base)
	case *Assertion:
		base.op = opAssert
		base.assertion = typed.Kind
		regexpCompiler.emit(base)
	case *Backreference:
		base.op = opReference
		if typed.Name != "" {
			base.references = regexpCompiler.p.names[typed.Name]
		} else {
			base.references = []int{typed.Index}
		}
		regexpCompiler.emit(base)
	case *Group:
		flags.IgnoreCase = (flags.IgnoreCase || typed.Enable.IgnoreCase) && !typed.Disable.IgnoreCase
		flags.Multiline = (flags.Multiline || typed.Enable.Multiline) && !typed.Disable.Multiline
		flags.DotAll = (flags.DotAll || typed.Enable.DotAll) && !typed.Disable.DotAll
		if typed.Kind >= PositiveLookahead {
			direction := 1
			if typed.Kind == PositiveLookbehind || typed.Kind == NegativeLookbehind {
				direction = -1
			}
			sub := &Program{flags: flags, captures: regexpCompiler.p.captures, names: regexpCompiler.p.names}
			child := compiler{p: sub, groups: regexpCompiler.groups, properties: regexpCompiler.properties}
			if disjunctionError := child.disjunction(typed.Body, flags, direction); disjunctionError != nil {
				return disjunctionError
			}
			child.emit(instruction{op: opAccept})
			base.op = opLook
			base.look = sub
			base.negative = typed.Kind == NegativeLookahead || typed.Kind == NegativeLookbehind
			regexpCompiler.emit(base)
		} else {
			id := regexpCompiler.groups[typed]
			if id != 0 {
				regexpCompiler.emit(instruction{op: opSave, x: 2*id + boolInt(dir < 0)})
			}
			if disjunctionError := regexpCompiler.disjunction(typed.Body, flags, dir); disjunctionError != nil {
				return disjunctionError
			}
			if id != 0 {
				regexpCompiler.emit(instruction{op: opSave, x: 2*id + boolInt(dir > 0)})
			}
		}
	case *Quantifier:
		id := regexpCompiler.p.repeats
		regexpCompiler.p.repeats++
		regexpCompiler.emit(instruction{op: opRepeatInit, x: id})
		choice := regexpCompiler.emit(
			instruction{op: opRepeat, x: id, min: typed.Min, max: typed.Max, greedy: typed.Greedy},
		)
		var collect func(Node)
		collect = func(node Node) {
			switch typed := node.(type) {
			case *Group:
				if group := regexpCompiler.groups[typed]; group != 0 {
					regexpCompiler.p.code[choice].clear = append(regexpCompiler.p.code[choice].clear, group)
				}
				collect(typed.Body)
			case *Quantifier:
				collect(typed.Atom)
			case *Disjunction:
				for _, alternative := range typed.Alternatives {
					for _, term := range alternative.Terms {
						collect(term)
					}
				}
			}
		}
		collect(typed.Atom)
		if nodeError := regexpCompiler.node(typed.Atom, flags, dir); nodeError != nil {
			return nodeError
		}
		regexpCompiler.emit(instruction{op: opRepeatEnd, x: id, y: choice, min: typed.Min})
		regexpCompiler.p.code[choice].y = len(regexpCompiler.p.code)
	default:
		return fmt.Errorf("regexp: unsupported syntax %T", typed)
	}
	return nil
}
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
func unicodeMode(flags Flags) bool { return flags.Unicode || flags.UnicodeSets }
func maxCharacter(flags Flags) rune {
	if unicodeMode(flags) {
		return 0x10ffff
	}
	return 0xffff
}

// RegExp owns JavaScript's mutable lastIndex. Programs can be shared by instances.
// LastIndex is already converted by ToLength by the embedding runtime.
type RegExp struct {
	Program        *Program
	LastIndex      uint64
	StepLimit      uint64
	DisableRegular bool
}

func (program *Program) New() *RegExp { return &RegExp{Program: program} }

// Exec accepts exact UTF-16, including lone surrogates. A zero StepLimit is unlimited.
func (exp *RegExp) Exec(input []uint16) (*Match, error) {
	program := exp.Program
	start := uint64(0)
	stateful := program.flags.Global || program.flags.Sticky
	if stateful {
		start = exp.LastIndex
	}
	budget := executionBudget{limit: exp.StepLimit}
	// V8 first rewinds a lastIndex inside a pair, but its assertion-only
	// paths can subsequently try the low-half position. ECMA-262 22.2.2.2
	// and 22.2.7.2 describe a code-point Input list with no such position.
	if unicodeMode(program.flags) &&
		start > 0 &&
		start < uint64(len(input)) &&
		high(input[start-1]) &&
		low(input[start]) {
		start--
	}
	stickyEnd := start
	if stateful &&
		exp.LastIndex == start+1 &&
		start+1 < uint64(len(input)) &&
		high(input[start]) &&
		low(input[start+1]) {
		stickyEnd++
	}
	for start <= uint64(len(input)) {
		position := int(start)
		if program.anchored && position != 0 {
			break
		}
		if program.regular != nil && exp.StepLimit == 0 && !exp.DisableRegular {
			position = program.regularFind(input, position, program.flags.Sticky)
			if position < 0 {
				break
			}
		}
		if !program.anchored && !program.flags.Sticky {
			for position < len(input) {
				c := input[position]
				possible := !program.filterFirst || c >= 128 || program.firstAscii[c/64]&(uint64(1)<<uint(c%64)) != 0
				prefix := len(program.prefix) == 0
				if len(program.prefix) > 0 && position+len(program.prefix) <= len(input) {
					prefix = true
					for index, character := range program.prefix {
						if input[position+index] != uint16(character) {
							prefix = false
							break
						}
					}
				}
				if possible && prefix {
					break
				}
				_, next, readable := readCharacter(input, position, 1, unicodeMode(program.flags))
				if !readable {
					next = position + 1
				}
				position = next
			}
		}
		caps := make([]int, 2*(program.captures+1))
		for index := range caps {
			caps[index] = -1
		}
		caps[0] = position
		state, ok, runError := program.run(input, position, caps, &budget)
		if runError != nil {
			return nil, runError
		}
		if ok {
			state.caps[1] = state.pos
			match := &Match{Captures: make([]Capture, program.captures+1), Groups: map[string]Capture{}}
			for index := range match.Captures {
				match.Captures[index] = Capture{state.caps[2*index], state.caps[2*index+1]}
			}
			for name, ids := range program.names {
				capture := Capture{-1, -1}
				for _, id := range ids {
					if match.Captures[id].Start >= 0 {
						capture = match.Captures[id]
					}
				}
				match.Groups[name] = capture
			}
			if stateful {
				exp.LastIndex = uint64(state.pos)
			}
			return match, nil
		}
		if program.anchored || program.flags.Sticky && start >= stickyEnd {
			break
		}
		// Irregexp searches UTF-16 positions even under u/v. A consuming
		// matcher rejects pair interiors; \B and negative lookaround can
		// accept there. Match Node rather than AdvanceStringIndex's spec walk.
		start = uint64(position + 1)
	}
	if stateful {
		exp.LastIndex = 0
	}
	return nil, nil
}

// ExecString is a convenience for well-formed UTF-8; use Exec for lone surrogates.
func (exp *RegExp) ExecString(input string) (*Match, error) {
	return exp.Exec(utf16.Encode([]rune(input)))
}

type repeatRegister struct {
	count *big.Int
	start int
}
type machineState struct {
	pc, pos int
	caps    []int
	repeats []repeatRegister
}

func (state machineState) clone() machineState {
	state.caps = slices.Clone(state.caps)
	state.repeats = slices.Clone(state.repeats)
	return state
}

type executionBudget struct{ limit, steps uint64 }

func (budget *executionBudget) step() error {
	if budget.limit != 0 && budget.steps >= budget.limit {
		return ErrStepLimit
	}
	budget.steps++
	return nil
}
func (program *Program) run(input []uint16, pos int, caps []int, budget *executionBudget) (machineState, bool, error) {
	state := machineState{pos: pos, caps: slices.Clone(caps), repeats: make([]repeatRegister, program.repeats)}
	var stack []machineState
	one := big.NewInt(1)
	for {
		if stepError := budget.step(); stepError != nil {
			return state, false, stepError
		}
		instruction := program.code[state.pc]
		failed := false
		switch instruction.op {
		case opAccept:
			return state, true, nil
		case opJump:
			state.pc = instruction.x
			continue
		case opSplit:
			alternative := state.clone()
			alternative.pc = instruction.y
			stack = append(stack, alternative)
			state.pc = instruction.x
			continue
		case opSave:
			state.caps[instruction.x] = state.pos
		case opSet:
			if len(instruction.set.strings) == 0 {
				character, next, exists := readCharacter(
					input,
					state.pos,
					instruction.direction,
					unicodeMode(instruction.flags),
				)
				if exists && instruction.set.contains(canonicalize(character, instruction.flags)) {
					state.pos = next
				} else {
					failed = true
				}
				break
			}
			positions := instruction.set.match(input, state.pos, instruction.direction, instruction.flags)
			if len(positions) == 0 {
				failed = true
			} else {
				for index := len(positions) - 1; index > 0; index-- {
					alt := state.clone()
					alt.pc++
					alt.pos = positions[index]
					stack = append(stack, alt)
				}
				state.pos = positions[0]
			}
		case opAssert:
			before := state.pos > 0 && lineTerminator(rune(input[state.pos-1]))
			after := state.pos < len(input) && lineTerminator(rune(input[state.pos]))
			switch instruction.assertion {
			case Start:
				failed = !(state.pos == 0 || instruction.flags.Multiline && before)
			case End:
				failed = !(state.pos == len(input) || instruction.flags.Multiline && after)
			default:
				left, _, l := readCharacter(input, state.pos, -1, unicodeMode(instruction.flags))
				right, _, r := readCharacter(input, state.pos, 1, unicodeMode(instruction.flags))
				boundary := (l &&
					wordCharacter(left, instruction.flags)) != (r && wordCharacter(right, instruction.flags))
				failed = boundary == (instruction.assertion == NotWordBoundary)
			}
		case opLook:
			sub, ok, runError := instruction.look.run(input, state.pos, state.caps, budget)
			if runError != nil {
				return state, false, runError
			}
			failed = ok == instruction.negative
			if ok && !instruction.negative {
				state.caps = sub.caps
			}
		case opReference:
			capture := Capture{-1, -1}
			for _, id := range instruction.references {
				if state.caps[2*id] >= 0 && state.caps[2*id+1] >= 0 {
					capture = Capture{state.caps[2*id], state.caps[2*id+1]}
				}
			}
			if capture.Start >= 0 {
				at, end := capture.Start, capture.End
				if instruction.direction < 0 {
					at, end = end, at
				}
				for at != end {
					a, next, ok := readCharacter(input, at, instruction.direction, unicodeMode(instruction.flags))
					b, target, exists := readCharacter(
						input,
						state.pos,
						instruction.direction,
						unicodeMode(instruction.flags),
					)
					if !ok || !exists || canonicalize(a, instruction.flags) != canonicalize(b, instruction.flags) {
						failed = true
						break
					}
					at = next
					state.pos = target
				}
			}
		case opRepeatInit:
			state.repeats[instruction.x] = repeatRegister{count: new(big.Int)}
		case opRepeat:
			reg := state.repeats[instruction.x]
			canExit := reg.count.Cmp(instruction.min) >= 0
			canBody := instruction.max == nil || reg.count.Cmp(instruction.max) < 0
			if !canBody && !canExit {
				failed = true
				break
			}
			enter := func(state *machineState) {
				state.pc++
				state.repeats[instruction.x].start = state.pos
				for _, id := range instruction.clear {
					state.caps[2*id] = -1
					state.caps[2*id+1] = -1
				}
			}
			if canBody && canExit {
				alternative := state.clone()
				if instruction.greedy {
					alternative.pc = instruction.y
				} else {
					enter(&alternative)
				}
				stack = append(stack, alternative)
			}
			if canBody && (!canExit || instruction.greedy) {
				enter(&state)
			} else {
				state.pc = instruction.y
			}
			continue
		case opRepeatEnd:
			reg := state.repeats[instruction.x]
			if reg.count.Cmp(instruction.min) >= 0 && reg.start == state.pos {
				failed = true
			} else {
				state.repeats[instruction.x].count = new(big.Int).Add(reg.count, one)
				state.pc = instruction.y
				continue
			}
		}
		if failed {
			if len(stack) == 0 {
				return state, false, nil
			}
			state = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
		} else {
			state.pc++
		}
	}
}
func high(c uint16) bool { return c >= 0xd800 && c <= 0xdbff }
func low(c uint16) bool  { return c >= 0xdc00 && c <= 0xdfff }
func readCharacter(input []uint16, pos, dir int, unicode bool) (rune, int, bool) {
	if unicode && pos > 0 && pos < len(input) && high(input[pos-1]) && low(input[pos]) {
		return 0, pos, false
	}
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
func lineTerminator(character rune) bool {
	return character == 10 || character == 13 || character == 0x2028 || character == 0x2029
}
