package regexp

import "math/bits"

// The regular engine forgets capture history only while locating a candidate.
// Backreferences and lookaround therefore cannot enter it. The ordered VM runs
// once at the earliest successful start to recover ECMAScript's chosen end and
// captures against the full input (not a slice with different assertion context).
const regularLimit = 768

type regularInstruction struct {
	op           opcode
	x, y, source int
}
type regularProgram struct {
	code            []regularInstruction
	start           int
	classes         [128]uint8
	representatives []byte
}

func (program *Program) compileRegular() *regularProgram {
	for _, instruction := range program.code {
		// Unicode boundary assertions can match a pair interior in V8; retain its VM walk.
		if instruction.op == opAssert && instruction.assertion >= WordBoundary && unicodeMode(instruction.flags) {
			return nil
		}
		if instruction.op == opReference || instruction.op == opLook {
			return nil
		}
		if instruction.op == opSet && (instruction.direction != 1 || len(instruction.set.strings) != 0) {
			return nil
		}
		if instruction.op == opRepeat &&
			(!instruction.min.IsUint64() ||
				instruction.min.Uint64() > 32 ||
				instruction.max != nil && (!instruction.max.IsUint64() || instruction.max.Uint64() > 32)) {
			return nil
		}
	}
	regularProgram := &regularProgram{}
	add := func(instruction regularInstruction) int {
		if len(regularProgram.code) >= regularLimit {
			return -1
		}
		regularProgram.code = append(regularProgram.code, instruction)
		return len(regularProgram.code) - 1
	}
	accept := add(regularInstruction{op: opAccept})
	// Empty bounded bodies can expand exponentially without adding states.
	// Bound construction work as well as output size, then retain the VM.
	work := 0
	var segment func(int, int, int) int
	segment = func(lo, hi, next int) int {
		work++
		if work > regularLimit*8 {
			return -1
		}
		mapped := map[int]int{hi: next}
		for pc := hi - 1; pc >= lo; pc-- {
			instruction := program.code[pc]
			switch instruction.op {
			case opRepeatEnd:
				choice := instruction.y
				instruction := program.code[choice]
				if instruction.max == nil {
					loop := add(regularInstruction{op: opSplit, y: next})
					if loop < 0 {
						return -1
					}
					body := segment(choice+1, pc, loop)
					if body < 0 {
						return -1
					}
					regularProgram.code[loop].x = body
					next = loop
				} else {
					for count := instruction.max.Uint64(); count > instruction.min.Uint64(); count-- {
						body := segment(choice+1, pc, next)
						if body < 0 {
							return -1
						}
						next = add(regularInstruction{op: opSplit, x: body, y: next})
						if next < 0 {
							return -1
						}
					}
				}
				for count := uint64(0); count < instruction.min.Uint64(); count++ {
					next = segment(choice+1, pc, next)
					if next < 0 {
						return -1
					}
				}
				pc = choice - 1
			case opSet, opAssert:
				next = add(regularInstruction{op: instruction.op, x: next, source: pc})
			case opSplit:
				a, aok := mapped[instruction.x]
				b, bok := mapped[instruction.y]
				if !aok || !bok {
					return -1
				}
				next = add(regularInstruction{op: opSplit, x: a, y: b})
			case opJump:
				var ok bool
				next, ok = mapped[instruction.x]
				if !ok {
					return -1
				}
			case opSave, opReference:
				// opReference is unreachable unless the eligibility guard is mutated.
			case opAccept:
				next = accept
			default:
				return -1
			}
			if next < 0 {
				return -1
			}
			mapped[pc] = next
		}
		return next
	}
	regularProgram.start = segment(0, len(program.code), accept)
	if regularProgram.start < 0 {
		return nil
	}
	// ASCII alphabet classes preserve every consuming predicate and assertion.
	// Folding is performed at compile time with exactly the VM's Canonicalize.
	keys := map[string]uint8{}
	for c := 0; c < 128; c++ {
		key := make([]byte, (len(regularProgram.code)+7)/8+2)
		for pc, regularInstruction := range regularProgram.code {
			if regularInstruction.op == opSet {
				i := program.code[regularInstruction.source]
				if i.set.contains(canonicalize(rune(c), i.flags)) {
					key[pc/8] |= 1 << uint(pc%8)
				}
			}
		}
		if wordCharacter(rune(c), Flags{}) {
			key[len(key)-2] = 1
		}
		if lineTerminator(rune(c)) {
			key[len(key)-1] = 1
		}
		id, ok := keys[string(key)]
		if !ok {
			id = uint8(len(regularProgram.representatives))
			keys[string(key)] = id
			regularProgram.representatives = append(regularProgram.representatives, byte(c))
		}
		regularProgram.classes[c] = id
	}
	return regularProgram
}

