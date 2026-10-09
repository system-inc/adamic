package regexp

import (
	"fmt"
	"strings"
)

// nativeFirstASCII computes a necessary first-character condition from bytecode.
// Unknown or nullable paths disable it. Non-ASCII input always uses the VM.
func (p *Program) nativeFirstASCII() ([2]uint64, bool) {
	var visit func(int, map[int]bool) ([2]uint64, bool)
	visit = func(pc int, path map[int]bool) ([2]uint64, bool) {
		var bits [2]uint64
		if pc < 0 || pc >= len(p.code) || path[pc] {
			return bits, false
		}
		path[pc] = true
		defer delete(path, pc)
		i := p.code[pc]
		switch i.op {
		case opSave, opAssert, opLook, opRepeatInit:
			return visit(pc+1, path)
		case opJump:
			return visit(i.x, path)
		case opSplit:
			a, ok := visit(i.x, path)
			b, yes := visit(i.y, path)
			return [2]uint64{a[0] | b[0], a[1] | b[1]}, ok && yes
		case opRepeat:
			a, ok := visit(pc+1, path)
			if i.min.Sign() > 0 {
				return a, ok
			}
			b, yes := visit(i.y, path)
			return [2]uint64{a[0] | b[0], a[1] | b[1]}, ok && yes
		case opSet:
			if i.direction != 1 {
				return bits, false
			}
			for _, s := range i.set.strings {
				if len(s) == 0 {
					return bits, false
				}
			}
			for c := rune(0); c < 128; c++ {
				fold := canonicalize(c, i.flags)
				contains := false
				for _, r := range i.set.ranges {
					if fold >= r.From && fold <= r.To {
						contains = true
					}
				}
				for _, s := range i.set.strings {
					if fold == s[0] {
						contains = true
					}
				}
				if contains {
					bits[c/64] |= 1 << uint(c%64)
				}
			}
			return bits, true
		}
		return bits, false
	}
	return visit(0, map[int]bool{})
}
func (p *Program) nativePrefix() []rune {
	var prefix []rune
	for _, i := range p.code {
		switch i.op {
		case opSave, opAssert, opLook:
			continue
		case opSet:
			if i.direction != 1 || i.flags.IgnoreCase || len(i.set.strings) != 0 || len(i.set.ranges) != 1 || i.set.ranges[0].From != i.set.ranges[0].To || i.set.ranges[0].To >= 128 {
				return prefix
			}
			prefix = append(prefix, i.set.ranges[0].From)
		default:
			return prefix
		}
	}
	return prefix
}

// Small, forward ASCII programs have no choice points and need no VM state.
func (p *Program) nativeStraightLine(name string) string {
	if len(p.code) > 32 {
		return ""
	}
	for _, i := range p.code {
		switch i.op {
		case opAccept, opSave:
		case opAssert:
			if i.flags.Multiline || i.assertion > End {
				return ""
			}
		case opSet:
			if i.direction != 1 || i.flags.IgnoreCase || len(i.set.strings) != 0 {
				return ""
			}
			for _, r := range i.set.ranges {
				if r.To >= 128 {
					return ""
				}
			}
		default:
			return ""
		}
	}
	out := fmt.Sprintf("static bool %s_fast(const uint16_t *input,size_t length,ptrdiff_t *position,ptrdiff_t *captures,uint64_t *steps) {ptrdiff_t at=*position;\n", name)
	for pc, i := range p.code {
		fail := fmt.Sprintf("{*steps+=%d;return false;}", pc+1)
		switch i.op {
		case opSave:
			out += fmt.Sprintf("captures[%d]=at;\n", i.x)
		case opAssert:
			if i.assertion == Start {
				out += "if(at!=0)" + fail + "\n"
			} else {
				out += "if((size_t)at!=length)" + fail + "\n"
			}
		case opSet:
			out += "if((size_t)at>=length)" + fail + "\nif(!("
			for k, r := range i.set.ranges {
				if k > 0 {
					out += "||"
				}
				out += fmt.Sprintf("(input[at]>=%d && input[at]<=%d)", r.From, r.To)
			}
			if len(i.set.ranges) == 0 {
				out += "false"
			}
			out += "))" + fail + "\nat++;\n"
		case opAccept:
			out += fmt.Sprintf("*position=at;*steps+=%d;return true;\n", pc+1)
		}
	}
	return out + "}\n"
}

// A non-multiline start assertion before consumption permits only position zero.
func (p *Program) nativeAnchored() bool {
	for _, i := range p.code {
		switch i.op {
		case opSave, opLook:
			continue
		case opAssert:
			if i.assertion == Start && !i.flags.Multiline {
				return true
			}
		default:
			return false
		}
	}
	return false
}

// ASCII strings admit a separate straight-line runner even when Unicode folding
// or a wide character class requires the general VM on other strings.
func (p *Program) nativeStraightLineASCII(name string) string {
	copyProgram := *p
	copyProgram.code = append([]instruction(nil), p.code...)
	for pc, i := range copyProgram.code {
		if i.op != opSet {
			continue
		}
		if i.direction != 1 || len(i.set.strings) != 0 {
			return ""
		}
		var ranges []RuneRange
		for c := rune(0); c < 128; c++ {
			if i.set.contains(canonicalize(c, i.flags)) {
				ranges = append(ranges, RuneRange{c, c})
			}
		}
		i.set.ranges = normalizeRanges(ranges)
		i.flags.IgnoreCase = false
		copyProgram.code[pc] = i
	}
	body := copyProgram.nativeStraightLine(name)
	return strings.ReplaceAll(strings.ReplaceAll(body, name+"_fast", name+"_ascii"), "uint16_t", "unsigned char")
}
