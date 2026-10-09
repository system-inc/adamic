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

// Boolean calls need neither capture registers nor instruction accounting in
// unlimited mode. Emit their entire search, so flags and first sets are compile
// time constants rather than tests in the shared runtime loop.
func (p *Program) nativeBooleanASCII(name, first string) string {
	if p.nativeStraightLineASCII(name) == "" {
		return p.nativeRepeatedASCII(name)
	}
	var out strings.Builder
	minimum, endAnchored := p.nativeBooleanWidth()
	bits, filter := p.nativeFirstASCII()
	if filter && first == "NULL" && !endAnchored && !p.nativeAnchored() && !p.flags.Sticky {
		fmt.Fprintf(&out, "static const uint64_t %s_test_first[2]={UINT64_C(%d),UINT64_C(%d)};\n", name, bits[0], bits[1])
	}
	fmt.Fprintf(&out, "static bool %s_test(const unsigned char *input,size_t length,ptrdiff_t *position) {size_t start=(size_t)*position;(void)input;\n", name)
	if minimum > 0 {
		fmt.Fprintf(&out, "if(start>length||length-start<%d)return false;size_t last=length-%d;\n", minimum, minimum)
	} else {
		out.WriteString("if(start>length)return false;size_t last=length;\n")
	}
	if p.nativeAnchored() {
		out.WriteString("if(start!=0)return false;\n")
	}
	if endAnchored {
		// A trailing non-multiline end assertion and a fixed width admit only
		// this start. Sticky calls must already be at that start.
		if p.flags.Sticky {
			out.WriteString("if(start!=last)return false;\n")
		} else {
			out.WriteString("if(start>last)return false;start=last;\n")
		}
	}
	out.WriteString("while(start<=last){size_t at=start;\n")
	if !endAnchored && !p.nativeAnchored() && !p.flags.Sticky && first != "NULL" {
		fmt.Fprintf(&out, "ptrdiff_t candidate=%s(input,length,start);if(candidate<0)return false;start=(size_t)candidate;if(start>last)return false;at=start;\n", first)
	} else if filter && !endAnchored && !p.nativeAnchored() && !p.flags.Sticky {
		fmt.Fprintf(&out, "while(start<=last&&start<length&&!(%s_test_first[input[start]/64]&(UINT64_C(1)<<(input[start]%%64))))start++;if(start>last)return false;at=start;\n", name)
	}
	hasFailure := false
	firstSet := true
	firstProven := !endAnchored && !p.nativeAnchored() && !p.flags.Sticky && filter && (first == "NULL" || p.nativeLiteralWindow(name) == "")
	for _, i := range p.code {
		switch i.op {
		case opSave:
		case opAssert:
			hasFailure = true
			if i.assertion == Start {
				out.WriteString("if(at!=0)goto failed;\n")
			} else {
				out.WriteString("if(at!=length)goto failed;\n")
			}
		case opSet:
			if firstSet && firstProven {
				// The emitted first-set scan already proved this predicate.
				out.WriteString("at++;\n")
				firstSet = false
				continue
			}
			firstSet = false
			hasFailure = true
			var ranges []RuneRange
			for c := rune(0); c < 128; c++ {
				fold := canonicalize(c, i.flags)
				if i.set.contains(fold) {
					ranges = append(ranges, RuneRange{c, c})
				}
			}
			out.WriteString("if(at>=length||!(")
			for k, r := range normalizeRanges(ranges) {
				if k > 0 {
					out.WriteString("||")
				}
				fmt.Fprintf(&out, "(input[at]>=%d&&input[at]<=%d)", r.From, r.To)
			}
			if len(ranges) == 0 {
				out.WriteString("false")
			}
			out.WriteString("))goto failed;at++;\n")
		case opAccept:
			out.WriteString("*position=(ptrdiff_t)at;return true;\n")
		}
	}
	if hasFailure {
		out.WriteString("failed:;")
		if endAnchored || p.nativeAnchored() || p.flags.Sticky {
			out.WriteString("return false;")
		} else {
			out.WriteString("start++;")
		}
	}
	out.WriteString("}return false;}\n")
	return out.String()
}

