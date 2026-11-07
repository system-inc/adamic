package regexp

import (
	"fmt"
	"strings"
)

// nativeFirstASCII computes a necessary first-character condition from bytecode.
// Unknown or nullable paths disable it. Non-ASCII input always uses the VM.
func (program *Program) nativeFirstASCII() ([2]uint64, bool) {
	var visit func(int, map[int]bool) ([2]uint64, bool)
	visit = func(pc int, path map[int]bool) ([2]uint64, bool) {
		var bits [2]uint64
		if pc < 0 || pc >= len(program.code) || path[pc] {
			return bits, false
		}
		path[pc] = true
		defer delete(path, pc)
		instruction := program.code[pc]
		switch instruction.op {
		case opSave, opAssert, opLook, opRepeatInit:
			return visit(pc+1, path)
		case opJump:
			return visit(instruction.x, path)
		case opSplit:
			a, ok := visit(instruction.x, path)
			b, yes := visit(instruction.y, path)
			return [2]uint64{a[0] | b[0], a[1] | b[1]}, ok && yes
		case opRepeat:
			a, ok := visit(pc+1, path)
			if instruction.min.Sign() > 0 {
				return a, ok
			}
			b, yes := visit(instruction.y, path)
			return [2]uint64{a[0] | b[0], a[1] | b[1]}, ok && yes
		case opSet:
			if instruction.direction != 1 {
				return bits, false
			}
			for _, s := range instruction.set.strings {
				if len(s) == 0 {
					return bits, false
				}
			}
			for character := rune(0); character < 128; character++ {
				fold := canonicalize(character, instruction.flags)
				contains := false
				for _, runeRange := range instruction.set.ranges {
					if fold >= runeRange.From && fold <= runeRange.To {
						contains = true
					}
				}
				for _, s := range instruction.set.strings {
					if fold == s[0] {
						contains = true
					}
				}
				if contains {
					bits[character/64] |= 1 << uint(character%64)
				}
			}
			return bits, true
		}
		return bits, false
	}
	return visit(0, map[int]bool{})
}
func (program *Program) nativePrefix() []rune {
	var prefix []rune
	for _, instruction := range program.code {
		switch instruction.op {
		case opSave, opAssert, opLook:
			continue
		case opSet:
			if instruction.direction != 1 ||
				instruction.flags.IgnoreCase ||
				len(instruction.set.strings) != 0 ||
				len(instruction.set.ranges) != 1 ||
				instruction.set.ranges[0].From != instruction.set.ranges[0].To ||
				instruction.set.ranges[0].To >= 128 {
				return prefix
			}
			prefix = append(prefix, instruction.set.ranges[0].From)
		default:
			return prefix
		}
	}
	return prefix
}