func regularAssertion(instruction instruction, input []uint16, at int) bool {
	switch instruction.assertion {
	case Start:
		return at == 0 || instruction.flags.Multiline && lineTerminator(rune(input[at-1]))
	case End:
		return at == len(input) || instruction.flags.Multiline && lineTerminator(rune(input[at]))
	default:
		left, _, l := readCharacter(input, at, -1, unicodeMode(instruction.flags))
		right, _, r := readCharacter(input, at, 1, unicodeMode(instruction.flags))
		boundary := (l && wordCharacter(left, instruction.flags)) != (r && wordCharacter(right, instruction.flags))
		return boundary != (instruction.assertion == NotWordBoundary)
	}
}

// Merging threads keeps the earliest start at each state. With no references or
// lookaround their future languages are identical. Once a candidate accepts,
// only threads from earlier starts need finish before capture recovery.
func (program *Program) regularFind(input []uint16, start int, sticky bool) int {
	regularProgram := program.regular
	current := make([]int, len(regularProgram.code))
	next := make([]int, len(regularProgram.code))
	for index := range current {
		current[index] = -1
		next[index] = -1
	}
	best := -1
	for at := start; ; {
		var pending [12]uint64
		put := func(pc, origin int) {
			if current[pc] < 0 || origin < current[pc] {
				current[pc] = origin
				pending[pc/64] |= uint64(1) << uint(pc%64)
			}
		}
		if best < 0 && (at == start || !sticky && !program.anchored) {
			put(regularProgram.start, at)
		}
		for pc, origin := range current {
			if origin >= 0 {
				pending[pc/64] |= uint64(1) << uint(pc%64)
			}
		}
		for {
			pc := -1
			for index, word := range pending {
				if word != 0 {
					pc = index*64 + bits.TrailingZeros64(word)
					pending[index] &= pending[index] - 1
					break
				}
			}
			if pc < 0 {
				break
			}
			instruction := regularProgram.code[pc]
			origin := current[pc]
			if best >= 0 && origin >= best {
				continue
			}
			switch instruction.op {
			case opAccept:
				best = origin
			case opSplit:
				put(instruction.x, origin)
				put(instruction.y, origin)
			case opAssert:
				if regularAssertion(program.code[instruction.source], input, at) {
					put(instruction.x, origin)
				}
			}
		}
		if at == len(input) {
			return best
		}
		character, to, readable := readCharacter(input, at, 1, unicodeMode(program.flags))
		if !readable {
			to = at + 1
		}
		alive := false
		for pc, origin := range current {
			if origin < 0 || best >= 0 && origin >= best {
				continue
			}
			regularInstruction := regularProgram.code[pc]
			if regularInstruction.op == opSet && readable {
				i := program.code[regularInstruction.source]
				if i.set.contains(canonicalize(character, i.flags)) &&
					(next[regularInstruction.x] < 0 || origin < next[regularInstruction.x]) {
					next[regularInstruction.x] = origin
					alive = true
				}
			}
		}
		if !alive && (best >= 0 || sticky || program.anchored) {
			return best
		}
		current, next = next, current
		for index := range next {
			next[index] = -1
		}
		at = to
	}
}
