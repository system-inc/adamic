package regexp

import (
	"fmt"
	"strings"
)

func (program *Program) nativeRegular(name string) (string, string) {
	regularProgram := program.regular
	if regularProgram == nil {
		return "", "NULL"
	}
	var out strings.Builder
	fmt.Fprintf(&out, "static const adamic_regex_regular_instruction %s_regular_code[] = {", name)
	for _, instruction := range regularProgram.code {
		fmt.Fprintf(&out, "{%d,%d,%d,%d},", instruction.op, instruction.x, instruction.y, instruction.source)
	}
	out.WriteString("};\n")
	fmt.Fprintf(&out, "static const unsigned char %s_regular_classes[128] = {", name)
	for _, class := range regularProgram.classes {
		fmt.Fprintf(&out, "%d,", class)
	}
	out.WriteString("};\n")
	fmt.Fprintf(&out, "static const unsigned char %s_regular_characters[] = {", name)
	for _, representative := range regularProgram.representatives {
		fmt.Fprintf(&out, "%d,", representative)
	}
	out.WriteString("};\n")
	words := (len(regularProgram.code) + 63) / 64
	fmt.Fprintf(&out, "static const uint64_t %s_regular_masks[] = {", name)
	for _, representative := range regularProgram.representatives {
		masks := make([]uint64, words)
		for pc, regularInstruction := range regularProgram.code {
			if regularInstruction.op == opSet {
				i := program.code[regularInstruction.source]
				if i.set.contains(canonicalize(rune(representative), i.flags)) {
					masks[pc/64] |= uint64(1) << uint(pc%64)
				}
			}
		}
		for _, mask := range masks {
			fmt.Fprintf(&out, "UINT64_C(%d),", mask)
		}
	}
	out.WriteString("};\n")
	fmt.Fprintf(
		&out,
		"static const adamic_regex_regular %s_regular = {%s_regular_code,%d,%d,%s_regular_classes,%s_regular_characters,%d,%s_regular_masks,%d};\n",
		name,
		name,
		len(regularProgram.code),
		regularProgram.start,
		name,
		name,
		len(regularProgram.representatives),
		name,
		words,
	)
	return out.String(), "&" + name + "_regular"
}

// Tiny necessary first sets can reject ASCII input before UTF-16 conversion,
// and jump straight to a candidate in an existing one-pass runner.
func (program *Program) nativeSparseFirst(name string) (string, string) {
	if !program.nativeAnchored() {
		if body := program.nativeLiteralWindow(name); body != "" {
			return body, name + "_first"
		}
	}
	bits, filter := program.nativeFirstASCII()
	if !filter || program.nativeAnchored() {
		return "", "NULL"
	}
	var characters []int
	for c := 0; c < 128; c++ {
		if bits[c/64]&(uint64(1)<<uint(c%64)) != 0 {
			characters = append(characters, c)
		}
	}
	if len(characters) > 4 {
		return "", "NULL"
	}
	var out strings.Builder
	out.WriteString("#include <string.h>\n")
	fmt.Fprintf(&out, "static ptrdiff_t %s_first(const unsigned char *input,size_t length,size_t start) {\n", name)
	out.WriteString("if(start>=length)return -1;(void)input;ptrdiff_t first=-1;\n")
	for _, character := range characters {
		fmt.Fprintf(
			&out,
			"{const unsigned char *hit=memchr(input+start,%d,length-start);if(hit!=NULL&&(first<0||hit-input<first))first=hit-input;}\n",
			character,
		)
	}
	out.WriteString("return first;}\n")
	return out.String(), name + "_first"
}

// In an ASCII input every forward scalar set consumes one byte. A literal
// window after a fixed number of such sets is necessary even when the first
// set is broad. Stop before any variable-width or branching instruction.
func (program *Program) nativeLiteralWindow(name string) string {
	var best, run []rune
	offset, runOffset, bestOffset := 0, 0, 0
	for _, instruction := range program.code {
		switch instruction.op {
		case opSave, opAssert, opLook:
			continue
		case opSet:
			if instruction.direction != 1 || len(instruction.set.strings) != 0 {
				goto finished
			}
			if !instruction.flags.IgnoreCase &&
				len(instruction.set.ranges) == 1 &&
				instruction.set.ranges[0].From == instruction.set.ranges[0].To &&
				instruction.set.ranges[0].To < 128 {
				if len(run) == 0 {
					runOffset = offset
				}
				run = append(run, instruction.set.ranges[0].From)
				if runOffset > 0 && len(run) > len(best) {
					best = run
					bestOffset = runOffset
				}
			} else {
				run = nil
			}
			offset++
		default:
			goto finished
		}
	}
finished:
	if len(best) < 2 {
		return ""
	}
	var out strings.Builder
	out.WriteString("#include <string.h>\n")
	fmt.Fprintf(&out, "static const unsigned char %s_window[]={", name)
	for _, character := range best {
		fmt.Fprintf(&out, "%d,", character)
	}
	fmt.Fprintf(&out, "};\nstatic ptrdiff_t %s_first(const unsigned char *input,size_t length,size_t start){\n", name)
	fmt.Fprintf(
		&out,
		"if(start>length||length-start<%d)return -1;size_t scan=start+%d;\n",
		bestOffset+len(best),
		bestOffset,
	)
	fmt.Fprintf(
		&out,
		"while(length-scan>=%d){const unsigned char *hit=memchr(input+scan,%d,length-scan-%d+1);if(hit==NULL)return -1;scan=(size_t)(hit-input);\n",
		len(best),
		best[0],
		len(best),
	)
	fmt.Fprintf(
		&out,
		"if(memcmp(hit,%s_window,%d)==0)return (ptrdiff_t)(scan-%d);scan++;}return -1;}\n",
		name,
		len(best),
		bestOffset,
	)
	return out.String()
}
