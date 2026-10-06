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

func (p *Program) compileRegular() *regularProgram {
	for _, i := range p.code {
		if i.op == opReference || i.op == opLook {
			return nil
		}
		if i.op == opSet && (i.direction != 1 || len(i.set.strings) != 0) {
			return nil
		}
		if i.op == opRepeat && (!i.min.IsUint64() || i.min.Uint64() > 32 || i.max != nil && (!i.max.IsUint64() || i.max.Uint64() > 32)) {
			return nil
		}
	}
	r := &regularProgram{}
	add := func(i regularInstruction) int {
		if len(r.code) >= regularLimit {
			return -1
		}
		r.code = append(r.code, i)
		return len(r.code) - 1
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
			i := p.code[pc]
			switch i.op {
			case opRepeatEnd:
				choice := i.y
				q := p.code[choice]
				if q.max == nil {
					loop := add(regularInstruction{op: opSplit, y: next})
					if loop < 0 {
						return -1
					}
					body := segment(choice+1, pc, loop)
					if body < 0 {
						return -1
					}
					r.code[loop].x = body
					next = loop
				} else {
					for count := q.max.Uint64(); count > q.min.Uint64(); count-- {
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
				for count := uint64(0); count < q.min.Uint64(); count++ {
					next = segment(choice+1, pc, next)
					if next < 0 {
						return -1
					}
				}
				pc = choice - 1
			case opSet, opAssert:
				next = add(regularInstruction{op: i.op, x: next, source: pc})
			case opSplit:
				a, aok := mapped[i.x]
				b, bok := mapped[i.y]
				if !aok || !bok {
					return -1
				}
				next = add(regularInstruction{op: opSplit, x: a, y: b})
			case opJump:
				var ok bool
				next, ok = mapped[i.x]
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
	r.start = segment(0, len(p.code), accept)
	if r.start < 0 {
		return nil
	}
	// ASCII alphabet classes preserve every consuming predicate and assertion.
	// Folding is performed at compile time with exactly the VM's Canonicalize.
	keys := map[string]uint8{}
	for c := 0; c < 128; c++ {
		key := make([]byte, (len(r.code)+7)/8+2)
		for pc, n := range r.code {
			if n.op == opSet {
				i := p.code[n.source]
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
			id = uint8(len(r.representatives))
			keys[string(key)] = id
			r.representatives = append(r.representatives, byte(c))
		}
		r.classes[c] = id
	}
	return r
}

func regularAssertion(i instruction, input []uint16, at int) bool {
	switch i.assertion {
	case Start:
		return at == 0 || i.flags.Multiline && lineTerminator(rune(input[at-1]))
	case End:
		return at == len(input) || i.flags.Multiline && lineTerminator(rune(input[at]))
	default:
		left, _, l := readCharacter(input, at, -1, unicodeMode(i.flags))
		right, _, r := readCharacter(input, at, 1, unicodeMode(i.flags))
		boundary := (l && wordCharacter(left, i.flags)) != (r && wordCharacter(right, i.flags))
		return boundary != (i.assertion == NotWordBoundary)
	}
}

// Merging threads keeps the earliest start at each state. With no references or
// lookaround their future languages are identical. Once a candidate accepts,
// only threads from earlier starts need finish before capture recovery.
func (p *Program) regularFind(input []uint16, start int, sticky bool) int {
	r := p.regular
	current := make([]int, len(r.code))
	next := make([]int, len(r.code))
	for k := range current {
		current[k] = -1
		next[k] = -1
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
		if best < 0 && (at == start || !sticky && !p.anchored) {
			put(r.start, at)
		}
		for pc, origin := range current {
			if origin >= 0 {
				pending[pc/64] |= uint64(1) << uint(pc%64)
			}
		}
		for {
			pc := -1
			for k, word := range pending {
				if word != 0 {
					pc = k*64 + bits.TrailingZeros64(word)
					pending[k] &= pending[k] - 1
					break
				}
			}
			if pc < 0 {
				break
			}
			n := r.code[pc]
			origin := current[pc]
			if best >= 0 && origin >= best {
				continue
			}
			switch n.op {
			case opAccept:
				best = origin
			case opSplit:
				put(n.x, origin)
				put(n.y, origin)
			case opAssert:
				if regularAssertion(p.code[n.source], input, at) {
					put(n.x, origin)
				}
			}
		}
		if at == len(input) {
			return best
		}
		c, to, _ := readCharacter(input, at, 1, unicodeMode(p.flags))
		alive := false
		for pc, origin := range current {
			if origin < 0 || best >= 0 && origin >= best {
				continue
			}
			n := r.code[pc]
			if n.op == opSet {
				i := p.code[n.source]
				if i.set.contains(canonicalize(c, i.flags)) && (next[n.x] < 0 || origin < next[n.x]) {
					next[n.x] = origin
					alive = true
				}
			}
		}
		if !alive && (best >= 0 || sticky || p.anchored) {
			return best
		}
		current, next = next, current
		for k := range next {
			next[k] = -1
		}
		at = to
	}
}