func (p *Program) nativeBooleanWidth() (int, bool) {
	minimum, endOffset := 0, -1
	for _, i := range p.code {
		if i.op == opSet {
			minimum++
		} else if i.op == opAssert && i.assertion == End {
			endOffset = minimum
		}
	}
	return minimum, endOffset == minimum
}

// An anchored scalar star followed by one scalar needs no choice points for a
// stateless boolean result. Captures and stateful calls retain the VM, because
// their preferred end depends on greediness. Saves do not affect existence.
func (p *Program) nativeRepeatedASCII(name string) string {
	if p.flags.Global || p.flags.Sticky {
		return ""
	}
	pc := 0
	skipSaves := func() {
		for pc < len(p.code) && p.code[pc].op == opSave {
			pc++
		}
	}
	skipSaves()
	if pc >= len(p.code) || p.code[pc].op != opAssert || p.code[pc].assertion != Start || p.code[pc].flags.Multiline {
		return ""
	}
	pc++
	skipSaves()
	if pc+2 >= len(p.code) || p.code[pc].op != opRepeatInit {
		return ""
	}
	id := p.code[pc].x
	pc++
	choice := pc
	repeat := p.code[pc]
	if repeat.op != opRepeat || repeat.x != id || repeat.min == nil || repeat.min.Sign() != 0 || repeat.max != nil {
		return ""
	}
	pc++
	skipSaves()
	if pc >= len(p.code) {
		return ""
	}
	body := p.code[pc]
	pc++
	skipSaves()
	if pc >= len(p.code) || p.code[pc].op != opRepeatEnd || p.code[pc].x != id || p.code[pc].y != choice || repeat.y != pc+1 {
		return ""
	}
	pc++
	skipSaves()
	if pc >= len(p.code) {
		return ""
	}
	tail := p.code[pc]
	pc++
	skipSaves()
	if pc != len(p.code)-1 || p.code[pc].op != opAccept {
		return ""
	}
	var out strings.Builder
	for k, i := range []instruction{body, tail} {
		if i.op != opSet || i.direction != 1 || len(i.set.strings) != 0 {
			return ""
		}
		var bits [2]uint64
		for c := rune(0); c < 128; c++ {
			if i.set.contains(canonicalize(c, i.flags)) {
				bits[c/64] |= uint64(1) << uint(c%64)
			}
		}
		label := "repeat"
		if k == 1 {
			label = "tail"
		}
		fmt.Fprintf(&out, "static const uint64_t %s_%s[2]={UINT64_C(%d),UINT64_C(%d)};\n", name, label, bits[0], bits[1])
	}
	fmt.Fprintf(&out, "static bool %s_test(const unsigned char *input,size_t length,ptrdiff_t *position){if(*position!=0)return false;for(size_t at=0;at<length;at++){unsigned c=input[at];\n", name)
	// The tail may match after zero repetitions, even outside the body set.
	fmt.Fprintf(&out, "if(%s_tail[c/64]&(UINT64_C(1)<<(c%%64))){*position=(ptrdiff_t)at+1;return true;}\n", name)
	fmt.Fprintf(&out, "if(!(%s_repeat[c/64]&(UINT64_C(1)<<(c%%64))))return false;}return false;}\n", name)
	return out.String()
}

// A leading positive literal lookbehind supplies a necessary context filter.
// The VM still evaluates the assertion, including any captures. Negative,
// branching, folded and non-ASCII lookbehind remain completely general.
func (p *Program) nativeLookbehindPrefix() []rune {
	for _, i := range p.code {
		if i.op == opSave {
			continue
		}
		if i.op != opLook || i.negative {
			return nil
		}
		var reverse []rune
		for _, sub := range i.look.code {
			switch sub.op {
			case opSave, opAccept:
			case opSet:
				if sub.direction != -1 || sub.flags.IgnoreCase || len(sub.set.strings) != 0 || len(sub.set.ranges) != 1 || sub.set.ranges[0].From != sub.set.ranges[0].To || sub.set.ranges[0].To >= 128 {
					return nil
				}
				reverse = append(reverse, sub.set.ranges[0].From)
			default:
				return nil
			}
		}
		for left, right := 0, len(reverse)-1; left < right; left, right = left+1, right-1 {
			reverse[left], reverse[right] = reverse[right], reverse[left]
		}
		return reverse
	}
	return nil
}