// Small, forward ASCII programs have no choice points and need no VM state.
func (program *Program) nativeStraightLine(name string) string {
	if len(program.code) > 32 {
		return ""
	}
	for _, instruction := range program.code {
		switch instruction.op {
		case opAccept, opSave:
		case opAssert:
			if instruction.flags.Multiline || instruction.assertion > End {
				return ""
			}
		case opSet:
			if instruction.direction != 1 || instruction.flags.IgnoreCase || len(instruction.set.strings) != 0 {
				return ""
			}
			for _, runeRange := range instruction.set.ranges {
				if runeRange.To >= 128 {
					return ""
				}
			}
		default:
			return ""
		}
	}
	out := fmt.Sprintf(
		"static bool %s_fast(const uint16_t *input,size_t length,ptrdiff_t *position,ptrdiff_t *captures,uint64_t *steps) {ptrdiff_t at=*position;\n",
		name,
	)
	for pc, instruction := range program.code {
		fail := fmt.Sprintf("{*steps+=%d;return false;}", pc+1)
		switch instruction.op {
		case opSave:
			out += fmt.Sprintf("captures[%d]=at;\n", instruction.x)
		case opAssert:
			if instruction.assertion == Start {
				out += "if(at!=0)" + fail + "\n"
			} else {
				out += "if((size_t)at!=length)" + fail + "\n"
			}
		case opSet:
			out += "if((size_t)at>=length)" + fail + "\nif(!("
			for index, runeRange := range instruction.set.ranges {
				if index > 0 {
					out += "||"
				}
				out += fmt.Sprintf("(input[at]>=%d && input[at]<=%d)", runeRange.From, runeRange.To)
			}
			if len(instruction.set.ranges) == 0 {
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
func (program *Program) nativeAnchored() bool {
	for _, instruction := range program.code {
		switch instruction.op {
		case opSave, opLook:
			continue
		case opAssert:
			if instruction.assertion == Start && !instruction.flags.Multiline {
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
func (program *Program) nativeStraightLineASCII(name string) string {
	copyProgram := *program
	copyProgram.code = append([]instruction(nil), program.code...)
	for pc, instruction := range copyProgram.code {
		if instruction.op != opSet {
			continue
		}
		if instruction.direction != 1 || len(instruction.set.strings) != 0 {
			return ""
		}
		var ranges []RuneRange
		for character := rune(0); character < 128; character++ {
			if instruction.set.contains(canonicalize(character, instruction.flags)) {
				ranges = append(ranges, RuneRange{character, character})
			}
		}
		instruction.set.ranges = normalizeRanges(ranges)
		instruction.flags.IgnoreCase = false
		copyProgram.code[pc] = instruction
	}
	body := copyProgram.nativeStraightLine(name)
	return strings.ReplaceAll(strings.ReplaceAll(body, name+"_fast", name+"_ascii"), "uint16_t", "unsigned char")
}

// Boolean calls need neither capture registers nor instruction accounting in
// unlimited mode. Emit their entire search, so flags and first sets are compile
// time constants rather than tests in the shared runtime loop.
func (program *Program) nativeBooleanASCII(name, first string) string {
	if program.nativeStraightLineASCII(name) == "" {
		return program.nativeRepeatedASCII(name)
	}
	var out strings.Builder
	minimum, endAnchored := program.nativeBooleanWidth()
	bits, filter := program.nativeFirstASCII()
	if filter && first == "NULL" && !endAnchored && !program.nativeAnchored() && !program.flags.Sticky {
		fmt.Fprintf(
			&out,
			"static const uint64_t %s_test_first[2]={UINT64_C(%d),UINT64_C(%d)};\n",
			name,
			bits[0],
			bits[1],
		)
	}
	fmt.Fprintf(
		&out,
		"static bool %s_test(const unsigned char *input,size_t length,ptrdiff_t *position) {size_t start=(size_t)*position;(void)input;\n",
		name,
	)
	if minimum > 0 {
		fmt.Fprintf(&out, "if(start>length||length-start<%d)return false;size_t last=length-%d;\n", minimum, minimum)
	} else {
		out.WriteString("if(start>length)return false;size_t last=length;\n")
	}
	if program.nativeAnchored() {
		out.WriteString("if(start!=0)return false;\n")
	}
	if endAnchored {
		// A trailing non-multiline end assertion and a fixed width admit only
		// this start. Sticky calls must already be at that start.
		if program.flags.Sticky {
			out.WriteString("if(start!=last)return false;\n")
		} else {
			out.WriteString("if(start>last)return false;start=last;\n")
		}
	}
	out.WriteString("while(start<=last){size_t at=start;\n")
	if !endAnchored && !program.nativeAnchored() && !program.flags.Sticky && first != "NULL" {
		fmt.Fprintf(
			&out,
			"ptrdiff_t candidate=%s(input,length,start);if(candidate<0)return false;start=(size_t)candidate;if(start>last)return false;at=start;\n",
			first,
		)
	} else if filter && !endAnchored && !program.nativeAnchored() && !program.flags.Sticky {
		fmt.Fprintf(
			&out,
			"while(start<=last&&start<length&&!(%s_test_first[input[start]/64]&(UINT64_C(1)<<(input[start]%%64))))start++;if(start>last)return false;at=start;\n",
			name,
		)
	}
	hasFailure := false
	firstSet := true
	firstProven := !endAnchored &&
		!program.nativeAnchored() &&
		!program.flags.Sticky &&
		filter &&
		(first == "NULL" || program.nativeLiteralWindow(name) == "")
	for _, instruction := range program.code {
		switch instruction.op {
		case opSave:
		case opAssert:
			hasFailure = true
			if instruction.assertion == Start {
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
			for character := rune(0); character < 128; character++ {
				fold := canonicalize(character, instruction.flags)
				if instruction.set.contains(fold) {
					ranges = append(ranges, RuneRange{character, character})
				}
			}
			out.WriteString("if(at>=length||!(")
			for index, runeRange := range normalizeRanges(ranges) {
				if index > 0 {
					out.WriteString("||")
				}
				fmt.Fprintf(&out, "(input[at]>=%d&&input[at]<=%d)", runeRange.From, runeRange.To)
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
		if endAnchored || program.nativeAnchored() || program.flags.Sticky {
			out.WriteString("return false;")
		} else {
			out.WriteString("start++;")
		}
	}
	out.WriteString("}return false;}\n")
	return out.String()
}

func (program *Program) nativeBooleanWidth() (int, bool) {
	minimum, endOffset := 0, -1
	for _, instruction := range program.code {
		if instruction.op == opSet {
			minimum++
		} else if instruction.op == opAssert && instruction.assertion == End {
			endOffset = minimum
		}
	}
	return minimum, endOffset == minimum
}

// An anchored scalar star followed by one scalar needs no choice points for a
// stateless boolean result. Captures and stateful calls retain the VM, because
// their preferred end depends on greediness. Saves do not affect existence.
func (program *Program) nativeRepeatedASCII(name string) string {
	if program.flags.Global || program.flags.Sticky {
		return ""
	}
	pc := 0
	skipSaves := func() {
		for pc < len(program.code) && program.code[pc].op == opSave {
			pc++
		}
	}
	skipSaves()
	if pc >= len(program.code) ||
		program.code[pc].op != opAssert ||
		program.code[pc].assertion != Start ||
		program.code[pc].flags.Multiline {
		return ""
	}
	pc++
	skipSaves()
	if pc+2 >= len(program.code) || program.code[pc].op != opRepeatInit {
		return ""
	}
	id := program.code[pc].x
	pc++
	choice := pc
	repeat := program.code[pc]
	if repeat.op != opRepeat || repeat.x != id || repeat.min == nil || repeat.min.Sign() != 0 || repeat.max != nil {
		return ""
	}
	pc++
	skipSaves()
	if pc >= len(program.code) {
		return ""
	}
	body := program.code[pc]
	pc++
	skipSaves()
	if pc >= len(program.code) ||
		program.code[pc].op != opRepeatEnd ||
		program.code[pc].x != id ||
		program.code[pc].y != choice ||
		repeat.y != pc+1 {
		return ""
	}
	pc++
	skipSaves()
	if pc >= len(program.code) {
		return ""
	}
	tail := program.code[pc]
	pc++
	skipSaves()
	if pc != len(program.code)-1 || program.code[pc].op != opAccept {
		return ""
	}
	var out strings.Builder
	for index, instruction := range []instruction{body, tail} {
		if instruction.op != opSet || instruction.direction != 1 || len(instruction.set.strings) != 0 {
			return ""
		}
		var bits [2]uint64
		for character := rune(0); character < 128; character++ {
			if instruction.set.contains(canonicalize(character, instruction.flags)) {
				bits[character/64] |= uint64(1) << uint(character%64)
			}
		}
		label := "repeat"
		if index == 1 {
			label = "tail"
		}
		fmt.Fprintf(
			&out,
			"static const uint64_t %s_%s[2]={UINT64_C(%d),UINT64_C(%d)};\n",
			name,
			label,
			bits[0],
			bits[1],
		)
	}
	fmt.Fprintf(
		&out,
		"static bool %s_test(const unsigned char *input,size_t length,ptrdiff_t *position){if(*position!=0)return false;for(size_t at=0;at<length;at++){unsigned c=input[at];\n",
		name,
	)
	// The tail may match after zero repetitions, even outside the body set.
	fmt.Fprintf(&out, "if(%s_tail[c/64]&(UINT64_C(1)<<(c%%64))){*position=(ptrdiff_t)at+1;return true;}\n", name)
	fmt.Fprintf(&out, "if(!(%s_repeat[c/64]&(UINT64_C(1)<<(c%%64))))return false;}return false;}\n", name)
	return out.String()
}

// A leading positive literal lookbehind supplies a necessary context filter.
// The VM still evaluates the assertion, including any captures. Negative,
// branching, folded and non-ASCII lookbehind remain completely general.
func (program *Program) nativeLookbehindPrefix() []rune {
	for _, instruction := range program.code {
		if instruction.op == opSave {
			continue
		}
		if instruction.op != opLook || instruction.negative {
			return nil
		}
		var reverse []rune
		for _, sub := range instruction.look.code {
			switch sub.op {
			case opSave, opAccept:
			case opSet:
				if sub.direction != -1 ||
					sub.flags.IgnoreCase ||
					len(sub.set.strings) != 0 ||
					len(sub.set.ranges) != 1 ||
					sub.set.ranges[0].From != sub.set.ranges[0].To ||
					sub.set.ranges[0].To >= 128 {
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
